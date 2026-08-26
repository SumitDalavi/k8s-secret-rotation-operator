package tests

import (
	"testing"
	"time"

	operatorv1 "github.com/SumitDalavi/k8s-secret-rotation-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newSR(lastRotation *time.Time, interval time.Duration) *operatorv1.SecretRotation {
	sr := &operatorv1.SecretRotation{
		Spec: operatorv1.SecretRotationSpec{
			Provider:         "vault",
			SecretPath:       "secret/data/app/db",
			TargetSecret:     "db-creds",
			RotationInterval: metav1.Duration{Duration: interval},
		},
	}
	if lastRotation != nil {
		t := metav1.NewTime(*lastRotation)
		sr.Status.LastRotationTime = &t
	}
	return sr
}

func isRotationDue(sr *operatorv1.SecretRotation) bool {
	if sr.Status.LastRotationTime == nil {
		return true
	}
	next := sr.Status.LastRotationTime.Add(sr.Spec.RotationInterval.Duration)
	return time.Now().After(next)
}

func TestRotationDue_NeverRotated(t *testing.T) {
	sr := newSR(nil, 24*time.Hour)
	if !isRotationDue(sr) {
		t.Error("expected rotation due for never-rotated secret")
	}
}

func TestRotationDue_RecentRotation(t *testing.T) {
	now := time.Now()
	sr := newSR(&now, 24*time.Hour)
	if isRotationDue(sr) {
		t.Error("expected rotation not due for recently rotated secret")
	}
}

func TestRotationDue_OverdueRotation(t *testing.T) {
	old := time.Now().Add(-48 * time.Hour)
	sr := newSR(&old, 24*time.Hour)
	if !isRotationDue(sr) {
		t.Error("expected rotation due for 48h-old rotation with 24h interval")
	}
}
