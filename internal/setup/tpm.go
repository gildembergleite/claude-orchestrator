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

	// Install plugins
	installScript := filepath.Join(tpmDir, "bin", "install_plugins")
	if _, err := os.Stat(installScript); err == nil {
		cmd := exec.Command(installScript)
		if err := cmd.Run(); err != nil {
			// Non-fatal: plugins can be installed later via prefix + I
			return result, nil
		}
		result.PluginsInstalled = true
	}

	return result, nil
}
