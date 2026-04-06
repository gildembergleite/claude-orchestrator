package tui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewDirBrowser_UsesInitialDir(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, "subdir-a"), 0755)
	os.MkdirAll(filepath.Join(tmp, "subdir-b"), 0755)

	db := NewDirBrowser(tmp)
	if db.currentDir != tmp {
		t.Fatalf("expected currentDir=%s, got %s", tmp, db.currentDir)
	}
	// entries = selectEntry + ".." + 2 subdirs
	if len(db.entries) != 4 {
		t.Fatalf("expected 4 entries, got %d: %v", len(db.entries), db.entries)
	}
	if db.entries[0] != selectEntry {
		t.Fatalf("first entry should be selectEntry, got '%s'", db.entries[0])
	}
}

func TestNewDirBrowser_HidesHiddenDirs(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, ".hidden"), 0755)
	os.MkdirAll(filepath.Join(tmp, "visible"), 0755)

	db := NewDirBrowser(tmp)
	// entries = selectEntry + ".." + visible
	if len(db.entries) != 3 {
		t.Fatalf("expected 3 entries, got %d: %v", len(db.entries), db.entries)
	}
	if db.entries[2] != "visible" {
		t.Fatalf("expected 'visible', got '%s'", db.entries[2])
	}
}

func TestDirBrowser_FilterByPrefix(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, "alpha"), 0755)
	os.MkdirAll(filepath.Join(tmp, "beta"), 0755)
	os.MkdirAll(filepath.Join(tmp, "alpha-two"), 0755)

	db := NewDirBrowser(tmp)
	db.filter = "al"
	filtered := db.filteredEntries()
	// selectEntry + ".." + alpha + alpha-two
	if len(filtered) != 4 {
		t.Fatalf("expected 4 filtered entries, got %d: %v", len(filtered), filtered)
	}
}

func TestDirBrowser_DrillDown(t *testing.T) {
	tmp := t.TempDir()
	child := filepath.Join(tmp, "child")
	os.MkdirAll(filepath.Join(child, "grandchild"), 0755)

	db := NewDirBrowser(tmp)
	db.enterSelected("child")

	if db.currentDir != child {
		t.Fatalf("expected currentDir=%s, got %s", child, db.currentDir)
	}
	// entries = selectEntry + ".." + grandchild
	if len(db.entries) != 3 {
		t.Fatalf("expected 3 entries, got %d: %v", len(db.entries), db.entries)
	}
}

func TestDirBrowser_GoUp(t *testing.T) {
	tmp := t.TempDir()
	child := filepath.Join(tmp, "child")
	os.MkdirAll(child, 0755)

	db := NewDirBrowser(child)
	db.goUp()

	if db.currentDir != tmp {
		t.Fatalf("expected currentDir=%s, got %s", tmp, db.currentDir)
	}
}

func TestDirBrowser_Result(t *testing.T) {
	tmp := t.TempDir()
	db := NewDirBrowser(tmp)
	if db.Result() != tmp {
		t.Fatalf("expected Result()=%s, got %s", tmp, db.Result())
	}
}

func TestDirBrowser_SelectEntryConfirms(t *testing.T) {
	tmp := t.TempDir()
	db := NewDirBrowser(tmp)
	// cursor starts at 0 which is selectEntry
	if db.entries[db.cursor] != selectEntry {
		t.Fatalf("cursor should start on selectEntry")
	}
}
