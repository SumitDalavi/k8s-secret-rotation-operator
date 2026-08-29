# Decisions

## ADR-001: Kubernetes Native Secrets vs External Secrets Operator
**Date:** 2026-08-29  
**Status:** Accepted

**Context:**  
We need to manage and rotate secrets in Kubernetes. External Secrets Operator (ESO) is an industry standard for syncing secrets from external managers (Vault, AWS Secrets Manager) into K8s.

**Decision:**  
We chose to build a custom `SecretRotation` operator. This is to demonstrate deep understanding of the kubebuilder framework and operator patterns. It generates secrets natively or fetches from Vault and updates the K8s Secret directly, effectively acting as a lightweight, specialized alternative to ESO for auto-generation use cases.

**Consequences:**  
- ✅ Demonstrates API-level engineering skills.
- ⚠️ In a real enterprise, ESO might be preferred for managing external sync, while this operator is better suited for *auto-generating* new K8s-native secrets (like rotation of auto-generated DB passwords).
