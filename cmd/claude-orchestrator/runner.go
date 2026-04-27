package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gildembergleite/claude-orchestrator/internal/tmux"
)

type NewArgs struct {
	TmuxBin   string
	ClaudeBin string
	Dir       string
	Name      string
	Prompt    string
}

func runList(out io.Writer, jsonMode bool) error {
	list, err := tmux.ListRegistered()
	if err != nil {
		return fmt.Errorf("list registered: %w", err)
	}

	if jsonMode {
		type entry struct {
			Name           string            `json:"name"`
			Dir            string            `json:"dir"`
			CreatedAt      time.Time         `json:"created_at"`
			LastAttachedAt time.Time         `json:"last_attached_at"`
			Command        string            `json:"command,omitempty"`
			Env            map[string]string `json:"env,omitempty"`
			Tags           []string          `json:"tags,omitempty"`
			Workspace      string            `json:"workspace,omitempty"`
		}
		items := make([]entry, 0, len(list))
		for _, s := range list {
			items = append(items, entry{
				Name: s.Name, Dir: s.Dir,
				CreatedAt: s.CreatedAt, LastAttachedAt: s.LastAttachedAt,
				Command: s.Command, Env: s.Env, Tags: s.Tags, Workspace: s.Workspace,
			})
		}
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(items)
	}

	if len(list) == 0 {
		fmt.Fprintln(os.Stderr, "(nenhuma sessão registrada)")
		return nil
	}

	nameWidth := 4
	for _, s := range list {
		if len(s.Name) > nameWidth {
			nameWidth = len(s.Name)
		}
	}
	now := time.Now()
	for _, s := range list {
		idle := formatIdleSimple(now.Sub(s.LastAttachedAt))
		fmt.Fprintf(out, "%-*s  %-40s  %s\n", nameWidth, s.Name, collapseHome(s.Dir), idle)
	}
	return nil
}

func runNew(args NewArgs) error {
	if args.Dir == "" {
		return fmt.Errorf("--dir é obrigatório")
	}
	abs, err := filepath.Abs(args.Dir)
	if err != nil {
		return fmt.Errorf("resolver dir: %w", err)
	}
	if info, err := os.Stat(abs); err != nil {
		return fmt.Errorf("dir inválido: %w", err)
	} else if !info.IsDir() {
		return fmt.Errorf("não é um diretório: %s", abs)
	}

	name := args.Name
	if name == "" {
		name = filepath.Base(abs)
	}

	live, _ := tmux.ListSessions(args.TmuxBin)
	for _, n := range live {
		if n == name {
			return fmt.Errorf("sessão tmux '%s' já existe — use attach ou kill", name)
		}
	}

	claudeCmd := args.ClaudeBin
	if args.Prompt != "" {
		claudeCmd = fmt.Sprintf("%s %s", args.ClaudeBin, shellQuote(args.Prompt))
	}

	if err := tmux.NewSession(args.TmuxBin, name, abs, claudeCmd); err != nil {
		return fmt.Errorf("criar sessão tmux: %w", err)
	}

	opts := []tmux.RegisterOption{}
	if args.Prompt != "" {
		opts = append(opts, tmux.WithCommand(args.Prompt))
	}
	if err := tmux.RegisterSession(name, abs, opts...); err != nil {
		return fmt.Errorf("registrar sessão: %w", err)
	}

	if tmux.IsInsideTmux() {
		return tmux.SwitchSession(args.TmuxBin, name)
	}
	return tmux.AttachSession(args.TmuxBin, name)
}

func runAttach(tmuxBin, name string) error {
	live, _ := tmux.ListSessions(tmuxBin)
	found := false
	for _, n := range live {
		if n == name {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("sessão tmux '%s' não existe", name)
	}

	tmux.TouchSession(name)

	if tmux.IsInsideTmux() {
		return tmux.SwitchSession(tmuxBin, name)
	}
	return tmux.AttachSession(tmuxBin, name)
}

func runKill(tmuxBin, name string) error {
	tmux.KillSession(tmuxBin, name)
	return tmux.UnregisterSession(name)
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func collapseHome(p string) string {
	if p == "" {
		return ""
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return p
	}
	if p == home {
		return "~"
	}
	if strings.HasPrefix(p, home+string(os.PathSeparator)) {
		return "~" + p[len(home):]
	}
	return p
}

func formatIdleSimple(d time.Duration) string {
	if d < time.Minute {
		return "agora"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	days := int(d.Hours() / 24)
	if days < 7 {
		return fmt.Sprintf("%dd", days)
	}
	return fmt.Sprintf("%dw", days/7)
}
