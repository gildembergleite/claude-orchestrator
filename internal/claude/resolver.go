package claude

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Resolve finds the Claude Code binary by checking candidate paths in order.
// Pass nil for candidates to use the default search paths.
func Resolve(candidates []string) (string, error) {
	if candidates == nil {
		home, _ := os.UserHomeDir()
		candidates = []string{
			filepath.Join(home, ".npm-global", "bin", "claude"),
			filepath.Join(home, ".local", "bin", "claude"),
			"/usr/local/bin/claude",
		}

		// Check npm global root
		out, err := exec.Command("npm", "root", "-g").Output()
		if err == nil {
			npmBin := filepath.Join(strings.TrimSpace(string(out)), ".bin", "claude")
			candidates = append(candidates, npmBin)
		}
	}

	self, _ := os.Executable()
	selfReal, _ := filepath.EvalSymlinks(self)

	for _, bin := range candidates {
		info, err := os.Stat(bin)
		if err != nil {
			continue
		}
		if info.Mode()&0111 == 0 {
			continue
		}
		// Don't resolve to ourselves
		binReal, _ := filepath.EvalSymlinks(bin)
		if binReal == selfReal {
			continue
		}
		return bin, nil
	}

	// Fallback to npx
	if _, err := exec.LookPath("npx"); err == nil {
		return "npx @anthropic-ai/claude-code", nil
	}

	return "", fmt.Errorf("claude code not found — install with: npm install -g @anthropic-ai/claude-code")
}
