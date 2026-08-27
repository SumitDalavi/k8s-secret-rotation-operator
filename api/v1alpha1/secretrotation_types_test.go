package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeepCopy_SecretRotation(t *testing.T) {
	now := metav1.Now()
	sr := &SecretRotation{
		Spec: SecretRotationSpec{
			SecretRef:        SecretReference{Name: "my-secret", Namespace: "default"},
			RotationSchedule: "0 2 * * 0",
			RotationStrategy: "generate",
			KeyLength:        32,
		},
		Status: SecretRotationStatus{
			LastRotation:  &now,
			NextRotation:  &now,
			RotationCount: 5,
			Phase:         "Rotated",
			Message:       "success",
		},
	}
	sr.Labels = map[string]string{"key": "value"}

	out := new(SecretRotation)
	sr.DeepCopyInto(out)
	if out.Spec.KeyLength != 32 {
		t.Errorf("expected KeyLength 32, got %d", out.Spec.KeyLength)
	}
	if out.Spec.SecretRef.Name != "my-secret" {
		t.Errorf("expected secret name my-secret")
	}

	obj := sr.DeepCopyObject()
	if obj == nil {
		t.Fatal("expected non-nil deep copy object")
	}
}

func TestDeepCopy_NilSecretRotation(t *testing.T) {
	var sr *SecretRotation
	if sr.DeepCopyObject() != nil {
		t.Error("expected nil for nil input")
	}
}

func TestDeepCopy_SecretRotationList(t *testing.T) {
	srl := &SecretRotationList{
		Items: []SecretRotation{
			{Spec: SecretRotationSpec{RotationStrategy: "generate"}},
			{Spec: SecretRotationSpec{RotationStrategy: "external"}},
		},
	}
	out := new(SecretRotationList)
	srl.DeepCopyInto(out)
	if len(out.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(out.Items))
	}

	obj := srl.DeepCopyObject()
	if obj == nil {
		t.Fatal("expected non-nil deep copy object")
	}
}

func TestDeepCopy_NilSecretRotationList(t *testing.T) {
	var srl *SecretRotationList
	if srl.DeepCopyObject() != nil {
		t.Error("expected nil for nil input")
	}
}

func TestDeepCopy_NoItems(t *testing.T) {
	srl := &SecretRotationList{}
	out := new(SecretRotationList)
	srl.DeepCopyInto(out)
	if len(out.Items) != 0 {
		t.Errorf("expected 0 items")
	}
}
