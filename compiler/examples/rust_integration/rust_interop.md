# Rust-Osprey Interoperability

The Osprey compiler now supports calling Rust functions from Osprey code through a robust interoperability system.

## Overview

This system enables:
- Define external Rust functions in Osprey using `extern fn` declarations
- Call Rust functions from Osprey code with proper type safety
- Link Rust static libraries with compiled Osprey programs
- Leverage Rust's performance and ecosystem from within Osprey

## How It Works

1. **External Function Declaration**: Osprey code declares external functions using the `extern fn` syntax
2. **Rust Implementation**: Rust code implements these functions with C ABI compatibility
3. **Static Library**: Rust code is compiled into a static library (.a file)
4. **Linking**: The final executable links the Osprey object file with the Rust static library

## Example Implementation

### Osprey Code (`demo.osp`)
```osprey
// External function declarations
extern fn rust_add(a: int, b: int) -> int
extern fn rust_multiply(a: int, b: int) -> int
extern fn rust_factorial(n: int) -> int
extern fn rust_fibonacci(n: int) -> int
extern fn rust_is_prime(n: int) -> bool

// Using the external functions
let sum = rust_add(a: 15, b: 25)
let product = rust_multiply(a: 6, b: 7)
let factorial = rust_factorial(5)
let fib = rust_fibonacci(10)
let isPrime = rust_is_prime(17)

print("Rust add(15, 25) = ${sum}")
print("Rust multiply(6, 7) = ${product}")
print("Rust factorial(5) = ${factorial}")
print("Rust fibonacci(10) = ${fib}")
print("Rust is_prime(17) = ${isPrime}")
```

### Rust Implementation (`src/lib.rs`)
```rust
// Basic arithmetic functions
#[no_mangle]
pub extern "C" fn rust_add(a: i64, b: i64) -> i64 {
    a + b
}

#[no_mangle]
pub extern "C" fn rust_multiply(a: i64, b: i64) -> i64 {
    a * b
}

// More complex mathematical functions
#[no_mangle]
pub extern "C" fn rust_factorial(n: i64) -> i64 {
    if n <= 1 { 1 } else { n * rust_factorial(n - 1) }
}

#[no_mangle]
pub extern "C" fn rust_fibonacci(n: i64) -> i64 {
    match n {
        0 => 0,
        1 => 1,
        _ => rust_fibonacci(n - 1) + rust_fibonacci(n - 2)
    }
}

#[no_mangle]
pub extern "C" fn rust_is_prime(n: i64) -> bool {
    if n < 2 { return false; }
    for i in 2..=(n as f64).sqrt() as i64 {
        if n % i == 0 { return false; }
    }
    true
}
```

### Cargo Configuration (`Cargo.toml`)
```toml
[package]
name = "osprey_math_utils"
version = "0.1.0"
edition = "2021"

[lib]
name = "osprey_math_utils"
crate-type = ["staticlib"]
```

## Build Process

1. **Build Rust Library**: `cargo build --release`
2. Generate Osprey LLVM IR: `osprey demo.osp --llvm > demo.ll`
3. **Compile to Object**: `llc -filetype=obj demo.ll -o demo.o`
4. **Link with Rust**: `clang demo.o target/release/libosprey_math_utils.a -o demo`
5. **Run**: `./demo`

## Type Mapping

| Osprey Type | Rust Type | LLVM Type |
|-------------|-----------|-----------|
| `int`       | `i64`     | `i64`     |
| `bool`      | `bool`    | `i1`      |
| `string`    | `*const c_char` | `i8*` |

## Key Requirements

### Rust Side
- Functions must be marked with `#[no_mangle]`
- Functions must use `extern "C"` ABI
- Library must be compiled as `staticlib`
- Use appropriate C-compatible types

### Osprey Side
- External functions must be declared with `extern fn`
- All parameters must have explicit type annotations
- Function calls use named argument syntax for multi-parameter functions

### Linking
- Rust static library must be available at link time
- LLVM tools (llc, clang) must be available
- Proper library search paths must be configured

## Advanced Features

### Error Handling
```rust
#[no_mangle]
pub extern "C" fn rust_safe_divide(a: i64, b: i64) -> i64 {
    if b == 0 { -1 } else { a / b }  // Return -1 for division by zero
}
```

### String Handling
```rust
use std::ffi::{CStr, CString};
use std::os::raw::c_char;

#[no_mangle]
pub extern "C" fn rust_string_length(s: *const c_char) -> i64 {
    if s.is_null() { return 0; }
    unsafe {
        CStr::from_ptr(s).to_str().unwrap_or("").len() as i64
    }
}
```

## Directory Structure

```
rust_integration/
├── demo.osp            # Osprey code using Rust functions
├── run.sh              # Build and run script
├── Cargo.toml          # Rust project configuration
├── Cargo.lock          # Rust dependency lock file
├── src/
│   └── lib.rs          # Rust implementation
└── target/
    └── release/
        └── libosprey_math_utils.a  # Generated static library
```

## Testing

The integration includes comprehensive testing:

1. **Unit Tests**: Test Rust functions independently
2. **Integration Tests**: Test through Osprey calls
3. **Build Tests**: Verify compilation and linking
4. **Runtime Tests**: Verify correct execution

## Performance Benefits

- **Zero-cost abstractions**: Direct function calls with no overhead
- **Rust optimizations**: Benefit from Rust's advanced optimizations
- **Static linking**: No runtime dependencies
- **Type safety**: Compile-time verification of function signatures

## Future Enhancements

- Support for more complex types (structs, arrays)
- Async function support
- Error propagation mechanisms
- Dynamic library support
- Automatic binding generation

This interoperability system allows Osprey programs to leverage the entire Rust ecosystem while maintaining type safety and performance. 