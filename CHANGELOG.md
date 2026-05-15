# Changelog

All notable changes should be documented here. This project follows semantic version tags in the form `vX.Y.Z`.

## Unreleased

### Added

- Guarded release scripts and release workflow asset checks.
- Docker release build hardening with pinned release artifact selection and KasmVNC checksum verification.
- Community governance, support, security, and contribution documentation.
- Initial application i18n policy and locale structure.

### Changed

- Docker documentation recommends pinning version tags for production deployments.

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
