SHELL := /bin/sh

BIN_DIR := bin
BIN := $(BIN_DIR)/pkt
DIST_DIR := dist
CGO_ENABLED ?= 0

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo 0.1.0)

LDFLAGS := -s -w \
	-X github.com/MarkOne-dev/pkt-git/internal/config.Version=$(VERSION)

.PHONY: build test fmt fmt-check vet lint dist clean

build:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=$(CGO_ENABLED) go build -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/pkt

test:
	go test ./...

fmt:
	gofmt -w cmd internal

fmt-check:
	@files="$$(gofmt -l cmd internal)"; \
	if [ -n "$$files" ]; then \
		echo "gofmt required for:"; \
		echo "$$files"; \
		exit 1; \
	fi

vet:
	go vet ./...

lint: fmt-check vet

dist:
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/pkt_linux_amd64 ./cmd/pkt
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/pkt_linux_arm64 ./cmd/pkt
	CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/pkt_darwin_amd64 ./cmd/pkt
	CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/pkt_darwin_arm64 ./cmd/pkt
	CGO_ENABLED=$(CGO_ENABLED) GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/pkt_windows_amd64.exe ./cmd/pkt

clean:
	rm -rf $(BIN_DIR) $(DIST_DIR)
