#!/usr/bin/env bash

# Post-setup script for Subosity Installer dev container
# Installs specified Go version using gvm and sets up development tools

set -euo pipefail

# --- Colors ---
CYAN='\033[36m'
GREEN='\033[32m'
RED='\033[31m'
YELLOW='\033[33m'
GRAY='\033[90m'
BLUE='\033[34m'
RESET='\033[0m'

# --- Configuration ---
GO_VERSION="${GO_VERSION:-1.23.4}"
GVM_ROOT="${GVM_ROOT:-$HOME/.gvm}"

echo -e "${CYAN}[*] Starting Subosity Installer dev environment setup${RESET}"

# Check if running in dev container
if [[ "${SUBOSITY_DEV_MODE:-false}" != "true" ]]; then
    echo -e "${YELLOW}[!] Not running in dev container mode - skipping setup${RESET}"
    exit 0
fi

# --- Start Docker Daemon (DinD setup) ---
echo -e "${GREEN}[+] Starting Docker daemon${RESET}"
if ! pgrep dockerd > /dev/null; then
  sudo nohup dockerd > /tmp/dockerd.log 2>&1 &
  sleep 5
  echo -e "${BLUE}[i] Docker started successfully${RESET}"
else
  echo -e "${BLUE}[i] Docker already running${RESET}"
fi

# Source gvm if available
if [[ -f "$GVM_ROOT/scripts/gvm" ]]; then
    echo -e "${CYAN}[*] Loading Go Version Manager${RESET}"
    source "$GVM_ROOT/scripts/gvm"
else
    echo -e "${RED}[-] Go Version Manager not found at $GVM_ROOT${RESET}"
    echo -e "${YELLOW}[!] Attempting to install gvm...${RESET}"
    
    # Install gvm if not present
    curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer | bash
    source "$GVM_ROOT/scripts/gvm"
fi

# Install specified Go version
echo -e "${CYAN}[*] Installing Go $GO_VERSION using gvm${RESET}"

# Check if Go version is already installed
if gvm list | grep -q "go$GO_VERSION"; then
    echo -e "${GREEN}[+] Go $GO_VERSION already installed${RESET}"
else
    echo -e "${CYAN}[*] Installing Go $GO_VERSION...${RESET}"
    if ! gvm install "go$GO_VERSION" --binary; then
        echo -e "${YELLOW}[!] Binary installation failed, trying source compilation...${RESET}"
        # Install Go 1.20 first as bootstrap for newer versions
        if ! gvm list | grep -q "go1.20"; then
            gvm install go1.20 --binary
        fi
        gvm use go1.20
        gvm install "go$GO_VERSION"
    fi
fi

# Use the specified Go version as default
echo -e "${CYAN}[*] Setting Go $GO_VERSION as default${RESET}"
gvm use "go$GO_VERSION" --default

# Update environment variables for current session
export GOROOT="$GVM_ROOT/gos/go$GO_VERSION"
export GOPATH="/go"
export PATH="$GOROOT/bin:$GOPATH/bin:$PATH"

# Verify Go installation
echo -e "${CYAN}[*] Verifying Go installation${RESET}"
go_version=$(go version)
echo -e "${GRAY}    $go_version${RESET}"

# Install Go development tools with proper versions
echo -e "${CYAN}[*] Installing Go development tools${RESET}"

# Create GOPATH bin directory if it doesn't exist
mkdir -p "$GOPATH/bin"

# Install tools with error handling
install_go_tool() {
    local tool="$1"
    local package="$2"
    
    echo -e "${GRAY}    Installing $tool...${RESET}"
    if go install "$package"; then
        echo -e "${GREEN}[+] Successfully installed $tool${RESET}"
    else
        echo -e "${RED}[-] Failed to install $tool${RESET}"
        return 1
    fi
}

# Install core development tools
install_go_tool "goimports" "golang.org/x/tools/cmd/goimports@latest"
install_go_tool "golangci-lint" "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
install_go_tool "gosec" "github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"
install_go_tool "govulncheck" "golang.org/x/vuln/cmd/govulncheck@latest"

# Additional useful tools for the project
install_go_tool "gofumpt" "mvdan.cc/gofumpt@latest"
install_go_tool "staticcheck" "honnef.co/go/tools/cmd/staticcheck@latest"

# Set up gvm environment for future shells
echo -e "${CYAN}[*] Setting up gvm environment for future shells${RESET}"

# Add comprehensive environment setup to .bashrc
cat >> "$HOME/.bashrc" << 'EOF'

# === Go Version Manager ===
if [[ -s "$HOME/.gvm/scripts/gvm" ]]; then
    source "$HOME/.gvm/scripts/gvm"
    gvm use go1.23.4 &>/dev/null || true
fi

# === Development Aliases ===
alias ll="ls -la"
alias supa-start="supabase start"
alias supa-stop="supabase stop"
alias supa-db-reset="supabase db reset"
alias supa-db-push="supabase db push"
alias supa-functions-serve="supabase functions serve --env-file supabase/.env"

# === Go Development Aliases ===
alias go-build="go build -v ./..."
alias go-test="go test -v ./..."
alias go-test-coverage="go test -v -coverprofile=coverage.out ./... && go tool cover -html=coverage.out"
alias go-lint="golangci-lint run"
alias go-fmt="gofmt -s -w . && goimports -w ."
alias go-tidy="go mod tidy"

# === Installer Development Aliases ===
alias installer-build="make build"
alias installer-test="make test"
alias installer-run="make run"
alias installer-container="make build-container"

show_help() {
  echo
  echo    "╔══════════════════════════════════════════════════════════════════════════════╗"
  echo -e "║                          🚀 \033[36mSubosity Installer Dev Commands\033[0m                  ║"
  echo    "╠══════════════════════════════════════════════════════════════════════════════╣"
  echo -e "║  \033[32minstaller-build\033[0m      Build the installer binary                              ║"
  echo -e "║  \033[32minstaller-test\033[0m       Run all tests with coverage                            ║"
  echo -e "║  \033[32minstaller-run\033[0m        Run the installer for testing                          ║"
  echo -e "║  \033[32minstaller-container\033[0m  Build the smart container image                        ║"
  echo    "║                                                                              ║"
  echo -e "║  \033[32mgo-build\033[0m             Build all Go packages                                  ║"
  echo -e "║  \033[32mgo-test\033[0m              Run all tests                                          ║"
  echo -e "║  \033[32mgo-test-coverage\033[0m     Run tests with coverage report                         ║"
  echo -e "║  \033[32mgo-lint\033[0m              Run golangci-lint                                      ║"
  echo -e "║  \033[32mgo-fmt\033[0m               Format and organize imports                            ║"
  echo    "║                                                                              ║"
  echo -e "║  \033[32msupa-start\033[0m           Start local Supabase stack                             ║"
  echo -e "║  \033[32msupa-stop\033[0m            Stop local Supabase stack                              ║"
  echo -e "║  \033[32msupa-db-reset\033[0m        Reset and reapply all migrations                       ║"
  echo -e "║  \033[32msupa-db-push\033[0m         Push schema changes to local DB                        ║"
  echo    "║                                                                              ║"
  echo -e "║  \033[33mTip:\033[0m Use \033[32mdev-help\033[0m to show this message again                               ║"
  echo    "╚══════════════════════════════════════════════════════════════════════════════╝"
  echo
}

alias dev-help="show_help"

# === Colorful Prompt ===
PS1='\[\033[38;5;39m\]\u\[\033[0m\]@\[\033[38;5;42m\]\h\[\033[0m\] \[\033[38;5;244m\]\w\[\033[0m\]\n\$ '

# Show help on interactive terminals
if [[ $- == *i* ]]; then
    show_help
fi
EOF

# Verify tool installations
echo -e "${CYAN}[*] Verifying installed tools${RESET}"
tools=("goimports" "golangci-lint" "gosec" "govulncheck" "gofumpt" "staticcheck")

for tool in "${tools[@]}"; do
    if command -v "$tool" &>/dev/null; then
        version=$(eval "$tool --version 2>/dev/null || $tool version 2>/dev/null || echo 'installed'")
        echo -e "${GREEN}[+] $tool: ${GRAY}$version${RESET}"
    else
        echo -e "${RED}[-] $tool: not found${RESET}"
    fi
done

# Set correct permissions for Go workspace
echo -e "${CYAN}[*] Setting up Go workspace permissions${RESET}"
sudo chown -R vscode:vscode "$GOPATH" 2>/dev/null || true
chmod -R 755 "$GOPATH" 2>/dev/null || true

# Download Go dependencies if go.mod exists
if [[ -f "/workspaces/subosity-installer/go.mod" ]]; then
    echo -e "${GREEN}[+] Downloading Go dependencies${RESET}"
    cd /workspaces/subosity-installer
    go mod download
    echo -e "${BLUE}[i] Go dependencies downloaded${RESET}"
fi

echo -e "${GREEN}[+] Dev environment setup complete!${RESET}"
echo -e "${CYAN}[*] Go version: $(go version)${RESET}"
echo -e "${CYAN}[*] GOROOT: $GOROOT${RESET}"
echo -e "${CYAN}[*] GOPATH: $GOPATH${RESET}"

# Display available make targets for development
echo -e "${CYAN}[*] Available make targets:${RESET}"
if [[ -f "/workspaces/subosity-installer/Makefile" ]]; then
    make -f "/workspaces/subosity-installer/Makefile" help 2>/dev/null || \
    echo -e "${GRAY}    Run 'make' to see available targets${RESET}"
else
    echo -e "${GRAY}    Makefile not found - you may need to create build targets${RESET}"
fi
