#!/bin/bash

# FINAL WORKING Rust-Osprey integration demo
# This script demonstrates Rust-Osprey interoperability

set -e  # Exit on error

# Get the script directory and compiler root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPILER_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "🦀 Starting Rust-Osprey Integration Demo"

echo "🔧 Checking required tools..."

# Check for rustc
if ! command -v rustc &> /dev/null; then
    echo "❌ rustc not found. Please install Rust: https://rustup.rs/"
    exit 1
fi

# Check for cargo
if ! command -v cargo &> /dev/null; then
    echo "❌ cargo not found. Please install Rust: https://rustup.rs/"
    exit 1
fi

# Check for clang  
if ! command -v clang &> /dev/null; then
    echo "❌ clang not found. Please install LLVM/Clang"
    exit 1
fi

echo "✅ All required tools are available"

echo "🧹 Cleaning up previous builds..."
cd "$SCRIPT_DIR"
rm -f librust_utils.a osprey.ll osprey.o final_rust_osprey_demo &>/dev/null || true

# Step 1: Build Rust library
echo "🦀 Building Rust library..."
cargo build --release

# Copy the Rust library to a simpler name
echo "📦 Copying Rust library..."
cp target/release/libosprey_math_utils.a librust_utils.a

# Step 2: Generate Osprey LLVM IR and compile to object
echo "🔧 Compiling Osprey code to LLVM IR..."

# Check if osprey binary exists
if [ ! -f "$COMPILER_ROOT/bin/osprey" ]; then
    echo "❌ Osprey binary not found at $COMPILER_ROOT/bin/osprey. Please run 'make build' first."
    exit 1
fi

cd "$COMPILER_ROOT"
./bin/osprey examples/rust_integration/demo.osp --llvm > examples/rust_integration/osprey.ll
if [ $? -ne 0 ]; then
    echo "❌ Failed to compile Osprey code"
    exit 1
fi
echo "✅ Osprey LLVM IR generated"

# Step 3: Compile LLVM IR to object file  
echo "🔧 Compiling LLVM IR to object file..."
cd "$SCRIPT_DIR"
clang -c osprey.ll -o osprey.o
if [ $? -ne 0 ]; then
    echo "❌ Failed to compile LLVM IR to object file"
    exit 1
fi
echo "✅ Object file created"

# Step 4: Link everything together
echo "🔗 Linking Rust and Osprey object files..."
clang osprey.o librust_utils.a -o final_rust_osprey_demo
if [ $? -ne 0 ]; then
    echo "❌ Failed to link object files"
    exit 1
fi
echo "✅ Executable created successfully"

# Step 5: Run the demo
echo "🚀 Running Rust-Osprey integration demo:"
echo "================================================"
./final_rust_osprey_demo
echo "================================================"

# Clean up
echo "🧹 Cleaning up build artifacts..."
rm -f osprey.ll osprey.o librust_utils.a final_rust_osprey_demo &>/dev/null

echo "✅ Rust-Osprey integration demo completed successfully!" 