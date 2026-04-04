package tmux

import (
	"fmt"
	"os"
	"os/exec"
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

// Session represents a tmux session.
type Session struct {
	Name string
}

// ListSessions returns all active tmux sessions.
func ListSessions(bin string) ([]Session, error) {
	out, err := exec.Command(bin, "ls", "-F", "#{session_name}").Output()
	if err != nil {
		// No server running = no sessions
		return nil, nil
	}
	var sessions []Session
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			sessions = append(sessions, Session{Name: line})
		}
	}
	return sessions, nil
}

// NewSession creates a new tmux session and returns after detaching.
func NewSession(bin, name, dir, command string) error {
	args := []string{"new-session", "-s", name, "-c", dir, "-d", command}
	return exec.Command(bin, args...).Run()
}

// AttachSession attaches to an existing tmux session.
// This replaces the current process.
func AttachSession(bin, name string) error {
	return execReplace(bin, "attach-session", "-t", name)
}

// KillSession kills a tmux session.
func KillSession(bin, name string) error {
	return exec.Command(bin, "kill-session", "-t", name).Run()
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
