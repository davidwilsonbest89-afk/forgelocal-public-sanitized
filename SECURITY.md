# Security Policy

BrowseForge controls browser profiles, local API tokens, MCP access, browser automation sessions, and remote Docker endpoints. Please report security issues privately and do not open a public issue for vulnerabilities.

## Reporting a Vulnerability

Send a private report to the project maintainer with:

- Affected version or commit.
- Platform and deployment mode.
- Reproduction steps.
- Impact and whether credentials, profiles, tokens, or remote browser control are exposed.
- Suggested mitigation, if known.

Do not include real user tokens, production profiles, private cookies, or secrets in reports.

## Supported Versions

Security fixes target the latest released version and `main`. Older versions receive best-effort guidance only unless the project publishes an explicit long-term support policy.

## Security Boundaries

- Do not expose ports `19280`, `19281`, or `6901` directly to the public internet.
- Use VPN, SSH tunnels, or a hardened reverse proxy with HTTPS and access controls.
- Treat `data/.api-token`, browser profiles, backup ZIPs, and exported profiles as sensitive.
- MCP Streamable HTTP requires `Authorization: Bearer <token>`.

## Responsible Use

BrowseForge is intended for legitimate QA, automation, privacy research, compatibility testing, and controlled browser operations. Do not use it for unauthorized access, credential abuse, spam, fraud, or evasion of systems you do not own or have permission to test.
