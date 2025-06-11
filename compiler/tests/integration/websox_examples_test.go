package integration

// DO NOT EVER SKIP TESTS!!!!

import (
	"os"
	"strings"
	"testing"

	"github.com/christianfindlay/osprey/internal/codegen"
)

// TestWebsoxExamples tests WebSocket examples in the examples/tested/websox directory.
// These tests verify compilation only since network operations can hang in test environments.
func TestWebsoxExamples(t *testing.T) {
	checkLLVMTools(t)

	examplesDir := "../../examples/tested/websox"

	// Test compilation only for WebSocket examples to avoid network hangs
	websoxExamples := []string{
		"websocket_example.osp",
		"websocket_local_test.osp",
		"websocket_server_example.osp",
		"websocket_server_live.osp",
	}

	runWebsoxCompilationTests(t, examplesDir, websoxExamples)
}

// runWebsoxCompilationTests tests that the given WebSocket examples compile without errors.
func runWebsoxCompilationTests(t *testing.T, examplesDir string, examples []string) {
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

			t.Logf("✅ WebSocket example %s compiled successfully", example)
		})
	}
}
