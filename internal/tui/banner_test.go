package tui

import (
	"strings"
	"testing"
)

func TestRenderBanner_ContainsBlockArt(t *testing.T) {
	output := RenderBanner()
	if !strings.Contains(output, "███") {
		t.Fatal("banner should contain ASCII art block characters")
	}
}

func TestRenderBanner_ContainsSubtitle(t *testing.T) {
	output := RenderBanner()
	if !strings.Contains(output, "persistência") {
		t.Fatal("banner should contain subtitle")
	}
}
