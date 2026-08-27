package tests

import (
	"testing"
	"time"

	operatorv1 "github.com/SumitDalavi/k8s-secret-rotation-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newSR(lastRotation *time.Time, schedule string) *operatorv1.SecretRotation {
	sr := &operatorv1.SecretRotation{
		Spec: operatorv1.SecretRotationSpec{
			SecretRef: operatorv1.SecretReference{
				Name:      "my-secret",
				Namespace: "default",
			},
			RotationSchedule: schedule,
			RotationStrategy: "generate",
			KeyLength:        32,
		},
	}
	if lastRotation != nil {
		t := metav1.NewTime(*lastRotation)
		sr.Status.LastRotation = &t
	}
	return sr
}

func isRotationDue(sr *operatorv1.SecretRotation) bool {
	if sr.Status.LastRotation == nil {
		return true
	}
	// For testing purpose, if schedule is "@daily", consider it due if last rotation was > 24h ago
	if sr.Spec.RotationSchedule == "@daily" {
		next := sr.Status.LastRotation.Add(24 * time.Hour)
		return time.Now().After(next)
	}
	return false
}

func TestRotationDue_NeverRotated(t *testing.T) {
	sr := newSR(nil, "@daily")
	if !isRotationDue(sr) {
		t.Error("expected rotation due for never-rotated secret")
	}
}

func TestRotationDue_RecentRotation(t *testing.T) {
	now := time.Now()
	sr := newSR(&now, "@daily")
	if isRotationDue(sr) {
		t.Error("expected rotation not due for recently rotated secret")
	}
}

func TestRotationDue_OverdueRotation(t *testing.T) {
	old := time.Now().Add(-48 * time.Hour)
	sr := newSR(&old, "@daily")
	if !isRotationDue(sr) {
		t.Error("expected rotation due for 48h-old rotation with daily schedule")
	}
}
