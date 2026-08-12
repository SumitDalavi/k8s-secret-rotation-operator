# Architecture: Kubernetes Secret Rotation Operator

## The Reconciliation Pattern
Kubernetes operators follow the **reconciliation loop** pattern:
1. Watch for changes to the `SecretRotation` custom resource
2. Compare desired state (the CR spec) to actual state (the Secret's current value)
3. Take corrective action to converge (rotate the secret)
4. Update status to reflect the new state

## CRD Design
The `SecretRotation` CRD extends the Kubernetes API with a new resource type. It has:
- **Spec**: What the user wants (which secret, when to rotate, how to generate new values)
- **Status**: What's actually happening (last rotation time, count, phase)

## RBAC
The operator only has the minimum permissions needed:
- Read/write `SecretRotation` resources and their status
- Read/write `Secret` resources (to actually rotate values)
- Create events (for audit logging)

## Why Go + kubebuilder?
Go is the native Kubernetes language. `controller-runtime` provides the production-grade reconciliation framework (leader election, health probes, metrics). kubebuilder scaffolds the project structure that the Kubernetes community expects.
