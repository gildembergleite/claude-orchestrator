package setup

import (
	"fmt"
	"os"
	"os/exec"
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

// ConfigureShellAlias sets up the zarc alias/function for the detected shell.
// Returns a description of what was done, or empty string if skipped.
func ConfigureShellAlias(zarcBin string) (string, error) {
	// Check if zarc is already in PATH
	if path, err := exec.LookPath("zarc"); err == nil && path != "" {
		return "zarc já está no PATH", nil
	}

	shell := DetectShell()
	home, _ := os.UserHomeDir()

	switch shell {
	case "fish":
		dir := filepath.Join(home, ".config", "fish", "functions")
		if err := configureShellFish(dir, zarcBin); err != nil {
			return "", err
		}
		return "fish", nil

	case "zsh":
		rc := filepath.Join(home, ".zshrc")
		if err := configureShellRC(rc, zarcBin); err != nil {
			return "", err
		}
		return "zsh", nil

	case "bash":
		rc := filepath.Join(home, ".bashrc")
		if err := configureShellRC(rc, zarcBin); err != nil {
			return "", err
		}
		return "bash", nil

	default:
		return "", fmt.Errorf("shell não detectado — configure manualmente: alias zarc=\"%s\"", zarcBin)
	}
}

func configureShellFish(functionsDir, zarcBin string) error {
	if err := os.MkdirAll(functionsDir, 0755); err != nil {
		return err
	}

	fishFile := filepath.Join(functionsDir, "zarc.fish")

	// Idempotent check
	if _, err := os.Stat(fishFile); err == nil {
		return nil
	}

	content := fmt.Sprintf(`function zarc --description 'Claude Code + tmux session launcher'
  %s $argv
end
`, zarcBin)

	return os.WriteFile(fishFile, []byte(content), 0644)
}

func configureShellRC(rcPath, zarcBin string) error {
	existing, _ := os.ReadFile(rcPath)

	if strings.Contains(string(existing), "alias zarc=") {
		return nil // already configured
	}

	f, err := os.OpenFile(rcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	line := fmt.Sprintf("\n# zarc — Claude Code orchestrator\nalias zarc=\"%s\"\n", zarcBin)
	_, err = f.WriteString(line)
	return err
}
