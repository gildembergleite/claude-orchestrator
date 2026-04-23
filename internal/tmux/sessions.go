package tmux

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var sessionsFile = filepath.Join(configDir(), "sessions.json")

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "claude-orchestrator")
}

// RegisterSession saves a session name → directory mapping.
func RegisterSession(name, dir string) error {
	sessions, _ := LoadSessions()
	if sessions == nil {
		sessions = make(map[string]string)
	}
	sessions[name] = dir
	return saveSessions(sessions)
}

// UnregisterSession removes a session from the registry.
func UnregisterSession(name string) error {
	sessions, _ := LoadSessions()
	if sessions == nil {
		return nil
	}
	delete(sessions, name)
	return saveSessions(sessions)
}

// LoadSessions reads the sessions registry.
func LoadSessions() (map[string]string, error) {
	data, err := os.ReadFile(sessionsFile)
	if err != nil {
		return nil, err
	}
	var sessions map[string]string
	if err := json.Unmarshal(data, &sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}

func saveSessions(sessions map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(sessionsFile), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(sessionsFile, data, 0644)
}
