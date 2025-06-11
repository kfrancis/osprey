#!/bin/bash

# Test runner for Osprey VSCode Extension
# This script runs the comprehensive test suite to verify all features

set -e

echo "🧪 Osprey Extension Test Runner"
echo "================================"

# Check if we're in the right directory
if [ ! -f "package.json" ]; then
    echo "❌ Error: Must be run from the VS Code Extension directory"
    exit 1
fi

# Check if osprey compiler is available
if ! command -v osprey &> /dev/null; then
    echo "❌ Error: osprey compiler not found in PATH"
    echo "   Please build and install the compiler first:"
    echo "   cd ../compiler && ./build_and_install.sh"
    exit 1
fi

echo "✅ Osprey compiler found: $(which osprey)"

# Install dependencies if needed
if [ ! -d "node_modules" ]; then
    echo "📦 Installing dependencies..."
    npm install
fi

# Compile TypeScript
echo "🔨 Compiling TypeScript..."
npm run compile

# Build the extension
echo "📦 Building extension package..."
npm run package

# Run the tests
echo "🧪 Running comprehensive test suite..."
echo ""
echo "Tests include:"
echo "  ✅ Basic extension activation"
echo "  ✅ Language server integration" 
echo "  ✅ Hover documentation (from compiler)"
echo "  ✅ Built-in function documentation"
echo "  ✅ Pipe operator documentation"
echo "  ✅ Signature help"
echo "  ✅ Diagnostics and error reporting"
echo "  ✅ Document symbols"
echo "  ✅ Code completion"
echo "  ✅ Compiler integration verification"
echo ""

# Run tests with timeout
timeout 300 npm test || {
    echo "❌ Tests failed or timed out"
    exit 1
}

echo ""
echo "🎉 All tests completed successfully!"
echo ""
echo "📋 Test Coverage:"
echo "  ✅ Extension activation and basic functionality"
echo "  ✅ Language server startup and communication"
echo "  ✅ Dynamic documentation from compiler"
echo "  ✅ All built-in functions have hover support"
echo "  ✅ Pipe operator documentation"
echo "  ✅ Function signature help"
echo "  ✅ Syntax error diagnostics"
echo "  ✅ Symbol navigation"
echo "  ✅ Code completion"
echo "  ✅ Compiler integration verification"
echo ""
echo "🚀 Extension is ready for use!" 