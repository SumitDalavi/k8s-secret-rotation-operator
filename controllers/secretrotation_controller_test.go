package controllers

import (
	"context"
	"testing"

	operatorv1 "github.com/SumitDalavi/k8s-secret-rotation-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = operatorv1.AddToScheme(s)
	_ = corev1.AddToScheme(s)
	return s
}

func TestReconcile_NotFound(t *testing.T) {
	s := newScheme()
	fakeClient := fake.NewClientBuilder().WithScheme(s).Build()
	r := &SecretRotationReconciler{
		Client: fakeClient,
		Scheme: s,
	}

	req := ctrl.Request{NamespacedName: types.NamespacedName{Name: "nonexistent", Namespace: "default"}}
	res, err := r.Reconcile(context.Background(), req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if res.Requeue {
		t.Errorf("did not expect requeue")
	}
}

func TestReconcile_Success(t *testing.T) {
	s := newScheme()

	sr := &operatorv1.SecretRotation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-rotation",
			Namespace: "default",
		},
		Spec: operatorv1.SecretRotationSpec{
			SecretRef:        operatorv1.SecretReference{Name: "my-secret", Namespace: "default"},
			RotationSchedule: "0 2 * * 0",
			RotationStrategy: "generate",
			KeyLength:        32,
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(s).
		WithStatusSubresource(&operatorv1.SecretRotation{}).
		WithObjects(sr).
		Build()

	r := &SecretRotationReconciler{
		Client: fakeClient,
		Scheme: s,
	}
	req := ctrl.Request{NamespacedName: types.NamespacedName{Name: "my-rotation", Namespace: "default"}}

	res, err := r.Reconcile(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RequeueAfter == 0 {
		t.Errorf("expected RequeueAfter to be set")
	}

	// Verify status was updated
	updated := &operatorv1.SecretRotation{}
	if err := fakeClient.Get(context.Background(), req.NamespacedName, updated); err != nil {
		t.Fatalf("could not fetch updated object: %v", err)
	}
}

func TestReconcile_MultipleRotations(t *testing.T) {
	s := newScheme()

	sr := &operatorv1.SecretRotation{
		ObjectMeta: metav1.ObjectMeta{Name: "multi-rot", Namespace: "default"},
		Spec: operatorv1.SecretRotationSpec{
			SecretRef:        operatorv1.SecretReference{Name: "sec", Namespace: "default"},
			RotationSchedule: "0 1 * * 0",
			RotationStrategy: "external",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(s).
		WithStatusSubresource(&operatorv1.SecretRotation{}).
		WithObjects(sr).
		Build()

	r := &SecretRotationReconciler{
		Client: fakeClient,
		Scheme: s,
	}
	req := ctrl.Request{NamespacedName: types.NamespacedName{Name: "multi-rot", Namespace: "default"}}

	for i := 0; i < 3; i++ {
		_, err := r.Reconcile(context.Background(), req)
		if err != nil {
			t.Fatalf("reconcile %d failed: %v", i+1, err)
		}
	}
}
