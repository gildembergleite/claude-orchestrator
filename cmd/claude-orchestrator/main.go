package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/gildembergleite/claude-orchestrator/internal/claude"
	"github.com/gildembergleite/claude-orchestrator/internal/setup"
	"github.com/gildembergleite/claude-orchestrator/internal/tmux"
	"github.com/gildembergleite/claude-orchestrator/internal/tui"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:     "claude-orchestrator",
		Short:   "Claude Code + tmux session launcher",
		Version: version,
		RunE:    runTUI,
	}

	var noAlias bool
	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Configure tmux, CLAUDE.md, and shell alias",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setup.Run(noAlias)
		},
	}
	setupCmd.Flags().BoolVar(&noAlias, "no-alias", false, "Skip alias prompt, use default 'claude-orchestrator'")

	rootCmd.AddCommand(setupCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runTUI(cmd *cobra.Command, args []string) error {
	tmuxBin, err := tmux.FindBinary()
	if err != nil {
		return err
	}

	claudeBin, err := claude.Resolve(nil)
	if err != nil {
		return err
	}

	insideTmux := tmux.IsInsideTmux()

	// Restore sessions from resurrect if none exist (e.g. after reboot)
	tmux.RestoreIfNeeded(tmuxBin)

	// Run TUI
	app := tui.NewApp(tmuxBin, claudeBin)
	p := tea.NewProgram(app)
	result, err := p.Run()
	if err != nil {
		return err
	}

	m := result.(tui.AppModel)

	switch m.Action {
	case "new":
		claudeCmd := claudeBin
		if len(args) > 0 {
			claudeCmd += " " + strings.Join(args, " ")
		}
		if err := tmux.NewSession(tmuxBin, m.SessionName, m.SessionDir, claudeCmd); err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}
		tmux.RegisterSession(m.SessionName, m.SessionDir)
		if insideTmux {
			return tmux.SwitchSession(tmuxBin, m.SessionName)
		}
		return tmux.AttachSession(tmuxBin, m.SessionName)

	case "attach":
		tmux.TouchSession(m.SessionName)
		if insideTmux {
			return tmux.SwitchSession(tmuxBin, m.SessionName)
		}
		return tmux.AttachSession(tmuxBin, m.SessionName)

	}

	return nil
}
