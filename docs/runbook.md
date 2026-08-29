# Runbook — k8s-secret-rotation-operator
> Last updated: 2026-08-29

## Prerequisites
| Tool | Required Version | How to check |
|---|---|---|
| Go | >= 1.21 | `go version` |
| kubectl | >= 1.28 | `kubectl version --client` |
| kind | Latest | `kind version` |
| Vault CLI | Latest | `vault version` (Optional, for external backend tests) |

## Quick Start
```bash
# Start a cluster
kind create cluster --name secret-rotation

# Apply CRD and RBAC
kubectl apply -f config/crd/secretrotation-crd.yaml
kubectl apply -f config/rbac/role.yaml

# Run operator
go run . &
```

## Run Tests
```bash
# Unit tests
go test ./... -v

# Integration Test (Vault Backend)
bash tests/integration/test_vault_rotation.sh
```

Expected output:
```
?       github.com/SumitDalavi/k8s-secret-rotation-operator     [no test files]
ok      github.com/SumitDalavi/k8s-secret-rotation-operator/api/v1alpha1        0.011s
```

## Environment Variables
| Variable | Default | Purpose |
|---|---|---|
| METRICS_ADDR | `:8080` | Port for prometheus metrics |
| VAULT_ADDR | `http://127.0.0.1:8200` | Address for Vault backend (if used) |

## Common Failure Modes
| Symptom | Cause | Fix |
|---|---|---|
| Secret not updating | CronSchedule incorrect | Check the `rotationIntervalMinutes` in the CRD |
| Unauthorized error on Vault | Missing Vault token | Ensure operator is configured with the correct VAULT_TOKEN |
