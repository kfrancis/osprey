package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/christianfindlay/osprey/internal/cli"
)

// TestDocumentationDeterministic verifies that documentation generation
// produces identical output across 5 consecutive runs.
func TestDocumentationDeterministic(t *testing.T) {
	const numRuns = 5

	// Create temporary directories for each run
	tempDirs := make([]string, numRuns)
	defer func() {
		// Clean up all temp directories
		for _, dir := range tempDirs {
			if dir != "" {
				_ = os.RemoveAll(dir)
			}
		}
	}()

	// Generate documentation 5 times
	for i := range numRuns {
		tempDir, err := os.MkdirTemp("", fmt.Sprintf("osprey-docs-test-%d-", i))
		if err != nil {
			t.Fatalf("Failed to create temp directory for run %d: %v", i, err)
		}
		tempDirs[i] = tempDir

		// Generate documentation
		result := cli.RunCommand("", "docs", tempDir)
		if !result.Success {
			t.Fatalf("Documentation generation failed on run %d: %s", i, result.ErrorMsg)
		}
	}

	// Compare all runs against the first run
	firstRunHashes := getDirectoryHashes(t, tempDirs[0])

	for i := 1; i < numRuns; i++ {
		currentRunHashes := getDirectoryHashes(t, tempDirs[i])

		// Compare the hash maps
		if !hashMapsEqual(firstRunHashes, currentRunHashes) {
			t.Errorf("Documentation output differs between run 1 and run %d", i+1)

			// Print detailed differences for debugging
			printHashDifferences(t, firstRunHashes, currentRunHashes, 1, i+1)
		}
	}

	t.Logf("All %d documentation generation runs produced identical output", numRuns)
}

// TestFunctionsIndexDeterministic specifically tests the functions index file
// that was mentioned in the user's request.
func TestFunctionsIndexDeterministic(t *testing.T) {
	const numRuns = 5

	// Store file contents from each run
	var contents []string

	for i := range numRuns {
		tempDir, err := os.MkdirTemp("", fmt.Sprintf("osprey-functions-test-%d-", i))
		if err != nil {
			t.Fatalf("Failed to create temp directory for run %d: %v", i, err)
		}
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Generate documentation
		result := cli.RunCommand("", "docs", tempDir)
		if !result.Success {
			t.Fatalf("Documentation generation failed on run %d: %s", i, result.ErrorMsg)
		}

		// Read the functions index file
		functionsIndexPath := filepath.Join(tempDir, "functions", "index.md")
		content, err := os.ReadFile(functionsIndexPath)
		if err != nil {
			t.Fatalf("Failed to read functions index file on run %d: %v", i, err)
		}

		contents = append(contents, string(content))
	}

	// Compare all runs against the first run
	firstContent := contents[0]
	for i := 1; i < numRuns; i++ {
		if contents[i] != firstContent {
			t.Errorf("Functions index file differs between run 1 and run %d", i+1)

			// Show first few lines of difference for debugging
			t.Logf("First run content (first 200 chars): %s...",
				truncateString(firstContent, 200))
			t.Logf("Run %d content (first 200 chars): %s...",
				i+1, truncateString(contents[i], 200))
		}
	}

	t.Logf("Functions index file is identical across all %d runs", numRuns)
}

// getDirectoryHashes recursively walks a directory and returns a map of
// relative file paths to their SHA256 hashes.
func getDirectoryHashes(t *testing.T, dirPath string) map[string]string {
	hashes := make(map[string]string)

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		// Calculate SHA256 hash
		hash := sha256.Sum256(content)

		// Store with relative path
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		hashes[relPath] = hex.EncodeToString(hash[:])
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to walk directory %s: %v", dirPath, err)
	}

	return hashes
}

// hashMapsEqual compares two hash maps for equality.
func hashMapsEqual(map1, map2 map[string]string) bool {
	if len(map1) != len(map2) {
		return false
	}

	for key, value1 := range map1 {
		if value2, exists := map2[key]; !exists || value1 != value2 {
			return false
		}
	}

	return true
}

// printHashDifferences prints detailed differences between two hash maps.
func printHashDifferences(t *testing.T, map1, map2 map[string]string, run1, run2 int) {
	t.Logf("=== Hash differences between run %d and run %d ===", run1, run2)

	// Files only in map1
	for file := range map1 {
		if _, exists := map2[file]; !exists {
			t.Logf("File only in run %d: %s", run1, file)
		}
	}

	// Files only in map2
	for file := range map2 {
		if _, exists := map1[file]; !exists {
			t.Logf("File only in run %d: %s", run2, file)
		}
	}

	// Files with different hashes
	for file, hash1 := range map1 {
		if hash2, exists := map2[file]; exists && hash1 != hash2 {
			t.Logf("File differs: %s", file)
			t.Logf("  Run %d hash: %s", run1, hash1)
			t.Logf("  Run %d hash: %s", run2, hash2)
		}
	}
}

// truncateString truncates a string to the specified length.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
