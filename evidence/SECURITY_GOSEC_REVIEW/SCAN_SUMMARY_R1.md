# Scan summary R1
 
| Scan | Scope | Exit | Result |
|---|---|---:|---|
| Gosec | ./cmd/... ./internal/... | 1 | 176 findings, gate open |
| Govulncheck | ./cmd/... ./internal/... | 0 | no vulnerabilities found |
| Gitleaks | repository | 0 | no leaks found |
| Trivy | filesystem vuln/secret/misconfig | 0 | 0 vulnerabilities, 0 misconfigurations, 0 secrets |
| OSV | go.mod | 1 | 46 vulnerabilities; remediation review required |
| OSV | forge-dashboard/pnpm-lock.yaml | 0 | 0 vulnerabilities |
| OSV | recursive repository | 1 | 98 matches including historical SBOMs; not used as current-manifest count |
| Semgrep/Grype/Shellcheck/Yamllint | environment | n/a | unavailable |

Generated_UTC=2026-08-26T01:49:36Z
Tools:
Version: dev
Scanner: govulncheck@v1.1.4
8.18.4
osv-scanner version: 1.9.2
commit: n/a
built at: n/a
Version: 0.74.0
