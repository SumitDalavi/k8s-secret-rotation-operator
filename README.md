# Kubernetes Secret Rotation Operator 🔄🔐

> A custom Kubernetes operator built with kubebuilder that automatically rotates secrets on a configurable schedule — moving past Helm/Rancher-level K8s into controller/API-level engineering.

## The Problem

Kubernetes Secrets are static by default. Once created, they never change unless someone manually updates them. In enterprise environments, compliance frameworks (SOC2, PCI-DSS) require credentials to be rotated regularly. Doing this manually is error-prone and doesn't scale.

## The Solution

This operator introduces a Custom Resource Definition (CRD) called `SecretRotation` that watches for secrets matching a label selector and rotates them on a configurable schedule. It demonstrates:

- **Custom Resource Definitions (CRDs)**: Extending the Kubernetes API
- **Controller Pattern**: Reconciliation loop that converges actual state to desired state
- **RBAC**: Operator-specific service accounts and roles

```yaml
apiVersion: secretops.io/v1alpha1
kind: SecretRotation
metadata:
  name: db-credentials-rotation
spec:
  secretRef:
    name: db-credentials
    namespace: production
  rotationSchedule: "0 2 * * 0"  # Weekly at 2am Sunday
  rotationStrategy: generate     # Auto-generate new value
  keyLength: 32
```

## Why This Over the Obvious Alternative

Most K8s portfolios show Helm charts and `kubectl apply`. This project operates **at the Kubernetes API level** — defining custom resources, writing reconciliation controllers, and managing RBAC. It's the difference between *using* Kubernetes and *extending* it, which is what platform-engineer technical rounds test.

## 📁 Project Structure

```
├── api/v1alpha1/
│   └── secretrotation_types.go   # CRD type definitions
├── controllers/
│   └── secretrotation_controller.go  # Reconciliation logic
├── config/
│   ├── crd/                      # Auto-generated CRD manifests
│   ├── rbac/                     # RBAC roles and bindings
│   └── manager/                  # Controller manager deployment
├── Dockerfile
├── Makefile
├── go.mod
├── docs/ARCHITECTURE.md
└── README.md
```

## 🛠️ Tech Stack

- **Language**: Go
- **Framework**: kubebuilder
- **Libraries**: controller-runtime, client-go
- **CRDs**: Custom Resource Definitions

## Decision Log

| Decision | Rationale |
|----------|-----------|
| kubebuilder over Operator SDK | kubebuilder is the upstream CNCF scaffolding tool; Operator SDK wraps it |
| Go over Python (kopf) | Go is the native Kubernetes language; controller-runtime is first-class |
| Secret rotation over other operators | Directly maps to compliance requirements (SOC2 credential rotation) |
| CronJob-style scheduling | Familiar cron syntax, enterprise-ready scheduling patterns |

## 👨‍💻 Author

*Built to demonstrate Kubernetes API-level engineering: CRDs, controllers, and the reconciliation pattern.*
