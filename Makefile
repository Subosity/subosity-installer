# Subosity Installer Makefile
.PHONY: all clean build test lint help build-binary build-container run-tests

# Build configuration
BINARY_NAME := subosity-installer
CONTAINER_IMAGE := subosity/installer
VERSION := 1.0.0-dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go build flags
LDFLAGS := -ldflags "-X github.com/subosity/subosity-installer/shared/constants.AppVersion=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"
BUILD_FLAGS := -trimpath -mod=readonly

# Docker build configuration
DOCKER_BUILD_ARGS := --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT)

# Default target
all: clean lint test build build-container

# Help target
help: ## Show this help message
	@echo "Subosity Installer Build System"
	@echo "==============================="
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	rm -rf dist/
	go clean -cache -testcache
	docker rmi $(CONTAINER_IMAGE):dev 2>/dev/null || true
	docker rmi $(CONTAINER_IMAGE):latest 2>/dev/null || true

# Download dependencies
deps: ## Download and verify dependencies
	@echo "📦 Downloading dependencies..."
	go mod download
	go mod verify
	go mod tidy

# Run linting
lint: ## Run code linting
	@echo "🔍 Running linting..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "❌ golangci-lint not found. Install it with:"; \
		echo "   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi
	golangci-lint run --config .golangci.yml

# Run tests
test: ## Run tests
	@echo "🧪 Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

# Run tests with coverage report
test-coverage: test ## Run tests with HTML coverage report
	@echo "📊 Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build host binary
build-binary: ## Build the host binary
	@echo "🔨 Building host binary..."
	mkdir -p dist/
	CGO_ENABLED=0 go build $(BUILD_FLAGS) $(LDFLAGS) -o dist/$(BINARY_NAME) ./main.go

# Cross-compile binaries
build-all: ## Cross-compile for all supported platforms
	@echo "🌍 Cross-compiling for all platforms..."
	mkdir -p dist/
	
	# Linux x86_64
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) $(LDFLAGS) \
		-o dist/$(BINARY_NAME)-linux-amd64 ./main.go
	
	# Linux ARM64
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(BUILD_FLAGS) $(LDFLAGS) \
		-o dist/$(BINARY_NAME)-linux-arm64 ./main.go
	
	# Generate checksums
	cd dist && sha256sum $(BINARY_NAME)-* > checksums.sha256
	
	@echo "✅ Cross-compilation complete:"
	@ls -la dist/

# Build container image
build-container: ## Build the container image
	@echo "🐳 Building container image..."
	docker build $(DOCKER_BUILD_ARGS) -f container/Dockerfile -t $(CONTAINER_IMAGE):dev .
	docker tag $(CONTAINER_IMAGE):dev $(CONTAINER_IMAGE):$(VERSION)

# Build both binary and container
build: build-binary build-container ## Build both binary and container

# Run development version
run-dev: build-binary ## Run development version
	@echo "🚀 Running development version..."
	./dist/$(BINARY_NAME) --help

# Test the binary
test-binary: build-binary ## Test the binary installation
	@echo "🧪 Testing binary..."
	./dist/$(BINARY_NAME) version
	./dist/$(BINARY_NAME) setup --help

# Test the container
test-container: build-container ## Test the container
	@echo "🧪 Testing container..."
	docker run --rm $(CONTAINER_IMAGE):dev --help || true

# Development setup
dev-setup: ## Set up development environment
	@echo "🔧 Setting up development environment..."
	
	# Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	
	# Install pre-commit hooks
	@if [ -d .git ]; then \
		echo "Setting up pre-commit hooks..."; \
		echo "#!/bin/bash" > .git/hooks/pre-commit; \
		echo "make lint test" >> .git/hooks/pre-commit; \
		chmod +x .git/hooks/pre-commit; \
	fi
	
	@echo "✅ Development environment ready!"

# Development setup (alias for devcontainer compatibility)
setup-dev: dev-setup ## Alias for dev-setup (used by devcontainer)

# Format code
fmt: ## Format code
	@echo "📝 Formatting code..."
	go fmt ./...
	goimports -w .

# Security scan
security: ## Run security scan
	@echo "🔒 Running security scan..."
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	gosec ./...

# Vulnerability check
vuln-check: ## Check for vulnerabilities
	@echo "🛡️  Checking for vulnerabilities..."
	@if ! command -v govulncheck >/dev/null 2>&1; then \
		echo "Installing govulncheck..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
	fi
	govulncheck ./...

# Integration tests
test-integration: build build-container ## Run integration tests
	@echo "🔗 Running integration tests..."
	@echo "⚠️  Integration tests not yet implemented (Phase 1)"

# Package for release
package: build-all ## Package binaries for release
	@echo "📦 Packaging for release..."
	mkdir -p dist/packages/
	
	# Create tar.gz packages
	cd dist && tar -czf packages/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	cd dist && tar -czf packages/$(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64
	
	# Copy checksums
	cp dist/checksums.sha256 dist/packages/
	
	@echo "✅ Packages created in dist/packages/"

# Quick build and test cycle
quick: lint test build ## Quick build and test cycle

# Full CI pipeline
ci: clean deps lint test security vuln-check build test-binary test-container ## Full CI pipeline

# Show project info
info: ## Show project information
	@echo "Subosity Installer Build Information"
	@echo "==================================="
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(shell go version)"
	@echo "Platform:   $(shell go env GOOS)/$(shell go env GOARCH)"
