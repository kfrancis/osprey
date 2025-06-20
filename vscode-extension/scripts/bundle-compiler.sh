#!/bin/bash

# Bundle Osprey Compiler for VSCode Extension
# This script downloads or builds the Osprey compiler and packages it with the extension

# Script parameters
OSPREY_VERSION="0.1.0"
DOWNLOAD_URL=""
BUILD_FROM_SOURCE=false
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Parse command line arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --version) OSPREY_VERSION="$2"; shift ;;
        --url) DOWNLOAD_URL="$2"; shift ;;
        --build) BUILD_FROM_SOURCE=true ;;
        *) echo "Unknown parameter: $1"; exit 1 ;;
    esac
    shift
done

echo "🦅 Bundling Osprey compiler for VSCode extension..."

# Create the bin directory if it doesn't exist
BIN_DIR="$SCRIPT_DIR/../bin"
mkdir -p "$BIN_DIR"
echo "Created bin directory at: $BIN_DIR"

# Function to download the Osprey compiler
download_osprey_compiler() {
    TEMP_FILE=$(mktemp)
    
    if [ -z "$DOWNLOAD_URL" ]; then
        # Determine platform and architecture
        PLATFORM=$(uname -s | tr '[:upper:]' '[:lower:]')
        ARCH=$(uname -m)
        
        # Normalize architecture
        case "$ARCH" in
            x86_64) ARCH="x64" ;;
            aarch64|arm64) ARCH="arm64" ;;
            *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
        esac
        
        # Normalize platform
        case "$PLATFORM" in
            linux) PLATFORM="linux" ;;
            darwin) PLATFORM="macos" ;;
            *) echo "Unsupported platform: $PLATFORM"; exit 1 ;;
        esac
        
        # If no URL is provided, construct a default one
        DOWNLOAD_URL="https://github.com/osprey/osprey/releases/download/v$OSPREY_VERSION/osprey-$OSPREY_VERSION-$PLATFORM-$ARCH.tar.gz"
    fi
    
    echo "Downloading Osprey compiler from: $DOWNLOAD_URL"
    
    # Download the file
    if command -v curl &> /dev/null; then
        curl -L "$DOWNLOAD_URL" -o "$TEMP_FILE"
    elif command -v wget &> /dev/null; then
        wget -O "$TEMP_FILE" "$DOWNLOAD_URL"
    else
        echo "Error: Neither curl nor wget found. Please install one of these utilities."
        return 1
    fi
    
    # Check if download was successful
    if [ $? -ne 0 ]; then
        echo "Error downloading Osprey compiler"
        return 1
    fi
    
    echo "Extracting compiler to bin directory..."
    
    # Extract based on file extension
    if [[ "$DOWNLOAD_URL" == *.zip ]]; then
        unzip -o "$TEMP_FILE" -d "$BIN_DIR"
    else
        tar -xzf "$TEMP_FILE" -C "$BIN_DIR"
    fi
    
    # Make sure the compiler is executable
    COMPILER_BIN="$BIN_DIR/osprey"
    if [ -f "$COMPILER_BIN" ]; then
        chmod +x "$COMPILER_BIN"
    else
        # Try to find the executable in extracted folders
        FOUND_EXEC=$(find "$BIN_DIR" -type f -name "osprey" | head -1)
        if [ -n "$FOUND_EXEC" ]; then
            cp "$FOUND_EXEC" "$COMPILER_BIN"
            chmod +x "$COMPILER_BIN"
            echo "Copied compiler executable to: $COMPILER_BIN"
        else
            echo "Warning: Could not find Osprey compiler executable in extracted files"
            return 1
        fi
    fi
    
    # Clean up
    rm -f "$TEMP_FILE"
    
    return 0
}

# Function to create a mock compiler for testing
create_mock_compiler() {
    echo "Creating mock Osprey compiler for development..."
    
    COMPILER_PATH="$BIN_DIR/osprey.js"
    
    cat > "$COMPILER_PATH" << 'EOF'
#!/usr/bin/env node

/**
 * Mock Osprey compiler for VSCode extension development
 * This is a simple JavaScript implementation to simulate the Osprey compiler
 */

const fs = require('fs');
const path = require('path');

// Parse command line arguments
const args = process.argv.slice(2);
let runMode = false;
let inputFile = null;

// Process arguments
args.forEach(arg => {
  if (arg === '--run') {
    runMode = true;
  } else if (!inputFile && !arg.startsWith('--')) {
    inputFile = arg;
  }
});

// Check if we have an input file
if (!inputFile) {
  console.error('Error: No input file specified');
  process.exit(1);
}

// Read the file content
try {
  const fileContent = fs.readFileSync(inputFile, 'utf8');
  
  // Simple syntax check - detect common errors
  let hasErrors = false;
  const lines = fileContent.split('\n');
  
  lines.forEach((line, index) => {
    // Example checks:
    
    // Check for missing equals in function definition
    if (line.trim().startsWith('fn ') && !line.includes('=')) {
      console.error(`${inputFile}:${index + 1}:${line.indexOf('fn') + 3}: Error: Function declaration missing '='`);
      hasErrors = true;
    }
    
    // Check for unclosed brackets
    if ((line.includes('{') && !fileContent.includes('}')) || 
        (line.includes('[') && !fileContent.includes(']')) ||
        (line.includes('(') && !fileContent.includes(')'))) {
      console.error(`${inputFile}:${index + 1}:${line.indexOf('{') || line.indexOf('[') || line.indexOf('(')}: Error: Unclosed bracket`);
      hasErrors = true;
    }
  });
  
  // If in run mode and no errors, "execute" the code
  if (runMode && !hasErrors) {
    // Extract and run the main function if it exists
    if (fileContent.includes('fn main()')) {
      console.log('Executing Osprey program...');
      
      // Find print statements and simulate their output
      const printMatches = fileContent.match(/print\(["'](.*?)["']\)/g);
      if (printMatches) {
        printMatches.forEach(match => {
          // Extract the content inside the quotes
          const content = match.match(/["'](.*?)["']/)[1];
          console.log(content);
        });
      }
      
      console.log('Program execution complete');
    } else {
      console.log('No main function found, nothing to execute');
    }
  }
  
  // Exit with error code if errors were found
  if (hasErrors) {
    process.exit(1);
  }
} catch (err) {
  console.error(`Error reading or processing file: ${err.message}`);
  process.exit(1);
}
EOF
    
    # Make the mock compiler executable
    chmod +x "$COMPILER_PATH"
    
    # Create a shell script wrapper
    WRAPPER_PATH="$BIN_DIR/osprey"
    cat > "$WRAPPER_PATH" << EOF
#!/bin/bash
node "$(dirname "$0")/osprey.js" "\$@"
EOF
    
    chmod +x "$WRAPPER_PATH"
    
    echo "Created mock compiler at: $COMPILER_PATH and wrapper at: $WRAPPER_PATH"
    return 0
}

# Main script execution
if [ "$BUILD_FROM_SOURCE" = true ]; then
    echo "Build from source option not implemented yet. Using mock compiler instead."
    create_mock_compiler
else
    # Try downloading first, fall back to mock compiler
    if ! download_osprey_compiler; then
        echo "Failed to download compiler. Creating mock compiler instead..."
        create_mock_compiler
    fi
fi

# Update package.json to include the compiler path
echo "Updating package.json with compiler path..."
PACKAGE_JSON="$SCRIPT_DIR/../package.json"

# Check if jq is installed for proper JSON manipulation
if command -v jq &> /dev/null; then
    # Create a temporary file with the updated JSON
    jq '.contributes.configuration.properties["osprey.server.compilerPath"] = {"type": "string", "default": "${workspaceFolder}/bin/osprey", "description": "Path to the Osprey compiler executable"}' "$PACKAGE_JSON" > "$PACKAGE_JSON.tmp"
    mv "$PACKAGE_JSON.tmp" "$PACKAGE_JSON"
else
    echo "Warning: jq not installed. Package.json not updated automatically."
    echo "Please manually add the osprey.server.compilerPath setting to your package.json"
fi

echo "✅ Osprey compiler bundling complete!"
