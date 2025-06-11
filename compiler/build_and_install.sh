#!/bin/bash

# Build and Install Osprey Extension Script
# This script builds the compiler, packages the extension, and installs it in both VSCode and Cursor

set -e  # Exit on any error

echo "🚀 Osprey Build and Install Script"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_step() {
    echo -e "${BLUE}📋 $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Get the project root directory (where this script is located - Go project)
PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
EXTENSION_DIR="$PROJECT_ROOT/../vscode-extension"
VSIX_FILE="$EXTENSION_DIR/osprey-language-support-0.1.0.vsix"

print_step "Go project root: $PROJECT_ROOT"
print_step "Extension directory: $EXTENSION_DIR"

# Step 1: Build the Osprey Go compiler
print_step "Building Osprey Go compiler..."
cd "$PROJECT_ROOT"

if [ -f "Makefile" ]; then
    print_step "Running 'make build' in Go project..."
    make build
    print_success "Go compiler build completed"
    
    print_step "Running 'make install' in Go project..."
    make install
    print_success "Go compiler install completed"
else
    print_warning "No Makefile found in Go project, skipping compiler build"
fi

# Step 2: Build the VSCode extension (Node.js project)
print_step "Building VSCode extension (Node.js)..."
cd "$EXTENSION_DIR"

# Check if we're in the right directory
if [ ! -f "package.json" ]; then
    print_error "No package.json found in $EXTENSION_DIR"
    print_error "This script must be run from the osprey project root"
    exit 1
fi

# Install extension dependencies if needed
if [ ! -d "node_modules" ]; then
    print_step "Installing extension dependencies..."
    npm install
fi

# Install server dependencies if needed
if [ ! -d "server/node_modules" ]; then
    print_step "Installing server dependencies..."
    cd server
    npm install
    cd ..
fi

# Compile TypeScript for both client and server
print_step "Compiling client TypeScript..."
npm run compile

print_step "Compiling server TypeScript..."
cd server
npm run compile
cd ..

# Package the extension
print_step "Packaging extension..."
if command -v vsce &> /dev/null; then
    vsce package
else
    print_warning "vsce not found, installing..."
    npm install -g @vscode/vsce
    vsce package
fi

print_success "Extension packaged: $VSIX_FILE"

# Step 3: Install in VSCode
print_step "Installing extension in VSCode..."
if command -v code &> /dev/null; then
    # Uninstall old version first (ignore errors)
    code --uninstall-extension christianfindlay.osprey-language-support 2>/dev/null || true
    
    # Install new version
    code --install-extension "$VSIX_FILE" --force
    print_success "Extension installed in VSCode"
else
    print_warning "VSCode 'code' command not found, skipping VSCode installation"
    print_warning "You can manually install by opening VSCode and using Extensions > Install from VSIX"
fi

# Step 4: Install in Cursor
print_step "Installing extension in Cursor..."
if command -v cursor &> /dev/null; then
    # Uninstall old version first (ignore errors)
    cursor --uninstall-extension christianfindlay.osprey-language-support 2>/dev/null || true
    
    # Install new version
    cursor --install-extension "$VSIX_FILE" --force
    print_success "Extension installed in Cursor"
else
    print_warning "Cursor 'cursor' command not found, skipping Cursor installation"
    print_warning "You can manually install by opening Cursor and using Extensions > Install from VSIX"
fi

# Step 5: Verify installation
print_step "Verifying installation..."
echo ""
echo "📋 Installation Summary:"
echo "  • Go compiler: $(which osprey 2>/dev/null || echo 'Not found in PATH')"
echo "  • Extension VSIX: $VSIX_FILE"
echo "  • VSCode installation: $(if command -v code &> /dev/null; then echo "✅ Attempted"; else echo "⚠️ Skipped"; fi)"
echo "  • Cursor installation: $(if command -v cursor &> /dev/null; then echo "✅ Attempted"; else echo "⚠️ Skipped"; fi)"
echo ""

print_success "Build and installation complete!"
echo ""
echo "🔍 Next steps:"
echo "  1. Restart VSCode/Cursor if they were running"
echo "  2. Open a .osp file to test the extension"
echo "  3. Check that syntax highlighting works"
echo "  4. Verify the Osprey compiler is working: osprey --help"
echo ""

# Optional: Test the compiler
if command -v osprey &> /dev/null; then
    print_step "Testing Osprey compiler..."
    echo "Osprey version:"
    osprey --help 2>/dev/null || osprey 2>&1 | head -5 || echo "Compiler test failed"
else
    print_warning "Osprey compiler not found in PATH after installation"
fi

print_success "Script completed successfully! 🎉" 