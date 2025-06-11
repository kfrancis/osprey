package integration

// DO NOT EVER SKIP TESTS!!!!

import (
	"os"
	"strings"
	"testing"

	"github.com/christianfindlay/osprey/internal/codegen"
)

// TestHttpExamples tests HTTP examples in the examples/tested/http directory.
// These tests verify compilation only since network operations can hang in test environments.
func TestHttpExamples(t *testing.T) {
	checkLLVMTools(t)

	examplesDir := "../../examples/tested/http"

	// Test compilation only for HTTP examples to avoid network hangs
	httpExamples := []string{
		"http_create_client.osp",
		"http_client_example.osp",
		"http_server_example.osp",
		"http_advanced_example.osp",
	}

	runCompilationOnlyTests(t, examplesDir, httpExamples)
}

// runCompilationOnlyTests tests that the given examples compile without errors.
func runCompilationOnlyTests(t *testing.T, examplesDir string, examples []string) {
	for _, example := range examples {
		testName := strings.TrimSuffix(example, ".osp")
		t.Run(testName, func(t *testing.T) {
			filePath := examplesDir + "/" + example

			// Read and compile the file
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", filePath, err)
			}

			source := string(content)

			// Test compilation only - avoid network execution that can hang
			_, err = codegen.CompileToLLVM(source)
			if err != nil {
				t.Fatalf("Failed to compile %s: %v", filePath, err)
			}

			t.Logf("✅ HTTP example %s compiled successfully", example)
		})
	}
}
