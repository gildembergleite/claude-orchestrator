package tmux

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"syscall"
	"time"
)

const storeVersion = 1

type Session struct {
	Name           string            `json:"-"`
	Dir            string            `json:"dir"`
	CreatedAt      time.Time         `json:"created_at"`
	LastAttachedAt time.Time         `json:"last_attached_at"`
	Command        string            `json:"command,omitempty"`
	Env            map[string]string `json:"env,omitempty"`
	Tags           []string          `json:"tags,omitempty"`
	Workspace      string            `json:"workspace,omitempty"`
}

type store struct {
	Version  int                `json:"version"`
	Sessions map[string]Session `json:"sessions"`
}

func configDir() string {
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "claude-orchestrator")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "claude-orchestrator")
}

func sessionsPath() string { return filepath.Join(configDir(), "sessions.json") }
func lockPath() string     { return filepath.Join(configDir(), "sessions.lock") }

func RegisterSession(name, dir string) error {
	return mutate(func(s *store) {
		now := time.Now().UTC()
		existing, ok := s.Sessions[name]
		if !ok {
			s.Sessions[name] = Session{Dir: dir, CreatedAt: now, LastAttachedAt: now}
			return
		}
		existing.Dir = dir
		existing.LastAttachedAt = now
		s.Sessions[name] = existing
	})
}

func UnregisterSession(name string) error {
	return mutate(func(s *store) { delete(s.Sessions, name) })
}

func TouchSession(name string) error {
	return mutate(func(s *store) {
		sess, ok := s.Sessions[name]
		if !ok {
			return
		}
		sess.LastAttachedAt = time.Now().UTC()
		s.Sessions[name] = sess
	})
}

func GetSession(name string) (Session, bool, error) {
	s, err := load()
	if err != nil {
		return Session{}, false, err
	}
	sess, ok := s.Sessions[name]
	if !ok {
		return Session{}, false, nil
	}
	sess.Name = name
	return sess, true, nil
}

func ListRegistered() ([]Session, error) {
	s, err := load()
	if err != nil {
		return nil, err
	}
	out := make([]Session, 0, len(s.Sessions))
	for name, sess := range s.Sessions {
		sess.Name = name
		out = append(out, sess)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].LastAttachedAt.After(out[j].LastAttachedAt)
	})
	return out, nil
}

func mutate(fn func(*store)) error {
	if err := os.MkdirAll(configDir(), 0700); err != nil {
		return fmt.Errorf("mkdir config: %w", err)
	}
	lock, err := acquireLock()
	if err != nil {
		return err
	}
	defer releaseLock(lock)

	s, err := load()
	if err != nil {
		return err
	}
	fn(s)
	return save(s)
}

func load() (*store, error) {
	path := sessionsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &store{Version: storeVersion, Sessions: map[string]Session{}}, nil
		}
		return nil, fmt.Errorf("read sessions: %w", err)
	}

	var s store
	if err := json.Unmarshal(data, &s); err == nil && s.Version >= 1 && s.Sessions != nil {
		return &s, nil
	}

	var legacy map[string]string
	if err := json.Unmarshal(data, &legacy); err == nil {
		now := time.Now().UTC()
		migrated := &store{Version: storeVersion, Sessions: make(map[string]Session, len(legacy))}
		for name, dir := range legacy {
			migrated.Sessions[name] = Session{Dir: dir, CreatedAt: now, LastAttachedAt: now}
		}
		if err := save(migrated); err != nil {
			return nil, fmt.Errorf("persist migration: %w", err)
		}
		return migrated, nil
	}

	return nil, fmt.Errorf("unrecognized sessions.json shape")
}

func save(s *store) error {
	if s.Sessions == nil {
		s.Sessions = map[string]Session{}
	}
	s.Version = storeVersion

	if err := os.MkdirAll(configDir(), 0700); err != nil {
		return fmt.Errorf("mkdir config: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal sessions: %w", err)
	}

	final := sessionsPath()
	tmp, err := os.CreateTemp(filepath.Dir(final), "sessions-*.json.tmp")
	if err != nil {
		return fmt.Errorf("create tmp: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write tmp: %w", err)
	}
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return fmt.Errorf("chmod tmp: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync tmp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close tmp: %w", err)
	}
	if err := os.Rename(tmpName, final); err != nil {
		return fmt.Errorf("rename tmp: %w", err)
	}
	cleanup = false
	return nil
}

func acquireLock() (*os.File, error) {
	f, err := os.OpenFile(lockPath(), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("open lock: %w", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, fmt.Errorf("flock: %w", err)
	}
	return f, nil
}

func releaseLock(f *os.File) {
	syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	f.Close()
}
