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
	if len(db.entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(db.entries))
	}
}

func TestNewDirBrowser_HidesHiddenDirs(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, ".hidden"), 0755)
	os.MkdirAll(filepath.Join(tmp, "visible"), 0755)

	db := NewDirBrowser(tmp)
	if len(db.entries) != 1 {
		t.Fatalf("expected 1 visible entry, got %d", len(db.entries))
	}
	if db.entries[0] != "visible" {
		t.Fatalf("expected 'visible', got '%s'", db.entries[0])
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
	if len(filtered) != 2 {
		t.Fatalf("expected 2 filtered entries, got %d", len(filtered))
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
	if len(db.entries) != 1 {
		t.Fatalf("expected 1 entry (grandchild), got %d", len(db.entries))
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
