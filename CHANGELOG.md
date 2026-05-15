# Changelog

All notable changes should be documented here. This project follows semantic version tags in the form `vX.Y.Z`.

## Unreleased

### Added

- Guarded release scripts and release workflow asset checks.
- Docker release build hardening with pinned release artifact selection and KasmVNC checksum verification.
- Community governance, support, security, and contribution documentation.
- Initial application i18n policy and locale structure.
- English-first README and API reference with Traditional Chinese counterparts.
- English public docs for platform support, Linux server deployment, and Playwright patch status.
- i18n coverage checker for Dashboard and WebExtension locale key parity.
- Marketing-oriented product positioning, audience, trust, and deployment messaging in README.
- Dual-browser anti-detection architecture documentation in English and Traditional Chinese.
- Opt-in CloakBrowser runtime spike harness for the Playwright Bind endpoint path.

### Changed

- Docker documentation recommends pinning version tags for production deployments.
- Replaced remaining early Camoufox-only tool naming in local scripts and clarified current dual-browser fingerprint behavior.
- Release preflight runs the CloakBrowser Bind spike when a local binary is available and can enforce it with `REQUIRE_CLOAKBROWSER=1`.

## v1.7.0 - 2026-05-15

### Added

- Playwright 1.60 integration through the project fork.
- MCP Streamable HTTP authentication with Bearer tokens.
- Camoufox runtime spike coverage for the Playwright Bind endpoint path.

### Changed

- Removed the previous Playwright 1.59.1 hotfix path and now uses the Playwright 1.60 `browser.Bind()` endpoint directly.
- Improved startup, token, browser-download, profile-store, backup/restore, and session request error handling.

### Upgrade Notes

- External Playwright clients should use Playwright 1.60.x for `browserType.connect()`.
- Existing `config.json`, `data/.api-token`, and `profiles/` remain compatible.
- MCP HTTP clients must send `Authorization: Bearer <token>`.
