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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// ChaosClusterIssuer is the Schema for the chaosclusterissuers API
type ChaosClusterIssuer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ChaosIssuerSpec   `json:"spec,omitempty"`
	Status ChaosIssuerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ChaosClusterIssuerList contains a list of ChaosClusterIssuer
type ChaosClusterIssuerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ChaosClusterIssuer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ChaosClusterIssuer{}, &ChaosClusterIssuerList{})
}
