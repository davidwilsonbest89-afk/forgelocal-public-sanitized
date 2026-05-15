# Release Process

BrowseForge releases are tag driven. Do not create or push release tags by hand.

## Prepare

1. Update the version references:

   - `docker/docker-compose.yml`
   - `docker/README.md`
   - `docs/linux-server.md`

2. Commit and push `main`.

3. Run preflight from a clean `main` checkout:

   ```bash
   scripts/release-preflight.sh v1.7.3
   ```

The preflight checks the working tree, version references, Go tests, vet, UI i18n syntax, workflow YAML, Docker Compose config, the Camoufox Bind spike, the CloakBrowser Bind spike when a local binary is available, and a real `linux/amd64` Docker build.

For release machines that must prove both browser engines before publishing, enforce the CloakBrowser spike:

```bash
REQUIRE_CLOAKBROWSER=1 \
CLOAKBROWSER_PATH=/path/to/Chromium \
scripts/release-preflight.sh v1.7.3
```

## Publish

After preflight passes:

```bash
scripts/release-push.sh v1.7.3
gh run watch
```

The pushed tag starts the GitHub release workflow:

1. Verify
2. Cross-platform binary packages
3. GitHub release asset verification
4. GitHub release
5. GHCR Docker image

## Platform Policy

The release workflow publishes binary zip assets for:

- `linux-x64`
- `linux-arm64`
- `macos-x64`
- `macos-arm64`
- `windows-x64`

The Docker image is currently published as `linux/amd64` only. Apple Silicon runs it through Docker emulation. Native `linux/arm64` Docker images should only be enabled after KasmVNC, Camoufox, and CloakBrowser runtime checks pass inside an ARM container.
