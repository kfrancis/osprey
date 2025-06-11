# Osprey Web Compiler Testing

This directory contains test scripts to validate the Osprey web compiler API functionality.

## Test Scripts

### 🚀 Quick Test (`quick-test.sh`)
**Duration:** ~10 seconds  
**Purpose:** Essential functionality validation

```bash
./quick-test.sh
```

**Tests:**
- ✅ Health check endpoint
- ✅ Successful program execution (string interpolation, print)
- ❌ Failing program (field access error detection)
- 🧮 Mathematical computation (demonstrates correct calculation)

### 🧪 Comprehensive Test Suite (`test-api.sh`)
**Duration:** ~30 seconds  
**Purpose:** Full API validation with detailed reporting

```bash
./test-api.sh
```

**Tests:**
- Health check
- Successful vs failing program execution
- Compile-only checks (AST generation)
- Mathematical computations
- Pattern matching with union types
- Return value handling (0 vs non-zero)

## Test Cases

### Successful Programs
- **String interpolation:** `let x = 5; print("x = ${x}")`
- **Mathematical functions:** `fn add(x: int, y: int) -> int = x + y`
- **Pattern matching:** Union types with match expressions
- **Zero return:** Programs that return 0 (success)

### Failing Programs
- **Field access errors:** `let user = 42; print("${user.name}")`
- **Type mismatches:** Accessing fields on primitive types
- **Compilation errors:** Caught during LLVM IR generation

## API Endpoints Tested

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api` | GET | Health check |
| `/api/compile` | POST | AST generation only |
| `/api/run` | POST | Full compile and execute |

## Expected Results

### ✅ Success Cases
- Health check returns JSON with status "ok"
- Valid programs compile and execute
- Programs returning 0 show `"success": true`
- String interpolation and print statements work

### ❌ Expected "Failures" 
- Programs returning non-zero values show `"success": false` but may be computing correctly
- Field access on primitives caught during code generation
- Parse errors vs semantic errors properly distinguished

## Notes

- **Exit Status Interpretation:** Programs that return non-zero values (like 42) are flagged as "errors" by the API, but this indicates the computation succeeded and returned that value
- **Error Stages:** The compiler properly distinguishes between parse errors (syntax) and semantic errors (type checking, code generation)
- **Container Management:** Scripts automatically start/stop Docker containers and handle cleanup

## Requirements

- Docker with `osprey-web-compiler` image built
- `curl` command available
- `python3` for JSON formatting (optional, graceful fallback)
- Port 3001 available

## Usage Examples

```bash
# Quick validation
./quick-test.sh

# Comprehensive testing
./test-api.sh

# Build and test
docker build -f Dockerfile -t osprey-web-compiler .. && ./quick-test.sh
``` 