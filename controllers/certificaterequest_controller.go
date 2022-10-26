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

package controllers

import (
	"context"
	"fmt"
	"errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/clock"
	"k8s.io/apimachinery/pkg/types"
	"github.com/cert-manager/cert-manager/pkg/util/kube"
	"github.com/kxk-4498/Venafi-test-wizard/issuer/signer"

	api "github.com/kxk-4498/Venafi-test-wizard/api/v1alpha1"
	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	apiutil "github.com/cert-manager/cert-manager/pkg/api/util"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	cmerrors "github.com/cert-manager/cert-manager/pkg/util/errors"
	issuerutil "github.com/kxk-4498/Venafi-test-wizard/issuer/util"
)

// CertificateRequestReconciler reconciles a CertificateRequest object
type CertificateRequestReconciler struct {
	Scheme                   *runtime.Scheme
	client.Client
	//SignerBuilder          signer.SignerBuilder
	Log                      logr.Logger
	Recorder 				 record.EventRecorder
	Clock                    clock.Clock
	CheckApprovedCondition   bool
}
// annotation for generating RBAC role for writing events
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=self-signed-issuer.chaos.ch,resources=certificaterequests,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=self-signed-issuer.chaos.ch,resources=certificaterequests/status,verbs=get;update;patch



// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile

func (r *CertificateRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	log := r.Log.WithValues("certificaterequests", req.NamespacedName)

	// Fetch the CertificateRequest resource being reconciled.
	// Just ignore the request if the certificate request has been deleted.
	cr := cmapi.CertificateRequest{}
	if err := r.Client.Get(ctx, req.NamespacedName, cr); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "failed to retrieve CertificateRequest resource")
		return ctrl.Result{}, err
	}

	// Check the CertificateRequest's issuerRef and if it does not match the api
	// group name, log a message at a debug level and stop processing.
	if cr.Spec.IssuerRef.Group != "" && cr.Spec.IssuerRef.Group != api.GroupVersion.Group {
		log.V(4).Info("resource does not specify an issuerRef group name that we are responsible for", "group", cr.Spec.IssuerRef.Group) 
		return ctrl.Result{}, nil
	}

	//requestShouldBeProcessed is function given below to check for different conditions of Certificate Request
	shouldProcess, err := r.requestShouldBeProcessed(ctx, log, cr)
	if err != nil || !shouldProcess {
		return ctrl.Result{}, err
	}

	// If the certificate data is already set then we skip this request as it
	// has already been completed in the past.
	if len(cr.Status.Certificate) > 0 {
		log.V(4).Info("existing certificate data found in status, skipping already completed CertificateRequest") 
		return ctrl.Result{}, nil
	}

	// Chaos CA does not support online signing of CA certificate at this time
	if cr.Spec.IsCA {
		log.Info("Chaos CA does not support online signing of CA certificates")
		return ctrl.Result{}, nil
	}

	//##############################################################
	//#######################ISSUER-LOGIC###########################
	//##############################################################

	// Ignore but log an error if the issuerRef.Kind is unrecognised
	issuerGVK := api.GroupVersion.WithKind(cr.Spec.IssuerRef.Kind)
	issuerRO, err := r.Scheme.New(issuerGVK)
	if err != nil {
		err = fmt.Errorf("%w: %v", errIssuerRef, err)
		log.Error(err, "Unrecognised kind. Ignoring.")
		r.setStatus(ctx, cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonDenied, err.Error())
		return ctrl.Result{}, nil
	}
	issuer := issuerRO.(client.Object)
	// Create a Namespaced name for Issuer and a non-Namespaced name for ClusterIssuer
	issuerName := types.NamespacedName{
		Name: cr.Spec.IssuerRef.Name,
	}
	var secretNamespace string
	switch t := issuer.(type) {
	case *api.ChaosIssuer:
		issuerName.Namespace = cr.Namespace
		secretNamespace = cr.Namespace
		log = log.WithValues("chaosissuer", issuerName)
	case *api.ChaosClusterIssuer:
		secretNamespace = r.ClusterResourceNamespace
		log = log.WithValues("chaosclusterissuer", issuerName)
	default:
		err := fmt.Errorf("unexpected issuer type: %v", t)
		log.Error(err, "The issuerRef referred to a registered Kind which is not yet handled. Ignoring.")
		r.setStatus(ctx, cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonDenied, err.Error())
		return ctrl.Result{}, nil
	}

	// Get the Issuer or ClusterIssuer
	if err := r.Get(ctx, issuerName, issuer); err != nil {
		return ctrl.Result{}, fmt.Errorf("%w: %v", errGetIssuer, err)
	}

	issuerSpec, issuerStatus, err := issuerutil.GetSpecAndStatus(issuer)
	if err != nil {
		log.Error(err, "Unable to get the IssuerStatus. Ignoring.")
		r.setStatus(ctx, cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonDenied, err.Error())
		return ctrl.Result{}, nil
	}

	if !issuerutil.IsReady(issuerStatus) {
		err := errors.New("issuer is not ready")
		return ctrl.Result{}, err
	}

	//getting temp secret-name created by cert-manager for which holds the temp-private key
	secretName, ok := cr.ObjectMeta.Annotations[cmapi.CertificateRequestPrivateKeyAnnotationKey]
	if !ok || secretName == "" {
		message := fmt.Sprintf("Annotation %q missing or reference empty",
			cmapi.CertificateRequestPrivateKeyAnnotationKey)
		err := errors.New("secret name missing")		
		log.Error(err, message)
		r.setStatus(ctx, cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonDenied, err.Error())
		return nil, nil
	}

	//fetching the temp-private key from the secret
	secretsLister := ctx.KubeSharedInformerFactory.Core().V1().Secrets().Lister()
	privatekey, err := kube.SecretTLSKey(ctx, r.secretsLister, cr.Namespace, secretName)
	if k8sErrors.IsNotFound(err) {
		message := fmt.Sprintf("Referenced secret %s/%s not found", cr.Namespace, secretName)
		log.Error(err, message)
		r.setStatus(ctx, cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonDenied, err.Error())
		return nil, nil
	}
	
	if cmerrors.IsInvalidData(err) {
		message := fmt.Sprintf("Failed to get key %q referenced in annotation %q",
			secretName, cmapi.CertificateRequestPrivateKeyAnnotationKey)
		log.Error(err, message)
		r.setStatus(ctx, cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonDenied, err.Error())

		return nil, nil
	}

	if err != nil {
		// We are probably in a network error here so we should backoff and retry
		message := fmt.Sprintf("Failed to get certificate key pair from secret %s/%s", resourceNamespace, secretName)
		log.Error(err, message)
		return nil, r.setStatus(ctx, cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonDenied, err.Error())
	}

	//generating x509 certificate from the certificate request data to be signed later
	template, err := signer.GenerateTemplateFromCertificateRequest(cr)
	if err != nil {
		message := "Error generating certificate template"
		log.Error(err, message)
		r.setStatus(ctx, cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonDenied, err.Error())
		return nil, nil
	}

	template.CRLDistributionPoints = issuerSpec.SelfSigned.CRLDistributionPoints

	if template.Subject.String() == "" {
		// RFC 5280 (https://tools.ietf.org/html/rfc5280#section-4.1.2.4) says that:
		// "The issuer field MUST contain a non-empty distinguished name (DN)."
		// Since we're creating a self-signed cert, the issuer will match whatever is
		// in the template's subject DN.
		log.V(logf.DebugLevel).Info("issued cert will have an empty issuer DN, which contravenes RFC 5280. emitting warning event")
	}

	// extract the public component of the key
	publickey, err := signer.PublicKeyForPrivateKey(privatekey)
	if err != nil {
		message := "Failed to get public key from private key"
		s.reporter.Failed(cr, err, "ErrorPublicKey", message)
		log.Error(err, message)
		return nil, nil
	}

	//signing the certificate
	signedPEM, _, err := pkiutil.SignCertificate(template, template, publickey, privatekey)
	if err != nil {
		log.Error(err, "failed signing certificate")
		err := r.setStatus(ctx, log, &cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonFailed, "Failed to sign certificate: %v", err)
		return ctrl.Result{}, err
	}

	// Store the signed certificate data in the status
	cr.Status.Certificate = signedPEM
	// copy the CA data from the CA secret
	// We set the CA to the returned certificate here since this is self signed.
	cr.Status.CA = signedPEM
	// Finally, update the status as signed
	return ctrl.Result{}, r.setStatus(ctx, log, &cr, cmmeta.ConditionTrue, cmapi.CertificateRequestReasonIssued, "Successfully issued certificate")
}

// SetupWithManager initialises the CertificateRequest controller into the
// controller runtime.
func (r *CertificateRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cmapi.CertificateRequest{}).
		Complete(r)
}

// requestShouldBeProcessed will return false if the conditions on the request
// mean that it should not be processed. If the request has been denied, it
// will set the request failure time and add a Ready=False condition.
func (r *CertificateRequestReconciler) requestShouldBeProcessed(ctx context.Context, log logr.Logger, cr *cmapi.CertificateRequest) (bool, error) {
	dbg := log.V(4)//4 is debug level

	// Ignore CertificateRequest if it is already Ready
	if apiutil.CertificateRequestHasCondition(cr, cmapi.CertificateRequestCondition{
		Type:   cmapi.CertificateRequestConditionReady,
		Status: cmmeta.ConditionTrue,
	}) {
		dbg.Info("CertificateRequest is Ready. Ignoring.")
		return false, nil
	}
	// Ignore CertificateRequest if it is already Failed
	if apiutil.CertificateRequestHasCondition(cr, cmapi.CertificateRequestCondition{
		Type:   cmapi.CertificateRequestConditionReady,
		Status: cmmeta.ConditionFalse,
		Reason: cmapi.CertificateRequestReasonFailed,
	}) {
		dbg.Info("CertificateRequest is Failed. Ignoring.")
		return false, nil
	}
	// Ignore CertificateRequest if it already has a Denied Ready Reason
	if apiutil.CertificateRequestHasCondition(cr, cmapi.CertificateRequestCondition{
		Type:   cmapi.CertificateRequestConditionReady,
		Status: cmmeta.ConditionFalse,
		Reason: cmapi.CertificateRequestReasonDenied,
	}) {
		dbg.Info("CertificateRequest already has a Ready condition with Denied Reason. Ignoring.")
		return false, nil
	}

	// If CertificateRequest has been denied, mark the CertificateRequest as
	// Ready=Denied and set FailureTime if not already.
	if apiutil.CertificateRequestIsDenied(cr) {
		dbg.Info("CertificateRequest has been denied. Marking as failed.")

		if cr.Status.FailureTime == nil {
			nowTime := metav1.NewTime(r.Clock.Now())
			cr.Status.FailureTime = &nowTime
		}

		message := "The CertificateRequest was denied by an approval controller"
		return false, r.setStatus(ctx, cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonDenied, message)
	}

	if r.CheckApprovedCondition {
		// If CertificateRequest has not been approved, exit early.
		if !apiutil.CertificateRequestIsApproved(cr) {
			dbg.Info("certificate request has not been approved yet, ignoring")
			return false, nil
		}
	}

	return true, nil
}

//sets the condition of certificate request
func (r *CertificateRequestReconciler) setStatus(ctx context.Context, cr *cmapi.CertificateRequest, status cmmeta.ConditionStatus, reason, message string, args ...interface{}) error {
	// Format the message and update the localCA variable with the new Condition
	completeMessage := fmt.Sprintf(message, args...)
	apiutil.SetCertificateRequestCondition(cr, cmapi.CertificateRequestConditionReady, status, reason, completeMessage)

	// Fire an Event to additionally inform users of the change
	eventType := core.EventTypeNormal
	if status == cmmeta.ConditionFalse {
		eventType = core.EventTypeWarning
	}
	r.Recorder.Event(cr, eventType, reason, completeMessage)
	log.Info(completeMessage)

	return r.Status().Update(ctx, cr) //return r.Client.Status().Update(ctx, cr)??
}