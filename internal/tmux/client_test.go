package tmux

import (
	"testing"
)

func TestIsInsideTmux_False(t *testing.T) {
	t.Setenv("TMUX", "")
	if IsInsideTmux() {
		t.Fatal("expected false when TMUX env is empty")
	}
}

func TestIsInsideTmux_True(t *testing.T) {
	t.Setenv("TMUX", "/tmp/tmux-501/default,12345,0")
	if !IsInsideTmux() {
		t.Fatal("expected true when TMUX env is set")
	}
}

func TestFindBinary_ReturnsPathOrError(t *testing.T) {
	bin, err := FindBinary()
	// On CI or systems without tmux, this will error — that's fine
	if err != nil {
		t.Skipf("tmux not installed, skipping: %v", err)
	}
	if bin == "" {
		t.Fatal("expected non-empty path")
	}
}
