# GOSEC-R3 Lot A matrix

The matrix compares the 34 baseline G703/G304/G305/G122 entries with the R3-A post-change scan. No suppression was used.

| Status | Count |
|---|---:|
| `CORRECTED_AND_VERIFIED` | 2 |
| `MITIGATED_CONTROL_SCANNER_OPEN` | 1 |
| `NEEDS_MANUAL_REVIEW` | 31 |

Post-scan counts: G703=12, G304=19, G305=1, G122=0, G110=0.

| ID | Rule | File | Line | Status | Rationale |
|---|---|---|---:|---|---|
| R3A-034 | `G703` | `cmd/server/cli_runtime.go` | 914 | `NEEDS_MANUAL_REVIEW` | CLI path is an explicit local operator argument; no new global allowlist or suppression added. |
| R3A-035 | `G703` | `cmd/server/cli_runtime.go` | 883 | `NEEDS_MANUAL_REVIEW` | CLI path is an explicit local operator argument; no new global allowlist or suppression added. |
| R3A-036 | `G703` | `cmd/server/cli_runtime.go` | 880 | `NEEDS_MANUAL_REVIEW` | CLI path is an explicit local operator argument; no new global allowlist or suppression added. |
| R3A-037 | `G703` | `cmd/server/cli_runtime.go` | 768 | `NEEDS_MANUAL_REVIEW` | CLI path is an explicit local operator argument; no new global allowlist or suppression added. |
| R3A-038 | `G703` | `cmd/server/cli_runtime.go` | 744 | `NEEDS_MANUAL_REVIEW` | CLI path is an explicit local operator argument; no new global allowlist or suppression added. |
| R3A-039 | `G703` | `cmd/server/cli_runtime.go` | 741 | `NEEDS_MANUAL_REVIEW` | CLI path is an explicit local operator argument; no new global allowlist or suppression added. |
| R3A-040 | `G703` | `cmd/server/cli_runtime.go` | 731 | `NEEDS_MANUAL_REVIEW` | CLI path is an explicit local operator argument; no new global allowlist or suppression added. |
| R3A-041 | `G703` | `cmd/server/cli_runtime.go` | 692 | `NEEDS_MANUAL_REVIEW` | CLI path is an explicit local operator argument; no new global allowlist or suppression added. |
| R3A-042 | `G703` | `cmd/server/cli_runtime.go` | 594 | `NEEDS_MANUAL_REVIEW` | CLI path is an explicit local operator argument; no new global allowlist or suppression added. |
| R3A-043 | `G703` | `cmd/server/cli_runtime.go` | 591 | `NEEDS_MANUAL_REVIEW` | CLI path is an explicit local operator argument; no new global allowlist or suppression added. |
| R3A-044 | `G703` | `cmd/server/cli_runtime.go` | 581 | `NEEDS_MANUAL_REVIEW` | CLI path is an explicit local operator argument; no new global allowlist or suppression added. |
| R3A-045 | `G703` | `cmd/server/cli_runtime.go` | 457 | `NEEDS_MANUAL_REVIEW` | CLI path is an explicit local operator argument; no new global allowlist or suppression added. |
| R3A-047 | `G122` | `cmd/server/cli_runtime.go` | 658 | `CORRECTED_AND_VERIFIED` | addPathToTar now opens regular entries through os.Root; G122 after scan is zero. |
| R3A-053 | `G304` | `internal/workflow/engine.go` | 53 | `NEEDS_MANUAL_REVIEW` | path is supplied by a runtime/config/CLI integration and requires component-level policy review; finding remains visible. |
| R3A-054 | `G304` | `internal/mcp/advanced_tools.go` | 481 | `NEEDS_MANUAL_REVIEW` | path is supplied by a runtime/config/CLI integration and requires component-level policy review; finding remains visible. |
| R3A-055 | `G304` | `internal/fingerprint/pool.go` | 25 | `NEEDS_MANUAL_REVIEW` | path is supplied by a runtime/config/CLI integration and requires component-level policy review; finding remains visible. |
| R3A-056 | `G304` | `internal/config/config.go` | 172 | `NEEDS_MANUAL_REVIEW` | path is supplied by a runtime/config/CLI integration and requires component-level policy review; finding remains visible. |
| R3A-057 | `G304` | `internal/config/config.go` | 91 | `NEEDS_MANUAL_REVIEW` | path is supplied by a runtime/config/CLI integration and requires component-level policy review; finding remains visible. |
| R3A-058 | `G304` | `internal/browser/launch_chromium.go` | 201 | `NEEDS_MANUAL_REVIEW` | path is supplied by a runtime/config/CLI integration and requires component-level policy review; finding remains visible. |
| R3A-059 | `G304` | `internal/browser/download.go` | 532 | `NEEDS_MANUAL_REVIEW` | path is supplied by a runtime/config/CLI integration and requires component-level policy review; finding remains visible. |
| R3A-060 | `G304` | `internal/browser/download.go` | 136 | `NEEDS_MANUAL_REVIEW` | path is supplied by a runtime/config/CLI integration and requires component-level policy review; finding remains visible. |
| R3A-061 | `G304` | `internal/browser/archive_security.go` | 209 | `NEEDS_MANUAL_REVIEW` | path is supplied by a runtime/config/CLI integration and requires component-level policy review; finding remains visible. |
| R3A-062 | `G304` | `internal/browser/archive_security.go` | 168 | `NEEDS_MANUAL_REVIEW` | path is supplied by a runtime/config/CLI integration and requires component-level policy review; finding remains visible. |
| R3A-063 | `G304` | `internal/browser/archive_security.go` | 141 | `NEEDS_MANUAL_REVIEW` | path is supplied by a runtime/config/CLI integration and requires component-level policy review; finding remains visible. |
| R3A-064 | `G304` | `internal/api/router.go` | 831 | `NEEDS_MANUAL_REVIEW` | path is supplied by a runtime/config/CLI integration and requires component-level policy review; finding remains visible. |
| R3A-065 | `G304` | `cmd/server/cli_runtime.go` | 914 | `NEEDS_MANUAL_REVIEW` | path is supplied by a runtime/config/CLI integration and requires component-level policy review; finding remains visible. |
| R3A-066 | `G304` | `cmd/server/cli_runtime.go` | 883 | `NEEDS_MANUAL_REVIEW` | path is supplied by a runtime/config/CLI integration and requires component-level policy review; finding remains visible. |
| R3A-067 | `G304` | `cmd/server/cli_runtime.go` | 744 | `NEEDS_MANUAL_REVIEW` | path is supplied by a runtime/config/CLI integration and requires component-level policy review; finding remains visible. |
| R3A-068 | `G304` | `cmd/server/cli_runtime.go` | 692 | `NEEDS_MANUAL_REVIEW` | path is supplied by a runtime/config/CLI integration and requires component-level policy review; finding remains visible. |
| R3A-069 | `G304` | `cmd/server/cli_runtime.go` | 658 | `CORRECTED_AND_VERIFIED` | the baseline os.Open(path) callback was replaced by root-scoped Root.Open; G304 count dropped 20 to 19. |
| R3A-070 | `G304` | `cmd/server/cli_runtime.go` | 594 | `NEEDS_MANUAL_REVIEW` | path is supplied by a runtime/config/CLI integration and requires component-level policy review; finding remains visible. |
| R3A-071 | `G304` | `cmd/server/cli.go` | 856 | `NEEDS_MANUAL_REVIEW` | path is supplied by a runtime/config/CLI integration and requires component-level policy review; finding remains visible. |
| R3A-072 | `G304` | `cmd/server/cli.go` | 572 | `NEEDS_MANUAL_REVIEW` | path is supplied by a runtime/config/CLI integration and requires component-level policy review; finding remains visible. |
| R3A-074 | `G305` | `cmd/server/cli_runtime.go` | 764 | `MITIGATED_CONTROL_SCANNER_OPEN` | restore extractor retains path/type/size confinement and negative tests; scanner still reports the archive extraction pattern. |
