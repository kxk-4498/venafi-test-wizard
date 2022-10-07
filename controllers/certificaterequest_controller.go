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
	"github.com/kxk-4498/Venafi-test-wizard/issuer"

	api "github.com/kxk-4498/Venafi-test-wizard/api/v1alpha1"
	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	apiutil "github.com/cert-manager/cert-manager/pkg/api/util"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	issuerutil "github.com/kxk-4498/Venafi-test-wizard/util"
)

var (
	errIssuerRef      = errors.New("error interpreting issuerRef")
	errGetIssuer      = errors.New("error getting issuer")
	errIssuerNotReady = errors.New("issuer is not ready")
	errSignerBuilder  = errors.New("failed to build the signer")
	errSignerSign     = errors.New("failed to sign")
)


// CertificateRequestReconciler reconciles a CertificateRequest object
type CertificateRequestReconciler struct {
	client.Client
	Scheme                   *runtime.Scheme

	Log                      logr.Logger
	Recorder 				 record.EventRecorder
	Clock                    clock.Clock
	CheckApprovedCondition   bool
}
// annotation for generating RBAC role for writing events
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=certmanager.chaos.ch,resources=certificaterequests,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=certmanager.chaos.ch,resources=certificaterequests/status,verbs=get;update;patch



// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CertificateRequest object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
// Reconcile will read and validate a ChaosIssuer resource associated to the
// CertificateRequest resource, and it will sign the CertificateRequest with the
// provisioner in the ChaosIssuer.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile

func (r *CertificateRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	log := r.Log.WithValues("certificaterequests", req.NamespacedName)

	// Fetch the CertificateRequest resource being reconciled.
	// Just ignore the request if the certificate request has been deleted.
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
	//##############################################################
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
		log = log.WithValues("issuer", issuerName)
	case *api.ChaosClusterIssuer:
		secretNamespace = r.ClusterResourceNamespace
		log = log.WithValues("clusterissuer", issuerName)
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
		return ctrl.Result{}, errIssuerNotReady
	}

	secretName := types.NamespacedName{
		Name:      issuerSpec.AuthSecretName,
		Namespace: secretNamespace,
	}

	var secret corev1.Secret
	if err := r.Get(ctx, secretName, &secret); err != nil {
		return ctrl.Result{}, fmt.Errorf("%w, secret name: %s, reason: %v", errGetAuthSecret, secretName, err)
	}

	signer, err := r.SignerBuilder(issuerSpec, secret.Data)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("%w: %v", errSignerBuilder, err)
	}

	signed, err := issuer.Sign(cr.Spec.Request)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("%w: %v", errSignerSign, err)
	}
	cr.Status.Certificate = signed

	r.setStatus(ctx, cr, cmmeta.ConditionTrue, cmapi.CertificateRequestReasonIssued, "Signed")
	return ctrl.Result{}, nil
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
	completeMessage := fmt.Sprintf(message, args...)
	apiutil.SetCertificateRequestCondition(cr, cmapi.CertificateRequestConditionReady, status, reason, completeMessage)

	// Fire an Event to additionally inform users of the change
	eventType := core.EventTypeNormal
	if status == cmmeta.ConditionFalse {
		eventType = core.EventTypeWarning
	}
	r.Recorder.Event(cr, eventType, reason, completeMessage)

	return r.Status().Update(ctx, cr) //return r.Client.Status().Update(ctx, cr)??
}