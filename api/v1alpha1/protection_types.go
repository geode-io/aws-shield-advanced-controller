/*
Copyright 2024.

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

// ProtectionSpec defines the desired state of Protection
type ProtectionSpec struct {
	// The resource ARN to protect with Shield Advanced
	ResourceArn string `json:"resourceArn,omitempty"`
}

// ProtectionStatus defines the observed state of Protection
type ProtectionStatus struct {
	// +kubebuilder:default=Inactive
	State         ProtectionState `json:"state,omitempty"`
	ProtectionArn string          `json:"protectionArn,omitempty"`
}

// ProtectionState describes the status of the protection in AWS Shield Advanced.
type ProtectionState string

const (
	// ProtectionStateActive indicates that the protection is active.
	ProtectionStateActive ProtectionState = "Active"

	// ProtectionStateInactive indicates that the protection is inactive
	ProtectionStateInactive ProtectionState = "Inactive"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Protection is the Schema for the protections API
type Protection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProtectionSpec   `json:"spec,omitempty"`
	Status ProtectionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProtectionList contains a list of Protection
type ProtectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Protection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Protection{}, &ProtectionList{})
}
