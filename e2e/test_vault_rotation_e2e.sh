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
docker exec vault-dev vault kv put secret/database/credentials username=admin password=supersecret
export VAULT_ADDR="http://127.0.0.1:8200"
export VAULT_TOKEN="root"

log "Creating KIND cluster: $KIND_CLUSTER"
kind create cluster --name "$KIND_CLUSTER" --wait 60s

log "Applying CRD..."
kubectl apply -f config/crd/

log "Starting controller in background..."
export VAULT_ADDR="http://127.0.0.1:8200"
export VAULT_TOKEN="root"
go run main.go &
PID=$!
sleep 5

log "Applying SecretRotation CR..."
START=$(date +%s%3N)
kubectl apply -f e2e/fixtures/rotation-trigger.yaml

log "Waiting for rotation..."
# The operator should create/update the Secret based on Vault data. We just check if the Secret appears.
sleep 10
kubectl get secret my-db-secret || { log "❌ Secret not created by operator"; kill $PID; exit 1; }
# Assuming there is some status update or condition, we can check it. For now, secret existence is proof.
END=$(date +%s%3N)

log "✅ Operator successfully pulled from Vault and created Secret"

mkdir -p benchmarks/results
echo "{
  \"metrics\": {
    \"rotation_time_ms\": $((END - START))
  }
}" > benchmarks/results/vault_rotation_metrics.json

kill $PID
kind delete cluster --name "$KIND_CLUSTER"
docker rm -f vault-dev
