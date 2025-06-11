#!/bin/bash

set -e

echo "🧪 Running comprehensive code coverage analysis..."

# Clean up any previous coverage files
rm -f coverage.out coverage.html

# Run tests with coverage for all packages (excluding ANTLR-generated parser files)
echo "📊 Running tests with coverage..."
go test -v -coverprofile=coverage.out -covermode=atomic -coverpkg=./cmd/...,./internal/...,./examples/... ./...

# Generate HTML coverage report
echo "📄 Generating HTML coverage report..."
go tool cover -html=coverage.out -o coverage.html

# Show coverage summary
echo "📈 Coverage Summary:"
go tool cover -func=coverage.out

# Show total coverage percentage
TOTAL_COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
echo ""
echo "🎯 Total Coverage: $TOTAL_COVERAGE"

# Open HTML report if on macOS
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "🌐 Opening HTML coverage report in browser..."
    open coverage.html
fi

echo "✅ Coverage analysis complete!"
echo "📁 HTML report saved to: coverage.html"
echo "📁 Raw coverage data saved to: coverage.out" 