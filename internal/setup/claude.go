package setup

import (
	"fmt"
	"os"
	"strings"

	"github.com/zarc-tech/claude-orchestrator/configs"
)

// ConfigureClaude ensures ~/.claude/CLAUDE.md has the memory session section.
func ConfigureClaude(claudeDir, claudeMDPath string) error {
	// Ensure ~/.claude/ exists
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create %s: %w", claudeDir, err)
	}

	existing, _ := os.ReadFile(claudeMDPath)

	// Check if already configured
	if strings.Contains(string(existing), "Memória de Sessão") {
		return nil
	}

	// If file doesn't exist, create with just the memory section
	if len(existing) == 0 {
		content := "# Global Preferences\n" + configs.ClaudeMemorySection
		return os.WriteFile(claudeMDPath, []byte(content), 0644)
	}

	// Append to existing file
	f, err := os.OpenFile(claudeMDPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", claudeMDPath, err)
	}
	defer f.Close()

	if _, err := f.WriteString(configs.ClaudeMemorySection); err != nil {
		return fmt.Errorf("failed to write to %s: %w", claudeMDPath, err)
	}

	return nil
}
