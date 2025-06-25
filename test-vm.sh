#!/bin/bash
# VM Test Script for Subosity Installer
# Run this on a clean Ubuntu/Debian VM to test the installer

set -e

echo "🧪 Subosity Installer VM Test Script"
echo "====================================="

# Check if we're on a supported OS
if [[ ! -f /etc/os-release ]]; then
    echo "❌ Cannot detect OS - /etc/os-release not found"
    exit 1
fi

source /etc/os-release
echo "📋 Detected OS: $PRETTY_NAME"

# Verify supported OS
if [[ "$ID" != "ubuntu" && "$ID" != "debian" ]]; then
    echo "⚠️  Warning: OS '$ID' is not officially supported (ubuntu/debian)"
    echo "   Continuing anyway for testing purposes..."
fi

# Update system
echo "🔄 Updating system packages..."
sudo apt-get update

# Check if installer is already built
if [[ ! -f "./dist/subosity-installer" ]]; then
    echo "❌ Installer binary not found at ./dist/subosity-installer"
    echo "   Run 'make build' in the dev container first"
    exit 1
fi

echo "✅ Found installer binary"

# Make sure it's executable
chmod +x ./dist/subosity-installer

# Show system info
echo ""
echo "🖥️  System Information:"
echo "   OS: $PRETTY_NAME"
echo "   Architecture: $(uname -m)"
echo "   Kernel: $(uname -r)"
echo "   Memory: $(free -h | grep '^Mem:' | awk '{print $2}')"
echo "   Disk: $(df -h / | tail -1 | awk '{print $4}') available"

# Check Docker status
echo ""
echo "🐳 Docker Status:"
if command -v docker >/dev/null 2>&1; then
    echo "   Docker is already installed: $(docker --version)"
    echo "   This will test the 'already installed' path"
else
    echo "   Docker not installed - this will test the installation path"
fi

echo ""
echo "🚀 Ready to test installer!"
echo ""
echo "Test commands you can run:"
echo "  ./dist/subosity-installer --help"
echo "  ./dist/subosity-installer version"
echo "  ./dist/subosity-installer status"
echo "  ./dist/subosity-installer setup --env dev --domain test.local --email test@example.com"
echo ""
echo "⚠️  Remember: This is a test VM - be prepared to restore/recreate if needed!"
