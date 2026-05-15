# Support

Use GitHub issues for reproducible bugs, documentation gaps, feature requests, and platform compatibility reports.

## Before Opening an Issue

Please include:

- BrowseForge version or commit.
- Operating system and architecture.
- Deployment mode: native binary, Docker, or source build.
- Browser engine: Camoufox, CloakBrowser, or both.
- Relevant logs without tokens, cookies, profile data, or secrets.
- Exact reproduction steps.

For security vulnerabilities, use [SECURITY.md](SECURITY.md) instead of a public issue.

## Platform Support

See [docs/platform-support.md](docs/platform-support.md) for the current support matrix.

Docker images are currently `linux/amd64`. Apple Silicon can run them through Docker emulation. Native `linux/arm64` Docker images are not considered supported until KasmVNC, Camoufox, and CloakBrowser runtime checks pass inside an ARM container.
