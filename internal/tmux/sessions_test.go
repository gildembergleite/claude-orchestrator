package tmux

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegisterAndLoadSession(t *testing.T) {
	tmp := t.TempDir()
	orig := sessionsFile
	sessionsFile = filepath.Join(tmp, "sessions.json")
	defer func() { sessionsFile = orig }()

	if err := RegisterSession("backend", "/home/dev/api"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sessions, err := LoadSessions()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sessions["backend"] != "/home/dev/api" {
		t.Fatalf("expected /home/dev/api, got %s", sessions["backend"])
	}
}

func TestUnregisterSession(t *testing.T) {
	tmp := t.TempDir()
	orig := sessionsFile
	sessionsFile = filepath.Join(tmp, "sessions.json")
	defer func() { sessionsFile = orig }()

	RegisterSession("backend", "/home/dev/api")
	RegisterSession("frontend", "/home/dev/app")

	if err := UnregisterSession("backend"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sessions, _ := LoadSessions()
	if _, ok := sessions["backend"]; ok {
		t.Fatal("backend should have been removed")
	}
	if sessions["frontend"] != "/home/dev/app" {
		t.Fatal("frontend should still exist")
	}
}

func TestUnregisterNonExistent(t *testing.T) {
	tmp := t.TempDir()
	orig := sessionsFile
	sessionsFile = filepath.Join(tmp, "sessions.json")
	defer func() { sessionsFile = orig }()

	// Should not error on missing file
	if err := UnregisterSession("nope"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegisterOverwrites(t *testing.T) {
	tmp := t.TempDir()
	orig := sessionsFile
	sessionsFile = filepath.Join(tmp, "sessions.json")
	defer func() { sessionsFile = orig }()

	RegisterSession("backend", "/old/path")
	RegisterSession("backend", "/new/path")

	sessions, _ := LoadSessions()
	if sessions["backend"] != "/new/path" {
		t.Fatalf("expected /new/path, got %s", sessions["backend"])
	}
}

func TestLoadSessionsEmptyFile(t *testing.T) {
	tmp := t.TempDir()
	orig := sessionsFile
	sessionsFile = filepath.Join(tmp, "sessions.json")
	defer func() { sessionsFile = orig }()

	// No file exists
	sessions, err := LoadSessions()
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if sessions != nil {
		t.Fatal("expected nil sessions")
	}
}

func TestSessionsFileCreatesDir(t *testing.T) {
	tmp := t.TempDir()
	orig := sessionsFile
	sessionsFile = filepath.Join(tmp, "subdir", "deep", "sessions.json")
	defer func() { sessionsFile = orig }()

	if err := RegisterSession("test", "/tmp"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "subdir", "deep")); err != nil {
		t.Fatal("directory should have been created")
	}
}
