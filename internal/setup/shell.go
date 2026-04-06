package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DetectShell returns the current user's shell: "fish", "zsh", "bash", or "unknown".
func DetectShell() string {
	// Check SHELL env var
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

// ConfigureShellAlias sets up a single zarc alias (backward compat).
func ConfigureShellAlias(zarcBin string) (string, error) {
	return ConfigureShellAliases(zarcBin, []string{"zarc"})
}

// ConfigureShellAliases sets up shell aliases for the given names.
// Each name in aliases will point to zarcBin.
func ConfigureShellAliases(zarcBin string, aliases []string) (string, error) {
	shell := DetectShell()
	home, _ := os.UserHomeDir()

	switch shell {
	case "fish":
		dir := filepath.Join(home, ".config", "fish", "functions")
		if err := configureShellFishAliases(dir, zarcBin, aliases); err != nil {
			return "", err
		}
		return "fish", nil

	case "zsh":
		rc := filepath.Join(home, ".zshrc")
		if err := configureShellRCAliases(rc, zarcBin, aliases); err != nil {
			return "", err
		}
		return "zsh", nil

	case "bash":
		rc := filepath.Join(home, ".bashrc")
		if err := configureShellRCAliases(rc, zarcBin, aliases); err != nil {
			return "", err
		}
		return "bash", nil

	default:
		return "", fmt.Errorf("shell não detectado — configure manualmente: alias zarc=\"%s\"", zarcBin)
	}
}

// configureShellFish creates a single zarc.fish function file (backward compat for tests).
func configureShellFish(functionsDir, zarcBin string) error {
	return configureShellFishAliases(functionsDir, zarcBin, []string{"zarc"})
}

func configureShellFishAliases(functionsDir, zarcBin string, aliases []string) error {
	if err := os.MkdirAll(functionsDir, 0755); err != nil {
		return err
	}

	for _, name := range aliases {
		fishFile := filepath.Join(functionsDir, name+".fish")
		if _, err := os.Stat(fishFile); err == nil {
			continue // already exists
		}
		content := fmt.Sprintf("function %s --description 'Claude Code + tmux session launcher'\n  %s $argv\nend\n", name, zarcBin)
		if err := os.WriteFile(fishFile, []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

// configureShellRC creates a single zarc alias in an RC file (backward compat for tests).
func configureShellRC(rcPath, zarcBin string) error {
	return configureShellRCAliases(rcPath, zarcBin, []string{"zarc"})
}

func configureShellRCAliases(rcPath, zarcBin string, aliases []string) error {
	existing, _ := os.ReadFile(rcPath)
	content := string(existing)

	f, err := os.OpenFile(rcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, name := range aliases {
		marker := fmt.Sprintf("alias %s=", name)
		if strings.Contains(content, marker) {
			continue // already configured
		}
		line := fmt.Sprintf("\n# %s — Claude Code orchestrator\nalias %s=\"%s\"\n", name, name, zarcBin)
		if _, err := f.WriteString(line); err != nil {
			return err
		}
	}
	return nil
}
