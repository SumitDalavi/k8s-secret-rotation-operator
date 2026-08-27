package controllers

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1 "github.com/SumitDalavi/k8s-secret-rotation-operator/api/v1alpha1"
)

// SecretRotationReconciler reconciles SecretRotation objects
type SecretRotationReconciler struct {
	client.Client
}

func (r *SecretRotationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// Fetch the SecretRotation CR
	sr := &operatorv1.SecretRotation{}
	if err := r.Get(ctx, req.NamespacedName, sr); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Update status
	now := metav1.Now()
	sr.Status.LastRotation = &now
	next := metav1.NewTime(now.Add(24 * time.Hour))
	sr.Status.NextRotation = &next
	sr.Status.Phase = "Rotated"
	sr.Status.RotationCount++

	if err := r.Status().Update(ctx, sr); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 24 * time.Hour}, nil
}

func (r *SecretRotationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1.SecretRotation{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
