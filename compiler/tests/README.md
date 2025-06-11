# Osprey Compiler Tests

This directory contains the consolidated test suite for the Osprey compiler, focusing on clean, efficient testing of all compiler functionality.

## Test Structure

### Integration Tests (`tests/integration/`)

**Main Test File: `examples_test.go`**

This consolidated test file contains all the essential testing functionality:

#### `TestRootLevelExamples`
- **Primary Test Function**: Compiles and runs all examples from `examples/tested/`
- Tests both compilation and execution output validation
- Verifies expected output for each example program
- The `failscompilation` folder is for testing failing compilating and verifying the error message
- Covers all language features including:
  - Basic syntax (variables, functions, expressions)
  - String interpolation and math operations
  - Pattern matching and type definitions
  - Iterator patterns and functional programming
  - Complex applications (space trader, adventure game)
  - Calculator and math operations

#### `TestBasicCompilation`
- Tests fundamental language syntax compilation
- Validates that basic constructs compile without errors
- Covers: variables, functions, patterns, types, interpolation

#### `TestErrorHandling`
- Ensures invalid syntax properly fails compilation
- Tests error conditions: undefined variables, malformed syntax, invalid operators

#### `TestFunctionArguments`
- Validates function argument requirements
- Tests both valid cases (single params, named arguments) and invalid cases (multi-param without names)

## Key Features

### Test Design
- **Single consolidated test file** instead of scattered test files
- **No code duplication** - common functions defined once
- **Focused testing** - each test has a clear, specific purpose

### Comprehensive Coverage
- **All examples tested** from the `examples/tested/` directory
- **Compilation verification** for every example
- **Output validation** for examples that produce output
- **Error case testing** to ensure graceful failure handling

### Efficient Execution
- Tests run quickly with parallel execution where appropriate
- LLVM tool availability checked before execution
- Graceful skipping of tests when tools are unavailable

## Running Tests

```bash
# Run all integration tests
cd tests/integration
go test -v

# Run specific test suites
go test -v -run TestRootLevelExamples
go test -v -run TestBasicCompilation
go test -v -run TestErrorHandling
go test -v -run TestFunctionArguments

# Run with timeout for longer examples
go test -v -timeout 30s
```

## Example Directory Structure

- `examples/tested/` - All validated, working examples
- `examples/broken/` - Non-functional examples for reference

The tests automatically discover and test all `.osp` files in the `tested` directory, ensuring new examples are automatically included in the test suite.

## Maintenance

When adding new examples:
1. Place working examples in `examples/tested/`
2. Add expected output to the `expectedOutputs` map in `TestRootLevelExamples` if the example produces output
3. Tests will automatically include the new example

The consolidated approach ensures:
- **No test file proliferation**
- **Easy maintenance and updates**
- **Clear test organization**
- **Comprehensive coverage with minimal complexity** 