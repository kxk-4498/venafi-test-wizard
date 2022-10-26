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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	selfsignedissuerv1alpha1 "github.com/kxk-4498/Venafi-test-wizard/api/v1alpha1"
	issuerutil "github.com/kxk-4498/Venafi-test-wizard/issuer/util"
)

const (
	issuerReadyConditionReason = "self-signed-issuer.ChaosIssuerController.Reconcile"
)

// ChaosIssuerReconciler reconciles a ChaosIssuer object
type ChaosIssuerReconciler struct {
	client.Client
	Kind   string
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=self-signed-issuer.chaos.ch,resources=chaosissuers;chaosclusterissuers,verbs=get;list;watch
//+kubebuilder:rbac:groups=self-signed-issuer.chaos.ch,resources=chaosissuers/status;chaosclusterissuers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile

func (r *ChaosIssuerReconciler) newIssuer() (client.Object, error) {
	issuerGVK := selfsignedissuerv1alpha1.GroupVersion.WithKind(r.Kind)
	ro, err := r.Scheme.New(issuerGVK)
	if err != nil {
		return nil, err
	}
	return ro.(client.Object), nil
}

func (r *ChaosIssuerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	issuer, err := r.newIssuer()
	if err != nil {
		log.Error(err, "Unrecognised issuer type")
		return ctrl.Result{}, nil
	}
	if err := r.Get(ctx, req.NamespacedName, issuer); err != nil {
		if err := client.IgnoreNotFound(err); err != nil {
			return ctrl.Result{}, fmt.Errorf("unexpected get error: %v", err)
		}
		log.Info("Not found. Ignoring.")
		return ctrl.Result{}, nil
	}

	issuerStatus, err := issuerutil.GetStatus(issuer)

	issuerutil.SetReadyCondition(issuerStatus, selfsignedissuerv1alpha1.ConditionTrue, issuerReadyConditionReason, "Success")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ChaosIssuerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&selfsignedissuerv1alpha1.ChaosIssuer{}).
		Complete(r)
}
