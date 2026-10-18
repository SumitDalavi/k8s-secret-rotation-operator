# Changelog

## [v0.1.0] — 2026-10-18
### Added
- Automated GitHub Actions release pipeline triggered on `v*.*.*` git tags (`.github/workflows/release.yml`).
- Multi-stage distroless container image packaging (`ghcr.io/sumitdalavi/k8s-secret-rotation-operator`).
- Container vulnerability scanning step via Aqua Security Trivy (`aquasecurity/trivy-action@v0.36.0`).
- SPDX-format Software Bill of Materials (SBOM) generation step (`anchore/sbom-action@v0`).
- Automated GitHub release publishing attaching `sbom.spdx.json` via `softprops/action-gh-release@v2`.

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
