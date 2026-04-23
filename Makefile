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
	@mkdir -p dist/CamoufoxMulti
	@cp dist/control-server dist/CamoufoxMulti/
	@cp -r extension/ dist/CamoufoxMulti/extension/
	@cp -r dist/data/ dist/CamoufoxMulti/data/
	@mkdir -p dist/CamoufoxMulti/profiles
	@mkdir -p dist/CamoufoxMulti/logs
	@cp config.default.json dist/CamoufoxMulti/config.json
	@cp scripts/start.sh dist/CamoufoxMulti/ && chmod +x dist/CamoufoxMulti/start.sh
	@cp README.md dist/CamoufoxMulti/
	cd dist && zip -r CamoufoxMulti-v$(VERSION)-$(PLATFORM).zip CamoufoxMulti/
	@echo "Package: dist/CamoufoxMulti-v$(VERSION)-$(PLATFORM).zip"

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
	@mkdir -p dist/CamoufoxMulti
	@echo "Download Camoufox $(CAMOUFOX_VERSION) for macOS ARM64..."
	@echo "Visit: https://github.com/daijro/camoufox/releases/tag/$(CAMOUFOX_VERSION)"
	@echo "Extract to: dist/CamoufoxMulti/camoufox/"
