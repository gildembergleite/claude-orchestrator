package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectShell_ReturnsShellName(t *testing.T) {
	shell := DetectShell()
	valid := map[string]bool{"fish": true, "zsh": true, "bash": true, "unknown": true}
	if !valid[shell] {
		t.Fatalf("unexpected shell: %s", shell)
	}
}

func TestConfigureShellAlias_Fish(t *testing.T) {
	tmpDir := t.TempDir()
	functionsDir := filepath.Join(tmpDir, ".config", "fish", "functions")

	err := configureShellFish(functionsDir, "/usr/local/bin/claude-orchestrator")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(functionsDir, "claude-orchestrator.fish"))
	if err != nil {
		t.Fatal("claude-orchestrator.fish not created")
	}
	if !strings.Contains(string(content), "/usr/local/bin/claude-orchestrator") {
		t.Fatal("should contain binary path")
	}
}

func TestConfigureShellAlias_FishIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	functionsDir := filepath.Join(tmpDir, ".config", "fish", "functions")

	configureShellFish(functionsDir, "/usr/local/bin/claude-orchestrator")
	configureShellFish(functionsDir, "/usr/local/bin/claude-orchestrator")

	content, _ := os.ReadFile(filepath.Join(functionsDir, "claude-orchestrator.fish"))
	count := strings.Count(string(content), "function claude-orchestrator")
	if count != 1 {
		t.Fatalf("expected 1 function definition, got %d", count)
	}
}

func TestConfigureShellAlias_Bash(t *testing.T) {
	tmpDir := t.TempDir()
	rcPath := filepath.Join(tmpDir, ".bashrc")
	os.WriteFile(rcPath, []byte("# existing\n"), 0644)

	err := configureShellRC(rcPath, "/usr/local/bin/claude-orchestrator")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(rcPath)
	s := string(content)
	if !strings.Contains(s, "alias claude-orchestrator=") {
		t.Fatal("should contain alias")
	}
	if !strings.Contains(s, "existing") {
		t.Fatal("should preserve existing content")
	}
}

func TestConfigureShellFish_MultipleAliases(t *testing.T) {
	tmpDir := t.TempDir()
	functionsDir := filepath.Join(tmpDir, ".config", "fish", "functions")

	err := configureShellFishAliases(functionsDir, "/usr/local/bin/claude-orchestrator", []string{"claude-orchestrator", "co"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, name := range []string{"claude-orchestrator", "co"} {
		content, err := os.ReadFile(filepath.Join(functionsDir, name+".fish"))
		if err != nil {
			t.Fatalf("%s.fish not created: %v", name, err)
		}
		if !strings.Contains(string(content), "/usr/local/bin/claude-orchestrator") {
			t.Fatalf("%s.fish should contain binary path", name)
		}
		if !strings.Contains(string(content), "function "+name) {
			t.Fatalf("%s.fish should contain function %s", name, name)
		}
	}
}

func TestConfigureShellRC_MultipleAliases(t *testing.T) {
	tmpDir := t.TempDir()
	rcPath := filepath.Join(tmpDir, ".bashrc")
	os.WriteFile(rcPath, []byte("# existing\n"), 0644)

	err := configureShellRCAliases(rcPath, "/usr/local/bin/claude-orchestrator", []string{"claude-orchestrator", "co"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(rcPath)
	s := string(content)
	if !strings.Contains(s, `alias claude-orchestrator=`) {
		t.Fatal("should contain claude-orchestrator alias")
	}
	if !strings.Contains(s, `alias co=`) {
		t.Fatal("should contain co alias")
	}
}

func TestConfigureShellRC_CustomAlias(t *testing.T) {
	tmpDir := t.TempDir()
	rcPath := filepath.Join(tmpDir, ".zshrc")
	os.WriteFile(rcPath, []byte(""), 0644)

	err := configureShellRCAliases(rcPath, "/usr/local/bin/claude-orchestrator", []string{"meu-cli"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(rcPath)
	if !strings.Contains(string(content), `alias meu-cli=`) {
		t.Fatal("should contain custom alias")
	}
}

func TestConfigureShellRC_RemovesPreviousAliases(t *testing.T) {
	tmpDir := t.TempDir()
	rcPath := filepath.Join(tmpDir, ".zshrc")
	os.WriteFile(rcPath, []byte("# user existing\nexport FOO=bar\n"), 0644)

	// First call: creates two aliases
	configureShellRCAliases(rcPath, "/usr/local/bin/claude-orchestrator", []string{"claude-orchestrator", "co"})

	// Second call: only claude-orchestrator — co should be removed
	configureShellRCAliases(rcPath, "/usr/local/bin/claude-orchestrator", []string{"claude-orchestrator"})

	content, _ := os.ReadFile(rcPath)
	s := string(content)
	if !strings.Contains(s, `alias claude-orchestrator=`) {
		t.Fatal("should contain claude-orchestrator alias")
	}
	if strings.Contains(s, `alias co=`) {
		t.Fatal("co alias should have been removed")
	}
	if !strings.Contains(s, "user existing") {
		t.Fatal("should preserve unrelated user content")
	}
	if !strings.Contains(s, "export FOO=bar") {
		t.Fatal("should preserve unrelated user content")
	}
}

func TestConfigureShellFish_RemovesPreviousAliases(t *testing.T) {
	tmpDir := t.TempDir()
	functionsDir := filepath.Join(tmpDir, ".config", "fish", "functions")

	// First call: creates two function files
	configureShellFishAliases(functionsDir, "/usr/local/bin/claude-orchestrator", []string{"claude-orchestrator", "co"})

	// Verify both exist
	if _, err := os.Stat(filepath.Join(functionsDir, "co.fish")); err != nil {
		t.Fatal("co.fish should exist after first call")
	}

	// Second call: only claude-orchestrator
	configureShellFishAliases(functionsDir, "/usr/local/bin/claude-orchestrator", []string{"claude-orchestrator"})

	// claude-orchestrator.fish should still exist
	if _, err := os.Stat(filepath.Join(functionsDir, "claude-orchestrator.fish")); err != nil {
		t.Fatal("claude-orchestrator.fish should still exist")
	}
	// co.fish should have been removed
	if _, err := os.Stat(filepath.Join(functionsDir, "co.fish")); err == nil {
		t.Fatal("co.fish should have been removed")
	}
}

func TestConfigureShellFish_PreservesUnrelatedFishFiles(t *testing.T) {
	tmpDir := t.TempDir()
	functionsDir := filepath.Join(tmpDir, ".config", "fish", "functions")
	os.MkdirAll(functionsDir, 0755)

	// Create an unrelated fish function the user wrote
	unrelated := filepath.Join(functionsDir, "my-helper.fish")
	os.WriteFile(unrelated, []byte("function my-helper\n  echo hi\nend\n"), 0644)

	// Run configure
	configureShellFishAliases(functionsDir, "/usr/local/bin/claude-orchestrator", []string{"claude-orchestrator"})

	// Unrelated file should still exist
	if _, err := os.Stat(unrelated); err != nil {
		t.Fatal("unrelated user file should be preserved")
	}
}
