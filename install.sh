#!/bin/bash

# Secrets Snapshot CLI Installer
# This script downloads and installs the latest version of secretsnap

set -e

BINARY="secretsnap"
REPO="secret-snap/cli"
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
    
    # Try to get latest release from GitHub API
    if command -v curl &> /dev/null; then
        VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    fi
    
    # Fallback to hardcoded version if API call fails
    if [ -z "$VERSION" ] || [ "$VERSION" = "null" ]; then
        echo -e "${YELLOW}⚠️  Could not fetch latest version, using fallback${NC}"
        VERSION="v1.0.0"
    fi
    
    echo -e "${GREEN}✅ Latest version: $VERSION${NC}"
}

# Download and install
install_binary() {
    local version=$1
    
    # Create temporary directory
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    echo -e "${BLUE}📦 Downloading secretsnap $version...${NC}"
    
    # Construct download URL
    ARCHIVE_NAME="secretsnap-$version-$OS-$ARCH.tar.gz"
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$version/$ARCHIVE_NAME"
    
    # Download the release
    if command -v curl &> /dev/null; then
        curl -sL "$DOWNLOAD_URL" -o "$ARCHIVE_NAME"
    elif command -v wget &> /dev/null; then
        wget -q "$DOWNLOAD_URL" -O "$ARCHIVE_NAME"
    else
        echo -e "${RED}❌ curl or wget is required${NC}"
        exit 1
    fi
    
    # Check if download was successful
    if [ ! -f "$ARCHIVE_NAME" ]; then
        echo -e "${RED}❌ Failed to download $ARCHIVE_NAME${NC}"
        echo -e "${YELLOW}💡 Falling back to source build...${NC}"
        install_from_source "$version"
        return
    fi
    
    # Extract the archive
    echo -e "${BLUE}📦 Extracting archive...${NC}"
    tar -xzf "$ARCHIVE_NAME"
    
    # Install binary
    echo -e "${BLUE}📦 Installing to $INSTALL_DIR...${NC}"
    sudo cp "$BINARY" "$INSTALL_DIR/"
    sudo chmod +x "$INSTALL_DIR/$BINARY"
    
    # Cleanup
    cd /
    rm -rf "$TEMP_DIR"
    
    echo -e "${GREEN}✅ secretsnap installed successfully!${NC}"
}

# Fallback: build from source
install_from_source() {
    local version=$1
    
    echo -e "${YELLOW}⚠️  Building from source (release download failed)${NC}"
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go is not installed. Please install Go 1.22+ first.${NC}"
        echo -e "${BLUE}📖 Visit: https://golang.org/doc/install${NC}"
        exit 1
    fi
    
    # Check if git is installed
    if ! command -v git &> /dev/null; then
        echo -e "${RED}❌ git is not installed${NC}"
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
