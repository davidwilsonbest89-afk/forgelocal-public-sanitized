.PHONY: dev build test test-back01 test-back01-race package clean spike download-browsers download-camoufox download-cloakbrowser core-dev dashboard-dev test-bootstrap-ro

GO_TOOLCHAIN ?= go1.25.13

# --- Development (in Docker) ---

dev:
	docker build --target base -t browseforge-dev .
	docker run -it --rm -v $(PWD):/app -p 19280:19280 browseforge-dev bash

# --- Build ---

build-server:
	docker build --target build-server -t browseforge-build .
	docker create --name bf-build browseforge-build
	docker cp bf-build:/out/control-server ./dist/control-server
	docker rm bf-build
	chmod +x ./dist/control-server

build-fingerprints:
	docker build --target build-fingerprints -t browseforge-fp .
	docker create --name bf-fp browseforge-fp
	docker cp bf-fp:/out/ ./dist/data/
	docker rm bf-fp

build: build-server build-fingerprints
	@echo "Build complete: ./dist/"

# --- Test (in Docker) ---

test:
	docker build --target test -t browseforge-test .

# --- ForgeLocal BACK-01 verification ---

test-back01:
	GOTOOLCHAIN=$(GO_TOOLCHAIN) go test ./internal/backup -count=1 -v

test-back01-race:
	GOTOOLCHAIN=$(GO_TOOLCHAIN) go test -race ./internal/backup -count=1

# --- BOOTSTRAP-RO-01 local execution evidence ---
# These targets require a separate dashboard checkout through DASHBOARD_DIR.
# They are loopback-only, use a temporary directory and disable all runtimes.

core-dev:
	@FORGELOCAL_E2E_BASE_DIR="$${FORGELOCAL_E2E_BASE_DIR:?FORGELOCAL_E2E_BASE_DIR requis}" \
	FORGELOCAL_E2E_PORT="$${FORGELOCAL_E2E_PORT:-19280}" \
	./scripts/core-dev.sh

dashboard-dev:
	@DASHBOARD_DIR="$${DASHBOARD_DIR:?DASHBOARD_DIR requis}" \
	FORGELOCAL_E2E_PORT="$${FORGELOCAL_E2E_PORT:-19280}" \
	FORGELOCAL_DASHBOARD_PORT="$${FORGELOCAL_DASHBOARD_PORT:-3001}" \
	./scripts/dashboard-dev.sh

test-bootstrap-ro:
	@DASHBOARD_DIR="$${DASHBOARD_DIR:?DASHBOARD_DIR requis}" \
	FORGELOCAL_E2E_BASE_DIR="$${FORGELOCAL_E2E_BASE_DIR:?FORGELOCAL_E2E_BASE_DIR requis}" \
	FORGELOCAL_E2E_PORT="$${FORGELOCAL_E2E_PORT:-19280}" \
	FORGELOCAL_DASHBOARD_PORT="$${FORGELOCAL_DASHBOARD_PORT:-3001}" \
	FORGELOCAL_HOSTED_DASHBOARD_URL="$${FORGELOCAL_HOSTED_DASHBOARD_URL:-https://forgelocal-d-c8wqrxmp.manus.space}" \
	./scripts/test-bootstrap-ro-playwright.sh

# --- Package (final ZIP) ---

PLATFORM ?= macos-arm64
VERSION ?= 0.1.0

package: build
	@mkdir -p dist/BrowseForge
	@cp dist/control-server dist/BrowseForge/
	@cp -r extension/ dist/BrowseForge/extension/
	@cp -r dist/data/ dist/BrowseForge/data/
	@mkdir -p dist/BrowseForge/profiles
	@mkdir -p dist/BrowseForge/logs
	@cp config.default.json dist/BrowseForge/config.json
	@cp scripts/start.sh dist/BrowseForge/ && chmod +x dist/BrowseForge/start.sh
	@cp README.md README.zh-TW.md API.md API.zh-TW.md dist/BrowseForge/
	cd dist && zip -r BrowseForge-v$(VERSION)-$(PLATFORM).zip BrowseForge/
	@echo "Package: dist/BrowseForge-v$(VERSION)-$(PLATFORM).zip"

# --- Spike tests ---

spike:
	docker build --target base -t browseforge-dev .
	docker run --rm -v $(PWD):/app browseforge-dev go test ./internal/spike/...

# --- Clean ---

clean:
	rm -rf dist/
	docker rmi browseforge-dev browseforge-build browseforge-fp browseforge-test 2>/dev/null || true

# --- Download browser binaries (manual links for packaged builds) ---

CAMOUFOX_VERSION ?= v135.0.1-beta.24
CLOAKBROWSER_VERSION ?= chromium-v146.0.7680.177.4

download-browsers: download-camoufox download-cloakbrowser

download-camoufox:
	@mkdir -p dist/BrowseForge
	@echo "Download Camoufox $(CAMOUFOX_VERSION) for macOS ARM64..."
	@echo "Visit: https://github.com/daijro/camoufox/releases/tag/$(CAMOUFOX_VERSION)"
	@echo "Extract to: dist/BrowseForge/camoufox/"

download-cloakbrowser:
	@mkdir -p dist/BrowseForge
	@echo "Download CloakBrowser $(CLOAKBROWSER_VERSION) for macOS ARM64..."
	@echo "Visit: https://github.com/CloakHQ/CloakBrowser/releases/tag/$(CLOAKBROWSER_VERSION)"
	@echo "Extract to: dist/BrowseForge/cloakbrowser/"
