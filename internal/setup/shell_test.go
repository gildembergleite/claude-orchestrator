package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectShell_ReturnsShellName(t *testing.T) {
	shell := DetectShell()
	// Should be one of: fish, zsh, bash, unknown
	valid := map[string]bool{"fish": true, "zsh": true, "bash": true, "unknown": true}
	if !valid[shell] {
		t.Fatalf("unexpected shell: %s", shell)
	}
}

func TestConfigureShellAlias_Fish(t *testing.T) {
	tmpDir := t.TempDir()
	functionsDir := filepath.Join(tmpDir, ".config", "fish", "functions")

	err := configureShellFish(functionsDir, "/usr/local/bin/zarc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(functionsDir, "zarc.fish"))
	if err != nil {
		t.Fatal("zarc.fish not created")
	}
	if !strings.Contains(string(content), "/usr/local/bin/zarc") {
		t.Fatal("should contain zarc binary path")
	}
}

func TestConfigureShellAlias_FishIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	functionsDir := filepath.Join(tmpDir, ".config", "fish", "functions")

	configureShellFish(functionsDir, "/usr/local/bin/zarc")
	configureShellFish(functionsDir, "/usr/local/bin/zarc") // again

	content, _ := os.ReadFile(filepath.Join(functionsDir, "zarc.fish"))
	count := strings.Count(string(content), "function zarc")
	if count != 1 {
		t.Fatalf("expected 1 function definition, got %d", count)
	}
}

func TestConfigureShellAlias_Bash(t *testing.T) {
	tmpDir := t.TempDir()
	rcPath := filepath.Join(tmpDir, ".bashrc")
	os.WriteFile(rcPath, []byte("# existing\n"), 0644)

	err := configureShellRC(rcPath, "/usr/local/bin/zarc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(rcPath)
	s := string(content)
	if !strings.Contains(s, "alias zarc=") {
		t.Fatal("should contain alias")
	}
	if !strings.Contains(s, "existing") {
		t.Fatal("should preserve existing content")
	}
}

func TestConfigureShellFish_MultipleAliases(t *testing.T) {
	tmpDir := t.TempDir()
	functionsDir := filepath.Join(tmpDir, ".config", "fish", "functions")

	err := configureShellFishAliases(functionsDir, "/usr/local/bin/zarc", []string{"zarc", "claude"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, name := range []string{"zarc", "claude"} {
		content, err := os.ReadFile(filepath.Join(functionsDir, name+".fish"))
		if err != nil {
			t.Fatalf("%s.fish not created: %v", name, err)
		}
		if !strings.Contains(string(content), "/usr/local/bin/zarc") {
			t.Fatalf("%s.fish should contain zarc binary path", name)
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

	err := configureShellRCAliases(rcPath, "/usr/local/bin/zarc", []string{"zarc", "claude"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(rcPath)
	s := string(content)
	if !strings.Contains(s, `alias zarc=`) {
		t.Fatal("should contain zarc alias")
	}
	if !strings.Contains(s, `alias claude=`) {
		t.Fatal("should contain claude alias")
	}
}

func TestConfigureShellRC_CustomAlias(t *testing.T) {
	tmpDir := t.TempDir()
	rcPath := filepath.Join(tmpDir, ".zshrc")
	os.WriteFile(rcPath, []byte(""), 0644)

	err := configureShellRCAliases(rcPath, "/usr/local/bin/zarc", []string{"meu-cli"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(rcPath)
	if !strings.Contains(string(content), `alias meu-cli=`) {
		t.Fatal("should contain custom alias")
	}
}
