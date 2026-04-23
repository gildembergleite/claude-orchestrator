package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigureTmux_CreatesZarcConf(t *testing.T) {
	tmpHome := t.TempDir()
	tmuxDir := filepath.Join(tmpHome, ".tmux")
	tmuxConf := filepath.Join(tmpHome, ".tmux.conf")

	err := ConfigureTmux(tmuxDir, tmuxConf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check claude-orchestrator.conf was created
	zarcConf := filepath.Join(tmuxDir, "claude-orchestrator.conf")
	content, err := os.ReadFile(zarcConf)
	if err != nil {
		t.Fatalf("claude-orchestrator.conf not created: %v", err)
	}
	if !strings.Contains(string(content), "tmux-resurrect") {
		t.Fatal("claude-orchestrator.conf should contain resurrect plugin")
	}
}

func TestConfigureTmux_AddsSourceLine(t *testing.T) {
	tmpHome := t.TempDir()
	tmuxDir := filepath.Join(tmpHome, ".tmux")
	tmuxConf := filepath.Join(tmpHome, ".tmux.conf")

	// Create existing tmux.conf
	os.WriteFile(tmuxConf, []byte("# my existing config\n"), 0644)

	err := ConfigureTmux(tmuxDir, tmuxConf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(tmuxConf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "source-file") {
		t.Fatal("tmux.conf should contain source-file line")
	}
	if !strings.Contains(string(content), "my existing config") {
		t.Fatal("tmux.conf should preserve existing content")
	}
}

func TestConfigureTmux_Idempotent(t *testing.T) {
	tmpHome := t.TempDir()
	tmuxDir := filepath.Join(tmpHome, ".tmux")
	tmuxConf := filepath.Join(tmpHome, ".tmux.conf")

	ConfigureTmux(tmuxDir, tmuxConf)
	ConfigureTmux(tmuxDir, tmuxConf) // run again

	content, _ := os.ReadFile(tmuxConf)
	count := strings.Count(string(content), "source-file")
	if count != 1 {
		t.Fatalf("expected 1 source-file line, got %d", count)
	}
}
