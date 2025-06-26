#!/usr/bin/env bash

# Post-setup script for Subosity Installer dev container
# Sets up Go development environment using pre-installed GVM

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
GVM_ROOT="$HOME/.gvm"

echo -e "${CYAN}[*] Starting Subosity Installer dev environment setup${RESET}"

# Start Docker daemon if not running
if ! docker info > /dev/null 2>&1; then
    echo -e "${CYAN}[*] Starting Docker daemon...${RESET}"
    sudo service docker start
    sleep 3
    echo -e "${GREEN}[+] Docker started successfully${RESET}"
else
    echo -e "${GREEN}[+] Docker already running${RESET}"
fi

# Set up locale
export LANG=C.UTF-8
export LC_ALL=C.UTF-8

# Load GVM (pre-installed in Dockerfile)
echo -e "${CYAN}[*] Loading Go Version Manager${RESET}"
# Temporarily relax error handling for GVM (it has some unbound variables)
set +euo pipefail
source "$GVM_ROOT/scripts/gvm" >/dev/null 2>&1 || true
set -euo pipefail

# Install specified Go version
echo -e "${CYAN}[*] Installing Go $GO_VERSION${RESET}"
set +euo pipefail

# Check if Go version is already installed
if gvm list 2>/dev/null | grep -q "go$GO_VERSION"; then
    echo -e "${GREEN}[+] Go $GO_VERSION already installed${RESET}"
else
    echo -e "${CYAN}[*] Installing Go $GO_VERSION...${RESET}"
    # Try binary installation first
    if gvm install "go$GO_VERSION" --binary >/dev/null 2>&1; then
        echo -e "${GREEN}[+] Go $GO_VERSION installed via binary${RESET}"
    else
        echo -e "${YELLOW}[!] Binary installation failed, trying source compilation...${RESET}"
        # Install bootstrap Go if needed
        if ! gvm list 2>/dev/null | grep -q "go1.20"; then
            gvm install go1.20 --binary >/dev/null 2>&1 || {
                echo -e "${RED}[-] Failed to install bootstrap Go 1.20${RESET}"
                exit 1
            }
        fi
        gvm use go1.20 >/dev/null 2>&1
        gvm install "go$GO_VERSION" >/dev/null 2>&1 || {
            echo -e "${RED}[-] Failed to install Go $GO_VERSION${RESET}"
            exit 1
        }
        echo -e "${GREEN}[+] Go $GO_VERSION installed via source${RESET}"
    fi
fi

# Set as default version
gvm use "go$GO_VERSION" --default >/dev/null 2>&1
set -euo pipefail

# Update environment for current session
export GOROOT="$GVM_ROOT/gos/go$GO_VERSION"
export GOPATH="/go"
export PATH="$GOROOT/bin:$GOPATH/bin:$PATH"

# Verify Go installation
echo -e "${CYAN}[*] Verifying Go installation${RESET}"
go_version=$(go version)
echo -e "${GREEN}[+] $go_version${RESET}"

# Install Go development tools
echo -e "${CYAN}[*] Installing Go development tools${RESET}"
mkdir -p "$GOPATH/bin"

tools=(
    "goimports|golang.org/x/tools/cmd/goimports@latest"
    "golangci-lint|github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    "gosec|github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"
    "govulncheck|golang.org/x/vuln/cmd/govulncheck@latest"
    "gofumpt|mvdan.cc/gofumpt@latest"
    "staticcheck|honnef.co/go/tools/cmd/staticcheck@latest"
)

for tool_info in "${tools[@]}"; do
    tool_name="${tool_info%|*}"
    tool_package="${tool_info#*|}"
    echo -e "${GRAY}    Installing $tool_name...${RESET}"
    if go install "$tool_package" >/dev/null 2>&1; then
        echo -e "${GREEN}[+] $tool_name installed${RESET}"
    else
        echo -e "${YELLOW}[!] Failed to install $tool_name${RESET}"
    fi
done

# Set up bash environment for future shells
echo -e "${CYAN}[*] Setting up bash environment${RESET}"

# Add environment setup to .bashrc (append, don't overwrite)
if ! grep -q "Subosity Installer Development Environment" "$HOME/.bashrc" 2>/dev/null; then
    cat >> "$HOME/.bashrc" << 'EOF'

# === Subosity Installer Development Environment ===
export LANG=C.UTF-8
export LC_ALL=C.UTF-8

# Load GVM
if [[ -s "$HOME/.gvm/scripts/gvm" ]]; then
    source "$HOME/.gvm/scripts/gvm" 2>/dev/null || true
    gvm use go1.23.4 &>/dev/null || true
fi

# Development aliases
alias ll="ls -la"
alias go-build="go build -v ./..."
alias go-test="go test -v ./..."
alias go-lint="golangci-lint run"
alias go-fmt="gofmt -s -w . && goimports -w ."

# Installer aliases
alias installer-build="make build"
alias installer-test="make test"
alias installer-container="make build-container"

# Colorful prompt
PS1='\[\033[38;5;39m\]\u\[\033[0m\]@\[\033[38;5;42m\]\h\[\033[0m\] \[\033[38;5;244m\]\w\[\033[0m\]\n\$ '
EOF
fi

# Set correct permissions and download dependencies
sudo chown -R vscode:vscode "$GOPATH" 2>/dev/null || true

if [[ -f "/workspaces/subosity-installer/go.mod" ]]; then
    echo -e "${GREEN}[+] Downloading Go dependencies${RESET}"
    # Temporarily disable strict error checking for GVM's cd function
    set +euo pipefail
    cd /workspaces/subosity-installer
    set -euo pipefail
    go mod download
fi

# Final verification and status report
echo -e "${CYAN}[*] Performing final verification checks${RESET}"

# Check Go environment
echo -e "${CYAN}[*] Verifying Go environment${RESET}"
if command -v go >/dev/null 2>&1; then
    go_version=$(go version 2>/dev/null || echo "unknown")
    echo -e "${GREEN}[+] Go: $go_version${RESET}"
    echo -e "${GRAY}    GOROOT: $GOROOT${RESET}"
    echo -e "${GRAY}    GOPATH: $GOPATH${RESET}"
else
    echo -e "${RED}[-] Go: not found in PATH${RESET}"
fi

# Check GVM
echo -e "${CYAN}[*] Verifying GVM environment${RESET}"
if [[ -s "$GVM_ROOT/scripts/gvm" ]]; then
    echo -e "${GREEN}[+] GVM: installed at $GVM_ROOT${RESET}"
    # Temporarily relax error handling for GVM list
    set +euo pipefail
    installed_versions=$(gvm list 2>/dev/null | grep -c "go[0-9]" || echo "0")
    set -euo pipefail
    echo -e "${GRAY}    Installed Go versions: $installed_versions${RESET}"
else
    echo -e "${RED}[-] GVM: not found${RESET}"
fi

# Check development tools
echo -e "${CYAN}[*] Verifying development tools${RESET}"
dev_tools=("goimports" "golangci-lint" "govulncheck" "gofumpt" "staticcheck")
for tool in "${dev_tools[@]}"; do
    if command -v "$tool" >/dev/null 2>&1; then
        tool_path=$(which "$tool" 2>/dev/null || echo "unknown")
        echo -e "${GREEN}[+] $tool: ${GRAY}$tool_path${RESET}"
    else
        echo -e "${YELLOW}[!] $tool: not found${RESET}"
    fi
done

# Check Docker
echo -e "${CYAN}[*] Verifying Docker environment${RESET}"
if command -v docker >/dev/null 2>&1; then
    if docker info >/dev/null 2>&1; then
        docker_version=$(docker --version 2>/dev/null || echo "unknown")
        echo -e "${GREEN}[+] Docker: $docker_version (daemon running)${RESET}"
    else
        echo -e "${YELLOW}[!] Docker: installed but daemon not accessible${RESET}"
    fi
else
    echo -e "${RED}[-] Docker: not found${RESET}"
fi

# Check Git configuration
echo -e "${CYAN}[*] Verifying Git configuration${RESET}"
if command -v git >/dev/null 2>&1; then
    git_version=$(git --version 2>/dev/null || echo "unknown")
    git_user=$(git config --global user.name 2>/dev/null || echo "not set")
    git_email=$(git config --global user.email 2>/dev/null || echo "not set")
    echo -e "${GREEN}[+] Git: $git_version${RESET}"
    echo -e "${GRAY}    User: $git_user${RESET}"
    echo -e "${GRAY}    Email: $git_email${RESET}"
else
    echo -e "${RED}[-] Git: not found${RESET}"
fi

# Check if aliases are working
echo -e "${CYAN}[*] Verifying development aliases${RESET}"
if grep -q "go-build" "$HOME/.bashrc" 2>/dev/null; then
    echo -e "${GREEN}[+] Development aliases: configured in .bashrc${RESET}"
    echo -e "${GRAY}    Available: go-build, go-test, go-lint, go-fmt${RESET}"
    echo -e "${GRAY}    Available: installer-build, installer-test, installer-container${RESET}"
else
    echo -e "${YELLOW}[!] Development aliases: not found in .bashrc${RESET}"
fi

# Test a simple Go command
echo -e "${CYAN}[*] Testing Go functionality${RESET}"
if command -v go >/dev/null 2>&1; then
    # Test go env command
    if go env GOROOT >/dev/null 2>&1; then
        goroot_test=$(go env GOROOT 2>/dev/null || echo "unknown")
        gopath_test=$(go env GOPATH 2>/dev/null || echo "unknown")
        echo -e "${GREEN}[+] Go commands: working correctly${RESET}"
        echo -e "${GRAY}    go env GOROOT: $goroot_test${RESET}"
        echo -e "${GRAY}    go env GOPATH: $gopath_test${RESET}"
    else
        echo -e "${YELLOW}[!] Go commands: installed but not responding${RESET}"
    fi
else
    echo -e "${RED}[-] Go commands: cannot test (go not found)${RESET}"
fi

# Summary
echo
echo -e "${GREEN}[+] Dev environment setup complete!${RESET}"
echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════════════════╗${RESET}"
echo -e "${BLUE}║${RESET}                       ${CYAN}🚀 Development Environment Ready${RESET}                        ${BLUE}║${RESET}"
echo -e "${BLUE}╠══════════════════════════════════════════════════════════════════════════════╣${RESET}"
echo -e "${BLUE}║${RESET} ${GRAY}• Start a new terminal to use the configured environment${RESET}               ${BLUE}║${RESET}"
echo -e "${BLUE}║${RESET} ${GRAY}• Use 'source ~/.bashrc' to load environment in current terminal${RESET}       ${BLUE}║${RESET}"
echo -e "${BLUE}║${RESET} ${GRAY}• Run 'make help' to see available build targets${RESET}                      ${BLUE}║${RESET}"
echo -e "${BLUE}║${RESET} ${GRAY}• Use development aliases: go-build, go-test, go-lint, etc.${RESET}            ${BLUE}║${RESET}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════════════════╝${RESET}"
echo
