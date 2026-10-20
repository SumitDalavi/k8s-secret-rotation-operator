# Kubernetes Secret Rotation Operator 🔄🔐

[![CI Pipeline](https://github.com/SumitDalavi/k8s-secret-rotation-operator/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/SumitDalavi/k8s-secret-rotation-operator/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/SumitDalavi/k8s-secret-rotation-operator?color=0969da&label=release&logo=github)](https://github.com/SumitDalavi/k8s-secret-rotation-operator/releases)
[![Release Workflow](https://github.com/SumitDalavi/k8s-secret-rotation-operator/actions/workflows/release.yml/badge.svg)](https://github.com/SumitDalavi/k8s-secret-rotation-operator/actions/workflows/release.yml)
[![GHCR Container](https://img.shields.io/badge/ghcr.io-v0.2.0-blue?logo=docker&logoColor=white)](https://github.com/SumitDalavi/k8s-secret-rotation-operator/pkgs/container/k8s-secret-rotation-operator)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)](https://golang.org)
[![SBOM](https://img.shields.io/badge/SBOM-SPDX%202.3-blueviolet?logo=spdx)](https://github.com/SumitDalavi/k8s-secret-rotation-operator/releases/tag/v0.2.0)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)

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

## 🚀 Progressive Credential Loading & Execution Tiers

This project implements **Progressive Credential Loading (PCL)**, allowing it to run immediately out-of-the-box without external dependencies while seamlessly upgrading as credentials are provided.

### Tier 1: Zero-Config Mock Mode (Default)
Ideal for CI, testing, and rapid local evaluation.
- **Execution:** Run with default configuration (`cp .env.example .env`). Zero API keys, databases, or cloud credentials required.
- **Fallback Behavior:** In-memory caches, deterministic offline heuristics, mock vector embeddings, and local mock storage are automatically used.
- **Health Verification:** Endpoints report `{"mode": "mock"}` for unconfigured dependencies.

### Tier 2: Local Development with Live Services
Connect specific external services while keeping others mocked.
- **Execution:** Provide credentials in `.env` for the services you want active (e.g., `VAULT_ADDR, AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY`).
- **Progressive Upgrade:** The runtime inspects environment variables on boot and dynamically promotes only configured modules to live mode.
- **Safe Fallback:** Any missing credentials automatically remain on mock fallbacks with no crashes or missing service errors.

### Tier 3: Production Hardened
Enterprise deployment with production-grade backends, security scans, and container distributions.
- **OCI Container Image:** Pre-built, multi-arch image on GitHub Container Registry:
  ```bash
  docker pull ghcr.io/sumitdalavi/k8s-secret-rotation-operator:v0.2.0
  ```
- **Production Orchestration:** Mount secrets via Kubernetes Secrets or cloud secret managers with atomic config reload support.
- **Release Verification:** Checksums and SPDX SBOM available on [GitHub Releases](https://github.com/SumitDalavi/k8s-secret-rotation-operator/releases).

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

## 🔒 Release & Security Verification

Every tagged release (`v*.*.*`) triggers an automated DevSecOps release workflow:
1. **Automated Testing:** Runs `go test -v ./...` unit tests.
2. **Container Image Build:** Builds production container image (`ghcr.io/sumitdalavi/k8s-secret-rotation-operator:<tag>`).
3. **Vulnerability Scanning:** Scans the image using [Aqua Security Trivy](https://github.com/aquasecurity/trivy-action) for `HIGH` and `CRITICAL` CVEs.
4. **SBOM Generation:** Generates a Software Bill of Materials (SBOM) in SPDX JSON format (`sbom.spdx.json`) using [Anchore Syft](https://github.com/anchore/sbom-action).
5. **GitHub Release Publishing:** Packages release notes and attaches `sbom.spdx.json` as a verifiable supply-chain asset.

### Verifying Releases via GitHub CLI
```bash
# View release details
gh release view v0.2.0 --repo SumitDalavi/k8s-secret-rotation-operator

# Download and verify SBOM
gh release download v0.2.0 --repo SumitDalavi/k8s-secret-rotation-operator --pattern "sbom.spdx.json"
cat sbom.spdx.json | jq .
```

## 📚 Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [Runbook](docs/runbook.md)
- [Decisions](docs/decisions.md)
- [Changelog](docs/changelog.md)

## 👨‍💻 Author

**Sumit Dalavi** — Senior DevSecOps / Platform Engineer
[GitHub](https://github.com/SumitDalavi) | [LinkedIn](https://in.linkedin.com/in/sumit-dalavi-762838129)

---
*Built with a focus on robust patterns, not toy demos.*
