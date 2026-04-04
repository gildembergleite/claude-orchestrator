package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:     "zarc",
		Short:   "Claude Code + tmux session launcher",
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("zarc TUI — coming soon")
			return nil
		},
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
