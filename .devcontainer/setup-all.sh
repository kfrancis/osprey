#!/usr/bin/env bash
# This script runs all setup scripts for the Vexels development environment

set -e

echo "🚀 Setting up complete Vexels development environment..."
echo ""

# Make all scripts executable
chmod +x /workspaces/vexels/.devcontainer/*.sh

# Build the compiler
echo "📦 Building compiler..."
/workspaces/vexels/.devcontainer/build-compiler.sh
echo ""

# Setup VS Code extension
echo "📦 Setting up VS Code extension..."
/workspaces/vexels/.devcontainer/setup-vscode-extension.sh
echo ""

# Test everything
echo "🧪 Testing setup..."
/workspaces/vexels/.devcontainer/test-setup.sh

echo ""
echo "✅ Complete development environment setup finished!"
echo ""
echo "🎯 You can now:"
echo "- Build the compiler: cd compiler && make build"
echo "- Run compiler tests: cd compiler && make test"
echo "- Develop VS Code extension: Press F5 in VS Code"
echo "- Package extension: cd vscode-extension && npm run package" 