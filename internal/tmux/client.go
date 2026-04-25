package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FindBinary locates the tmux binary.
func FindBinary() (string, error) {
	path, err := exec.LookPath("tmux")
	if err != nil {
		return "", fmt.Errorf("tmux not found — install with: brew install tmux")
	}
	return path, nil
}

// IsInsideTmux returns true if the current process is running inside a tmux session.
func IsInsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

// ListSessions returns the names of all active tmux sessions.
func ListSessions(bin string) ([]string, error) {
	out, err := exec.Command(bin, "ls", "-F", "#{session_name}").Output()
	if err != nil {
		// No server running = no sessions
		return nil, nil
	}
	var names []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			names = append(names, line)
		}
	}
	return names, nil
}

// NewSession creates a new tmux session and returns after detaching.
func NewSession(bin, name, dir, command string) error {
	args := []string{"new-session", "-s", name, "-c", dir, "-d", command}
	return exec.Command(bin, args...).Run()
}

// AttachSession attaches to an existing tmux session.
func AttachSession(bin, name string) error {
	cmd := exec.Command(bin, "attach-session", "-t", name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// SwitchSession switches to another session (used when already inside tmux).
func SwitchSession(bin, name string) error {
	return exec.Command(bin, "switch-client", "-t", name).Run()
}

// KillSession kills a tmux session.
func KillSession(bin, name string) error {
	return exec.Command(bin, "kill-session", "-t", name).Run()
}

// RestoreIfNeeded restores tmux sessions from tmux-resurrect save file
// when no sessions exist (e.g. after a reboot).
func RestoreIfNeeded(bin string) {
	sessions, _ := ListSessions(bin)
	if len(sessions) > 0 {
		return
	}

	home, _ := os.UserHomeDir()

	// Find resurrect save file
	resurrectDir := filepath.Join(home, ".local", "share", "tmux", "resurrect")
	lastFile := filepath.Join(resurrectDir, "last")
	if _, err := os.Stat(lastFile); err != nil {
		resurrectDir = filepath.Join(home, ".tmux", "resurrect")
		lastFile = filepath.Join(resurrectDir, "last")
		if _, err := os.Stat(lastFile); err != nil {
			return
		}
	}

	restoreScript := filepath.Join(home, ".tmux", "plugins", "tmux-resurrect", "scripts", "restore.sh")
	if _, err := os.Stat(restoreScript); err != nil {
		return
	}

	// Need a running server to restore — create a temp session
	tmpSession := "co-restore"
	if err := exec.Command(bin, "new-session", "-d", "-s", tmpSession).Run(); err != nil {
		return
	}

	// Set resurrect-dir explicitly so the restore script finds save files
	// (TPM may not have finished loading the plugin yet)
	exec.Command(bin, "set-option", "-g", "@resurrect-dir", resurrectDir).Run()

	// Run resurrect restore (blocks until done)
	exec.Command(bin, "run-shell", restoreScript).Run()

	// Remove temp session
	exec.Command(bin, "kill-session", "-t", tmpSession).Run()
}

// CurrentSessionName returns the name of the current tmux session.
func CurrentSessionName(bin string) string {
	out, err := exec.Command(bin, "display-message", "-p", "#S").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// ExecReplace replaces the current process with the given command.
// Used for launching claude directly when inside tmux.
func ExecReplace(bin string, args ...string) error {
	return execReplace(bin, args...)
}
