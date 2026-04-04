package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/zarc-tech/claude-orchestrator/internal/claude"
	"github.com/zarc-tech/claude-orchestrator/internal/tmux"
	"github.com/zarc-tech/claude-orchestrator/internal/tui"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:     "zarc",
		Short:   "Claude Code + tmux session launcher",
		Version: version,
		RunE:    runTUI,
	}

	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Configure tmux, CLAUDE.md, and shell alias",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("zarc setup — coming soon")
			return nil
		},
	}

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

	// If already inside tmux, launch claude directly
	if tmux.IsInsideTmux() {
		fmt.Printf("\033[32m>\033[0m Inside tmux \033[1m%s\033[0m. Launching Claude Code...\n\n",
			tmux.CurrentSessionName(tmuxBin))

		parts := strings.Fields(claudeBin)
		if len(parts) > 1 {
			// npx fallback
			c := exec.Command(parts[0], append(parts[1:], args...)...)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		}
		return tmux.ExecReplace(claudeBin, args...)
	}

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
		return tmux.AttachSession(tmuxBin, m.SessionName)

	case "attach":
		return tmux.AttachSession(tmuxBin, m.SessionName)

	case "kill":
		if err := tmux.KillSession(tmuxBin, m.SessionName); err != nil {
			return fmt.Errorf("failed to kill session: %w", err)
		}
		fmt.Printf("  \033[32mSessão '%s' encerrada.\033[0m\n", m.SessionName)
	}

	return nil
}
