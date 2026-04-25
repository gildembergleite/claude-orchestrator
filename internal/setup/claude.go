package setup

import (
	"fmt"
	"os"
	"strings"

	"github.com/gildembergleite/claude-orchestrator/configs"
)

// ConfigureClaude ensures ~/.claude/CLAUDE.md has the memory session section.
func ConfigureClaude(claudeDir, claudeMDPath string) error {
	// Ensure ~/.claude/ exists
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create %s: %w", claudeDir, err)
	}

	existing, _ := os.ReadFile(claudeMDPath)

	content := string(existing)

	// Add memory section if not present
	if !strings.Contains(content, "Memória de Sessão") {
		if len(existing) == 0 {
			content = "# Global Preferences\n"
		}
		content += configs.ClaudeMemorySection
	}

	// Add sessions section if not present
	if !strings.Contains(content, "Regras de memória claude-orchestrator") {
		content += configs.ClaudeSessionsSection
	}

	return os.WriteFile(claudeMDPath, []byte(content), 0644)
}
