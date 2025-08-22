#!/bin/bash

# Release script for Secrets Snapshot CLI
# This script builds releases for all platforms and creates tarballs

set -e

BINARY=secretsnap
PLATFORMS=("linux/amd64" "darwin/amd64" "darwin/arm64")
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")

echo "🚀 Building Secrets Snapshot CLI v$VERSION"

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -rf bin/ dist/

# Create directories
mkdir -p bin/ dist/

# Build for all platforms
for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$platform"
    echo "🔨 Building for $GOOS/$GOARCH..."
    
    # Build binary
    GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags="-s -w -X main.version=$VERSION" \
        -o "bin/$BINARY-$GOOS-$GOARCH" \
        ./main.go
    
    # Create tarball
    tar -czf "dist/$BINARY-$VERSION-$GOOS-$GOARCH.tar.gz" \
        -C bin/ \
        "$BINARY-$GOOS-$GOARCH"
    
    echo "✅ Built $BINARY-$VERSION-$GOOS-$GOARCH.tar.gz"
done

echo ""
echo "🎉 Release build complete!"
echo "📦 Artifacts in dist/:"
ls -la dist/

# Create checksums
echo "🔍 Creating checksums..."
cd dist/
for file in *.tar.gz; do
    shasum -a 256 "$file" > "$file.sha256"
done
cd ..

echo "✅ Release ready for distribution!"
