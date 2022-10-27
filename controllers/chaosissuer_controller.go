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
	"github.com/go-logr/logr"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	selfsignedissuerv1alpha1 "github.com/kxk-4498/Venafi-test-wizard/api/v1alpha1"
)

// ChaosIssuerReconciler reconciles a ChaosIssuer object
type ChaosIssuerReconciler struct {
	client.Client
	Log      logr.Logger
	Clock    clock.Clock
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=self-signed-issuer.chaos.ch,resources=chaosissuers;chaosclusterissuers,verbs=get;list;watch
//+kubebuilder:rbac:groups=self-signed-issuer.chaos.ch,resources=chaosissuers/status;chaosclusterissuers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

func (r *ChaosIssuerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("chaosIssuer", req.NamespacedName)

	//fetch the issuer being synced
	chaosIssuer := selfsignedissuerv1alpha1.ChaosIssuer{}
	if err := r.Client.Get(ctx, req.NamespacedName, &chaosIssuer); err != nil {
		log.Error(err, "failed to retrieve chaosIssuer resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return ctrl.Result{}, r.setChaosIssuerStatus(ctx, log, &chaosIssuer, selfsignedissuerv1alpha1.ConditionTrue, "Verified", "Signing ChaosIssuer verified and ready to issue certificates")
}

func (r *ChaosIssuerReconciler) setChaosIssuerStatus(ctx context.Context, log logr.Logger, chaosIssuer *selfsignedissuerv1alpha1.ChaosIssuer, status selfsignedissuerv1alpha1.ConditionStatus, reason, message string, args ...interface{}) error {
	// Format the message and update the issuer variables with the new Condition
	completeMessage := fmt.Sprintf(message, args...)
	r.setChaosIssuerCondition(log, chaosIssuer, selfsignedissuerv1alpha1.IssuerConditionReady, status, reason, completeMessage)
	// Fire an Event to additionally inform users of the change
	eventType := core.EventTypeNormal
	if status == selfsignedissuerv1alpha1.ConditionFalse {
		eventType = core.EventTypeWarning
	}
	r.Recorder.Event(chaosIssuer, eventType, reason, completeMessage)

	// Actually update the issuer resource
	return r.Client.Status().Update(ctx, chaosIssuer)
}

//		setChaosIssuerCondition will set a 'condition' on the given ChaosIssuer.
//	  - If no condition of the same type already exists, the condition will be
//	    inserted with the LastTransitionTime set to the current time.
//	  - If a condition of the same type and state already exists, the condition
//	    will be updated but the LastTransitionTime will not be modified.
//	  - If a condition of the same type and different state already exists, the
//	    condition will be updated and the LastTransitionTime set to the current
//	    time.
func (r *ChaosIssuerReconciler) setChaosIssuerCondition(log logr.Logger, chaosIssuer *selfsignedissuerv1alpha1.ChaosIssuer, conditionType selfsignedissuerv1alpha1.IssuerConditionType, status selfsignedissuerv1alpha1.ConditionStatus, reason, message string) {
	newCondition := selfsignedissuerv1alpha1.chaosIssuerCondition{
		Type:    conditionType,
		Status:  status,
		Reason:  reason,
		Message: message,
	}

	nowTime := metav1.NewTime(r.Clock.Now())
	newCondition.LastTransitionTime = &nowTime

	// Search through existing conditions
	for idx, cond := range chaosIssuer.Status.Conditions {
		// Skip unrelated conditions
		if cond.Type != conditionType {
			continue
		}

		// If this update doesn't contain a state transition, we don't update
		// the conditions LastTransitionTime to Now()
		if cond.Status == status {
			newCondition.LastTransitionTime = cond.LastTransitionTime
		} else {
			log.Info("found status change for chaosIssuer condition; setting lastTransitionTime", "condition", conditionType, "old_status", cond.Status, "new_status", status, "time", nowTime.Time)
		}

		// Overwrite the existing condition
		chaosIssuer.Status.Conditions[idx] = newCondition
		return
	}

	// If we've not found an existing condition of this type, we simply insert
	// the new condition into the slice.
	chaosIssuer.Status.Conditions = append(chaosIssuer.Status.Conditions, newCondition)
	log.Info("setting lastTransitionTime for chaosIssuer condition", "condition", conditionType, "time", nowTime.Time)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ChaosIssuerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&selfsignedissuerv1alpha1.ChaosIssuer{}).
		Complete(r)
}
