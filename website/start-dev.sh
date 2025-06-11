#!/bin/bash

echo "🌐 Starting Osprey Website (Dev/Local)..."
echo "================================================"
echo ""
echo "This will start the Osprey website development server:"
echo "- Static site generator (Eleventy/11ty)"
echo "- Hot reload for development"
echo "- Syntax highlighting for Osprey code"
echo ""
echo "Access the website at: http://localhost:8080"
echo ""
echo "================================================"

# Navigate to website directory
cd "$(dirname "$0")"

# Detect if we're in a dev container or local environment
if [ -f "/usr/lib/node_modules/npm/bin/npm-cli.js" ]; then
    NPM_CMD="/usr/lib/node_modules/npm/bin/npm-cli.js"
    echo "🐳 Detected dev container environment"
else
    NPM_CMD="npm"
    echo "💻 Detected local environment"
    
    # Check if Node.js and npm are installed locally
    if ! command -v node &> /dev/null; then
        echo "❌ Node.js is required but not installed!"
        echo "Install Node.js from: https://nodejs.org/"
        exit 1
    fi

    if ! command -v npm &> /dev/null; then
        echo "❌ npm is required but not installed!"
        echo "Install npm with Node.js from: https://nodejs.org/"
        exit 1
    fi
fi

# Check if dependencies are installed
if [ ! -d "node_modules" ]; then
    echo "📦 Installing website dependencies..."
    $NPM_CMD install
fi

# Start the development server
echo "🚀 Starting Eleventy development server..."
echo "   - Building site and watching for changes..."
echo "   - Server will start on port 8080 (accessible from host)"
echo ""
/usr/lib/node_modules/npm/bin/npm-cli.js run dev