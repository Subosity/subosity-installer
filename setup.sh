#!/bin/bash
# Subosity Installer Setup Script
# This is the entry point that users run to install Subosity
# Usage: curl -fsSL https://install.subosity.com | bash
# Or: curl -fsSL https://install.subosity.com | bash -s -- --env prod --domain myapp.com

set -euo pipefail

# Configuration
readonly GITHUB_REPO="subosity/subosity-installer"
readonly BINARY_NAME="subosity-installer"
readonly INSTALL_DIR="/usr/local/bin"
readonly TEMP_DIR="/tmp/subosity-install-$$"

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly CYAN='\033[0;36m'
readonly NC='\033[0m' # No Color

# Logging functions following PRD standards
log_info() {
    echo -e "${CYAN}[*]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[+]${NC} $1"
}

log_error() {
    echo -e "${RED}[-]${NC} $1" >&2
}

log_warn() {
    echo -e "${YELLOW}[!]${NC} $1"
}

# Error handling
error_exit() {
    log_error "$1"
    cleanup
    exit "${2:-1}"
}

# Cleanup function
cleanup() {
    if [[ -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
}

# Set up trap for cleanup
trap cleanup EXIT

# Detect architecture
detect_arch() {
    local arch
    arch=$(uname -m)
    case $arch in
        x86_64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        *)
            error_exit "Unsupported architecture: $arch" 2
            ;;
    esac
}

# Detect OS
detect_os() {
    if [[ ! -f /etc/os-release ]]; then
        error_exit "Cannot detect OS - /etc/os-release not found" 2
    fi
    
    source /etc/os-release
    case $ID in
        ubuntu|debian)
            log_info "Detected supported OS: $PRETTY_NAME"
            ;;
        *)
            error_exit "Unsupported OS: $ID (supported: ubuntu, debian)" 2
            ;;
    esac
}

# Check if running as root
check_root() {
    if [[ $EUID -eq 0 ]]; then
        log_warn "Running as root - this is not recommended for security"
        log_warn "Consider running as a regular user with sudo access"
    fi
}

# Check system requirements
check_requirements() {
    log_info "Checking system requirements..."
    
    # Check required commands
    local required_cmds=("curl" "wget" "sudo")
    for cmd in "${required_cmds[@]}"; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            error_exit "Required command not found: $cmd" 2
        fi
    done
    
    # Check memory (minimum 2GB)
    local mem_kb
    mem_kb=$(grep MemTotal /proc/meminfo | awk '{print $2}')
    local mem_gb=$((mem_kb / 1024 / 1024))
    if [[ $mem_gb -lt 2 ]]; then
        error_exit "Insufficient memory: ${mem_gb}GB (minimum 2GB required)" 2
    fi
    
    # Check disk space (minimum 10GB)
    local disk_gb
    disk_gb=$(df / | tail -1 | awk '{print int($4/1024/1024)}')
    if [[ $disk_gb -lt 10 ]]; then
        error_exit "Insufficient disk space: ${disk_gb}GB (minimum 10GB required)" 2
    fi
    
    log_success "System requirements check passed"
}

# Download and verify installer binary
download_installer() {
    local arch
    arch=$(detect_arch)
    
    log_info "Downloading Subosity installer for linux-${arch}..."
    
    # Create temp directory
    mkdir -p "$TEMP_DIR"
    cd "$TEMP_DIR"
    
    # Get latest release URL
    local download_url="https://github.com/${GITHUB_REPO}/releases/latest/download/${BINARY_NAME}-linux-${arch}"
    local checksum_url="https://github.com/${GITHUB_REPO}/releases/latest/download/${BINARY_NAME}-linux-${arch}.sha256"
    
    # Download binary
    if ! curl -fsSL "$download_url" -o "$BINARY_NAME"; then
        error_exit "Failed to download installer binary" 3
    fi
    
    # Download and verify checksum
    if ! curl -fsSL "$checksum_url" -o "${BINARY_NAME}.sha256"; then
        log_warn "Could not download checksum file - skipping verification"
    else
        if ! sha256sum -c "${BINARY_NAME}.sha256" >/dev/null 2>&1; then
            error_exit "Checksum verification failed - binary may be corrupted" 3
        fi
        log_success "Binary checksum verified"
    fi
    
    # Make executable
    chmod +x "$BINARY_NAME"
    
    log_success "Installer binary downloaded successfully"
}

# Install binary to system
install_binary() {
    log_info "Installing binary to $INSTALL_DIR..."
    
    # Check if we can write to install dir
    if [[ ! -w "$INSTALL_DIR" ]]; then
        if ! sudo cp "$TEMP_DIR/$BINARY_NAME" "$INSTALL_DIR/"; then
            error_exit "Failed to install binary to $INSTALL_DIR" 4
        fi
    else
        if ! cp "$TEMP_DIR/$BINARY_NAME" "$INSTALL_DIR/"; then
            error_exit "Failed to install binary to $INSTALL_DIR" 4
        fi
    fi
    
    log_success "Binary installed to $INSTALL_DIR/$BINARY_NAME"
}

# Run the installer with provided arguments
run_installer() {
    log_info "Running Subosity installer..."
    
    # If we have arguments, pass them to the installer
    if [[ $# -gt 0 ]]; then
        exec "$INSTALL_DIR/$BINARY_NAME" setup "$@"
    else
        # Interactive mode
        log_info "Starting interactive installation..."
        exec "$INSTALL_DIR/$BINARY_NAME" setup --interactive
    fi
}

# Main function
main() {
    echo "🚀 Subosity Installer Setup"
    echo "=========================="
    echo ""
    
    # Perform checks
    detect_os
    check_root
    check_requirements
    
    # Download and install
    download_installer
    install_binary
    
    # Run the installer
    run_installer "$@"
}

# Handle script being piped from curl
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
