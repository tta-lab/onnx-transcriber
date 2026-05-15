.PHONY: help build install setup run clean reset test tidy fmt lint all dev ci check-clean open-guide

BINARY := onnx-transcribe
CMD := ./cmd/onnx-transcribe
GOBIN := $(shell go env GOPATH)/bin

help:
	@echo "Available commands:"
	@echo "  make build       - Build the onnx-transcribe binary"
	@echo "  make install     - Install onnx-transcribe to GOPATH/bin"
	@echo "  make setup       - Install binary and print next steps"
	@echo "  make run         - Run onnx-transcribe (usage: make run ARGS='doctor')"
	@echo "  make clean       - Remove built binaries"
	@echo "  make reset       - Remove binaries"
	@echo "  make test        - Run tests"
	@echo "  make tidy        - Tidy go modules"
	@echo "  make fmt         - Format code with gofmt"
	@echo "  make lint        - Run golangci-lint"
	@echo "  make all         - Format, tidy, lint, and build"
	@echo "  make ci          - Run CI checks"
	@echo "  make check-clean - Check if working directory is clean"
	@echo "  make open-guide  - Open docs/try.html"

build:
	@echo "Building $(BINARY)..."
	@go build -o $(BINARY) $(CMD)
	@echo "✓ Build complete: ./$(BINARY)"

install:
	@echo "Installing $(BINARY)..."
	@go build -o $(GOBIN)/$(BINARY) $(CMD)
	@echo "✓ Installed to $(GOBIN)/$(BINARY)"

setup: install
	@echo "Setup complete"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Run: $(BINARY) doctor"
	@echo "  2. Run: $(BINARY) setup"
	@echo "  3. Run: sh scripts/smoke-mp4.sh"

run: build
	@./$(BINARY) $(ARGS)

clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY)
	@echo "✓ Cleaned build artifacts"

reset: clean
	@echo "✓ Reset complete"

test:
	@echo "Running tests..."
	@go test -v -count=1 ./...

tidy:
	@echo "Tidying go modules..."
	@go mod tidy
	@echo "✓ go mod tidy complete"

fmt:
	@echo "Formatting code..."
	@gofmt -w -s .
	@echo "✓ Code formatted"

lint:
	@echo "Running golangci-lint..."
	@golangci-lint run ./...
	@echo "✓ Lint complete"

all: fmt tidy lint build
	@echo "✓ All checks passed and binary built"

dev: all
	@echo "✓ Development build complete"

ci: lint test build
	@echo "✓ CI checks complete"

check-clean:
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Working directory is not clean"; \
		git status --short; \
		exit 1; \
	else \
		echo "✓ Working directory is clean"; \
	fi

open-guide:
	@open docs/try.html
