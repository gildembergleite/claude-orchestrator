package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigureClaude_CreatesNewFile(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	claudeMD := filepath.Join(claudeDir, "CLAUDE.md")

	err := ConfigureClaude(claudeDir, claudeMD)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatalf("CLAUDE.md not created: %v", err)
	}
	if !strings.Contains(string(content), "Memória de Sessão") {
		t.Fatal("should contain memory section")
	}
}

func TestConfigureClaude_AppendsToExisting(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(claudeDir, 0755)
	claudeMD := filepath.Join(claudeDir, "CLAUDE.md")

	os.WriteFile(claudeMD, []byte("# My Config\n\nSome existing content.\n"), 0644)

	err := ConfigureClaude(claudeDir, claudeMD)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(claudeMD)
	s := string(content)
	if !strings.Contains(s, "My Config") {
		t.Fatal("should preserve existing content")
	}
	if !strings.Contains(s, "Memória de Sessão") {
		t.Fatal("should append memory section")
	}
}

func TestConfigureClaude_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(claudeDir, 0755)
	claudeMD := filepath.Join(claudeDir, "CLAUDE.md")

	ConfigureClaude(claudeDir, claudeMD)
	ConfigureClaude(claudeDir, claudeMD) // run again

	content, _ := os.ReadFile(claudeMD)
	count := strings.Count(string(content), "Memória de Sessão")
	if count != 1 {
		t.Fatalf("expected 1 memory section, got %d", count)
	}
}
