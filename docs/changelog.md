# Changelog

## [2026-08-29] — Phase 2 Evidence
### Added
- Added `tests/integration/test_vault_rotation.sh` script to verify rotation against Vault Dev.
- Added simulated metric evidence for secret recovery `benchmarks/results/vault_rotation_metrics.json`.
- Standardized documentation (`runbook.md`, `decisions.md`, `ARCHITECTURE.md`).
- Added maturity badge and mock boundaries to `README.md`.

### Post-Release Hotfixes
- Resolved lingering CI/CD failures introduced during portfolio elevation.
- Fixed Docker build and permission errors across client/server components.
- Corrected Kubernetes controller GroupVersionKind mismatches and E2E Vault addressing.
- Repaired broken property-based test configurations.
