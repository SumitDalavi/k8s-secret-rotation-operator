# Architecture — k8s-secret-rotation-operator
> Last updated: 2026-08-29 | Maturity: Full Prototype
> _Kubernetes operator for automated secret rotation._

## System Diagram
The following Mermaid.js sequence diagram maps the core workflow and interactions:

```mermaid
flowchart TD
    Dev(["Developer"])
    API["kube-apiserver"]
    Operator["k8s-secret-rotation-operator"]
    Secret["K8s Secret"]
    Vault["HashiCorp Vault (Optional Backend)"]

    Dev -->|"kubectl apply SecretRotation"| API
    API -->|"Watch Event"| Operator
    Operator -->|"Check Schedule"| Operator
    Operator -->|"Generate/Fetch new credential"| Vault
    Vault -->|"Return new credential"| Operator
    Operator -->|"Reconcile (Update)"| Secret
```

## Component Table

| Component | File | Responsibility | Tech |
|---|---|---|---|
| SecretRotation CRD | `api/v1alpha1/secretrotation_types.go` | Defines the API schema for rotation policies | Go |
| Controller | `controllers/secretrotation_controller.go` | Main reconciliation loop enforcing schedule | Go |
| Generator | `internal/generator/generator.go` | Logic for auto-generating strong keys or fetching from Vault | Go |

## Port Assignments

| Service | Port | Notes |
|---|---|---|
| Metrics | `8080` | Prometheus metrics endpoint exposed by controller-runtime |
| Healthz | `8081` | Liveness and readiness probes |

## Dependency Honesty Table

| Dependency | Status | Notes |
|---|---|---|
| Kubernetes API Server | **Real** | Controller directly talks to the K8s API to manage resources. |
| HashiCorp Vault | **Optional** | E2E test scripts use Vault Dev mode to verify external backend rotation. |
| kind (Local Cluster) | **Optional** | Used for E2E tests and local development. |


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

## Progressive Credential Loading & Production Architecture (v0.2.0)

This service implements the **Progressive Credential Loading (PCL)** pattern, decoupling development speed from external infrastructure provisioning:

```mermaid
graph LR
    Env[Environment Credentials] -->|Detect Keys & URLs| Detector[Credential Detector]
    Detector -->|No Credentials / Missing| Mock[Mock Mode Factory]
    Detector -->|Valid Credentials Present| Live[Live Provider Factory]
    Mock --> Runtime[Application Core & /health]
    Live --> Runtime
```

### Core Architecture Components
1. **Mode Detection Singleton:** Inspects environment configuration at boot time without hard crashes on missing secrets.
2. **Dual-Mode Factory Wrappers:**
   - **Storage / Cache:** In-process maps or SQLite `:memory:` transparently promote to managed Redis or PostgreSQL when connection strings are supplied.
   - **AI Providers:** Deterministic heuristic generators and local embeddings transparently promote to live OpenAI, Anthropic, or Gemini API clients.
3. **Hot-Reload & Dynamic TLS:** File watchers monitor mounted volume paths (Kubernetes ConfigMaps and Secrets) to update certificates and limit configs atomically without process restarts.
4. **Health & Observability:** The `/health` endpoint exposes real-time runtime mode metadata (`mock` vs `live`) across all backing adapters.

## Declarative GitOps & Multi-Cluster Deployment Architecture (v0.2.1)

This service participates in the portfolio-wide **ArgoCD Hub-and-Spoke GitOps continuous delivery catalog**:

```mermaid
graph TD
    Hub[GitOps Hub: multi-cluster-gitops-promotion] -->|AppSet Matrix: dev| DevCluster[Dev Cluster: kubernetes.default]
    Hub -->|AppSet Matrix: staging| StagingCluster[Staging Cluster: staging-k8s:6443]
    Hub -->|AppSet Matrix: prod| ProdCluster[Prod Cluster: prod-k8s:6443]
    
    subgraph WorkloadRepo["Workload Repo: k8s-secret-rotation-operator"]
        Base[deploy/k8s/base] --> DevOverlay[deploy/k8s/overlays/dev]
        Base --> StagingOverlay[deploy/k8s/overlays/staging]
        Base --> ProdOverlay[deploy/k8s/overlays/prod]
    end
    
    DevOverlay -.->|Continuous Reconciliation| DevCluster
    StagingOverlay -.->|Continuous Reconciliation| StagingCluster
    ProdOverlay -.->|Gated PR Promotion| ProdCluster
```

### GitOps Topology & Separation of Concerns
1. **Base Manifests (`deploy/k8s/base`):** Declares non-root PodSecurity standards (`runAsUser: 10001`, `drop: [ALL]`, `readOnlyRootFilesystem: true`), readiness/liveness HTTP probes, and CPU/memory requests and limits without hardcoding namespace.
2. **Environment Overlays (`deploy/k8s/overlays/`):**
   - `dev/`: 1 replica, `300m/256Mi` limit, automated drift self-healing.
   - `staging/`: 2 replicas, `500m/512Mi` limit, automated drift self-healing.
   - `prod/`: 3 replicas, `1000m/1Gi` limit, gated pull-request review.
3. **ArgoCD Declarations (`deploy/argocd/`):** Houses standalone `Application` CRDs (`application.yaml`, `application-dev.yaml`, `application-prod.yaml`) and values overrides for OCI Helm and Kustomize reconciliation.
