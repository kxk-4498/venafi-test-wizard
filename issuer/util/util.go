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
package util

import (
	"fmt"

	api "github.com/kxk-4498/Venafi-test-wizard/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetSpecAndStatus(issuer client.Object) (*api.ChaosIssuerSpec, *api.ChaosIssuerStatus, error) {
	switch t := issuer.(type) {
	case *api.ChaosIssuer:
		return &t.Spec, &t.Status, nil
	case *api.ChaosClusterIssuer:
		return &t.Spec, &t.Status, nil
	default:
		return nil, nil, fmt.Errorf("not an issuer type: %t", t)
	}
}
func GetSpec(issuer client.Object) (*api.ChaosIssuerSpec, error) {
	switch t := issuer.(type) {
	case *api.ChaosIssuer:
		return &t.Spec, nil
	case *api.ChaosClusterIssuer:
		return &t.Spec, nil
	default:
		return nil, fmt.Errorf("not an issuer type: %t", t)
	}
}
