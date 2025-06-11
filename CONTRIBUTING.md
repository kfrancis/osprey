# Contributing to Osprey

Thanks for your interest in contributing! This guide will help you get started.

## Ways to Contribute

1. **Implement language features** - Check [compiler/spec.md](compiler/spec.md) for unimplemented features
2. **Write tests** - Add examples to `compiler/examples/` directory
3. **Fix bugs** - Look for issues labeled "good first issue"
4. **Improve documentation** - Enhance error messages and examples

## Quick Start

### Prerequisites
- Docker (Dev Containers recommended)
- GitHub account

### Setup
1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/osprey.git`
3. Open in VS Code and use "Reopen in Devcontainer"
4. Create feature branch: `git checkout -b feature/your-feature-name`

### Development
```bash
cd compiler
go test ./...        # Run tests
make build          # Build compiler
make test           # Test with examples
```

### Submit Changes
1. Commit with clear message describing what and why
2. Push to your fork: `git push origin feature/your-feature-name`
3. Create pull request with:
   - Clear description of changes
   - Testing performed
   - Any breaking changes

## Finding Work

- Browse issues with "good first issue" or "help wanted" labels
- Search for `TODO` comments in codebase
- Check `compiler/examples/failscompilation/` for features that should work
- Review [spec.md](compiler/spec.md) for "NOT IMPLEMENTED" sections

## Code Guidelines

- Follow existing patterns and style
- Include tests for new functionality
- Keep changes focused and atomic
- Document complex logic with comments
- Fix linter errors before submitting

## Using AI Tools

AI tools like Claude can help you:
- Understand complex compiler concepts
- Implement language features
- Debug issues and write tests
- Learn patterns from existing code

Ask questions like: "How do I add a new operator to this parser?" or "What's the pattern for implementing pattern matching?"

## Getting Help

- Open an issue for discussion
- Check existing issues and documentation
- Review similar implementations in the codebase

## License

By contributing, you agree that your contributions will be licensed under the MIT License. 