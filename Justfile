# SPDX-License-Identifier: MIT
# Copyright (c) 2025-2026 Jason Giese (Bl4cky99)

out_dir   := "bin"
coverfile := "coverage.out"

# List all available recipes
help:
    @just --list

# Build a local Caddy binary with this module via xcaddy
build:
    mkdir -p {{out_dir}}
    xcaddy build --with github.com/Bl4cky99/caddy-redirector=. --output {{out_dir}}/caddy

# Start Caddy with the example Caddyfile (requires a built binary)
run: build
    ./{{out_dir}}/caddy run --config example/Caddyfile --adapter caddyfile

# Run the unit test suite
test:
    go test -race -timeout=2m ./internal/...

# Run the unit test suite with verbose Ginkgo output
test-verbose:
    ginkgo -v ./internal/...

# Run the integration test (builds a caddy binary first if needed)
test-integration: build
    go test -race -tags=integration -timeout=30s ./internal/...

# Run benchmarks
bench:
    go test -run '^$' -bench . -benchmem -tags=bench ./internal/...

# Run tests with coverage and show summary
cover:
    go test -coverprofile={{coverfile}} ./internal/...
    go tool cover -func={{coverfile}} | tail -n 1

# Generate HTML coverage report and open it
cover-html: cover
    mkdir -p {{out_dir}}
    go tool cover -html={{coverfile}} -o {{out_dir}}/coverage.html

# Format code with go fmt (and goimports if installed)
fmt:
    go fmt ./...
    -goimports -w .

# Static analysis with go vet
vet:
    go vet ./...

# Sync go.mod/go.sum
tidy:
    go mod tidy

# Remove build artifacts
clean:
    rm -rf {{out_dir}} {{coverfile}}
