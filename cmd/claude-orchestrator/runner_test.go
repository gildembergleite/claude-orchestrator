package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gildembergleite/claude-orchestrator/internal/tmux"
)

func setupStore(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
}

func TestShellQuote(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", "''"},
		{"hello", "'hello'"},
		{"hello world", "'hello world'"},
		{`it's`, `'it'\''s'`},
		{`$(rm -rf /)`, `'$(rm -rf /)'`},
		{"line1\nline2", "'line1\nline2'"},
		{`"`, `'"'`},
		{`a;b`, `'a;b'`},
	}
	for _, tt := range tests {
		got := shellQuote(tt.in)
		if got != tt.want {
			t.Errorf("shellQuote(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestRunListEmptyStorePretty(t *testing.T) {
	setupStore(t)
	var buf bytes.Buffer
	if err := runList(&buf, false); err != nil {
		t.Fatalf("runList: %v", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", buf.String())
	}
}

func TestRunListEmptyStoreJSON(t *testing.T) {
	setupStore(t)
	var buf bytes.Buffer
	if err := runList(&buf, true); err != nil {
		t.Fatalf("runList: %v", err)
	}
	var parsed []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}
	if len(parsed) != 0 {
		t.Fatalf("expected empty array, got %d items", len(parsed))
	}
}

func TestRunListPopulatedJSON(t *testing.T) {
	setupStore(t)
	tmux.RegisterSession("backend", "/dev/api", tmux.WithCommand("init"))
	tmux.RegisterSession("frontend", "/dev/app", tmux.WithTags("ui", "react"))

	var buf bytes.Buffer
	if err := runList(&buf, true); err != nil {
		t.Fatalf("runList json: %v", err)
	}
	var parsed []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(parsed) != 2 {
		t.Fatalf("expected 2 items, got %d", len(parsed))
	}
	names := map[string]bool{}
	for _, item := range parsed {
		names[item["name"].(string)] = true
	}
	if !names["backend"] || !names["frontend"] {
		t.Fatalf("missing expected names: %v", names)
	}
}

func TestRunListPopulatedPretty(t *testing.T) {
	setupStore(t)
	tmux.RegisterSession("alpha", "/tmp/a")
	tmux.RegisterSession("beta", "/tmp/b")

	var buf bytes.Buffer
	if err := runList(&buf, false); err != nil {
		t.Fatalf("runList: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
		t.Fatalf("expected names in output: %s", out)
	}
	if !strings.Contains(out, "/tmp/a") && !strings.Contains(out, "~") {
		t.Fatalf("expected dir in output: %s", out)
	}
}

func TestRunNewMissingDir(t *testing.T) {
	setupStore(t)
	err := runNew(NewArgs{TmuxBin: "tmux", ClaudeBin: "claude"})
	if err == nil || !strings.Contains(err.Error(), "--dir") {
		t.Fatalf("expected --dir error, got %v", err)
	}
}

func TestRunNewInvalidDir(t *testing.T) {
	setupStore(t)
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	err := runNew(NewArgs{TmuxBin: "tmux", ClaudeBin: "claude", Dir: missing})
	if err == nil {
		t.Fatal("expected error for missing dir")
	}
}

func TestRunKillNonExistentIsIdempotent(t *testing.T) {
	setupStore(t)
	if err := runKill("/nonexistent/tmux", "nope"); err != nil {
		t.Fatalf("kill should be idempotent: %v", err)
	}
}

func TestCollapseHome(t *testing.T) {
	if collapseHome("") != "" {
		t.Fatal("empty input should return empty")
	}
	if got := collapseHome("/etc/hosts"); got != "/etc/hosts" {
		t.Errorf("expected unchanged, got %q", got)
	}
}
