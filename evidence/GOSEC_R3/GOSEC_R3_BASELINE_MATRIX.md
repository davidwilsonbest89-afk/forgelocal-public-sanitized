# GOSEC-R3 baseline matrix

This matrix is generated from the source-only scan at the final overnight HEAD. No finding is suppressed or treated as closed.

| Rule | Count | Initial status |
|---|---:|---|
| `G101` | 1 | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| `G104` | 36 | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| `G115` | 8 | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| `G118` | 1 | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| `G122` | 1 | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| `G301` | 17 | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| `G302` | 4 | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| `G304` | 20 | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| `G305` | 1 | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| `G306` | 1 | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| `G404` | 17 | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| `G703` | 12 | `OPEN_BASELINE_REQUIRES_TRIAGE` |

**Selected findings:** 119. **All baseline findings:** 132.

| ID | Rule | Severity | Confidence | File | Line | Details | Status |
|---|---|---|---|---|---:|---|---|
| R3-BL-001 | `G115` | HIGH | MEDIUM | `internal/launch/id.go` | 56 | integer overflow conversion uint64 -> byte | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-002 | `G115` | HIGH | MEDIUM | `internal/launch/id.go` | 55 | integer overflow conversion uint64 -> byte | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-003 | `G115` | HIGH | MEDIUM | `internal/launch/id.go` | 53 | integer overflow conversion uint64 -> byte | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-004 | `G115` | HIGH | MEDIUM | `internal/launch/id.go` | 50 | integer overflow conversion uint64 -> byte | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-005 | `G115` | HIGH | MEDIUM | `internal/browser/launch_chromium.go` | 1015 | integer overflow conversion int64 -> uint64 | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-006 | `G115` | HIGH | MEDIUM | `internal/browser/archive_security.go` | 244 | integer overflow conversion int64 -> uint64 | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-007 | `G115` | HIGH | MEDIUM | `internal/browser/archive_security.go` | 209 | integer overflow conversion int64 -> uint32 | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-008 | `G115` | HIGH | MEDIUM | `cmd/server/cli_runtime.go` | 828 | integer overflow conversion int64 -> uint32 | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-009 | `G404` | HIGH | MEDIUM | `internal/humanize/mouse.go` | 89 | Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand) | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-010 | `G404` | HIGH | MEDIUM | `internal/humanize/mouse.go` | 37 | Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand) | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-011 | `G404` | HIGH | MEDIUM | `internal/humanize/mouse.go` | 36 | Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand) | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-012 | `G404` | HIGH | MEDIUM | `internal/humanize/math.go` | 128 | Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand) | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-013 | `G404` | HIGH | MEDIUM | `internal/humanize/math.go` | 120 | Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand) | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-014 | `G404` | HIGH | MEDIUM | `internal/humanize/math.go` | 115 | Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand) | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-015 | `G404` | HIGH | MEDIUM | `internal/humanize/math.go` | 108 | Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand) | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-016 | `G404` | HIGH | MEDIUM | `internal/humanize/math.go` | 98 | Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand) | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-017 | `G404` | HIGH | MEDIUM | `internal/humanize/math.go` | 97 | Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand) | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-018 | `G404` | HIGH | MEDIUM | `internal/humanize/math.go` | 77 | Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand) | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-019 | `G404` | HIGH | MEDIUM | `internal/humanize/math.go` | 76 | Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand) | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-020 | `G404` | HIGH | MEDIUM | `internal/humanize/math.go` | 59 | Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand) | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-021 | `G404` | HIGH | MEDIUM | `internal/humanize/math.go` | 55 | Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand) | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-022 | `G404` | HIGH | MEDIUM | `internal/humanize/math.go` | 49 | Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand) | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-023 | `G404` | HIGH | MEDIUM | `internal/humanize/keyboard.go` | 84 | Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand) | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-024 | `G404` | HIGH | MEDIUM | `internal/humanize/keyboard.go` | 56 | Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand) | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-025 | `G404` | HIGH | MEDIUM | `internal/fingerprint/pool.go` | 64 | Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand) | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-033 | `G101` | HIGH | LOW | `internal/api/admin_token.go` | 22 | Potential hardcoded credentials | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-034 | `G703` | HIGH | HIGH | `cmd/server/cli_runtime.go` | 914 | Path traversal via taint analysis | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-035 | `G703` | HIGH | HIGH | `cmd/server/cli_runtime.go` | 883 | Path traversal via taint analysis | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-036 | `G703` | HIGH | HIGH | `cmd/server/cli_runtime.go` | 880 | Path traversal via taint analysis | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-037 | `G703` | HIGH | HIGH | `cmd/server/cli_runtime.go` | 768 | Path traversal via taint analysis | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-038 | `G703` | HIGH | HIGH | `cmd/server/cli_runtime.go` | 744 | Path traversal via taint analysis | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-039 | `G703` | HIGH | HIGH | `cmd/server/cli_runtime.go` | 741 | Path traversal via taint analysis | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-040 | `G703` | HIGH | HIGH | `cmd/server/cli_runtime.go` | 731 | Path traversal via taint analysis | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-041 | `G703` | HIGH | HIGH | `cmd/server/cli_runtime.go` | 692 | Path traversal via taint analysis | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-042 | `G703` | HIGH | HIGH | `cmd/server/cli_runtime.go` | 594 | Path traversal via taint analysis | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-043 | `G703` | HIGH | HIGH | `cmd/server/cli_runtime.go` | 591 | Path traversal via taint analysis | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-044 | `G703` | HIGH | HIGH | `cmd/server/cli_runtime.go` | 581 | Path traversal via taint analysis | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-045 | `G703` | HIGH | HIGH | `cmd/server/cli_runtime.go` | 457 | Path traversal via taint analysis | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-046 | `G118` | HIGH | MEDIUM | `internal/launch/launch.go` | 192 | Goroutine uses context.Background/TODO while request-scoped context is available | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-047 | `G122` | HIGH | MEDIUM | `cmd/server/cli_runtime.go` | 658 | Filesystem operation in filepath.Walk/WalkDir callback uses race-prone path; consider root-scoped APIs (e.g. os.Root) to prevent symlink TOCTOU traversal | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-053 | `G304` | MEDIUM | HIGH | `internal/workflow/engine.go` | 53 | Potential file inclusion via variable | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-054 | `G304` | MEDIUM | HIGH | `internal/mcp/advanced_tools.go` | 481 | Potential file inclusion via variable | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-055 | `G304` | MEDIUM | HIGH | `internal/fingerprint/pool.go` | 25 | Potential file inclusion via variable | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-056 | `G304` | MEDIUM | HIGH | `internal/config/config.go` | 172 | Potential file inclusion via variable | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-057 | `G304` | MEDIUM | HIGH | `internal/config/config.go` | 91 | Potential file inclusion via variable | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-058 | `G304` | MEDIUM | HIGH | `internal/browser/launch_chromium.go` | 201 | Potential file inclusion via variable | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-059 | `G304` | MEDIUM | HIGH | `internal/browser/download.go` | 532 | Potential file inclusion via variable | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-060 | `G304` | MEDIUM | HIGH | `internal/browser/download.go` | 136 | Potential file inclusion via variable | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-061 | `G304` | MEDIUM | HIGH | `internal/browser/archive_security.go` | 209 | Potential file inclusion via variable | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-062 | `G304` | MEDIUM | HIGH | `internal/browser/archive_security.go` | 168 | Potential file inclusion via variable | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-063 | `G304` | MEDIUM | HIGH | `internal/browser/archive_security.go` | 141 | Potential file inclusion via variable | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-064 | `G304` | MEDIUM | HIGH | `internal/api/router.go` | 831 | Potential file inclusion via variable | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-065 | `G304` | MEDIUM | HIGH | `cmd/server/cli_runtime.go` | 914 | Potential file inclusion via variable | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-066 | `G304` | MEDIUM | HIGH | `cmd/server/cli_runtime.go` | 883 | Potential file inclusion via variable | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-067 | `G304` | MEDIUM | HIGH | `cmd/server/cli_runtime.go` | 744 | Potential file inclusion via variable | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-068 | `G304` | MEDIUM | HIGH | `cmd/server/cli_runtime.go` | 692 | Potential file inclusion via variable | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-069 | `G304` | MEDIUM | HIGH | `cmd/server/cli_runtime.go` | 658 | Potential file inclusion via variable | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-070 | `G304` | MEDIUM | HIGH | `cmd/server/cli_runtime.go` | 594 | Potential file inclusion via variable | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-071 | `G304` | MEDIUM | HIGH | `cmd/server/cli.go` | 856 | Potential file inclusion via variable | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-072 | `G304` | MEDIUM | HIGH | `cmd/server/cli.go` | 572 | Potential file inclusion via variable | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-074 | `G305` | MEDIUM | HIGH | `cmd/server/cli_runtime.go` | 764 | File traversal when extracting zip/tar archive | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-075 | `G302` | MEDIUM | HIGH | `internal/browser/download.go` | 473 | Expect file permissions to be 0600 or less | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-076 | `G302` | MEDIUM | HIGH | `internal/browser/download.go` | 394 | Expect file permissions to be 0600 or less | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-077 | `G302` | MEDIUM | HIGH | `internal/browser/download.go` | 342 | Expect file permissions to be 0600 or less | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-078 | `G302` | MEDIUM | HIGH | `cmd/server/cli.go` | 763 | Expect file permissions to be 0600 or less | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-079 | `G301` | MEDIUM | HIGH | `internal/groups/store.go` | 45 | Expect directory permissions to be 0750 or less | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-080 | `G301` | MEDIUM | HIGH | `internal/browser/launch_chromium.go` | 1913 | Expect directory permissions to be 0750 or less | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-081 | `G301` | MEDIUM | HIGH | `internal/browser/launch_chromium.go` | 64 | Expect directory permissions to be 0750 or less | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-082 | `G301` | MEDIUM | HIGH | `internal/browser/download.go` | 449 | Expect directory permissions to be 0750 or less | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-083 | `G301` | MEDIUM | HIGH | `internal/browser/download.go` | 368 | Expect directory permissions to be 0750 or less | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-084 | `G301` | MEDIUM | HIGH | `internal/browser/download.go` | 307 | Expect directory permissions to be 0750 or less | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-085 | `G301` | MEDIUM | HIGH | `internal/browser/archive_security.go` | 206 | Expect directory permissions to be 0750 or less | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-086 | `G301` | MEDIUM | HIGH | `internal/browser/archive_security.go` | 199 | Expect directory permissions to be 0750 or less | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-087 | `G301` | MEDIUM | HIGH | `internal/browser/archive_security.go` | 138 | Expect directory permissions to be 0750 or less | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-088 | `G301` | MEDIUM | HIGH | `internal/browser/archive_security.go` | 130 | Expect directory permissions to be 0750 or less | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-089 | `G301` | MEDIUM | HIGH | `internal/browser/archive_security.go` | 23 | Expect directory permissions to be 0750 or less | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-090 | `G301` | MEDIUM | HIGH | `cmd/server/cli_runtime.go` | 880 | Expect directory permissions to be 0750 or less | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-091 | `G301` | MEDIUM | HIGH | `cmd/server/cli_runtime.go` | 768 | Expect directory permissions to be 0750 or less | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-092 | `G301` | MEDIUM | HIGH | `cmd/server/cli_runtime.go` | 741 | Expect directory permissions to be 0750 or less | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-093 | `G301` | MEDIUM | HIGH | `cmd/server/cli_runtime.go` | 731 | Expect directory permissions to be 0750 or less | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-094 | `G301` | MEDIUM | HIGH | `cmd/server/cli_runtime.go` | 457 | Expect directory permissions to be 0750 or less | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-095 | `G301` | MEDIUM | HIGH | `cmd/server/cli.go` | 778 | Expect directory permissions to be 0750 or less | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-096 | `G306` | MEDIUM | HIGH | `internal/browser/download.go` | 144 | Expect WriteFile permissions to be 0600 or less | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-097 | `G104` | LOW | HIGH | `internal/mcp/stdio.go` | 56 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-098 | `G104` | LOW | HIGH | `internal/mcp/server.go` | 818 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-099 | `G104` | LOW | HIGH | `internal/mcp/server.go` | 667 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-100 | `G104` | LOW | HIGH | `internal/mcp/protocol.go` | 40 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-101 | `G104` | LOW | HIGH | `internal/humanize/keyboard.go` | 73-75 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-102 | `G104` | LOW | HIGH | `internal/humanize/keyboard.go` | 60 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-103 | `G104` | LOW | HIGH | `internal/humanize/keyboard.go` | 33 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-104 | `G104` | LOW | HIGH | `internal/fingerprint/pool.go` | 72 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-105 | `G104` | LOW | HIGH | `internal/browser/socks5relay.go` | 144 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-106 | `G104` | LOW | HIGH | `internal/browser/socks5relay.go` | 142 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-107 | `G104` | LOW | HIGH | `internal/browser/socks5relay.go` | 135 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-108 | `G104` | LOW | HIGH | `internal/browser/socks5relay.go` | 129 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-109 | `G104` | LOW | HIGH | `internal/browser/socks5relay.go` | 122 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-110 | `G104` | LOW | HIGH | `internal/browser/socks5relay.go` | 95 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-111 | `G104` | LOW | HIGH | `internal/browser/socks5relay.go` | 87 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-112 | `G104` | LOW | HIGH | `internal/browser/socks5relay.go` | 49 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-113 | `G104` | LOW | HIGH | `internal/browser/manager.go` | 399 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-114 | `G104` | LOW | HIGH | `internal/browser/manager.go` | 391 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-115 | `G104` | LOW | HIGH | `internal/browser/manager.go` | 376 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-116 | `G104` | LOW | HIGH | `internal/browser/manager.go` | 373 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-117 | `G104` | LOW | HIGH | `internal/browser/launch_firefox.go` | 153 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-118 | `G104` | LOW | HIGH | `internal/browser/launch_chromium.go` | 333 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-119 | `G104` | LOW | HIGH | `internal/browser/launch_chromium.go` | 312 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-120 | `G104` | LOW | HIGH | `internal/browser/download.go` | 473 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-121 | `G104` | LOW | HIGH | `internal/browser/download.go` | 466 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-122 | `G104` | LOW | HIGH | `internal/browser/download.go` | 463 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-123 | `G104` | LOW | HIGH | `internal/browser/download.go` | 449 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-124 | `G104` | LOW | HIGH | `internal/browser/download.go` | 394 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-125 | `G104` | LOW | HIGH | `internal/browser/download.go` | 382 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-126 | `G104` | LOW | HIGH | `internal/browser/download.go` | 368 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-127 | `G104` | LOW | HIGH | `internal/browser/download.go` | 342 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-128 | `G104` | LOW | HIGH | `internal/browser/download.go` | 323 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-129 | `G104` | LOW | HIGH | `internal/browser/download.go` | 144 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-130 | `G104` | LOW | HIGH | `internal/api/router.go` | 707 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-131 | `G104` | LOW | HIGH | `internal/api/dashboard.go` | 13 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
| R3-BL-132 | `G104` | LOW | HIGH | `cmd/server/main.go` | 429 | Errors unhandled | `OPEN_BASELINE_REQUIRES_TRIAGE` |
