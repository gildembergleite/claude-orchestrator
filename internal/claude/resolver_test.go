package claude

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolve_FindsExistingBinary(t *testing.T) {
	// Create a temp directory with a fake "claude" binary
	tmpDir := t.TempDir()
	fakeBin := filepath.Join(tmpDir, "claude")
	if err := os.WriteFile(fakeBin, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}

	result, err := Resolve([]string{fakeBin})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != fakeBin {
		t.Fatalf("expected %s, got %s", fakeBin, result)
	}
}

func TestResolve_SkipsNonExecutable(t *testing.T) {
	tmpDir := t.TempDir()
	fakeBin := filepath.Join(tmpDir, "claude")
	if err := os.WriteFile(fakeBin, []byte("#!/bin/sh\n"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Resolve([]string{fakeBin})
	if err != nil {
		// No npx fallback available — error is fine
		return
	}
	// If npx is available, result should NOT be the non-executable path
	if result == fakeBin {
		t.Fatal("should have skipped non-executable file")
	}
}

func TestResolve_ReturnsNpxFallback(t *testing.T) {
	result, err := Resolve([]string{"/nonexistent/path/claude"})
	if err != nil {
		t.Fatalf("expected npx fallback, got error: %v", err)
	}
	if result != "npx @anthropic-ai/claude-code" {
		t.Fatalf("expected npx fallback, got %s", result)
	}
}
