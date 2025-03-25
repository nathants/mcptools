BINARY_NAME=mcptools
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

.PHONY: build clean install test

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/mcptools

build-all: clean
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)_darwin_amd64 ./cmd/mcptools
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)_darwin_arm64 ./cmd/mcptools
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)_linux_amd64 ./cmd/mcptools
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)_linux_arm64 ./cmd/mcptools
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)_windows_amd64.exe ./cmd/mcptools

install:
	go install $(LDFLAGS) ./cmd/mcptools

clean:
	rm -rf bin/

test:
	go test -v ./...

lint:
	golangci-lint run ./... 