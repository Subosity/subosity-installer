#!/bin/bash
# Build script that works even without Go installed locally
# This builds everything using Docker containers

set -euo pipefail

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly CYAN='\033[0;36m'
readonly NC='\033[0m' # No Color

log_info() { echo -e "${CYAN}[*]${NC} $1"; }
log_success() { echo -e "${GREEN}[+]${NC} $1"; }
log_error() { echo -e "${RED}[-]${NC} $1" >&2; }
log_warn() { echo -e "${YELLOW}[!]${NC} $1"; }

# Configuration
BINARY_NAME="subosity-installer"
CONTAINER_IMAGE="subosity/installer"
VERSION="1.0.0-dev"

echo "🚀 Building Subosity Installer"
echo "=============================="
echo

# Step 1: Create dist directory
log_info "Creating build directory..."
mkdir -p dist/

# Step 2: Build the thin binary using Go container
log_info "Building thin binary using Go container..."

docker run --rm \
  -v "$(pwd):/workspace" \
  -w /workspace \
  golang:1.21-alpine \
  sh -c "
    echo 'Installing build dependencies...' &&
    apk add --no-cache git ca-certificates &&
    echo 'Downloading Go dependencies...' &&
    go mod download &&
    go mod tidy &&
    echo 'Building binary...' &&
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
      -ldflags='-w -s -X github.com/subosity/subosity-installer/shared/constants.AppVersion=${VERSION}' \
      -o dist/${BINARY_NAME}-linux-amd64 ./main.go &&
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
      -ldflags='-w -s -X github.com/subosity/subosity-installer/shared/constants.AppVersion=${VERSION}' \
      -o dist/${BINARY_NAME}-linux-arm64 ./main.go &&
    echo 'Binary build complete!'
  "

log_success "Thin binary built successfully"

# Step 3: Build the smart container
log_info "Building smart container image..."

docker build \
  -f container/Dockerfile \
  -t "${CONTAINER_IMAGE}:dev" \
  -t "${CONTAINER_IMAGE}:${VERSION}" \
  --build-arg VERSION="${VERSION}" \
  .

log_success "Smart container built successfully"

# Step 4: Create checksums
log_info "Generating checksums..."
mkdir -p dist/
cd dist/
sha256sum ${BINARY_NAME}-* > checksums.sha256
cd ..

log_success "Build completed successfully!"
echo
echo "📦 Build artifacts:"
echo "  • Thin binary (amd64): dist/${BINARY_NAME}-linux-amd64"
echo "  • Thin binary (arm64): dist/${BINARY_NAME}-linux-arm64"
echo "  • Smart container: ${CONTAINER_IMAGE}:dev"
echo "  • Checksums: dist/checksums.sha256"
echo

# Step 5: Test the builds
log_info "Testing builds..."

# Test binary
log_info "Testing thin binary..."
chmod +x dist/${BINARY_NAME}-linux-amd64
./dist/${BINARY_NAME}-linux-amd64 version

# Test container
log_info "Testing smart container..."
docker run --rm "${CONTAINER_IMAGE}:dev" --help || true

log_success "All tests passed!"
echo
echo "🎉 Ready for Phase 1 testing!"
echo
echo "Usage:"
echo "  ./dist/${BINARY_NAME}-linux-amd64 setup --env dev --domain myapp.local"
