package setup

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	checkMark = "\033[32m\u2713\033[0m"
	crossMark = "\033[31m\u2717\033[0m"
	warnMark  = "\033[33m\u26a0\033[0m"
)

// Run executes all setup steps in order.
func Run() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	// Step 1: Check dependencies
	fmt.Println()
	deps := CheckDeps()
	allFound := true
	for _, d := range deps {
		if d.Found {
			fmt.Printf(" %s %s (%s)\n", checkMark, d.Name, d.Path)
		} else {
			fmt.Printf(" %s %s — instale com: %s\n", crossMark, d.Name, d.HelpMsg)
			allFound = false
		}
	}
	if !allFound {
		return fmt.Errorf("dependências faltando — instale e rode 'zarc setup' novamente")
	}
	fmt.Printf(" %s Dependências verificadas\n", checkMark)

	// Step 2: Configure tmux
	tmuxDir := filepath.Join(home, ".tmux")
	tmuxConf := filepath.Join(home, ".tmux.conf")
	if err := ConfigureTmux(tmuxDir, tmuxConf); err != nil {
		return fmt.Errorf("tmux configuration failed: %w", err)
	}
	fmt.Printf(" %s tmux configurado (~/.tmux/zarc.conf)\n", checkMark)

	// Step 3: Install tpm + plugins
	pluginsDir := filepath.Join(home, ".tmux", "plugins")
	tpmResult, err := InstallTPM(pluginsDir)
	if err != nil {
		fmt.Printf(" %s tpm — %v (instale manualmente com prefix+I no tmux)\n", warnMark, err)
	} else {
		if tpmResult.AlreadyInstalled {
			fmt.Printf(" %s tpm já instalado\n", checkMark)
		} else {
			fmt.Printf(" %s tpm + plugins instalados\n", checkMark)
		}
	}

	// Step 4: Configure CLAUDE.md
	claudeDir := filepath.Join(home, ".claude")
	claudeMD := filepath.Join(claudeDir, "CLAUDE.md")
	if err := ConfigureClaude(claudeDir, claudeMD); err != nil {
		return fmt.Errorf("CLAUDE.md configuration failed: %w", err)
	}
	fmt.Printf(" %s CLAUDE.md configurado (memória persistente)\n", checkMark)

	// Step 5: Configure shell alias
	zarcBin, _ := os.Executable()
	shellResult, err := ConfigureShellAlias(zarcBin)
	if err != nil {
		fmt.Printf(" %s Alias — %v\n", warnMark, err)
	} else {
		fmt.Printf(" %s Alias configurado (%s)\n", checkMark, shellResult)
	}

	fmt.Printf("\n Pronto! Execute 'zarc' para começar.\n\n")
	return nil
}
