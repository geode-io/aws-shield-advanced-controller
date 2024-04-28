package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProtectionSpec defines the desired state of Protection
type ProtectionSpec struct {
	// The resource ARN to protect with Shield Advanced
	ResourceArn string `json:"resourceArn,omitempty"`
}

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
