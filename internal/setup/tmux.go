package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gildembergleite/claude-orchestrator/configs"
)

// ConfigureTmux creates ~/.tmux/claude-orchestrator.conf and adds a source-file line to tmux.conf.
func ConfigureTmux(tmuxDir, tmuxConfPath string) error {
	// Ensure ~/.tmux/ exists
	if err := os.MkdirAll(tmuxDir, 0755); err != nil {
		return fmt.Errorf("failed to create %s: %w", tmuxDir, err)
	}

	// Write claude-orchestrator.conf
	zarcConfPath := filepath.Join(tmuxDir, "claude-orchestrator.conf")
	if err := os.WriteFile(zarcConfPath, []byte(configs.TmuxConfig), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", zarcConfPath, err)
	}

	// Add source-file to tmux.conf
	sourceLine := fmt.Sprintf("source-file %s", zarcConfPath)

	existing, _ := os.ReadFile(tmuxConfPath)
	if strings.Contains(string(existing), "claude-orchestrator.conf") {
		return nil // already configured
	}

	f, err := os.OpenFile(tmuxConfPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", tmuxConfPath, err)
	}
	defer f.Close()

	content := "\n# ─── claude-orchestrator ────────────────────────────────────────\n"
	content += sourceLine + "\n"

	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("failed to write to %s: %w", tmuxConfPath, err)
	}

	return nil
}
