package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCollapsePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("no home dir: %v", err)
	}

	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{home, "~"},
		{filepath.Join(home, "workspace", "api"), "~" + string(os.PathSeparator) + filepath.Join("workspace", "api")},
		{"/etc/hosts", "/etc/hosts"},
		{"/var/log", "/var/log"},
	}

	for _, tt := range tests {
		got := collapsePath(tt.in)
		if got != tt.want {
			t.Errorf("collapsePath(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFormatIdle(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "agora"},
		{30 * time.Second, "agora"},
		{-5 * time.Second, "agora"},
		{1 * time.Minute, "1m"},
		{59 * time.Minute, "59m"},
		{1 * time.Hour, "1h"},
		{23*time.Hour + 59*time.Minute, "23h"},
		{24 * time.Hour, "1d"},
		{6 * 24 * time.Hour, "6d"},
		{7 * 24 * time.Hour, "1w"},
		{30 * 24 * time.Hour, "4w"},
	}

	for _, tt := range tests {
		got := formatIdle(tt.d)
		if got != tt.want {
			t.Errorf("formatIdle(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestTruncateLeft(t *testing.T) {
	tests := []struct {
		s    string
		max  int
		want string
	}{
		{"", 5, ""},
		{"abc", 5, "abc"},
		{"abcdef", 6, "abcdef"},
		{"abcdef", 5, "...ef"},
		{"abcdef", 4, "...f"},
		{"abcdef", 3, "def"},
		{"abcdef", 2, "ef"},
		{"abcdef", 0, ""},
	}

	for _, tt := range tests {
		got := truncateLeft(tt.s, tt.max)
		if got != tt.want {
			t.Errorf("truncateLeft(%q, %d) = %q, want %q", tt.s, tt.max, got, tt.want)
		}
	}
}

func TestPadRight(t *testing.T) {
	if got := padRight("abc", 5); got != "abc  " {
		t.Errorf("padRight(abc, 5) = %q", got)
	}
	if got := padRight("abcdef", 3); got != "abcdef" {
		t.Errorf("padRight(abcdef, 3) = %q", got)
	}
	if got := padRight("", 3); !strings.HasPrefix(got, "   ") || len(got) != 3 {
		t.Errorf("padRight(empty, 3) = %q", got)
	}
}

func TestPadLeft(t *testing.T) {
	if got := padLeft("abc", 5); got != "  abc" {
		t.Errorf("padLeft(abc, 5) = %q", got)
	}
	if got := padLeft("abcdef", 3); got != "abcdef" {
		t.Errorf("padLeft(abcdef, 3) = %q", got)
	}
}
