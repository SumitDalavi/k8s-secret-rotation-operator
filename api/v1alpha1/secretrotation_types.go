package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// SecretRotationSpec defines the desired state of SecretRotation
type SecretRotationSpec struct {
	// SecretRef is a reference to the Kubernetes Secret to rotate
	SecretRef SecretReference `json:"secretRef"`

	// RotationSchedule is a cron expression for when to rotate (e.g., "0 2 * * 0")
	RotationSchedule string `json:"rotationSchedule"`

	// RotationStrategy defines how new values are generated
	// +kubebuilder:validation:Enum=generate;external
	RotationStrategy string `json:"rotationStrategy"`

	// KeyLength is the length of auto-generated secret values (used when strategy=generate)
	// +kubebuilder:default=32
	KeyLength int `json:"keyLength,omitempty"`
}

// SecretReference contains the name and namespace of the target secret
type SecretReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// SecretRotationStatus defines the observed state of SecretRotation
type SecretRotationStatus struct {
	// LastRotation is the timestamp of the last successful rotation
	LastRotation *metav1.Time `json:"lastRotation,omitempty"`

	// NextRotation is the calculated timestamp of the next rotation
	NextRotation *metav1.Time `json:"nextRotation,omitempty"`

	// RotationCount is the total number of rotations performed
	RotationCount int `json:"rotationCount"`

	// Phase is the current state of the rotation
	// +kubebuilder:validation:Enum=Pending;Active;Failed
	Phase string `json:"phase,omitempty"`

	// Message provides additional detail about the current phase
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Secret",type=string,JSONPath=".spec.secretRef.name"
// +kubebuilder:printcolumn:name="Schedule",type=string,JSONPath=".spec.rotationSchedule"
// +kubebuilder:printcolumn:name="Last Rotated",type=date,JSONPath=".status.lastRotation"
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=".status.phase"

// SecretRotation is the Schema for the secretrotations API
type SecretRotation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecretRotationSpec   `json:"spec,omitempty"`
	Status SecretRotationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecretRotationList contains a list of SecretRotation
type SecretRotationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretRotation `json:"items"`
}

func (in *SecretRotation) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	out := new(SecretRotation)
	in.DeepCopyInto(out)
	return out
}

func (in *SecretRotation) DeepCopyInto(out *SecretRotation) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

func (in *SecretRotationList) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	out := new(SecretRotationList)
	in.DeepCopyInto(out)
	return out
}

func (in *SecretRotationList) DeepCopyInto(out *SecretRotationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SecretRotation, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}
