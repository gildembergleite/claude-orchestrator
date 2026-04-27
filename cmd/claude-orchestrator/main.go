package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gildembergleite/claude-orchestrator/internal/claude"
	"github.com/gildembergleite/claude-orchestrator/internal/setup"
	"github.com/gildembergleite/claude-orchestrator/internal/tmux"
	"github.com/gildembergleite/claude-orchestrator/internal/tui"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:     "claude-orchestrator",
		Short:   "Claude Code + tmux session launcher",
		Version: version,
		RunE:    runTUI,
	}

	rootCmd.AddCommand(newSetupCmd())
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newNewCmd())
	rootCmd.AddCommand(newAttachCmd())
	rootCmd.AddCommand(newKillCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newSetupCmd() *cobra.Command {
	var noAlias bool
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Configure tmux, CLAUDE.md, and shell alias",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setup.Run(noAlias)
		},
	}
	cmd.Flags().BoolVar(&noAlias, "no-alias", false, "Skip alias prompt, use default 'claude-orchestrator'")
	return cmd
}

func newListCmd() *cobra.Command {
	var jsonMode bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista sessões registradas",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(os.Stdout, jsonMode)
		},
	}
	cmd.Flags().BoolVar(&jsonMode, "json", false, "Saída em JSON")
	return cmd
}

func newNewCmd() *cobra.Command {
	var dir, name, prompt string
	cmd := &cobra.Command{
		Use:   "new",
		Short: "Cria sessão tmux + Claude sem TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			tmuxBin, err := tmux.FindBinary()
			if err != nil {
				return err
			}
			claudeBin, err := claude.Resolve(nil)
			if err != nil {
				return err
			}
			return runNew(NewArgs{
				TmuxBin: tmuxBin, ClaudeBin: claudeBin,
				Dir: dir, Name: name, Prompt: prompt,
			})
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "Diretório da sessão (obrigatório)")
	cmd.Flags().StringVar(&name, "name", "", "Nome da sessão (default: basename de --dir)")
	cmd.Flags().StringVar(&prompt, "prompt", "", "Prompt inicial passado ao claude")
	cmd.MarkFlagRequired("dir")
	return cmd
}

func newAttachCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "attach <name>",
		Short: "Attach numa sessão existente",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tmuxBin, err := tmux.FindBinary()
			if err != nil {
				return err
			}
			return runAttach(tmuxBin, args[0])
		},
	}
}

func newKillCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "kill <name>",
		Short: "Mata sessão tmux e desregistra",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tmuxBin, err := tmux.FindBinary()
			if err != nil {
				return err
			}
			return runKill(tmuxBin, args[0])
		},
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

	tmux.RestoreIfNeeded(tmuxBin)

	app := tui.NewApp(tmuxBin, claudeBin)
	p := tea.NewProgram(app, tea.WithAltScreen())
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
