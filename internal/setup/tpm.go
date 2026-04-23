package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// TPMResult describes what happened during tpm installation.
type TPMResult struct {
	AlreadyInstalled bool
	Cloned           bool
	PluginsInstalled bool
}

// InstallTPM clones tpm if not present and installs plugins.
func InstallTPM(pluginsDir string) (TPMResult, error) {
	tpmDir := filepath.Join(pluginsDir, "tpm")
	result := TPMResult{}

	// Check if tpm already exists
	if _, err := os.Stat(tpmDir); err == nil {
		result.AlreadyInstalled = true
	} else {
		// Clone tpm
		if err := os.MkdirAll(pluginsDir, 0755); err != nil {
			return result, fmt.Errorf("failed to create plugins dir: %w", err)
		}

		cmd := exec.Command("git", "clone", "https://github.com/tmux-plugins/tpm", tpmDir)
		if err := cmd.Run(); err != nil {
			return result, fmt.Errorf("failed to clone tpm: %w", err)
		}
		result.Cloned = true
	}

	// Install plugins — requires a tmux server running with our config
	installScript := filepath.Join(tpmDir, "bin", "install_plugins")
	if _, err := os.Stat(installScript); err == nil {
		// Start a temporary tmux server to install plugins
		home, _ := os.UserHomeDir()
		tmuxConf := filepath.Join(home, ".tmux.conf")
		tmpSession := "co-tpm-install"

		// Start detached session with our config
		startCmd := exec.Command("tmux", "-f", tmuxConf, "new-session", "-d", "-s", tmpSession)
		if err := startCmd.Run(); err == nil {
			// Give tpm a moment to initialize
			cmd := exec.Command(installScript)
			if err := cmd.Run(); err == nil {
				result.PluginsInstalled = true
			}
			// Kill the temporary session
			exec.Command("tmux", "kill-session", "-t", tmpSession).Run()
		} else {
			// Fallback: try without server (may fail silently)
			cmd := exec.Command(installScript)
			if err := cmd.Run(); err == nil {
				result.PluginsInstalled = true
			}
		}
	}

	return result, nil
}
