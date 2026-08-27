# BrowseForge Development Environment
# All build tools in Docker, host stays clean

FROM golang:1.26-bookworm AS base

# Node.js 22 LTS for fingerprint-suite and web-ext
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - && \
    apt-get install -y --no-install-recommends nodejs && \
    npm install -g web-ext

# Playwright system dependencies
RUN npx playwright install-deps firefox chromium

# Working directory
WORKDIR /app

# Go dependencies cache
COPY go.mod go.sum* ./
RUN go mod download 2>/dev/null || true

# Node dependencies cache
COPY package.json package-lock.json* ./
RUN npm install 2>/dev/null || true

COPY . .

# --- Build targets ---

# Build Go server for macOS ARM64
FROM base AS build-server
RUN CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
    -ldflags="-s -w" \
    -o /out/control-server \
    ./cmd/server

# Generate fingerprint pool
FROM base AS build-fingerprints
RUN node scripts/generate-fingerprints.js \
    --browser firefox --os windows --count 500 \
    --output /out/fingerprints-firefox-windows.json && \
    node scripts/generate-fingerprints.js \
    --browser firefox --os macos --count 500 \
    --output /out/fingerprints-firefox-macos.json && \
    node scripts/generate-fingerprints.js \
    --browser chrome --os windows --count 500 \
    --output /out/fingerprints-chrome-windows.json && \
    node scripts/generate-fingerprints.js \
    --browser chrome --os macos --count 500 \
    --output /out/fingerprints-chrome-macos.json

# Run tests
FROM base AS test
RUN go test -race ./...
