# Color definitions
BLUE=\033[0;34m
GREEN=\033[0;32m
YELLOW=\033[0;33m
RED=\033[0;31m
NC=\033[0m # No Color

BINARY_NAME=mcp
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Default Makefile step
default: setup
	@echo "$(GREEN)All setup steps completed successfully!$(NC)"

# Setup for Cocoapods
setup: \
	check-go \
	build \
	clean \
	install \
	test
	@echo "$(GREEN)Setup process completed$(NC)"

check-go:
	@echo "$(BLUE)Checking Go installation and version...$(NC)"
	chmod +x ./scripts/check_go.bash
	./scripts/check_go.bash

build:
	@echo "$(YELLOW)Building $(BINARY_NAME)...$(NC)"
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/mcptools

build-all: clean
	@echo "$(YELLOW)Building all platform binaries...$(NC)"
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)_darwin_amd64 ./cmd/mcptools
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)_darwin_arm64 ./cmd/mcptools
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)_linux_amd64 ./cmd/mcptools
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)_linux_arm64 ./cmd/mcptools
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)_windows_amd64.exe ./cmd/mcptools

install:
	@echo "$(BLUE)Installing $(BINARY_NAME)...$(NC)"
	go install $(LDFLAGS) ./cmd/mcptools

clean:
	@echo "$(RED)Cleaning build artifacts...$(NC)"
	

test: check-go
	@echo "$(YELLOW)Running tests...$(NC)"
	go test -v ./...

lint: check_go
	@echo "$(BLUE)Running linter...$(NC)"
	golangci-lint run ./...