#!/usr/bin/env bash
set -euo pipefail

KIND_CLUSTER="vault-rotation-e2e"
log() { echo "[e2e] $*"; }

# Start Vault Dev Server in Docker
log "Starting Vault Dev Server..."
docker rm -f vault-dev 2>/dev/null || true
docker run -d --name vault-dev -p 8200:8200 -e VAULT_DEV_ROOT_TOKEN_ID=root hashicorp/vault:1.15 server -dev

# Initialize a secret in Vault
export VAULT_ADDR='http://127.0.0.1:8200'
sleep 3
docker exec -e VAULT_ADDR="http://127.0.0.1:8200" -e VAULT_TOKEN="root" vault-dev vault kv put secret/database/credentials username=admin password=supersecret
export VAULT_ADDR="http://127.0.0.1:8200"
export VAULT_TOKEN="root"

if kubectl cluster-info >/dev/null 2>&1; then
  log "Using existing Kubernetes cluster"
  CREATED_CLUSTER=false
else
  log "Creating KIND cluster: $KIND_CLUSTER"
  kind create cluster --name "$KIND_CLUSTER" --wait 60s
  CREATED_CLUSTER=true
fi

log "Applying CRD..."
kubectl apply -f config/crd/

log "Building operator binary..."
go build -o bin/manager main.go

log "Starting controller in background..."
export VAULT_ADDR="http://127.0.0.1:8200"
export VAULT_TOKEN="root"
./bin/manager > /tmp/operator.log 2>&1 &
PID=$!

log "Waiting for operator to start..."
for i in {1..20}; do
  if grep -q "k8s-secret-rotation-operator starting" /tmp/operator.log 2>/dev/null; then
    log "Operator running (PID: $PID)"
    break
  fi
  sleep 1
done

log "Applying SecretRotation CR..."
START=$(date +%s%3N)
kubectl apply -f e2e/fixtures/rotation-trigger.yaml

log "Waiting for rotation..."
FOUND=false
for i in {1..20}; do
  if kubectl get secret my-db-secret >/dev/null 2>&1; then
    FOUND=true
    log "Secret my-db-secret detected!"
    break
  fi
  sleep 2
done

if [ "$FOUND" != "true" ]; then
  log "❌ Secret not created by operator"
  log "Operator log:"
  cat /tmp/operator.log 2>/dev/null || true
  kill $PID 2>/dev/null || true
  exit 1
fi
END=$(date +%s%3N)

log "✅ Operator successfully pulled from Vault and created Secret"

mkdir -p benchmarks/results
echo "{
  \"metrics\": {
    \"rotation_time_ms\": $((END - START))
  }
}" > benchmarks/results/vault_rotation_metrics.json

kill $PID 2>/dev/null || true
if [ "$CREATED_CLUSTER" = "true" ]; then
  kind delete cluster --name "$KIND_CLUSTER" 2>/dev/null || true
fi
docker rm -f vault-dev 2>/dev/null || true
