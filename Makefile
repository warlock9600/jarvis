BINARY := jarvis
PKG := ./cmd/jarvis
INSTALL_DIR := /opt/homebrew/sbin
HOST_OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
HOST_ARCH_RAW := $(shell uname -m)
HOST_ARCH := $(if $(filter $(HOST_ARCH_RAW),x86_64),amd64,$(if $(filter $(HOST_ARCH_RAW),aarch64 arm64),arm64,$(HOST_ARCH_RAW)))

.PHONY: build install test lint run clean build-docker test-docker lint-docker

build:
	docker run --rm -v $(PWD):/src -w /src golang:1.25 sh -lc 'export PATH=/usr/local/go/bin:$$PATH && mkdir -p bin && CGO_ENABLED=0 GOOS=$(HOST_OS) GOARCH=$(HOST_ARCH) go build -buildvcs=false -trimpath -ldflags="-s -w" -o bin/$(BINARY) $(PKG)'

install: build
	mkdir -p $(INSTALL_DIR)
	install -m 0755 bin/$(BINARY) $(INSTALL_DIR)/$(BINARY)

test:
	go test ./...

lint:
	golangci-lint run ./...

run:
	go run $(PKG) $(ARGS)

clean:
	rm -rf bin

build-docker:
	make build

test-docker:
	docker run --rm -v $(PWD):/src -w /src golang:1.25 sh -lc 'export PATH=/usr/local/go/bin:$$PATH && go test ./...'

lint-docker:
	docker run --rm -v $(PWD):/src -w /src golang:1.25 sh -lc 'export PATH=/usr/local/go/bin:$$PATH && go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8 && $$(go env GOPATH)/bin/golangci-lint run ./...'
