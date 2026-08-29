# Kubernetes Secret Rotation Operator 🔄🔐

[![CI Pipeline](https://github.com/SumitDalavi/k8s-secret-rotation-operator/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/SumitDalavi/k8s-secret-rotation-operator/actions/workflows/ci.yml)

> **Maturity:** Full Prototype
> _A custom Kubernetes operator built with kubebuilder that automatically rotates secrets on a configurable schedule — moving past Helm/Rancher-level K8s into controller/API-level engineering._

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

## Mock Boundaries (Honest Scope)

| What | Status | Details |
|---|---|---|
| Kubernetes API | **Real** | Directly reconciles native K8s Secrets against API Server. |
| HashiCorp Vault | **Optional** | Tested against local Vault Dev server for external backend rotation. |
| kind Cluster | **Optional** | Tested on `kind` cluster locally, can deploy to any K8s cluster. |

## 📚 Documentation

- [Architecture](docs/ARCHITECTURE.md) — System diagram and component details
- [Runbook](docs/runbook.md) — Setup, commands, and expected outputs
- [Decisions](docs/decisions.md) — ADRs for operator pattern choices
- [Changelog](docs/changelog.md) — Change history

## Decision Log

| Decision | Rationale |
|----------|-----------|
| kubebuilder over Operator SDK | kubebuilder is the upstream CNCF scaffolding tool; Operator SDK wraps it |
| Go over Python (kopf) | Go is the native Kubernetes language; controller-runtime is first-class |
| Secret rotation over other operators | Directly maps to compliance requirements (SOC2 credential rotation) |
| CronJob-style scheduling | Familiar cron syntax, enterprise-ready scheduling patterns |


## 📋 Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| [Go](https://go.dev/) | >= 1.21 | Build the operator |
| [kubectl](https://kubernetes.io/docs/tasks/tools/) | >= 1.28 | Kubernetes CLI |
| [kind](https://kind.sigs.k8s.io/) or [minikube](https://minikube.sigs.k8s.io/) | Latest | Local K8s cluster |
| [Docker](https://www.docker.com/) | >= 24.x | Container runtime |

## 🚀 Step-by-Step Setup

### Option A: Local Cluster (kind)

```bash
# 1. Clone the repository
git clone https://github.com/SumitDalavi/k8s-secret-rotation-operator.git
cd k8s-secret-rotation-operator

# 2. Create a local cluster
kind create cluster --name secret-rotation

# 3. Apply the CRD
kubectl apply -f config/crd/secretrotation-crd.yaml

# 4. Apply RBAC
kubectl apply -f config/rbac/role.yaml

# 5. Build and run the controller locally
go run . &
```

### Option B: Build as container

```bash
docker build -t secret-rotation-operator:latest .
kind load docker-image secret-rotation-operator:latest --name secret-rotation
```


> **💡 Note on secret consumers:** After rotation, consumer pods must be restarted or use a sidecar (e.g., Reloader) to pick up the new secret value. In production, pair this operator with [stakater/Reloader](https://github.com/stakater/Reloader) for automatic rolling restarts.
## 🧪 Usage & Demo

### Step 1: Create a Kubernetes Secret to rotate
```bash
kubectl create secret generic db-credentials \
  --from-literal=password=initial-password-v1
```

### Step 2: Create a SecretRotation custom resource
```bash
cat <<EOF | kubectl apply -f -
apiVersion: security.example.com/v1alpha1
kind: SecretRotation
metadata:
  name: rotate-db-creds
spec:
  secretName: db-credentials
  rotationIntervalMinutes: 5
  key: password
EOF
```

### Step 3: Observe automatic rotation
```bash
# Watch the secret value change over time
kubectl get secret db-credentials -o jsonpath='{.data.password}' | base64 -d; echo
# Wait for the rotation interval, then check again
```

### Step 4: Check controller logs
```bash
# The controller logs rotation events
# Look for "Rotating secret" messages in the controller output
```

## ✅ Verification

| Check | Command | Expected |
|-------|---------|----------|
| CRD registered | `kubectl get crds \| grep secretrotation` | CRD present |
| CR created | `kubectl get secretrotations` | Resources listed |
| Secret exists | `kubectl get secret db-credentials` | Secret present |
| Rotation works | Check secret value after interval | Value has changed |

```bash
# Cleanup
kind delete cluster --name secret-rotation
```

## 👨‍💻 Author

**Sumit Dalavi** — Senior DevSecOps / Platform Engineer
[GitHub](https://github.com/SumitDalavi) | [LinkedIn](https://in.linkedin.com/in/sumit-dalavi-762838129)

---

*Built with a focus on robust patterns, not toy demos.*

## CI & Reliability Updates (August 2026)

- **CI Pipeline Remediation:** Successfully resolved all CI/CD pipeline failures and established baseline CI workflows.
- **Specific Fix:** Added and configured robust GitHub Actions workflows for automated testing, linting, and formatting.
- **Status:** 🟩 Passing
