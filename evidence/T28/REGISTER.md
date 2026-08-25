# REGISTER — T28 Extensions locales contrôlées

**Decision:** `T28_IMPLEMENTED_VERIFIABLE_LOCAL_PENDING_INDEPENDENT_REVIEW`

**Branch:** `feature/t28-local-extensions-controlled`

**HEAD at implementation publication:** `4f0f6201e1d8f8da44d82c4245bd9b7dfee44578`

**Baseline:** `999374d99b7996504ba91e421850a2fe84afb78d` / tag `t00-t42-v6-local-qualified-2026-08-25`

## Evidence index

| Evidence | Purpose | Status |
|---|---|---|
| `BASELINE_DISCOVERY_RAW.log` | Baseline, remotes, lineage, fsck, tools, disk and pre-code search | Captured |
| `GO_GLOBAL_VALIDATION_POST_IMPLEMENTATION_RAW.log` | Initial global test/vet/build record with CWD-corrected rerun | Captured; global runtime finding preserved |
| `FINAL_VALIDATION_RAW.log` | Final HEAD validation commands and exit codes | Captured |
| `SCANS_RAW.log` | Raw Gitleaks/Gosec/Govulncheck/Syft/OSV diagnostics, including corrected commands | Captured |
| `GITLEAKS_FINAL_EXTRACTION.json` | Fresh worktree extraction scan | Empty report / PASS |
| `GITLEAKS_FINAL_DIFF.json` | Git log-option scan result | Empty report; tool reports zero commits, so not used alone as range proof |
| `GITLEAKS_DIFF_POST.json` | Corrected extraction/diff report | Empty report |
| `GOSEC_FINAL_T28_EXTENSIONS.json` | T28 repository security scan | `found=0` |
| `GOSEC_HEAD_CORRECTED.json` | Whole repository Gosec result from correct module CWD | 182 historical findings; no T28 file finding |
| `SBOM_FINAL.spdx.json` | SPDX inventory from current worktree | Generated |

## Checks

The T28-targeted race tests pass. The full repository test suite keeps one pre-existing runtime configuration failure in `internal/runtime/runtime_test.go` (`BrowseForge Chromium binary/enabled`); it is not attributed to T28 and no runtime was launched. `go vet ./...`, `go build ./...`, and `govulncheck ./...` pass. `osv-scanner` is absent and therefore remains an environment limitation with exit 127, not a simulated pass.

## Closed scope

No extension package is downloaded, loaded or executed. No browser, Camoufox, Chromium, proxy, real cookie, SystemVault native integration, migration, production runtime or release operation is included. The API uses the pre-existing bearer, loopback and origin guards.

## Review gate

A fresh independent review of the T28 branch/package is required before changing the decision to any approved state.
