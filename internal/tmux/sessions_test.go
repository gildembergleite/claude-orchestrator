package tmux

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
)

func setupTempStore(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	return filepath.Join(tmp, "claude-orchestrator", "sessions.json")
}

func TestRegisterAndGet(t *testing.T) {
	path := setupTempStore(t)

	if err := RegisterSession("backend", "/home/dev/api"); err != nil {
		t.Fatalf("register: %v", err)
	}

	sess, ok, err := GetSession("backend")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !ok {
		t.Fatal("expected session to exist")
	}
	if sess.Name != "backend" || sess.Dir != "/home/dev/api" {
		t.Fatalf("unexpected session: %+v", sess)
	}
	if sess.CreatedAt.IsZero() || sess.LastAttachedAt.IsZero() {
		t.Fatal("timestamps should be populated")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat sessions.json: %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Fatalf("expected mode 0600, got %v", info.Mode().Perm())
	}
}

func TestUnregisterSession(t *testing.T) {
	setupTempStore(t)

	RegisterSession("backend", "/home/dev/api")
	RegisterSession("frontend", "/home/dev/app")

	if err := UnregisterSession("backend"); err != nil {
		t.Fatalf("unregister: %v", err)
	}

	if _, ok, _ := GetSession("backend"); ok {
		t.Fatal("backend should be gone")
	}
	if _, ok, _ := GetSession("frontend"); !ok {
		t.Fatal("frontend should remain")
	}
}

func TestUnregisterNonExistent(t *testing.T) {
	setupTempStore(t)
	if err := UnregisterSession("nope"); err != nil {
		t.Fatalf("unregister non-existent: %v", err)
	}
}

func TestRegisterOverwritesDirAndTouchesAttachment(t *testing.T) {
	setupTempStore(t)

	RegisterSession("backend", "/old/path")
	first, _, _ := GetSession("backend")

	time.Sleep(2 * time.Millisecond)
	RegisterSession("backend", "/new/path")
	second, _, _ := GetSession("backend")

	if second.Dir != "/new/path" {
		t.Fatalf("expected /new/path, got %s", second.Dir)
	}
	if !second.CreatedAt.Equal(first.CreatedAt) {
		t.Fatal("CreatedAt should be preserved on overwrite")
	}
	if !second.LastAttachedAt.After(first.LastAttachedAt) {
		t.Fatal("LastAttachedAt should advance on overwrite")
	}
}

func TestRegisterWithCommand(t *testing.T) {
	setupTempStore(t)

	if err := RegisterSession("a", "/d", WithCommand("hello")); err != nil {
		t.Fatalf("register with command: %v", err)
	}
	sess, _, _ := GetSession("a")
	if sess.Command != "hello" {
		t.Fatalf("expected Command='hello', got %q", sess.Command)
	}
}

func TestRegisterOptionsOverwriteOnUpdate(t *testing.T) {
	setupTempStore(t)

	RegisterSession("a", "/d", WithCommand("first"), WithTags("x", "y"), WithWorkspace("ws1"))
	RegisterSession("a", "/d", WithCommand("second"))

	sess, _, _ := GetSession("a")
	if sess.Command != "second" {
		t.Fatalf("expected Command='second', got %q", sess.Command)
	}
	// Tags e Workspace devem ser preservadas (não passamos opção pra elas)
	if len(sess.Tags) != 2 || sess.Workspace != "ws1" {
		t.Fatalf("expected tags/workspace preserved, got tags=%v ws=%q", sess.Tags, sess.Workspace)
	}
}

func TestRegisterOptionsClearWithEmpty(t *testing.T) {
	setupTempStore(t)

	RegisterSession("a", "/d", WithCommand("set"))
	RegisterSession("a", "/d", WithCommand(""))

	sess, _, _ := GetSession("a")
	if sess.Command != "" {
		t.Fatalf("expected Command cleared, got %q", sess.Command)
	}
}

func TestTouchSession(t *testing.T) {
	setupTempStore(t)

	RegisterSession("backend", "/home/dev/api")
	before, _, _ := GetSession("backend")

	time.Sleep(2 * time.Millisecond)
	if err := TouchSession("backend"); err != nil {
		t.Fatalf("touch: %v", err)
	}
	after, _, _ := GetSession("backend")

	if !after.LastAttachedAt.After(before.LastAttachedAt) {
		t.Fatal("LastAttachedAt should advance on touch")
	}
	if !after.CreatedAt.Equal(before.CreatedAt) {
		t.Fatal("CreatedAt should not change on touch")
	}
}

func TestTouchNonExistentIsNoOp(t *testing.T) {
	setupTempStore(t)
	if err := TouchSession("nope"); err != nil {
		t.Fatalf("touch non-existent: %v", err)
	}
	if _, ok, _ := GetSession("nope"); ok {
		t.Fatal("touch must not create entry")
	}
}

func TestListRegisteredOrderedByLastAttachedDesc(t *testing.T) {
	setupTempStore(t)

	RegisterSession("a", "/a")
	time.Sleep(2 * time.Millisecond)
	RegisterSession("b", "/b")
	time.Sleep(2 * time.Millisecond)
	RegisterSession("c", "/c")

	list, err := ListRegistered()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(list))
	}
	if list[0].Name != "c" || list[1].Name != "b" || list[2].Name != "a" {
		t.Fatalf("expected [c, b, a], got [%s, %s, %s]", list[0].Name, list[1].Name, list[2].Name)
	}
}

func TestEmptyStoreLoadsCleanly(t *testing.T) {
	setupTempStore(t)

	list, err := ListRegistered()
	if err != nil {
		t.Fatalf("list empty: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected empty list, got %d", len(list))
	}
	if _, ok, err := GetSession("nope"); err != nil || ok {
		t.Fatal("get on empty store should be (Session{}, false, nil)")
	}
}

func TestLegacyMigration(t *testing.T) {
	path := setupTempStore(t)

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	legacy := []byte(`{"backend":"/home/dev/api","frontend":"/home/dev/app"}`)
	if err := os.WriteFile(path, legacy, 0644); err != nil {
		t.Fatalf("seed legacy: %v", err)
	}

	list, err := ListRegistered()
	if err != nil {
		t.Fatalf("list after legacy: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 sessions migrated, got %d", len(list))
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read after migration: %v", err)
	}
	var s store
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatalf("unmarshal v1: %v", err)
	}
	if s.Version != 1 || len(s.Sessions) != 2 {
		t.Fatalf("file not v1 after migration: %+v", s)
	}
	for name, sess := range s.Sessions {
		if sess.Dir == "" {
			t.Fatalf("session %s has empty dir", name)
		}
		if sess.CreatedAt.IsZero() || sess.LastAttachedAt.IsZero() {
			t.Fatalf("session %s missing timestamps", name)
		}
	}
}

func TestRegisterIsConcurrencySafe(t *testing.T) {
	setupTempStore(t)

	const n = 20
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			name := []byte("sess-")
			name = append(name, byte('a'+i%26))
			RegisterSession(string(name), "/d")
		}()
	}
	wg.Wait()

	list, err := ListRegistered()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("expected at least one session after concurrent writes")
	}
}

func TestXDGConfigHomeRespected(t *testing.T) {
	setupTempStore(t)

	if err := RegisterSession("x", "/d"); err != nil {
		t.Fatalf("register: %v", err)
	}
	if got := configDir(); got != filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "claude-orchestrator") {
		t.Fatalf("XDG not honored: %s", got)
	}
}

func TestCleanupOrphansRemovesMissingDirs(t *testing.T) {
	setupTempStore(t)

	alive := t.TempDir()
	dead := filepath.Join(t.TempDir(), "gone")

	RegisterSession("live", alive)
	RegisterSession("orphan", dead)

	if _, err := os.Stat(dead); err == nil {
		t.Fatal("dead dir should not exist for this test")
	}

	removed, err := CleanupOrphans()
	if err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	if removed != 1 {
		t.Fatalf("expected 1 removed, got %d", removed)
	}
	if _, ok, _ := GetSession("orphan"); ok {
		t.Fatal("orphan should be gone")
	}
	if _, ok, _ := GetSession("live"); !ok {
		t.Fatal("live should remain")
	}
}

func TestCleanupOrphansEmpty(t *testing.T) {
	setupTempStore(t)

	removed, err := CleanupOrphans()
	if err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	if removed != 0 {
		t.Fatalf("expected 0 removed on empty store, got %d", removed)
	}
}

func TestCleanupOrphansAllAlive(t *testing.T) {
	setupTempStore(t)

	a := t.TempDir()
	b := t.TempDir()
	RegisterSession("a", a)
	RegisterSession("b", b)

	removed, err := CleanupOrphans()
	if err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	if removed != 0 {
		t.Fatalf("expected 0 removed, got %d", removed)
	}
}

func TestNoLeakedTempFiles(t *testing.T) {
	path := setupTempStore(t)
	RegisterSession("a", "/a")
	RegisterSession("b", "/b")
	UnregisterSession("a")

	dir := filepath.Dir(path)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		name := e.Name()
		if name == "sessions.json" || name == "sessions.lock" {
			continue
		}
		t.Fatalf("unexpected residual file: %s", name)
	}
}
