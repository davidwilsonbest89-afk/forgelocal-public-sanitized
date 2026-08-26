# GOSEC-R3 final individual matrix

The matrix contains every finding from the 132-finding source-only baseline. No `nosec`, `nolint`, skip or global allowlist was used.

| Classification | Count |
|---|---:|
| `CORRECTED_AND_VERIFIED` | 25 |
| `MITIGATED_CONTROL_SCANNER_OPEN` | 16 |
| `NEEDS_MANUAL_REVIEW` | 91 |

| Rule | Baseline | Final scan |
|---|---:|---:|
| `G101` | 1 | 1 |
| `G104` | 36 | 27 |
| `G107` | 1 | 1 |
| `G115` | 8 | 4 |
| `G118` | 1 | 0 |
| `G122` | 1 | 0 |
| `G204` | 5 | 5 |
| `G301` | 17 | 0 |
| `G302` | 4 | 5 |
| `G304` | 20 | 19 |
| `G305` | 1 | 1 |
| `G306` | 1 | 0 |
| `G404` | 17 | 17 |
| `G703` | 12 | 12 |
| `G704` | 7 | 7 |

| ID | Rule | File | Line | Status | Rationale |
|---|---|---|---:|---|---|
| R3-001 | `G115` | `internal/launch/id.go` | 56 | `CORRECTED_AND_VERIFIED` | The four implicit byte truncations were replaced with explicit binary.BigEndian encoding. |
| R3-002 | `G115` | `internal/launch/id.go` | 55 | `CORRECTED_AND_VERIFIED` | The four implicit byte truncations were replaced with explicit binary.BigEndian encoding. |
| R3-003 | `G115` | `internal/launch/id.go` | 53 | `CORRECTED_AND_VERIFIED` | The four implicit byte truncations were replaced with explicit binary.BigEndian encoding. |
| R3-004 | `G115` | `internal/launch/id.go` | 50 | `CORRECTED_AND_VERIFIED` | The four implicit byte truncations were replaced with explicit binary.BigEndian encoding. |
| R3-005 | `G115` | `internal/browser/launch_chromium.go` | 1015 | `MITIGATED_CONTROL_SCANNER_OPEN` | The conversion is bounded by a preceding range/size check or a positive-value contract; Gosec remains visible for manual review. |
| R3-006 | `G115` | `internal/browser/archive_security.go` | 244 | `MITIGATED_CONTROL_SCANNER_OPEN` | The conversion is bounded by a preceding range/size check or a positive-value contract; Gosec remains visible for manual review. |
| R3-007 | `G115` | `internal/browser/archive_security.go` | 209 | `MITIGATED_CONTROL_SCANNER_OPEN` | The conversion is bounded by a preceding range/size check or a positive-value contract; Gosec remains visible for manual review. |
| R3-008 | `G115` | `cmd/server/cli_runtime.go` | 828 | `MITIGATED_CONTROL_SCANNER_OPEN` | The conversion is bounded by a preceding range/size check or a positive-value contract; Gosec remains visible for manual review. |
| R3-009 | `G404` | `internal/humanize/mouse.go` | 89 | `NEEDS_MANUAL_REVIEW` | math/rand is used for humanization/fingerprint selection rather than cryptographic authorization; replacement requires behavior review. |
| R3-010 | `G404` | `internal/humanize/mouse.go` | 37 | `NEEDS_MANUAL_REVIEW` | math/rand is used for humanization/fingerprint selection rather than cryptographic authorization; replacement requires behavior review. |
| R3-011 | `G404` | `internal/humanize/mouse.go` | 36 | `NEEDS_MANUAL_REVIEW` | math/rand is used for humanization/fingerprint selection rather than cryptographic authorization; replacement requires behavior review. |
| R3-012 | `G404` | `internal/humanize/math.go` | 128 | `NEEDS_MANUAL_REVIEW` | math/rand is used for humanization/fingerprint selection rather than cryptographic authorization; replacement requires behavior review. |
| R3-013 | `G404` | `internal/humanize/math.go` | 120 | `NEEDS_MANUAL_REVIEW` | math/rand is used for humanization/fingerprint selection rather than cryptographic authorization; replacement requires behavior review. |
| R3-014 | `G404` | `internal/humanize/math.go` | 115 | `NEEDS_MANUAL_REVIEW` | math/rand is used for humanization/fingerprint selection rather than cryptographic authorization; replacement requires behavior review. |
| R3-015 | `G404` | `internal/humanize/math.go` | 108 | `NEEDS_MANUAL_REVIEW` | math/rand is used for humanization/fingerprint selection rather than cryptographic authorization; replacement requires behavior review. |
| R3-016 | `G404` | `internal/humanize/math.go` | 98 | `NEEDS_MANUAL_REVIEW` | math/rand is used for humanization/fingerprint selection rather than cryptographic authorization; replacement requires behavior review. |
| R3-017 | `G404` | `internal/humanize/math.go` | 97 | `NEEDS_MANUAL_REVIEW` | math/rand is used for humanization/fingerprint selection rather than cryptographic authorization; replacement requires behavior review. |
| R3-018 | `G404` | `internal/humanize/math.go` | 77 | `NEEDS_MANUAL_REVIEW` | math/rand is used for humanization/fingerprint selection rather than cryptographic authorization; replacement requires behavior review. |
| R3-019 | `G404` | `internal/humanize/math.go` | 76 | `NEEDS_MANUAL_REVIEW` | math/rand is used for humanization/fingerprint selection rather than cryptographic authorization; replacement requires behavior review. |
| R3-020 | `G404` | `internal/humanize/math.go` | 59 | `NEEDS_MANUAL_REVIEW` | math/rand is used for humanization/fingerprint selection rather than cryptographic authorization; replacement requires behavior review. |
| R3-021 | `G404` | `internal/humanize/math.go` | 55 | `NEEDS_MANUAL_REVIEW` | math/rand is used for humanization/fingerprint selection rather than cryptographic authorization; replacement requires behavior review. |
| R3-022 | `G404` | `internal/humanize/math.go` | 49 | `NEEDS_MANUAL_REVIEW` | math/rand is used for humanization/fingerprint selection rather than cryptographic authorization; replacement requires behavior review. |
| R3-023 | `G404` | `internal/humanize/keyboard.go` | 84 | `NEEDS_MANUAL_REVIEW` | math/rand is used for humanization/fingerprint selection rather than cryptographic authorization; replacement requires behavior review. |
| R3-024 | `G404` | `internal/humanize/keyboard.go` | 56 | `NEEDS_MANUAL_REVIEW` | math/rand is used for humanization/fingerprint selection rather than cryptographic authorization; replacement requires behavior review. |
| R3-025 | `G404` | `internal/fingerprint/pool.go` | 64 | `NEEDS_MANUAL_REVIEW` | math/rand is used for humanization/fingerprint selection rather than cryptographic authorization; replacement requires behavior review. |
| R3-026 | `G704` | `internal/api/sessions.go` | 478 | `MITIGATED_CONTROL_SCANNER_OPEN` | CLI and WebSocket paths enforce loopback, scheme, port, userinfo/query/fragment and timeout controls; static taint remains visible. |
| R3-027 | `G704` | `cmd/server/cli_runtime.go` | 924 | `MITIGATED_CONTROL_SCANNER_OPEN` | CLI and WebSocket paths enforce loopback, scheme, port, userinfo/query/fragment and timeout controls; static taint remains visible. |
| R3-028 | `G704` | `cmd/server/cli_runtime.go` | 872 | `MITIGATED_CONTROL_SCANNER_OPEN` | CLI and WebSocket paths enforce loopback, scheme, port, userinfo/query/fragment and timeout controls; static taint remains visible. |
| R3-029 | `G704` | `cmd/server/cli_runtime.go` | 867 | `MITIGATED_CONTROL_SCANNER_OPEN` | CLI and WebSocket paths enforce loopback, scheme, port, userinfo/query/fragment and timeout controls; static taint remains visible. |
| R3-030 | `G704` | `cmd/server/cli.go` | 1088 | `MITIGATED_CONTROL_SCANNER_OPEN` | CLI and WebSocket paths enforce loopback, scheme, port, userinfo/query/fragment and timeout controls; static taint remains visible. |
| R3-031 | `G704` | `cmd/server/cli.go` | 1004 | `MITIGATED_CONTROL_SCANNER_OPEN` | CLI and WebSocket paths enforce loopback, scheme, port, userinfo/query/fragment and timeout controls; static taint remains visible. |
| R3-032 | `G704` | `cmd/server/cli.go` | 985 | `MITIGATED_CONTROL_SCANNER_OPEN` | CLI and WebSocket paths enforce loopback, scheme, port, userinfo/query/fragment and timeout controls; static taint remains visible. |
| R3-033 | `G101` | `internal/api/admin_token.go` | 22 | `NEEDS_MANUAL_REVIEW` | The finding is the admin-token metadata identifier/context, not an embedded secret; it remains open pending independent secret-pattern review. |
| R3-034 | `G703` | `cmd/server/cli_runtime.go` | 914 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-035 | `G703` | `cmd/server/cli_runtime.go` | 883 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-036 | `G703` | `cmd/server/cli_runtime.go` | 880 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-037 | `G703` | `cmd/server/cli_runtime.go` | 768 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-038 | `G703` | `cmd/server/cli_runtime.go` | 744 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-039 | `G703` | `cmd/server/cli_runtime.go` | 741 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-040 | `G703` | `cmd/server/cli_runtime.go` | 731 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-041 | `G703` | `cmd/server/cli_runtime.go` | 692 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-042 | `G703` | `cmd/server/cli_runtime.go` | 594 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-043 | `G703` | `cmd/server/cli_runtime.go` | 591 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-044 | `G703` | `cmd/server/cli_runtime.go` | 581 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-045 | `G703` | `cmd/server/cli_runtime.go` | 457 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-046 | `G118` | `internal/launch/launch.go` | 192 | `CORRECTED_AND_VERIFIED` | The source-only post-R3 scan contains zero findings for this rule; the bounded context/root-scoped/permission fix is covered by targeted and full gates. |
| R3-047 | `G122` | `cmd/server/cli_runtime.go` | 658 | `CORRECTED_AND_VERIFIED` | The source-only post-R3 scan contains zero findings for this rule; the bounded context/root-scoped/permission fix is covered by targeted and full gates. |
| R3-048 | `G204` | `internal/browser/download.go` | 466 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-049 | `G204` | `internal/browser/download.go` | 352 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-050 | `G204` | `cmd/server/main.go` | 559 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-051 | `G204` | `cmd/server/main.go` | 557 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-052 | `G204` | `cmd/server/main.go` | 555 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-053 | `G304` | `internal/workflow/engine.go` | 53 | `NEEDS_MANUAL_REVIEW` | Finding remains open and requires individual review. |
| R3-054 | `G304` | `internal/mcp/advanced_tools.go` | 481 | `NEEDS_MANUAL_REVIEW` | Finding remains open and requires individual review. |
| R3-055 | `G304` | `internal/fingerprint/pool.go` | 25 | `NEEDS_MANUAL_REVIEW` | Finding remains open and requires individual review. |
| R3-056 | `G304` | `internal/config/config.go` | 172 | `NEEDS_MANUAL_REVIEW` | Finding remains open and requires individual review. |
| R3-057 | `G304` | `internal/config/config.go` | 91 | `NEEDS_MANUAL_REVIEW` | Finding remains open and requires individual review. |
| R3-058 | `G304` | `internal/browser/launch_chromium.go` | 201 | `NEEDS_MANUAL_REVIEW` | Finding remains open and requires individual review. |
| R3-059 | `G304` | `internal/browser/download.go` | 532 | `NEEDS_MANUAL_REVIEW` | Finding remains open and requires individual review. |
| R3-060 | `G304` | `internal/browser/download.go` | 136 | `NEEDS_MANUAL_REVIEW` | Finding remains open and requires individual review. |
| R3-061 | `G304` | `internal/browser/archive_security.go` | 209 | `NEEDS_MANUAL_REVIEW` | Finding remains open and requires individual review. |
| R3-062 | `G304` | `internal/browser/archive_security.go` | 168 | `NEEDS_MANUAL_REVIEW` | Finding remains open and requires individual review. |
| R3-063 | `G304` | `internal/browser/archive_security.go` | 141 | `NEEDS_MANUAL_REVIEW` | Finding remains open and requires individual review. |
| R3-064 | `G304` | `internal/api/router.go` | 831 | `NEEDS_MANUAL_REVIEW` | Finding remains open and requires individual review. |
| R3-065 | `G304` | `cmd/server/cli_runtime.go` | 914 | `NEEDS_MANUAL_REVIEW` | Finding remains open and requires individual review. |
| R3-066 | `G304` | `cmd/server/cli_runtime.go` | 883 | `NEEDS_MANUAL_REVIEW` | Finding remains open and requires individual review. |
| R3-067 | `G304` | `cmd/server/cli_runtime.go` | 744 | `NEEDS_MANUAL_REVIEW` | Finding remains open and requires individual review. |
| R3-068 | `G304` | `cmd/server/cli_runtime.go` | 692 | `NEEDS_MANUAL_REVIEW` | Finding remains open and requires individual review. |
| R3-069 | `G304` | `cmd/server/cli_runtime.go` | 658 | `CORRECTED_AND_VERIFIED` | The backup tar callback no longer opens the tainted filesystem path directly; it uses os.Root.Open. |
| R3-070 | `G304` | `cmd/server/cli_runtime.go` | 594 | `NEEDS_MANUAL_REVIEW` | Finding remains open and requires individual review. |
| R3-071 | `G304` | `cmd/server/cli.go` | 856 | `NEEDS_MANUAL_REVIEW` | Finding remains open and requires individual review. |
| R3-072 | `G304` | `cmd/server/cli.go` | 572 | `NEEDS_MANUAL_REVIEW` | Finding remains open and requires individual review. |
| R3-073 | `G107` | `internal/browser/download.go` | 523 | `NEEDS_MANUAL_REVIEW` | Network request context/control is outside the closed R3 lots and remains scanner-open. |
| R3-074 | `G305` | `cmd/server/cli_runtime.go` | 764 | `MITIGATED_CONTROL_SCANNER_OPEN` | Archive extraction retains lexical confinement, type rejection, size/count/depth limits and staging rollback; scanner still reports the pattern. |
| R3-075 | `G302` | `internal/browser/download.go` | 473 | `MITIGATED_CONTROL_SCANNER_OPEN` | Runtime executables intentionally use 0755 inside owner-only directories, while secrets/markers/downloads/backups use 0600; remaining scanner rows need manual policy review. |
| R3-076 | `G302` | `internal/browser/download.go` | 394 | `MITIGATED_CONTROL_SCANNER_OPEN` | Runtime executables intentionally use 0755 inside owner-only directories, while secrets/markers/downloads/backups use 0600; remaining scanner rows need manual policy review. |
| R3-077 | `G302` | `internal/browser/download.go` | 342 | `MITIGATED_CONTROL_SCANNER_OPEN` | Runtime executables intentionally use 0755 inside owner-only directories, while secrets/markers/downloads/backups use 0600; remaining scanner rows need manual policy review. |
| R3-078 | `G302` | `cmd/server/cli.go` | 763 | `MITIGATED_CONTROL_SCANNER_OPEN` | Runtime executables intentionally use 0755 inside owner-only directories, while secrets/markers/downloads/backups use 0600; remaining scanner rows need manual policy review. |
| R3-079 | `G301` | `internal/groups/store.go` | 45 | `CORRECTED_AND_VERIFIED` | All baseline G301 entries disappeared after owner-only runtime, archive, group and configuration directory modes. |
| R3-080 | `G301` | `internal/browser/launch_chromium.go` | 1913 | `CORRECTED_AND_VERIFIED` | All baseline G301 entries disappeared after owner-only runtime, archive, group and configuration directory modes. |
| R3-081 | `G301` | `internal/browser/launch_chromium.go` | 64 | `CORRECTED_AND_VERIFIED` | All baseline G301 entries disappeared after owner-only runtime, archive, group and configuration directory modes. |
| R3-082 | `G301` | `internal/browser/download.go` | 449 | `CORRECTED_AND_VERIFIED` | All baseline G301 entries disappeared after owner-only runtime, archive, group and configuration directory modes. |
| R3-083 | `G301` | `internal/browser/download.go` | 368 | `CORRECTED_AND_VERIFIED` | All baseline G301 entries disappeared after owner-only runtime, archive, group and configuration directory modes. |
| R3-084 | `G301` | `internal/browser/download.go` | 307 | `CORRECTED_AND_VERIFIED` | All baseline G301 entries disappeared after owner-only runtime, archive, group and configuration directory modes. |
| R3-085 | `G301` | `internal/browser/archive_security.go` | 206 | `CORRECTED_AND_VERIFIED` | All baseline G301 entries disappeared after owner-only runtime, archive, group and configuration directory modes. |
| R3-086 | `G301` | `internal/browser/archive_security.go` | 199 | `CORRECTED_AND_VERIFIED` | All baseline G301 entries disappeared after owner-only runtime, archive, group and configuration directory modes. |
| R3-087 | `G301` | `internal/browser/archive_security.go` | 138 | `CORRECTED_AND_VERIFIED` | All baseline G301 entries disappeared after owner-only runtime, archive, group and configuration directory modes. |
| R3-088 | `G301` | `internal/browser/archive_security.go` | 130 | `CORRECTED_AND_VERIFIED` | All baseline G301 entries disappeared after owner-only runtime, archive, group and configuration directory modes. |
| R3-089 | `G301` | `internal/browser/archive_security.go` | 23 | `CORRECTED_AND_VERIFIED` | All baseline G301 entries disappeared after owner-only runtime, archive, group and configuration directory modes. |
| R3-090 | `G301` | `cmd/server/cli_runtime.go` | 880 | `CORRECTED_AND_VERIFIED` | All baseline G301 entries disappeared after owner-only runtime, archive, group and configuration directory modes. |
| R3-091 | `G301` | `cmd/server/cli_runtime.go` | 768 | `CORRECTED_AND_VERIFIED` | All baseline G301 entries disappeared after owner-only runtime, archive, group and configuration directory modes. |
| R3-092 | `G301` | `cmd/server/cli_runtime.go` | 741 | `CORRECTED_AND_VERIFIED` | All baseline G301 entries disappeared after owner-only runtime, archive, group and configuration directory modes. |
| R3-093 | `G301` | `cmd/server/cli_runtime.go` | 731 | `CORRECTED_AND_VERIFIED` | All baseline G301 entries disappeared after owner-only runtime, archive, group and configuration directory modes. |
| R3-094 | `G301` | `cmd/server/cli_runtime.go` | 457 | `CORRECTED_AND_VERIFIED` | All baseline G301 entries disappeared after owner-only runtime, archive, group and configuration directory modes. |
| R3-095 | `G301` | `cmd/server/cli.go` | 778 | `CORRECTED_AND_VERIFIED` | All baseline G301 entries disappeared after owner-only runtime, archive, group and configuration directory modes. |
| R3-096 | `G306` | `internal/browser/download.go` | 144 | `CORRECTED_AND_VERIFIED` | The source-only post-R3 scan contains zero findings for this rule; the bounded context/root-scoped/permission fix is covered by targeted and full gates. |
| R3-097 | `G104` | `internal/mcp/stdio.go` | 56 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-098 | `G104` | `internal/mcp/server.go` | 818 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-099 | `G104` | `internal/mcp/server.go` | 667 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-100 | `G104` | `internal/mcp/protocol.go` | 40 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-101 | `G104` | `internal/humanize/keyboard.go` | 73-75 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-102 | `G104` | `internal/humanize/keyboard.go` | 60 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-103 | `G104` | `internal/humanize/keyboard.go` | 33 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-104 | `G104` | `internal/fingerprint/pool.go` | 72 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-105 | `G104` | `internal/browser/socks5relay.go` | 144 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-106 | `G104` | `internal/browser/socks5relay.go` | 142 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-107 | `G104` | `internal/browser/socks5relay.go` | 135 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-108 | `G104` | `internal/browser/socks5relay.go` | 129 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-109 | `G104` | `internal/browser/socks5relay.go` | 122 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-110 | `G104` | `internal/browser/socks5relay.go` | 95 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-111 | `G104` | `internal/browser/socks5relay.go` | 87 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-112 | `G104` | `internal/browser/socks5relay.go` | 49 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-113 | `G104` | `internal/browser/manager.go` | 399 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-114 | `G104` | `internal/browser/manager.go` | 391 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-115 | `G104` | `internal/browser/manager.go` | 376 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-116 | `G104` | `internal/browser/manager.go` | 373 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-117 | `G104` | `internal/browser/launch_firefox.go` | 153 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-118 | `G104` | `internal/browser/launch_chromium.go` | 333 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-119 | `G104` | `internal/browser/launch_chromium.go` | 312 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-120 | `G104` | `internal/browser/download.go` | 473 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-121 | `G104` | `internal/browser/download.go` | 466 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-122 | `G104` | `internal/browser/download.go` | 463 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-123 | `G104` | `internal/browser/download.go` | 449 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-124 | `G104` | `internal/browser/download.go` | 394 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-125 | `G104` | `internal/browser/download.go` | 382 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-126 | `G104` | `internal/browser/download.go` | 368 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-127 | `G104` | `internal/browser/download.go` | 342 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-128 | `G104` | `internal/browser/download.go` | 323 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-129 | `G104` | `internal/browser/download.go` | 144 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-130 | `G104` | `internal/api/router.go` | 707 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-131 | `G104` | `internal/api/dashboard.go` | 13 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
| R3-132 | `G104` | `cmd/server/main.go` | 429 | `NEEDS_MANUAL_REVIEW` | The finding remains scanner-open; it requires component-level review rather than a global suppression or allowlist. |
