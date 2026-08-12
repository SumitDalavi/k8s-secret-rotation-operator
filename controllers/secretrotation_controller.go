package controllers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	secretopsv1alpha1 "github.com/example/k8s-secret-rotation-operator/api/v1alpha1"
)

// SecretRotationReconciler reconciles a SecretRotation object
type SecretRotationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=secretops.io,resources=secretrotations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=secretops.io,resources=secretrotations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;update;patch

func (r *SecretRotationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// 1. Fetch the SecretRotation resource
	var rotation secretopsv1alpha1.SecretRotation
	if err := r.Get(ctx, req.NamespacedName, &rotation); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	logger.Info("Reconciling SecretRotation",
		"secret", rotation.Spec.SecretRef.Name,
		"namespace", rotation.Spec.SecretRef.Namespace)

	// 2. Fetch the target Secret
	var secret corev1.Secret
	secretKey := types.NamespacedName{
		Name:      rotation.Spec.SecretRef.Name,
		Namespace: rotation.Spec.SecretRef.Namespace,
	}
	if err := r.Get(ctx, secretKey, &secret); err != nil {
		if errors.IsNotFound(err) {
			rotation.Status.Phase = "Failed"
			rotation.Status.Message = fmt.Sprintf("Target secret %s not found", rotation.Spec.SecretRef.Name)
			r.Status().Update(ctx, &rotation)
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
		}
		return ctrl.Result{}, err
	}

	// 3. Check if rotation is due
	now := metav1.Now()
	shouldRotate := rotation.Status.LastRotation == nil // First run

	if !shouldRotate && rotation.Status.NextRotation != nil {
		shouldRotate = now.After(rotation.Status.NextRotation.Time)
	}

	if !shouldRotate {
		logger.Info("Rotation not yet due, requeueing")
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
	}

	// 4. Perform rotation
	logger.Info("Rotating secret", "secret", secret.Name)

	if rotation.Spec.RotationStrategy == "generate" {
		newValue, err := generateRandomSecret(rotation.Spec.KeyLength)
		if err != nil {
			rotation.Status.Phase = "Failed"
			rotation.Status.Message = fmt.Sprintf("Failed to generate secret: %v", err)
			r.Status().Update(ctx, &rotation)
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
		}

		// Update all keys in the secret with rotated values
		for key := range secret.Data {
			secret.Data[key] = []byte(newValue)
		}
	}

	// 5. Apply the updated secret
	if err := r.Update(ctx, &secret); err != nil {
		rotation.Status.Phase = "Failed"
		rotation.Status.Message = fmt.Sprintf("Failed to update secret: %v", err)
		r.Status().Update(ctx, &rotation)
		return ctrl.Result{}, err
	}

	// 6. Update status
	rotation.Status.LastRotation = &now
	nextRotation := metav1.NewTime(now.Add(7 * 24 * time.Hour)) // Default: weekly
	rotation.Status.NextRotation = &nextRotation
	rotation.Status.RotationCount++
	rotation.Status.Phase = "Active"
	rotation.Status.Message = "Secret rotated successfully"

	if err := r.Status().Update(ctx, &rotation); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Secret rotated successfully",
		"rotationCount", rotation.Status.RotationCount,
		"nextRotation", nextRotation.Time)

	return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
}

func generateRandomSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

func (r *SecretRotationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&secretopsv1alpha1.SecretRotation{}).
		Complete(r)
}
