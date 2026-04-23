package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// fishMarker identifies fish function files created by claude-orchestrator.
const fishMarker = "Claude Code + tmux session launcher"

// rcMarker identifies alias lines in rc files created by claude-orchestrator.
const rcMarker = "Claude Code orchestrator"

// DetectShell returns the current user's shell: "fish", "zsh", "bash", or "unknown".
func DetectShell() string {
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "fish") {
		return "fish"
	}
	if strings.Contains(shell, "zsh") {
		return "zsh"
	}
	if strings.Contains(shell, "bash") {
		return "bash"
	}
	return "unknown"
}

// ConfigureShellAlias sets up a single claude-orchestrator alias (backward compat).
func ConfigureShellAlias(bin string) (string, error) {
	return ConfigureShellAliases(bin, []string{"claude-orchestrator"})
}

// ConfigureShellAliases sets up shell aliases for the given names.
// Removes any previously-configured aliases that are not in the new list.
func ConfigureShellAliases(bin string, aliases []string) (string, error) {
	shell := DetectShell()
	home, _ := os.UserHomeDir()

	switch shell {
	case "fish":
		dir := filepath.Join(home, ".config", "fish", "functions")
		if err := configureShellFishAliases(dir, bin, aliases); err != nil {
			return "", err
		}
		return "fish", nil

	case "zsh":
		rc := filepath.Join(home, ".zshrc")
		if err := configureShellRCAliases(rc, bin, aliases); err != nil {
			return "", err
		}
		return "zsh", nil

	case "bash":
		rc := filepath.Join(home, ".bashrc")
		if err := configureShellRCAliases(rc, bin, aliases); err != nil {
			return "", err
		}
		return "bash", nil

	default:
		return "", fmt.Errorf("shell não detectado — configure manualmente: alias claude-orchestrator=\"%s\"", bin)
	}
}

// configureShellFish creates a single function file (backward compat for tests).
func configureShellFish(functionsDir, bin string) error {
	return configureShellFishAliases(functionsDir, bin, []string{"claude-orchestrator"})
}

func configureShellFishAliases(functionsDir, bin string, aliases []string) error {
	if err := os.MkdirAll(functionsDir, 0755); err != nil {
		return err
	}

	// Build set of new aliases for quick lookup
	wanted := make(map[string]bool)
	for _, name := range aliases {
		wanted[name] = true
	}

	// Remove existing fish function files that belong to us but aren't in the new list
	entries, _ := os.ReadDir(functionsDir)
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".fish") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".fish")
		if wanted[name] {
			continue // keep
		}
		path := filepath.Join(functionsDir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		// Only remove if it's one of ours (has our marker)
		if strings.Contains(string(content), fishMarker) {
			os.Remove(path)
		}
	}

	// Create or update the desired aliases
	for _, name := range aliases {
		fishFile := filepath.Join(functionsDir, name+".fish")
		content := fmt.Sprintf("function %s --description '%s'\n  %s $argv\nend\n", name, fishMarker, bin)
		if err := os.WriteFile(fishFile, []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

// configureShellRC creates a single alias in an RC file (backward compat for tests).
func configureShellRC(rcPath, bin string) error {
	return configureShellRCAliases(rcPath, bin, []string{"claude-orchestrator"})
}

func configureShellRCAliases(rcPath, bin string, aliases []string) error {
	existing, _ := os.ReadFile(rcPath)
	content := string(existing)

	// Remove existing claude-orchestrator alias blocks
	content = removeOurAliasBlocks(content)

	// Append the new aliases
	for _, name := range aliases {
		line := fmt.Sprintf("\n# %s — %s\nalias %s=\"%s\"\n", name, rcMarker, name, bin)
		content += line
	}

	return os.WriteFile(rcPath, []byte(content), 0644)
}

// removeOurAliasBlocks strips all "# <name> — Claude Code orchestrator" blocks
// (the comment line + the alias line that follows) from the rc file content.
func removeOurAliasBlocks(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	skipNext := false
	for _, line := range lines {
		if skipNext {
			skipNext = false
			continue
		}
		if strings.Contains(line, rcMarker) && strings.HasPrefix(strings.TrimSpace(line), "#") {
			skipNext = true // skip the alias line that follows
			continue
		}
		result = append(result, line)
	}
	// Remove trailing empty lines
	out := strings.Join(result, "\n")
	out = strings.TrimRight(out, "\n") + "\n"
	return out
}
