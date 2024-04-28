package v1alpha1

// ProtectionStatus defines the observed state of a protection
type ProtectionStatus struct {
	// +kubebuilder:default=Inactive
	State         ProtectionState `json:"state,omitempty"`
	ProtectionArn string          `json:"protectionArn,omitempty"`
	ResourceArn   string          `json:"resourceArn,omitempty"`
}

// ProtectionState describes the status of the protection in AWS Shield Advanced.
type ProtectionState string

const (
	// ProtectionStateActive indicates that the protection is active.
	ProtectionStateActive ProtectionState = "Active"

	// ProtectionStateInactive indicates that the protection is inactive
	ProtectionStateInactive ProtectionState = "Inactive"
)
