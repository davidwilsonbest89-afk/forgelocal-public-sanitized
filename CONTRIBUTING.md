# Contributing

BrowseForge aims to be useful to an international community of operators, QA engineers, automation developers, and browser-runtime researchers. Contributions should improve reliability, portability, security, documentation, or user experience without weakening those goals.

## Development Setup

Requirements:

- Go 1.22+
- Node.js 22+
- Docker with linux/amd64 build support
- `rg` for release preflight checks
- Camoufox installed under `browsers/camoufox/` for runtime spike tests, or set `CAMOUFOX_PATH`
- CloakBrowser installed under `browsers/cloakbrowser/` for Chromium-family runtime spike tests, or set `CLOAKBROWSER_PATH`

Common checks:

```bash
go test -count=1 -timeout 2m $(go list ./... | grep -v '/internal/spike$')
go vet ./...
bash -n scripts/release-preflight.sh scripts/release-push.sh
node scripts/check-i18n.js
node scripts/check-doc-language.js
docker compose -f docker/docker-compose.yml config
```

Camoufox Bind spike:

```bash
CAMOUFOX_PATH=/path/to/camoufox go test -count=1 -run '^TestPlaywrightBindEndpointWithCamoufox$' -v ./internal/spike
```

CloakBrowser Bind spike:

```bash
CLOAKBROWSER_SPIKE=1 CLOAKBROWSER_PATH=/path/to/Chromium go test -count=1 -timeout 45s -run '^TestPlaywrightBindEndpointWithCloakBrowser$' -v ./internal/spike
```

## Pull Request Expectations

Every PR should include:

- A clear problem statement and implementation summary.
- Tests or a concrete explanation for why tests are not practical.
- Platform notes when touching browser launch, Docker, Playwright integration, MCP, REST API, profiles, or release packaging.
- Documentation updates for user-visible changes.
- Public documentation kept English-first, with Traditional Chinese content in `.zh-TW.md` companions when localized.
- i18n updates for new user-facing UI text.

Do not remove tests to make CI pass. Fix the root cause, narrow the test, or document the platform-specific limitation.

## Release Process

Do not create or push release tags by hand. Follow [docs/release.md](docs/release.md):

```bash
scripts/release-preflight.sh vX.Y.Z
scripts/release-push.sh vX.Y.Z
gh run watch
```

## Internationalization

User-facing UI text must be localizable. Keep locale keys stable and update both English and Traditional Chinese strings when adding Dashboard or extension UI text. Public documentation uses English canonical files and `.zh-TW.md` Traditional Chinese companions. See [docs/i18n.md](docs/i18n.md).
