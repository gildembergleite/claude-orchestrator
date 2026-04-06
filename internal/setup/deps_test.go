package setup

import (
	"testing"
)

func TestCheckDeps_ReturnsResults(t *testing.T) {
	results := CheckDeps()
	if len(results) != 3 {
		t.Fatalf("expected 3 dependency checks, got %d", len(results))
	}

	names := map[string]bool{}
	for _, r := range results {
		names[r.Name] = true
		// Each result must have a name and a status
		if r.Name == "" {
			t.Fatal("dependency name should not be empty")
		}
	}

	for _, expected := range []string{"tmux", "claude", "git"} {
		if !names[expected] {
			t.Fatalf("expected dependency check for %s", expected)
		}
	}
}
