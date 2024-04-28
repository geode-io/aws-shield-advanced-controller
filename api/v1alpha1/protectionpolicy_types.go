package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProtectionPolicySpec defines the desired state of ProtectionPolicy
type ProtectionPolicySpec struct {

	// MatchResourceTypes is a list of resource types to match
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	MatchResourceTypes []ResourceType `json:"matchResourceTypes"`

	// MatchRegions is a list of regions to match
	// +kubebuilder:validation:MinItems=1
	MatchRegions []string `json:"matchRegions,omitempty"`
}

// ResourceType identifies the type of resource to match
// +kubebuilder:validation:Enum=cloudfront/distribution;route53/hostedzone;globalaccelerator/accelerator;ec2/eip;elasticloadbalancing/loadbalancer/app;elasticloadbalancing/loadbalancer/classic
type ResourceType string

// ProtectionPolicyStatus defines the observed state of ProtectionPolicy
type ProtectionPolicyStatus struct {
	Protections []ProtectionStatus `json:"protections,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ProtectionPolicy is the Schema for the protectionpolicies API
type ProtectionPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProtectionPolicySpec   `json:"spec,omitempty"`
	Status ProtectionPolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProtectionPolicyList contains a list of ProtectionPolicy
type ProtectionPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProtectionPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProtectionPolicy{}, &ProtectionPolicyList{})
}
