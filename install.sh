#!/bin/bash

# Secrets Snapshot CLI Installer
# This script downloads and installs the latest version of secretsnap

set -e

BINARY="secretsnap"
REPO="secretsnap/cli"  # Update this to your actual repo
INSTALL_DIR="/usr/local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Detect OS and architecture
detect_platform() {
    case "$(uname -s)" in
        Linux*)     OS="linux";;
        Darwin*)    OS="darwin";;
        *)          echo -e "${RED}❌ Unsupported operating system${NC}"; exit 1;;
    esac
    
    case "$(uname -m)" in
        x86_64)     ARCH="amd64";;
        arm64|aarch64) ARCH="arm64";;
        *)          echo -e "${RED}❌ Unsupported architecture${NC}"; exit 1;;
    esac
    
    echo -e "${BLUE}📋 Detected platform: $OS/$ARCH${NC}"
}

# Get latest version from GitHub
get_latest_version() {
    echo -e "${BLUE}🔍 Checking for latest version...${NC}"
    
    # For now, use a hardcoded version since we don't have releases yet
    VERSION="v0.1.0"
    
    echo -e "${GREEN}✅ Latest version: $VERSION${NC}"
}

# Download and install
install_binary() {
    local version=$1
    
    # Create temporary directory
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    echo -e "${BLUE}📦 Downloading secretsnap $version...${NC}"
    
    # For now, we'll build from source since we don't have releases
    echo -e "${YELLOW}⚠️  Building from source (no releases available yet)${NC}"
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go is not installed. Please install Go 1.22+ first.${NC}"
        echo -e "${BLUE}📖 Visit: https://golang.org/doc/install${NC}"
        exit 1
    fi
    
    # Clone and build
    echo -e "${BLUE}🔨 Building secretsnap...${NC}"
    git clone https://github.com/$REPO.git
    cd cli
    make build
    
    # Install binary
    echo -e "${BLUE}📦 Installing to $INSTALL_DIR...${NC}"
    sudo cp bin/$BINARY "$INSTALL_DIR/"
    sudo chmod +x "$INSTALL_DIR/$BINARY"
    
    # Cleanup
    cd /
    rm -rf "$TEMP_DIR"
    
    echo -e "${GREEN}✅ secretsnap installed successfully!${NC}"
}

# Verify installation
verify_installation() {
    if command -v $BINARY &> /dev/null; then
        echo -e "${GREEN}✅ Installation verified!${NC}"
        echo -e "${BLUE}📖 Run '$BINARY --help' to get started${NC}"
    else
        echo -e "${RED}❌ Installation failed${NC}"
        exit 1
    fi
}

# Main installation flow
main() {
    echo -e "${BLUE}🚀 Installing Secrets Snapshot CLI...${NC}"
    echo ""
    
    detect_platform
    get_latest_version
    install_binary "$VERSION"
    verify_installation
    
    echo ""
    echo -e "${GREEN}🎉 Installation complete!${NC}"
    echo -e "${BLUE}📖 Quick start:${NC}"
    echo -e "   $BINARY init"
    echo -e "   $BINARY bundle .env --out secrets.envsnap"
    echo -e "   $BINARY unbundle secrets.envsnap --out .env"
}

# Run main function
main "$@"
