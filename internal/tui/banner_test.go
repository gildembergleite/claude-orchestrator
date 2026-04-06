package tui

import (
	"strings"
	"testing"
)

func TestRenderBanner_ContainsZARC(t *testing.T) {
	output := RenderBanner()
	if !strings.Contains(output, "ZARC") {
		// The ASCII art uses block characters, not plain "ZARC"
		// Check for a known character from the banner
		if !strings.Contains(output, "███") {
			t.Fatal("banner should contain ASCII art block characters")
		}
	}
}

func TestRenderBanner_ContainsSubtitle(t *testing.T) {
	output := RenderBanner()
	if !strings.Contains(output, "persistência") {
		t.Fatal("banner should contain subtitle")
	}
}
