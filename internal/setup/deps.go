package setup

import (
	"os/exec"

	"github.com/zarc-tech/zarc-claude-orchestrator/internal/claude"
	"github.com/zarc-tech/zarc-claude-orchestrator/internal/tmux"
)

// DepResult represents the result of a dependency check.
type DepResult struct {
	Name    string
	Found   bool
	Path    string
	HelpMsg string
}

// CheckDeps checks if required dependencies are installed.
func CheckDeps() []DepResult {
	results := make([]DepResult, 0, 3)

	// tmux
	tmuxBin, err := tmux.FindBinary()
	if err != nil {
		results = append(results, DepResult{
			Name: "tmux", Found: false,
			HelpMsg: "brew install tmux",
		})
	} else {
		results = append(results, DepResult{
			Name: "tmux", Found: true, Path: tmuxBin,
		})
	}

	// claude
	claudeBin, err := claude.Resolve(nil)
	if err != nil {
		results = append(results, DepResult{
			Name: "claude", Found: false,
			HelpMsg: "npm install -g @anthropic-ai/claude-code",
		})
	} else {
		results = append(results, DepResult{
			Name: "claude", Found: true, Path: claudeBin,
		})
	}

	// git
	gitPath, err := exec.LookPath("git")
	if err != nil {
		results = append(results, DepResult{
			Name: "git", Found: false,
			HelpMsg: "brew install git",
		})
	} else {
		results = append(results, DepResult{
			Name: "git", Found: true, Path: gitPath,
		})
	}

	return results
}
