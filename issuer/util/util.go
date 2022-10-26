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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	api "github.com/kxk-4498/Venafi-test-wizard/api/v1alpha1"
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

func SetReadyCondition(status *api.ChaosIssuerStatus, conditionStatus api.ConditionStatus, reason, message string) {
	ready := GetReadyCondition(status)
	if ready == nil {
		ready = &api.IssuerCondition{
			Type: api.IssuerConditionReady,
		}
		status.Conditions = append(status.Conditions, *ready)
	}
	if ready.Status != conditionStatus {
		ready.Status = conditionStatus
		now := metav1.Now()
		ready.LastTransitionTime = &now
	}
	ready.Reason = reason
	ready.Message = message

	for i, c := range status.Conditions {
		if c.Type == api.IssuerConditionReady {
			status.Conditions[i] = *ready
			return
		}
	}
}

func GetReadyCondition(status *api.ChaosIssuerStatus) *api.IssuerCondition {
	for _, c := range status.Conditions {
		if c.Type == api.IssuerConditionReady {
			return &c
		}
	}
	return nil
}

func IsReady(status *api.ChaosIssuerStatus) bool {
	if c := GetReadyCondition(status); c != nil {
		return c.Status == api.ConditionTrue
	}
	return false
}