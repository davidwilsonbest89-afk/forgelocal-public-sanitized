.PHONY: dev build test package clean spike

# --- Development (in Docker) ---

dev:
	docker build --target base -t camoufoxmulti-dev .
	docker run -it --rm -v $(PWD):/app -p 19280:19280 camoufoxmulti-dev bash

# --- Build ---

build-server:
	docker build --target build-server -t camoufoxmulti-build .
	docker create --name cmfx-build camoufoxmulti-build
	docker cp cmfx-build:/out/control-server ./dist/control-server
	docker rm cmfx-build
	chmod +x ./dist/control-server

build-fingerprints:
	docker build --target build-fingerprints -t camoufoxmulti-fp .
	docker create --name cmfx-fp camoufoxmulti-fp
	docker cp cmfx-fp:/out/ ./dist/data/
	docker rm cmfx-fp

build: build-server build-fingerprints
	@echo "Build complete: ./dist/"

# --- Test (in Docker) ---

test:
	docker build --target test -t camoufoxmulti-test .

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
	docker build --target base -t camoufoxmulti-dev .
	docker run --rm -v $(PWD):/app camoufoxmulti-dev go test ./internal/spike/...

# --- Clean ---

clean:
	rm -rf dist/
	docker rmi camoufoxmulti-dev camoufoxmulti-build camoufoxmulti-fp camoufoxmulti-test 2>/dev/null || true

# --- Download Camoufox binary (macOS ARM64) ---

CAMOUFOX_VERSION ?= v135.0.1-beta.24

download-camoufox:
	@mkdir -p dist/BrowseForge
	@echo "Download Camoufox $(CAMOUFOX_VERSION) for macOS ARM64..."
	@echo "Visit: https://github.com/daijro/camoufox/releases/tag/$(CAMOUFOX_VERSION)"
	@echo "Extract to: dist/BrowseForge/camoufox/"
