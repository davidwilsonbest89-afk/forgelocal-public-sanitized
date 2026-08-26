# Gosec source-only classification

Total findings: **177**. Input: `/home/ubuntu/forgelocal_self_validation_v6/operational_validation_v1/repo/evidence/SECURITY_GOSEC_REVIEW/gosec_source_only.json`.

## Method

The scan covered only `./cmd/... ./internal/...`; generated/portable bundle source under `artifacts/` was excluded. No rule, analyzer, skip, allowlist, or nolint was used.

## Summary

| Dimension | Count |
|---|---:|
| severity `HIGH` | 49 |
| severity `LOW` | 53 |
| severity `MEDIUM` | 75 |
| confidence `HIGH` | 144 |
| confidence `LOW` | 2 |
| confidence `MEDIUM` | 31 |
| zone `executed-core-path` | 134 |
| zone `other-core-source` | 43 |

## Findings by rule

| Rule | Count | Severity | Confidence | Representative paths |
|---|---:|---|---|---|
| `G104` | 53 | LOW | HIGH | `cmd/server/cli_runtime.go; cmd/server/main.go; internal/api/dashboard.go` |
| `G301` | 25 | MEDIUM | HIGH | `cmd/server/cli.go; cmd/server/cli_runtime.go; internal/browser/download.go` |
| `G304` | 22 | MEDIUM | HIGH | `cmd/server/cli.go; cmd/server/cli_runtime.go; internal/api/artifacts.go` |
| `G404` | 17 | HIGH | MEDIUM | `internal/fingerprint/pool.go; internal/humanize/keyboard.go; internal/humanize/math.go` |
| `G703` | 14 | HIGH | HIGH | `cmd/server/cli_runtime.go; internal/browser/download.go` |
| `G306` | 9 | MEDIUM | HIGH | `cmd/server/cli.go; cmd/server/main.go; internal/browser/download.go` |
| `G115` | 8 | HIGH | MEDIUM | `cmd/server/cli_runtime.go; internal/browser/download.go; internal/browser/launch_chromium.go` |
| `G704` | 7 | HIGH | HIGH | `cmd/server/cli.go; cmd/server/cli_runtime.go; internal/api/sessions.go` |
| `G204` | 6 | MEDIUM | HIGH | `cmd/server/main.go; internal/browser/download.go` |
| `G302` | 5 | MEDIUM | HIGH | `cmd/server/cli.go; internal/browser/download.go; internal/config/config.go` |
| `G110` | 3 | MEDIUM | MEDIUM | `cmd/server/cli_runtime.go; internal/browser/download.go` |
| `G305` | 3 | MEDIUM | HIGH | `internal/browser/download.go` |
| `G101` | 1 | HIGH | LOW | `internal/api/admin_token.go` |
| `G118` | 1 | HIGH | MEDIUM | `internal/launch/launch.go` |
| `G122` | 1 | HIGH | MEDIUM | `cmd/server/cli_runtime.go` |
| `G112` | 1 | MEDIUM | LOW | `cmd/server/main.go` |
| `G107` | 1 | MEDIUM | MEDIUM | `internal/browser/download.go` |

## Executed-core-path review queue

The following findings are in the server/API/browser/secrets/proxy paths used by the local smoke and must be reviewed first. This is a prioritization queue, not a claim of exploitability.

| File | Line | Rule | Severity | Confidence | Details |
|---|---:|---|---|---|---|
| `cmd/server/cli.go` | 572 | `G304` | MEDIUM | HIGH | Potential file inclusion via variable |
| `cmd/server/cli.go` | 620 | `G306` | MEDIUM | HIGH | Expect WriteFile permissions to be 0600 or less |
| `cmd/server/cli.go` | 625 | `G306` | MEDIUM | HIGH | Expect WriteFile permissions to be 0600 or less |
| `cmd/server/cli.go` | 763 | `G302` | MEDIUM | HIGH | Expect file permissions to be 0600 or less |
| `cmd/server/cli.go` | 778 | `G301` | MEDIUM | HIGH | Expect directory permissions to be 0750 or less |
| `cmd/server/cli.go` | 787 | `G306` | MEDIUM | HIGH | Expect WriteFile permissions to be 0600 or less |
| `cmd/server/cli.go` | 856 | `G304` | MEDIUM | HIGH | Potential file inclusion via variable |
| `cmd/server/cli.go` | 982 | `G704` | HIGH | HIGH | SSRF via taint analysis |
| `cmd/server/cli.go` | 998 | `G704` | HIGH | HIGH | SSRF via taint analysis |
| `cmd/server/cli.go` | 1042 | `G704` | HIGH | HIGH | SSRF via taint analysis |
| `cmd/server/cli_runtime.go` | 453 | `G703` | HIGH | HIGH | Path traversal via taint analysis |
| `cmd/server/cli_runtime.go` | 453 | `G301` | MEDIUM | HIGH | Expect directory permissions to be 0750 or less |
| `cmd/server/cli_runtime.go` | 577 | `G703` | HIGH | HIGH | Path traversal via taint analysis |
| `cmd/server/cli_runtime.go` | 587 | `G703` | HIGH | HIGH | Path traversal via taint analysis |
| `cmd/server/cli_runtime.go` | 587 | `G301` | MEDIUM | HIGH | Expect directory permissions to be 0750 or less |
| `cmd/server/cli_runtime.go` | 590 | `G703` | HIGH | HIGH | Path traversal via taint analysis |
| `cmd/server/cli_runtime.go` | 590 | `G304` | MEDIUM | HIGH | Potential file inclusion via variable |
| `cmd/server/cli_runtime.go` | 645 | `G122` | HIGH | MEDIUM | Filesystem operation in filepath.Walk/WalkDir callback uses race-prone path; consider root-scoped APIs (e.g. os.Root) to prevent symlink TOCTOU traversal |
| `cmd/server/cli_runtime.go` | 645 | `G304` | MEDIUM | HIGH | Potential file inclusion via variable |
| `cmd/server/cli_runtime.go` | 656 | `G703` | HIGH | HIGH | Path traversal via taint analysis |
| `cmd/server/cli_runtime.go` | 656 | `G304` | MEDIUM | HIGH | Potential file inclusion via variable |
| `cmd/server/cli_runtime.go` | 687 | `G115` | HIGH | MEDIUM | integer overflow conversion int64 -> uint32 |
| `cmd/server/cli_runtime.go` | 687 | `G703` | HIGH | HIGH | Path traversal via taint analysis |
| `cmd/server/cli_runtime.go` | 691 | `G703` | HIGH | HIGH | Path traversal via taint analysis |
| `cmd/server/cli_runtime.go` | 691 | `G301` | MEDIUM | HIGH | Expect directory permissions to be 0750 or less |
| `cmd/server/cli_runtime.go` | 694 | `G115` | HIGH | MEDIUM | integer overflow conversion int64 -> uint32 |
| `cmd/server/cli_runtime.go` | 694 | `G703` | HIGH | HIGH | Path traversal via taint analysis |
| `cmd/server/cli_runtime.go` | 694 | `G304` | MEDIUM | HIGH | Potential file inclusion via variable |
| `cmd/server/cli_runtime.go` | 698 | `G110` | MEDIUM | MEDIUM | Potential DoS vulnerability via decompression bomb |
| `cmd/server/cli_runtime.go` | 699 | `G104` | LOW | HIGH | Errors unhandled |
| `cmd/server/cli_runtime.go` | 702 | `G104` | LOW | HIGH | Errors unhandled |
| `cmd/server/cli_runtime.go` | 707 | `G703` | HIGH | HIGH | Path traversal via taint analysis |
| `cmd/server/cli_runtime.go` | 707 | `G301` | MEDIUM | HIGH | Expect directory permissions to be 0750 or less |
| `cmd/server/cli_runtime.go` | 710 | `G703` | HIGH | HIGH | Path traversal via taint analysis |
| `cmd/server/cli_runtime.go` | 740 | `G704` | HIGH | HIGH | SSRF via taint analysis |
| `cmd/server/cli_runtime.go` | 745 | `G704` | HIGH | HIGH | SSRF via taint analysis |
| `cmd/server/cli_runtime.go` | 753 | `G703` | HIGH | HIGH | Path traversal via taint analysis |
| `cmd/server/cli_runtime.go` | 753 | `G301` | MEDIUM | HIGH | Expect directory permissions to be 0750 or less |
| `cmd/server/cli_runtime.go` | 756 | `G703` | HIGH | HIGH | Path traversal via taint analysis |
| `cmd/server/cli_runtime.go` | 756 | `G304` | MEDIUM | HIGH | Potential file inclusion via variable |
| `cmd/server/cli_runtime.go` | 784 | `G703` | HIGH | HIGH | Path traversal via taint analysis |
| `cmd/server/cli_runtime.go` | 784 | `G304` | MEDIUM | HIGH | Potential file inclusion via variable |
| `cmd/server/cli_runtime.go` | 794 | `G704` | HIGH | HIGH | SSRF via taint analysis |
| `cmd/server/main.go` | 325 | `G306` | MEDIUM | HIGH | Expect WriteFile permissions to be 0600 or less |
| `cmd/server/main.go` | 414 | `G112` | MEDIUM | LOW | Potential Slowloris Attack because ReadHeaderTimeout is not configured in the http.Server |
| `cmd/server/main.go` | 429 | `G104` | LOW | HIGH | Errors unhandled |
| `cmd/server/main.go` | 555 | `G204` | MEDIUM | HIGH | Subprocess launched with variable |
| `cmd/server/main.go` | 557 | `G204` | MEDIUM | HIGH | Subprocess launched with variable |
| `cmd/server/main.go` | 559 | `G204` | MEDIUM | HIGH | Subprocess launched with variable |
| `internal/api/admin_token.go` | 22 | `G101` | HIGH | LOW | Potential hardcoded credentials |
| `internal/api/artifacts.go` | 31 | `G304` | MEDIUM | HIGH | Potential file inclusion via variable |
| `internal/api/dashboard.go` | 13 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/api/router.go` | 707 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/api/router.go` | 831 | `G304` | MEDIUM | HIGH | Potential file inclusion via variable |
| `internal/api/sessions.go` | 477 | `G704` | HIGH | HIGH | SSRF via taint analysis |
| `internal/browser/download.go` | 137 | `G304` | MEDIUM | HIGH | Potential file inclusion via variable |
| `internal/browser/download.go` | 145 | `G306` | MEDIUM | HIGH | Expect WriteFile permissions to be 0600 or less |
| `internal/browser/download.go` | 145 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/download.go` | 308 | `G301` | MEDIUM | HIGH | Expect directory permissions to be 0750 or less |
| `internal/browser/download.go` | 324 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/download.go` | 334 | `G204` | MEDIUM | HIGH | Subprocess launched with variable |
| `internal/browser/download.go` | 334 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/download.go` | 341 | `G302` | MEDIUM | HIGH | Expect file permissions to be 0600 or less |
| `internal/browser/download.go` | 341 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/download.go` | 357 | `G301` | MEDIUM | HIGH | Expect directory permissions to be 0750 or less |
| `internal/browser/download.go` | 357 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/download.go` | 371 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/download.go` | 374 | `G204` | MEDIUM | HIGH | Subprocess launched with variable |
| `internal/browser/download.go` | 374 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/download.go` | 381 | `G302` | MEDIUM | HIGH | Expect file permissions to be 0600 or less |
| `internal/browser/download.go` | 381 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/download.go` | 436 | `G301` | MEDIUM | HIGH | Expect directory permissions to be 0750 or less |
| `internal/browser/download.go` | 436 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/download.go` | 450 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/download.go` | 453 | `G204` | MEDIUM | HIGH | Subprocess launched with variable |
| `internal/browser/download.go` | 453 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/download.go` | 460 | `G302` | MEDIUM | HIGH | Expect file permissions to be 0600 or less |
| `internal/browser/download.go` | 460 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/download.go` | 510 | `G107` | MEDIUM | MEDIUM | Potential HTTP request made with variable url |
| `internal/browser/download.go` | 519 | `G304` | MEDIUM | HIGH | Potential file inclusion via variable |
| `internal/browser/download.go` | 568 | `G305` | MEDIUM | HIGH | File traversal when extracting zip/tar archive |
| `internal/browser/download.go` | 573 | `G301` | MEDIUM | HIGH | Expect directory permissions to be 0750 or less |
| `internal/browser/download.go` | 573 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/download.go` | 576 | `G301` | MEDIUM | HIGH | Expect directory permissions to be 0750 or less |
| `internal/browser/download.go` | 576 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/download.go` | 578 | `G304` | MEDIUM | HIGH | Potential file inclusion via variable |
| `internal/browser/download.go` | 584 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/download.go` | 587 | `G110` | MEDIUM | MEDIUM | Potential DoS vulnerability via decompression bomb |
| `internal/browser/download.go` | 588 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/download.go` | 589 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/download.go` | 598 | `G304` | MEDIUM | HIGH | Potential file inclusion via variable |
| `internal/browser/download.go` | 619 | `G305` | MEDIUM | HIGH | File traversal when extracting zip/tar archive |
| `internal/browser/download.go` | 625 | `G301` | MEDIUM | HIGH | Expect directory permissions to be 0750 or less |
| `internal/browser/download.go` | 625 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/download.go` | 627 | `G301` | MEDIUM | HIGH | Expect directory permissions to be 0750 or less |
| `internal/browser/download.go` | 627 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/download.go` | 628 | `G115` | HIGH | MEDIUM | integer overflow conversion int64 -> uint32 |
| `internal/browser/download.go` | 629 | `G304` | MEDIUM | HIGH | Potential file inclusion via variable |
| `internal/browser/download.go` | 633 | `G110` | MEDIUM | MEDIUM | Potential DoS vulnerability via decompression bomb |
| `internal/browser/download.go` | 634 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/download.go` | 639 | `G301` | MEDIUM | HIGH | Expect directory permissions to be 0750 or less |
| `internal/browser/download.go` | 639 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/download.go` | 640 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/download.go` | 643 | `G305` | MEDIUM | HIGH | File traversal when extracting zip/tar archive |
| `internal/browser/download.go` | 644 | `G304` | MEDIUM | HIGH | Potential file inclusion via variable |
| `internal/browser/download.go` | 645 | `G703` | HIGH | HIGH | Path traversal via taint analysis |
| `internal/browser/download.go` | 645 | `G306` | MEDIUM | HIGH | Expect WriteFile permissions to be 0600 or less |
| `internal/browser/download.go` | 645 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/launch_chromium.go` | 64 | `G301` | MEDIUM | HIGH | Expect directory permissions to be 0750 or less |
| `internal/browser/launch_chromium.go` | 191 | `G301` | MEDIUM | HIGH | Expect directory permissions to be 0750 or less |
| `internal/browser/launch_chromium.go` | 196 | `G301` | MEDIUM | HIGH | Expect directory permissions to be 0750 or less |
| `internal/browser/launch_chromium.go` | 201 | `G304` | MEDIUM | HIGH | Potential file inclusion via variable |
| `internal/browser/launch_chromium.go` | 217 | `G306` | MEDIUM | HIGH | Expect WriteFile permissions to be 0600 or less |
| `internal/browser/launch_chromium.go` | 312 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/launch_chromium.go` | 333 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/launch_chromium.go` | 949 | `G301` | MEDIUM | HIGH | Expect directory permissions to be 0750 or less |
| `internal/browser/launch_chromium.go` | 957 | `G306` | MEDIUM | HIGH | Expect WriteFile permissions to be 0600 or less |
| `internal/browser/launch_chromium.go` | 1015 | `G115` | HIGH | MEDIUM | integer overflow conversion int64 -> uint64 |
| `internal/browser/launch_chromium.go` | 1913 | `G301` | MEDIUM | HIGH | Expect directory permissions to be 0750 or less |
| `internal/browser/launch_firefox.go` | 66 | `G301` | MEDIUM | HIGH | Expect directory permissions to be 0750 or less |
| `internal/browser/launch_firefox.go` | 80 | `G301` | MEDIUM | HIGH | Expect directory permissions to be 0750 or less |
| `internal/browser/launch_firefox.go` | 153 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/manager.go` | 373 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/manager.go` | 376 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/manager.go` | 391 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/manager.go` | 399 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/socks5relay.go` | 49 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/socks5relay.go` | 87 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/socks5relay.go` | 95 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/socks5relay.go` | 122 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/socks5relay.go` | 129 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/socks5relay.go` | 135 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/socks5relay.go` | 142 | `G104` | LOW | HIGH | Errors unhandled |
| `internal/browser/socks5relay.go` | 144 | `G104` | LOW | HIGH | Errors unhandled |

## Decision

No finding is closed by this inventory. The next code changes must be limited to findings confirmed exploitable or materially relevant to the executed path, with a regression test and a fresh scan. Historical or non-exploitable findings may be classified with rationale, but must remain visible and must not be suppressed.
