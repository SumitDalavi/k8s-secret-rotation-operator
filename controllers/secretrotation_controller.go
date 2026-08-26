package controllers

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1 "github.com/SumitDalavi/k8s-secret-rotation-operator/api/v1alpha1"
	"github.com/SumitDalavi/k8s-secret-rotation-operator/internal/vault"
	"github.com/SumitDalavi/k8s-secret-rotation-operator/internal/aws"
)

// SecretRotationReconciler reconciles SecretRotation objects
type SecretRotationReconciler struct {
	client.Client
	VaultClient *vault.Client
}

func (r *SecretRotationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the SecretRotation CR
	sr := &operatorv1.SecretRotation{}
	if err := r.Get(ctx, req.NamespacedName, sr); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Check if rotation is due
	if !r.isRotationDue(sr) {
		next := sr.Status.NextRotationTime.Time.Sub(time.Now())
		if next < 0 {
			next = time.Minute
		}
		return ctrl.Result{RequeueAfter: next}, nil
	}

	log.Info("Rotating secret", "name", sr.Name, "namespace", sr.Namespace)

	// Perform rotation based on provider
	var secretData map[string]string
	var rotateErr error

	switch sr.Spec.Provider {
	case "vault":
		secretData, rotateErr = r.VaultClient.GetSecret(ctx, sr.Spec.SecretPath)
	case "aws":
		rotateErr = aws.RotateSecret(ctx, sr.Spec.SecretPath)
	default:
		rotateErr = fmt.Errorf("unsupported provider: %s", sr.Spec.Provider)
	}

	if rotateErr != nil {
		log.Error(rotateErr, "Rotation failed")
		r.setCondition(sr, "RotationFailed", metav1.ConditionTrue, "RotationError", rotateErr.Error())
		sr.Status.Phase = "Failed"
		r.Status().Update(ctx, sr)
		return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
	}

	// Update Kubernetes Secret with new values
	if secretData != nil {
		if err := r.updateK8sSecret(ctx, sr, secretData); err != nil {
			return ctrl.Result{}, err
		}
		// Rolling restart of annotated workloads
		if err := r.triggerRollingRestart(ctx, sr); err != nil {
			log.Error(err, "Rolling restart failed")
		}
	}

	// Update status
	now := metav1.Now()
	sr.Status.LastRotationTime = &now
	next := metav1.NewTime(now.Add(sr.Spec.RotationInterval.Duration))
	sr.Status.NextRotationTime = &next
	sr.Status.Phase = "Rotated"
	sr.Status.RotationCount++
	r.setCondition(sr, "LastRotated", metav1.ConditionTrue, "RotationSucceeded", "Secret rotated successfully")

	if err := r.Status().Update(ctx, sr); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: sr.Spec.RotationInterval.Duration}, nil
}

func (r *SecretRotationReconciler) isRotationDue(sr *operatorv1.SecretRotation) bool {
	if sr.Status.LastRotationTime == nil {
		return true // never rotated
	}
	next := sr.Status.LastRotationTime.Add(sr.Spec.RotationInterval.Duration)
	return time.Now().After(next)
}

func (r *SecretRotationReconciler) updateK8sSecret(ctx context.Context, sr *operatorv1.SecretRotation, data map[string]string) error {
	secret := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{Name: sr.Spec.TargetSecret, Namespace: sr.Namespace}, secret)
	if errors.IsNotFound(err) {
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: sr.Spec.TargetSecret, Namespace: sr.Namespace},
			Type:       corev1.SecretTypeOpaque,
		}
	}
	if secret.StringData == nil {
		secret.StringData = make(map[string]string)
	}
	for k, v := range data {
		secret.StringData[k] = v
	}
	if errors.IsNotFound(err) {
		return r.Create(ctx, secret)
	}
	return r.Update(ctx, secret)
}

func (r *SecretRotationReconciler) triggerRollingRestart(ctx context.Context, sr *operatorv1.SecretRotation) error {
	// Patch each workload deployment with a restart annotation to trigger rolling update
	for _, ref := range sr.Spec.WorkloadRefs {
		patch := client.MergeFrom(&corev1.Pod{})
		_ = r.Patch(ctx, &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: ref.Name, Namespace: sr.Namespace,
				Annotations: map[string]string{
					"secret-rotation-operator/restart-at": time.Now().Format(time.RFC3339),
				},
			},
		}, patch)
	}
	return nil
}

func (r *SecretRotationReconciler) setCondition(sr *operatorv1.SecretRotation, condType string, status metav1.ConditionStatus, reason, msg string) {
	sr.Status.Conditions = append(sr.Status.Conditions, metav1.Condition{
		Type: condType, Status: status, Reason: reason, Message: msg,
		LastTransitionTime: metav1.Now(),
	})
}

func (r *SecretRotationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1.SecretRotation{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
