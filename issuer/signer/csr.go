/*
Copyright 2022 CMU-SV.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package signer

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiutil "github.com/cert-manager/cert-manager/pkg/api/util"
	v1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
)

var serialNumberLimit = new(big.Int).Lsh(big.NewInt(1), 128)

// DefaultCertDuration returns d.Duration if set, otherwise returns
// cert-manager's default certificate duration (90 days).
func DefaultCertDuration(d *metav1.Duration) time.Duration {
	certDuration := v1.DefaultCertificateDuration
	if d != nil {
		certDuration = d.Duration
	}

	return certDuration
}

func BuildKeyUsages(usages []v1.KeyUsage, isCA bool) (ku x509.KeyUsage, eku []x509.ExtKeyUsage, err error) {
	var unk []v1.KeyUsage
	if isCA {
		ku |= x509.KeyUsageCertSign
	}
	if len(usages) == 0 {
		usages = append(usages, v1.DefaultKeyUsages()...)
	}
	for _, u := range usages {
		if kuse, ok := apiutil.KeyUsageType(u); ok {
			ku |= kuse
		} else if ekuse, ok := apiutil.ExtKeyUsageType(u); ok {
			eku = append(eku, ekuse)
		} else {
			unk = append(unk, u)
		}
	}
	if len(unk) > 0 {
		err = fmt.Errorf("unknown key usages: %v", unk)
	}
	return
}

func GenerateTemplateFromCSRPEMWithUsages(csrPEM []byte, duration time.Duration, isCA bool, keyUsage x509.KeyUsage, extKeyUsage []x509.ExtKeyUsage) (*x509.Certificate, error) {
	block, _ := pem.Decode(csrPEM)
	if block == nil {
		return nil, errors.New("failed to decode csr")
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, err
	}

	if err := csr.CheckSignature(); err != nil {
		return nil, err
	}

	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %s", err.Error())
	}

	return &x509.Certificate{
		// Version must be 2 according to RFC5280.
		// A version value of 2 confusingly means version 3.
		// This value isn't used by Go at the time of writing.
		// https://datatracker.ietf.org/doc/html/rfc5280#section-4.1.2.1
		Version:               2,
		BasicConstraintsValid: true,
		SerialNumber:          serialNumber,
		PublicKeyAlgorithm:    csr.PublicKeyAlgorithm,
		PublicKey:             csr.PublicKey,
		IsCA:                  isCA,
		Subject:               csr.Subject,
		RawSubject:            csr.RawSubject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(duration),
		// see http://golang.org/pkg/crypto/x509/#KeyUsage
		KeyUsage:       keyUsage,
		ExtKeyUsage:    extKeyUsage,
		DNSNames:       csr.DNSNames,
		IPAddresses:    csr.IPAddresses,
		EmailAddresses: csr.EmailAddresses,
		URIs:           csr.URIs,
	}, nil
}

// GenerateTemplate will create a x509.Certificate for the given
// CertificateRequest resource
func GenerateTemplateFromCertificateRequest(cr *v1.CertificateRequest) (*x509.Certificate, error) {
	certDuration := DefaultCertDuration(cr.Spec.Duration)
	keyUsage, extKeyUsage, err := BuildKeyUsages(cr.Spec.Usages, cr.Spec.IsCA)
	if err != nil {
		return nil, err
	}
	return GenerateTemplateFromCSRPEMWithUsages(cr.Spec.Request, certDuration, cr.Spec.IsCA, keyUsage, extKeyUsage)
}

