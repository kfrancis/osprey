#!/bin/bash

# Build Osprey Release Package for Homebrew
# Run this script to create a release tarball for Homebrew submission

set -e

echo "🚀 Building Osprey Release Package"
echo "=================================="

# Get directories
HOMEBREW_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$HOMEBREW_DIR/.." && pwd)"
COMPILER_DIR="$PROJECT_ROOT/compiler"
RELEASE_DIR="$HOMEBREW_DIR/release"

echo "📁 Homebrew directory: $HOMEBREW_DIR"
echo "📁 Project root: $PROJECT_ROOT"
echo "📁 Compiler directory: $COMPILER_DIR"
echo "📁 Release directory: $RELEASE_DIR"

# Clean and create release directory
rm -rf "$RELEASE_DIR"
mkdir -p "$RELEASE_DIR"

# Build the compiler
echo "🔨 Building Osprey compiler..."
cd "$COMPILER_DIR"

if [ ! -f "Makefile" ]; then
    echo "❌ No Makefile found in $COMPILER_DIR"
    exit 1
fi

# Build everything
make clean
make build

# Verify binaries exist
if [ ! -f "bin/osprey" ]; then
    echo "❌ Compiler binary not found at bin/osprey"
    exit 1
fi

if [ ! -f "bin/libfiber_runtime.a" ]; then
    echo "❌ Fiber runtime library not found"
    exit 1
fi

if [ ! -f "bin/libhttp_runtime.a" ]; then
    echo "❌ HTTP runtime library not found"
    exit 1
fi

# Build for all architectures
ARCHITECTURES=("arm64" "amd64")

echo "🏗️  Building for ALL architectures: ${ARCHITECTURES[*]}"

for DARWIN_ARCH in "${ARCHITECTURES[@]}"; do
    echo ""
    echo "📦 Creating release package for $DARWIN_ARCH..."
    PACKAGE_NAME="osprey-darwin-$DARWIN_ARCH"
    PACKAGE_DIR="$RELEASE_DIR/$PACKAGE_NAME"
    mkdir -p "$PACKAGE_DIR"

    # Copy binaries and libraries from the compiler directory
    cp "$COMPILER_DIR/bin/osprey" "$PACKAGE_DIR/"
    cp "$COMPILER_DIR/bin/libfiber_runtime.a" "$PACKAGE_DIR/"
    cp "$COMPILER_DIR/bin/libhttp_runtime.a" "$PACKAGE_DIR/"

    # Create tarball
    cd "$RELEASE_DIR"
    tar -czf "$PACKAGE_NAME.tar.gz" "$PACKAGE_NAME/"

    # Calculate SHA256
    CHECKSUM=$(shasum -a 256 "$PACKAGE_NAME.tar.gz" | cut -d' ' -f1)
    
    echo "✅ $PACKAGE_NAME.tar.gz created (SHA256: $CHECKSUM)"
done

cd "$RELEASE_DIR"
echo ""
echo "🎉 ALL RELEASE PACKAGES CREATED:"
echo ""
for DARWIN_ARCH in "${ARCHITECTURES[@]}"; do
    PACKAGE_NAME="osprey-darwin-$DARWIN_ARCH"
    CHECKSUM=$(shasum -a 256 "$PACKAGE_NAME.tar.gz" | cut -d' ' -f1)
    echo "📦 $PACKAGE_NAME.tar.gz"
    echo "   SHA256: $CHECKSUM"
    echo ""
done

echo "🚀 Next steps:"
echo "  1. Upload ALL .tar.gz files to GitHub Releases"
echo "  2. Update osprey.rb with URLs for both architectures"
echo "  3. Submit to Homebrew"
echo ""

echo "🎉 Done!" 