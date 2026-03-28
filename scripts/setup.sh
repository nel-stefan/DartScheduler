#!/usr/bin/env bash
# setup.sh — install all tools required to develop and build DartScheduler
# Supports: macOS (Homebrew) and Debian/Ubuntu Linux
set -euo pipefail

###############################################################################
# Helpers
###############################################################################
info()    { echo "[INFO]  $*"; }
ok()      { echo "[OK]    $*"; }
warn()    { echo "[WARN]  $*"; }
die()     { echo "[ERROR] $*" >&2; exit 1; }

have() { command -v "$1" &>/dev/null; }

###############################################################################
# Detect OS / package manager
###############################################################################
OS="$(uname -s)"

if [[ "$OS" == "Darwin" ]]; then
    PM="brew"
    have brew || die "Homebrew is required on macOS. Install it from https://brew.sh and re-run."
elif [[ "$OS" == "Linux" ]]; then
    if have apt-get; then
        PM="apt"
    else
        die "Unsupported Linux distribution. Only Debian/Ubuntu (apt) is supported."
    fi
else
    die "Unsupported OS: $OS"
fi

###############################################################################
# Required versions (loosely checked)
###############################################################################
GO_MIN="1.22"   # go.mod declares go 1.26; toolchain 1.22+ satisfies the module graph
NODE_MIN="20"

###############################################################################
# 1. Go
###############################################################################
install_go() {
    if [[ "$PM" == "brew" ]]; then
        brew install go
    else
        # Install via the official binary tarball so we can control the version
        GO_VERSION="1.24.1"
        ARCH="$(uname -m)"
        [[ "$ARCH" == "x86_64" ]] && GOARCH="amd64" || GOARCH="arm64"
        TARBALL="go${GO_VERSION}.linux-${GOARCH}.tar.gz"
        info "Downloading Go ${GO_VERSION} …"
        curl -fsSL "https://go.dev/dl/${TARBALL}" -o "/tmp/${TARBALL}"
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf "/tmp/${TARBALL}"
        rm "/tmp/${TARBALL}"
        # Ensure PATH is updated for this session
        export PATH="/usr/local/go/bin:$PATH"
        info "Add the following line to your shell profile if not already present:"
        info '  export PATH="/usr/local/go/bin:$PATH"'
    fi
}

if have go; then
    CURRENT_GO="$(go version | awk '{print $3}' | sed 's/go//')"
    info "Go $CURRENT_GO already installed."
    # Loose version check: compare major.minor only
    CURRENT_MAJOR_MINOR="$(echo "$CURRENT_GO" | cut -d. -f1,2)"
    MIN_MAJOR_MINOR="$(echo "$GO_MIN" | cut -d. -f1,2)"
    if [[ "$(printf '%s\n%s' "$MIN_MAJOR_MINOR" "$CURRENT_MAJOR_MINOR" | sort -V | head -1)" != "$MIN_MAJOR_MINOR" ]]; then
        warn "Go $CURRENT_GO is older than $GO_MIN — upgrading."
        install_go
    else
        ok "Go version OK."
    fi
else
    info "Go not found — installing."
    install_go
fi

###############################################################################
# 2. Node.js (via nvm or package manager)
###############################################################################
install_node() {
    if [[ "$PM" == "brew" ]]; then
        # Use node@20 to pin the LTS version used in the Dockerfile
        brew install node@20
        brew link --overwrite --force node@20 || true
    else
        # Use NodeSource setup script
        info "Installing Node.js $NODE_MIN via NodeSource …"
        curl -fsSL "https://deb.nodesource.com/setup_${NODE_MIN}.x" | sudo -E bash -
        sudo apt-get install -y nodejs
    fi
}

if have node; then
    CURRENT_NODE_MAJOR="$(node --version | sed 's/v//' | cut -d. -f1)"
    info "Node.js v$(node --version | sed 's/v//') already installed."
    if (( CURRENT_NODE_MAJOR < NODE_MIN )); then
        warn "Node.js major version $CURRENT_NODE_MAJOR is older than required $NODE_MIN — upgrading."
        install_node
    else
        ok "Node.js version OK."
    fi
else
    info "Node.js not found — installing."
    install_node
fi

###############################################################################
# 3. npm (bundled with Node; just verify)
###############################################################################
have npm || die "npm not found after Node.js installation — check your PATH."
ok "npm $(npm --version) available."

###############################################################################
# 4. Angular CLI (@angular/cli 17)
###############################################################################
if have ng && ng version 2>/dev/null | grep -q "Angular CLI"; then
    NG_VER="$(ng version 2>/dev/null | grep 'Angular CLI' | awk '{print $NF}')"
    info "Angular CLI $NG_VER already installed."
    ok "Angular CLI OK."
else
    info "Angular CLI not found — installing globally."
    npm install -g @angular/cli@17
fi

###############################################################################
# 5. make
###############################################################################
if have make; then
    ok "make available."
else
    info "make not found — installing."
    if [[ "$PM" == "brew" ]]; then
        brew install make
    else
        sudo apt-get install -y make
    fi
fi

###############################################################################
# 6. gcc / C toolchain (required for CGO in go-sqlite3)
###############################################################################
if have gcc; then
    ok "gcc available (CGO support)."
else
    info "gcc not found — installing C toolchain for CGO."
    if [[ "$PM" == "brew" ]]; then
        # Xcode Command Line Tools provide gcc on macOS
        xcode-select --install 2>/dev/null || true
        info "If prompted, complete the Xcode Command Line Tools installation and re-run this script."
    else
        sudo apt-get install -y build-essential
    fi
fi

###############################################################################
# 7. Docker (optional — needed for 'make docker')
###############################################################################
if have docker; then
    ok "Docker $(docker --version | awk '{print $3}' | tr -d ',') available."
else
    warn "Docker not found. It is optional but required for 'make docker'."
    if [[ "$PM" == "brew" ]]; then
        info "Install Docker Desktop from https://www.docker.com/products/docker-desktop/"
    else
        info "Install Docker by following https://docs.docker.com/engine/install/"
    fi
fi

###############################################################################
# 8. Go module dependencies
###############################################################################
info "Downloading Go module dependencies (go mod download) …"
(cd "$(dirname "$0")/.." && go mod download)
ok "Go modules ready."

###############################################################################
# 9. Frontend npm dependencies
###############################################################################
info "Installing frontend npm dependencies (npm ci) …"
(cd "$(dirname "$0")/../frontend" && npm ci)
ok "Frontend dependencies ready."

###############################################################################
# Done
###############################################################################
echo ""
echo "========================================"
echo "  Setup complete. Quick-start commands:"
echo "========================================"
echo "  make dev        — start Go backend (port 8080)"
echo "  cd frontend && npm start"
echo "                  — Angular dev server (port 4200)"
echo "  make frontend   — production frontend build"
echo "  make test       — run Go tests"
echo "  make docker     — build & run via Docker Compose"
echo "========================================"
