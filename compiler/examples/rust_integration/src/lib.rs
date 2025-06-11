//! Math utilities written in Rust for Osprey integration
//! This module provides simple mathematical functions that can be called from Osprey

/// Adds two 64-bit integers
/// 
/// # Arguments
/// * `a` - First integer
/// * `b` - Second integer
/// 
/// # Returns
/// The sum of a and b
#[no_mangle]
pub extern "C" fn rust_add(a: i64, b: i64) -> i64 {
    a + b
}

/// Multiplies two 64-bit integers
/// 
/// # Arguments
/// * `a` - First integer
/// * `b` - Second integer
/// 
/// # Returns
/// The product of a and b
#[no_mangle]
pub extern "C" fn rust_multiply(a: i64, b: i64) -> i64 {
    a * b
}

/// Calculates the factorial of a number
/// 
/// # Arguments
/// * `n` - The number to calculate factorial for
/// 
/// # Returns
/// The factorial of n, or 1 if n <= 0
#[no_mangle]
pub extern "C" fn rust_factorial(n: i64) -> i64 {
    if n <= 0 {
        1
    } else {
        (1..=n).product()
    }
}

/// Calculates the nth Fibonacci number
/// 
/// # Arguments
/// * `n` - The position in the Fibonacci sequence
/// 
/// # Returns
/// The nth Fibonacci number
#[no_mangle]
pub extern "C" fn rust_fibonacci(n: i64) -> i64 {
    if n <= 1 {
        n
    } else {
        let mut a = 0;
        let mut b = 1;
        for _ in 2..=n {
            let temp = a + b;
            a = b;
            b = temp;
        }
        b
    }
}

/// Checks if a number is prime
/// 
/// # Arguments
/// * `n` - The number to check
/// 
/// # Returns
/// 1 if the number is prime, 0 otherwise
#[no_mangle]
pub extern "C" fn rust_is_prime(n: i64) -> i64 {
    if n < 2 {
        return 0;
    }
    if n == 2 {
        return 1;
    }
    if n % 2 == 0 {
        return 0;
    }
    
    let limit = (n as f64).sqrt() as i64;
    for i in (3..=limit).step_by(2) {
        if n % i == 0 {
            return 0;
        }
    }
    1
} 