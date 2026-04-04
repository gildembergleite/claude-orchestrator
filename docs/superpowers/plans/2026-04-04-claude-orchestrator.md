# claude-orchestrator (zarc) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go CLI called `zarc` that manages Claude Code + tmux sessions with a TUI, and provides a `zarc setup` command for automated environment configuration.

**Architecture:** Cobra-based CLI with two commands: root (TUI via bubbletea) and setup (sequential idempotent steps). The TUI handles session creation/management, while setup configures tmux, tpm, CLAUDE.md, and shell aliases.

**Tech Stack:** Go 1.25, cobra, bubbletea, lipgloss, goreleaser, GitHub Actions

---

## File Map

| File | Responsibility |
|------|---------------|
| `cmd/zarc/main.go` | Entrypoint, cobra root + setup commands |
| `internal/claude/resolver.go` | Locate Claude Code binary on the system |
| `internal/tmux/client.go` | Execute tmux commands (list, new, attach, kill) |
| `internal/tui/banner.go` | ASCII banner rendering with lipgloss |
| `internal/tui/model.go` | Main bubbletea model, state machine, view routing |
| `internal/tui/menu.go` | Reusable menu list component |
| `internal/tui/input.go` | Text input component for directory/session name |
| `internal/setup/setup.go` | Orchestrate all setup steps with status output |
| `internal/setup/deps.go` | Check tmux, claude, git availability |
| `internal/setup/tmux.go` | Write zarc.conf, source-file in tmux.conf |
| `internal/setup/tpm.go` | Clone tpm, install plugins |
| `internal/setup/claude.go` | Append memory section to CLAUDE.md |
| `internal/setup/shell.go` | Detect shell, configure alias/function |
| `configs/zarc.tmux.conf` | Embedded tmux config template |
| `configs/claude-memory.md` | Embedded CLAUDE.md memory section template |
| `.goreleaser.yml` | Build + Homebrew tap config |
| `.github/workflows/ci.yml` | Lint + test on push/PR |
| `.github/workflows/release.yml` | GoReleaser on tag push |
| `Makefile` | Dev shortcuts: build, test, lint, install |

---

### Task 1: Project Bootstrap — go.mod, deps, entrypoint

**Files:**
- Create: `go.mod`
- Create: `cmd/zarc/main.go`
- Create: `Makefile`

- [ ] **Step 1: Initialize Go module**

```bash
cd /Users/gildembergleite/workspace/zarc-claude-orchestrator
go mod init github.com/zarc-tech/claude-orchestrator
```

- [ ] **Step 2: Install dependencies**

```bash
go get github.com/spf13/cobra@latest
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
```

- [ ] **Step 3: Create entrypoint**

Create `cmd/zarc/main.go`:

```go
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
```

- [ ] **Step 4: Create Makefile**

Create `Makefile`:

```makefile
BINARY=zarc
CMD=./cmd/zarc

.PHONY: build run test lint install clean

build:
	go build -o bin/$(BINARY) $(CMD)

run: build
	./bin/$(BINARY)

test:
	go test ./... -v

lint:
	go vet ./...

install: build
	cp bin/$(BINARY) ~/.local/bin/$(BINARY)

clean:
	rm -rf bin/
```

- [ ] **Step 5: Verify build works**

```bash
make build && ./bin/zarc --version
```

Expected: `zarc version dev`

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum cmd/ Makefile
git commit -m "feat: bootstrap project with cobra entrypoint and makefile"
```

---

### Task 2: Claude Code Resolver

**Files:**
- Create: `internal/claude/resolver.go`
- Create: `internal/claude/resolver_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/claude/resolver_test.go`:

```go
package claude

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolve_FindsExistingBinary(t *testing.T) {
	// Create a temp directory with a fake "claude" binary
	tmpDir := t.TempDir()
	fakeBin := filepath.Join(tmpDir, "claude")
	if err := os.WriteFile(fakeBin, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}

	result, err := Resolve([]string{fakeBin})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != fakeBin {
		t.Fatalf("expected %s, got %s", fakeBin, result)
	}
}

func TestResolve_SkipsNonExecutable(t *testing.T) {
	tmpDir := t.TempDir()
	fakeBin := filepath.Join(tmpDir, "claude")
	if err := os.WriteFile(fakeBin, []byte("#!/bin/sh\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Resolve([]string{fakeBin})
	if err == nil {
		t.Fatal("expected error for non-executable, got nil")
	}
}

func TestResolve_ReturnsNpxFallback(t *testing.T) {
	result, err := Resolve([]string{"/nonexistent/path/claude"})
	if err != nil {
		t.Fatalf("expected npx fallback, got error: %v", err)
	}
	if result != "npx @anthropic-ai/claude-code" {
		t.Fatalf("expected npx fallback, got %s", result)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/claude/ -v
```

Expected: FAIL — `Resolve` not defined.

- [ ] **Step 3: Write implementation**

Create `internal/claude/resolver.go`:

```go
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
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/claude/ -v
```

Expected: 3 PASS

- [ ] **Step 5: Commit**

```bash
git add internal/claude/
git commit -m "feat: add claude code binary resolver with tests"
```

---

### Task 3: tmux Client

**Files:**
- Create: `internal/tmux/client.go`
- Create: `internal/tmux/client_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/tmux/client_test.go`:

```go
package tmux

import (
	"testing"
)

func TestIsInsideTmux_False(t *testing.T) {
	t.Setenv("TMUX", "")
	if IsInsideTmux() {
		t.Fatal("expected false when TMUX env is empty")
	}
}

func TestIsInsideTmux_True(t *testing.T) {
	t.Setenv("TMUX", "/tmp/tmux-501/default,12345,0")
	if !IsInsideTmux() {
		t.Fatal("expected true when TMUX env is set")
	}
}

func TestFindBinary_ReturnsPathOrError(t *testing.T) {
	bin, err := FindBinary()
	// On CI or systems without tmux, this will error — that's fine
	if err != nil {
		t.Skipf("tmux not installed, skipping: %v", err)
	}
	if bin == "" {
		t.Fatal("expected non-empty path")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/tmux/ -v
```

Expected: FAIL — functions not defined.

- [ ] **Step 3: Write implementation**

Create `internal/tmux/client.go`:

```go
package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// FindBinary locates the tmux binary.
func FindBinary() (string, error) {
	path, err := exec.LookPath("tmux")
	if err != nil {
		return "", fmt.Errorf("tmux not found — install with: brew install tmux")
	}
	return path, nil
}

// IsInsideTmux returns true if the current process is running inside a tmux session.
func IsInsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

// Session represents a tmux session.
type Session struct {
	Name string
}

// ListSessions returns all active tmux sessions.
func ListSessions(bin string) ([]Session, error) {
	out, err := exec.Command(bin, "ls", "-F", "#{session_name}").Output()
	if err != nil {
		// No server running = no sessions
		return nil, nil
	}
	var sessions []Session
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			sessions = append(sessions, Session{Name: line})
		}
	}
	return sessions, nil
}

// NewSession creates a new tmux session and returns after detaching.
func NewSession(bin, name, dir, command string) error {
	args := []string{"new-session", "-s", name, "-c", dir, "-d", command}
	return exec.Command(bin, args...).Run()
}

// AttachSession attaches to an existing tmux session.
// This replaces the current process.
func AttachSession(bin, name string) error {
	return execReplace(bin, "attach-session", "-t", name)
}

// KillSession kills a tmux session.
func KillSession(bin, name string) error {
	return exec.Command(bin, "kill-session", "-t", name).Run()
}

// CurrentSessionName returns the name of the current tmux session.
func CurrentSessionName(bin string) string {
	out, err := exec.Command(bin, "display-message", "-p", "#S").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
```

- [ ] **Step 4: Create the exec helper for syscall.Exec (unix only)**

Create `internal/tmux/exec_unix.go`:

```go
//go:build !windows

package tmux

import (
	"os"
	"syscall"
)

// execReplace replaces the current process with tmux attach.
func execReplace(bin string, args ...string) error {
	argv := append([]string{bin}, args...)
	return syscall.Exec(bin, argv, os.Environ())
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./internal/tmux/ -v
```

Expected: PASS (or skip if tmux not installed)

- [ ] **Step 6: Commit**

```bash
git add internal/tmux/
git commit -m "feat: add tmux client with session management"
```

---

### Task 4: TUI — Banner

**Files:**
- Create: `internal/tui/banner.go`
- Create: `internal/tui/banner_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/tui/banner_test.go`:

```go
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
	if !strings.Contains(output, "Claude Code") {
		t.Fatal("banner should contain 'Claude Code' subtitle")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/tui/ -v
```

Expected: FAIL — `RenderBanner` not defined.

- [ ] **Step 3: Write implementation**

Create `internal/tui/banner.go`:

```go
package tui

import "github.com/charmbracelet/lipgloss"

var bannerArt = ` ███████╗ █████╗ ██████╗  ██████╗
 ╚══███╔╝██╔══██╗██╔══██╗██╔════╝
   ███╔╝ ███████║██████╔╝██║
  ███╔╝  ██╔══██║██╔══██╗██║
 ███████╗██║  ██║██║  ██║╚██████╗
 ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝`

var (
	bannerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")). // cyan
			Bold(true)

	subtitleStyle = lipgloss.NewStyle().
			Faint(true)
)

// RenderBanner returns the styled ZARC ASCII banner with subtitle.
func RenderBanner() string {
	banner := bannerStyle.Render(bannerArt)
	subtitle := subtitleStyle.Render(" Claude Code · tmux session launcher")
	return banner + "\n" + subtitle + "\n"
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/tui/ -v
```

Expected: 2 PASS

- [ ] **Step 5: Commit**

```bash
git add internal/tui/banner.go internal/tui/banner_test.go
git commit -m "feat: add ZARC ASCII banner with lipgloss styling"
```

---

### Task 5: TUI — Menu Component

**Files:**
- Create: `internal/tui/menu.go`

- [ ] **Step 1: Write the menu component**

Create `internal/tui/menu.go`:

```go
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	selectedStyle   = lipgloss.NewStyle().Bold(true).Reverse(true)
	unselectedStyle = lipgloss.NewStyle().Faint(true)
	headerStyle     = lipgloss.NewStyle().Bold(true)
)

// MenuItem represents an option in the menu.
type MenuItem struct {
	Label string
	ID    string
}

// MenuModel is a reusable arrow-key menu.
type MenuModel struct {
	Title    string
	Items    []MenuItem
	cursor   int
	chosen   bool
	quitting bool
}

// NewMenu creates a new menu with the given title and items.
func NewMenu(title string, items []MenuItem) MenuModel {
	return MenuModel{
		Title: title,
		Items: items,
	}
}

// Chosen returns the selected item, or nil if not yet chosen.
func (m MenuModel) Chosen() *MenuItem {
	if m.chosen && m.cursor < len(m.Items) {
		return &m.Items[m.cursor]
	}
	return nil
}

// IsQuitting returns true if the user pressed q.
func (m MenuModel) IsQuitting() bool {
	return m.quitting
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (MenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.Items)-1 {
				m.cursor++
			}
		case "enter":
			m.chosen = true
			return m, tea.Quit
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m MenuModel) View() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(m.Title))
	b.WriteString("  ")
	b.WriteString(unselectedStyle.Render("| arrows/jk navigate | enter confirm | q quit"))
	b.WriteString("\n\n")

	for i, item := range m.Items {
		if i == m.cursor {
			b.WriteString(selectedStyle.Render(fmt.Sprintf(" > %s ", item.Label)))
		} else {
			b.WriteString(unselectedStyle.Render(fmt.Sprintf("   %s", item.Label)))
		}
		b.WriteString("\n")
	}

	return b.String()
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/tui/
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/tui/menu.go
git commit -m "feat: add reusable arrow-key menu TUI component"
```

---

### Task 6: TUI — Input Component

**Files:**
- Create: `internal/tui/input.go`

- [ ] **Step 1: Write the input component**

Create `internal/tui/input.go`:

```go
package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	promptStyle = lipgloss.NewStyle().Bold(true)
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // yellow
	dimStyle    = lipgloss.NewStyle().Faint(true)
)

// InputModel handles text input with validation.
type InputModel struct {
	Prompt      string
	Placeholder string
	Value       string
	Err         string
	Validate    func(string) error
	done        bool
	quitting    bool
}

// NewInput creates a new input model.
func NewInput(prompt, placeholder string, validate func(string) error) InputModel {
	return InputModel{
		Prompt:      prompt,
		Placeholder: placeholder,
		Validate:    validate,
	}
}

// Done returns true when the user pressed enter with valid input.
func (m InputModel) Done() bool {
	return m.done
}

// IsQuitting returns true if the user pressed ctrl+c.
func (m InputModel) IsQuitting() bool {
	return m.quitting
}

// Result returns the final value (with ~ expanded).
func (m InputModel) Result() string {
	v := m.Value
	if strings.HasPrefix(v, "~/") {
		home, _ := os.UserHomeDir()
		v = filepath.Join(home, v[2:])
	}
	return v
}

func (m InputModel) Init() tea.Cmd {
	return nil
}

func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			value := m.Value
			if value == "" && m.Placeholder != "" {
				value = m.Placeholder
				m.Value = value
			}
			if m.Validate != nil {
				if err := m.Validate(m.Result()); err != nil {
					m.Err = err.Error()
					return m, nil
				}
			}
			m.done = true
			return m, tea.Quit
		case "backspace":
			if len(m.Value) > 0 {
				m.Value = m.Value[:len(m.Value)-1]
			}
			m.Err = ""
		default:
			if len(msg.String()) == 1 {
				m.Value += msg.String()
				m.Err = ""
			}
		}
	}
	return m, nil
}

func (m InputModel) View() string {
	var b strings.Builder

	b.WriteString(promptStyle.Render(m.Prompt))
	if m.Placeholder != "" {
		b.WriteString(" ")
		b.WriteString(dimStyle.Render(fmt.Sprintf("[%s]", m.Placeholder)))
	}
	b.WriteString(": ")
	b.WriteString(m.Value)
	b.WriteString("█\n") // cursor

	if m.Err != "" {
		b.WriteString(errorStyle.Render(fmt.Sprintf("  ⚠ %s", m.Err)))
		b.WriteString("\n")
	}

	return b.String()
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/tui/
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/tui/input.go
git commit -m "feat: add text input TUI component with validation"
```

---

### Task 7: TUI — Main App Model (State Machine)

**Files:**
- Create: `internal/tui/model.go`

- [ ] **Step 1: Write the main TUI model**

Create `internal/tui/model.go`:

```go
package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zarc-tech/claude-orchestrator/internal/tmux"
)

type state int

const (
	stateMainMenu state = iota
	stateNewSessionDir
	stateNewSessionName
	stateSubMenu
	stateConfirmKill
)

var successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // green

// AppModel is the top-level bubbletea model.
type AppModel struct {
	state      state
	tmuxBin    string
	claudeBin  string
	menu       MenuModel
	subMenu    MenuModel
	dirInput   InputModel
	nameInput  InputModel
	sessions   []tmux.Session
	selected   string // selected session name
	err        error
	quitting   bool

	// results to act on after tea.Program exits
	Action     string // "attach", "new", "kill", ""
	SessionDir string
	SessionName string
}

// NewApp creates the app model.
func NewApp(tmuxBin, claudeBin string) AppModel {
	m := AppModel{
		tmuxBin:   tmuxBin,
		claudeBin: claudeBin,
	}
	m.loadMainMenu()
	return m
}

func (m *AppModel) loadMainMenu() {
	sessions, _ := tmux.ListSessions(m.tmuxBin)
	m.sessions = sessions

	items := []MenuItem{{Label: "[+] Nova sessão", ID: "new"}}
	for _, s := range sessions {
		items = append(items, MenuItem{Label: s.Name, ID: "session:" + s.Name})
	}

	m.menu = NewMenu("Selecione uma sessão", items)
	m.state = stateMainMenu
}

func (m AppModel) Init() tea.Cmd {
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateMainMenu:
		return m.updateMainMenu(msg)
	case stateNewSessionDir:
		return m.updateDirInput(msg)
	case stateNewSessionName:
		return m.updateNameInput(msg)
	case stateSubMenu:
		return m.updateSubMenu(msg)
	case stateConfirmKill:
		return m.updateConfirmKill(msg)
	}
	return m, nil
}

func (m AppModel) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.menu, cmd = m.menu.Update(msg)

	if m.menu.IsQuitting() {
		m.quitting = true
		return m, tea.Quit
	}

	if chosen := m.menu.Chosen(); chosen != nil {
		if chosen.ID == "new" {
			m.dirInput = NewInput("Diretório", "", func(v string) error {
				expanded := v
				if strings.HasPrefix(v, "~/") {
					home, _ := os.UserHomeDir()
					expanded = filepath.Join(home, v[2:])
				}
				if expanded == "" {
					return fmt.Errorf("diretório obrigatório")
				}
				info, err := os.Stat(expanded)
				if err != nil {
					return fmt.Errorf("diretório não encontrado: %s", expanded)
				}
				if !info.IsDir() {
					return fmt.Errorf("não é um diretório: %s", expanded)
				}
				return nil
			})
			m.state = stateNewSessionDir
			return m, nil
		}
		// Existing session selected
		m.selected = strings.TrimPrefix(chosen.ID, "session:")
		m.subMenu = NewMenu(
			fmt.Sprintf("Sessão '%s'", m.selected),
			[]MenuItem{
				{Label: "Attach", ID: "attach"},
				{Label: "Kill", ID: "kill"},
				{Label: "Voltar", ID: "back"},
			},
		)
		m.state = stateSubMenu
		return m, nil
	}

	return m, cmd
}

func (m AppModel) updateDirInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.dirInput, cmd = m.dirInput.Update(msg)

	if m.dirInput.IsQuitting() {
		m.quitting = true
		return m, tea.Quit
	}

	if m.dirInput.Done() {
		m.SessionDir = m.dirInput.Result()
		defaultName := filepath.Base(m.SessionDir)
		m.nameInput = NewInput("Nome da sessão", defaultName, nil)
		m.state = stateNewSessionName
		return m, nil
	}

	return m, cmd
}

func (m AppModel) updateNameInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)

	if m.nameInput.IsQuitting() {
		m.quitting = true
		return m, tea.Quit
	}

	if m.nameInput.Done() {
		m.SessionName = m.nameInput.Value
		m.Action = "new"
		return m, tea.Quit
	}

	return m, cmd
}

func (m AppModel) updateSubMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.subMenu, cmd = m.subMenu.Update(msg)

	if m.subMenu.IsQuitting() {
		m.quitting = true
		return m, tea.Quit
	}

	if chosen := m.subMenu.Chosen(); chosen != nil {
		switch chosen.ID {
		case "attach":
			m.SessionName = m.selected
			m.Action = "attach"
			return m, tea.Quit
		case "kill":
			m.subMenu = NewMenu(
				fmt.Sprintf("Confirma kill '%s'?", m.selected),
				[]MenuItem{
					{Label: "Sim", ID: "yes"},
					{Label: "Não", ID: "no"},
				},
			)
			m.state = stateConfirmKill
			return m, nil
		case "back":
			m.loadMainMenu()
			return m, nil
		}
	}

	return m, cmd
}

func (m AppModel) updateConfirmKill(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.subMenu, cmd = m.subMenu.Update(msg)

	if m.subMenu.IsQuitting() {
		m.quitting = true
		return m, tea.Quit
	}

	if chosen := m.subMenu.Chosen(); chosen != nil {
		if chosen.ID == "yes" {
			m.SessionName = m.selected
			m.Action = "kill"
			return m, tea.Quit
		}
		// "no" — go back to main menu
		m.loadMainMenu()
		return m, nil
	}

	return m, cmd
}

func (m AppModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder
	b.WriteString(RenderBanner())
	b.WriteString("\n")

	switch m.state {
	case stateMainMenu:
		b.WriteString(m.menu.View())
	case stateNewSessionDir:
		b.WriteString(m.dirInput.View())
	case stateNewSessionName:
		b.WriteString(m.nameInput.View())
	case stateSubMenu, stateConfirmKill:
		b.WriteString(m.subMenu.View())
	}

	if m.err != nil {
		b.WriteString(errorStyle.Render(m.err.Error()))
		b.WriteString("\n")
	}

	return b.String()
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/tui/
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/tui/model.go
git commit -m "feat: add main TUI state machine with session management flow"
```

---

### Task 8: Wire TUI into Cobra Root Command

**Files:**
- Modify: `cmd/zarc/main.go`

- [ ] **Step 1: Update main.go to run the TUI**

Replace `cmd/zarc/main.go` with:

```go
package main

import (
	"fmt"
	"os"
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
			c := execCommand(parts[0], append(parts[1:], args...)...)
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
```

- [ ] **Step 2: Add ExecReplace as exported function in tmux package**

Add to `internal/tmux/client.go`:

```go
// ExecReplace replaces the current process with the given command.
// Used for launching claude directly when inside tmux.
func ExecReplace(bin string, args ...string) error {
	return execReplace(bin, args...)
}
```

- [ ] **Step 3: Add execCommand helper to main.go**

Add to bottom of `cmd/zarc/main.go`:

```go
func execCommand(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}
```

Add `"os/exec"` to imports.

- [ ] **Step 4: Verify it builds**

```bash
make build
```

Expected: binary compiles successfully

- [ ] **Step 5: Commit**

```bash
git add cmd/zarc/main.go internal/tmux/client.go
git commit -m "feat: wire TUI into cobra root command with session actions"
```

---

### Task 9: Embedded Config Templates

**Files:**
- Create: `configs/zarc.tmux.conf`
- Create: `configs/claude-memory.md`
- Create: `configs/embed.go`

- [ ] **Step 1: Create tmux config template**

Create `configs/zarc.tmux.conf`:

```
# ─── zarc: Terminal compatibility ────────────────────────────────
set -g extended-keys on
set -gs terminal-features 'xterm*:extkeys'

# ─── zarc: Plugins ──────────────────────────────────────────────
set -g @plugin 'tmux-plugins/tpm'
set -g @plugin 'tmux-plugins/tmux-resurrect'
set -g @plugin 'tmux-plugins/tmux-continuum'

# ─── zarc: Resurrect + Continuum ────────────────────────────────
set -g @continuum-restore 'on'
set -g @continuum-save-interval '5'
set -g @resurrect-capture-pane-contents 'on'

# ─── zarc: Initialize tpm (must be last) ────────────────────────
run '~/.tmux/plugins/tpm/tpm'
```

- [ ] **Step 2: Create CLAUDE.md memory section template**

Create `configs/claude-memory.md`:

```markdown

## Memória de Sessão

- **Ao iniciar qualquer nova sessão**, verificar se existe um arquivo de memória para o diretório/repositório atual no sistema de memória (`~/.claude/projects/<project-path>/memory/`).
- Se **não existir**, criar imediatamente um arquivo de memória do tipo `project` com o contexto inicial do repositório (nome do projeto, stack, objetivo geral, branch atual, etc.).
- **A cada iteração relevante na sessão**, atualizar o arquivo de memória do projeto com o que foi feito, decisões tomadas, e contexto importante — garantindo que sessões futuras nunca percam o histórico do que já foi realizado.
- O objetivo é manter continuidade entre sessões: qualquer nova conversa deve poder retomar de onde a anterior parou, sem que o usuário precise re-explicar o contexto.
```

- [ ] **Step 3: Create embed.go to expose templates**

Create `configs/embed.go`:

```go
package configs

import _ "embed"

//go:embed zarc.tmux.conf
var TmuxConfig string

//go:embed claude-memory.md
var ClaudeMemorySection string
```

- [ ] **Step 4: Verify it compiles**

```bash
go build ./configs/
```

Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add configs/
git commit -m "feat: add embedded config templates for tmux and claude memory"
```

---

### Task 10: Setup — Dependency Checker

**Files:**
- Create: `internal/setup/deps.go`
- Create: `internal/setup/deps_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/setup/deps_test.go`:

```go
package setup

import (
	"testing"
)

func TestCheckDeps_ReturnsResults(t *testing.T) {
	results := CheckDeps()
	if len(results) != 3 {
		t.Fatalf("expected 3 dependency checks, got %d", len(results))
	}

	names := map[string]bool{}
	for _, r := range results {
		names[r.Name] = true
		// Each result must have a name and a status
		if r.Name == "" {
			t.Fatal("dependency name should not be empty")
		}
	}

	for _, expected := range []string{"tmux", "claude", "git"} {
		if !names[expected] {
			t.Fatalf("expected dependency check for %s", expected)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/setup/ -v
```

Expected: FAIL — `CheckDeps` not defined.

- [ ] **Step 3: Write implementation**

Create `internal/setup/deps.go`:

```go
package setup

import (
	"github.com/zarc-tech/claude-orchestrator/internal/claude"
	"github.com/zarc-tech/claude-orchestrator/internal/tmux"
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
	gitResult := DepResult{Name: "git"}
	if path, err := findExecutable("git"); err == nil {
		gitResult.Found = true
		gitResult.Path = path
	} else {
		gitResult.HelpMsg = "brew install git"
	}
	results = append(results, gitResult)

	return results
}

func findExecutable(name string) (string, error) {
	import_exec_lookpath(name)
}
```

Wait — let me fix that. Create `internal/setup/deps.go`:

```go
package setup

import (
	"os/exec"

	"github.com/zarc-tech/claude-orchestrator/internal/claude"
	"github.com/zarc-tech/claude-orchestrator/internal/tmux"
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
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/setup/ -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/setup/deps.go internal/setup/deps_test.go
git commit -m "feat: add dependency checker for setup command"
```

---

### Task 11: Setup — tmux Configuration

**Files:**
- Create: `internal/setup/tmux.go`
- Create: `internal/setup/tmux_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/setup/tmux_test.go`:

```go
package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigureTmux_CreatesZarcConf(t *testing.T) {
	tmpHome := t.TempDir()
	tmuxDir := filepath.Join(tmpHome, ".tmux")
	tmuxConf := filepath.Join(tmpHome, ".tmux.conf")

	err := ConfigureTmux(tmuxDir, tmuxConf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check zarc.conf was created
	zarcConf := filepath.Join(tmuxDir, "zarc.conf")
	content, err := os.ReadFile(zarcConf)
	if err != nil {
		t.Fatalf("zarc.conf not created: %v", err)
	}
	if !strings.Contains(string(content), "tmux-resurrect") {
		t.Fatal("zarc.conf should contain resurrect plugin")
	}
}

func TestConfigureTmux_AddsSourceLine(t *testing.T) {
	tmpHome := t.TempDir()
	tmuxDir := filepath.Join(tmpHome, ".tmux")
	tmuxConf := filepath.Join(tmpHome, ".tmux.conf")

	// Create existing tmux.conf
	os.WriteFile(tmuxConf, []byte("# my existing config\n"), 0644)

	err := ConfigureTmux(tmuxDir, tmuxConf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(tmuxConf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "source-file") {
		t.Fatal("tmux.conf should contain source-file line")
	}
	if !strings.Contains(string(content), "my existing config") {
		t.Fatal("tmux.conf should preserve existing content")
	}
}

func TestConfigureTmux_Idempotent(t *testing.T) {
	tmpHome := t.TempDir()
	tmuxDir := filepath.Join(tmpHome, ".tmux")
	tmuxConf := filepath.Join(tmpHome, ".tmux.conf")

	ConfigureTmux(tmuxDir, tmuxConf)
	ConfigureTmux(tmuxDir, tmuxConf) // run again

	content, _ := os.ReadFile(tmuxConf)
	count := strings.Count(string(content), "source-file")
	if count != 1 {
		t.Fatalf("expected 1 source-file line, got %d", count)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/setup/ -v -run TestConfigureTmux
```

Expected: FAIL — `ConfigureTmux` not defined.

- [ ] **Step 3: Write implementation**

Create `internal/setup/tmux.go`:

```go
package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zarc-tech/claude-orchestrator/configs"
)

// ConfigureTmux creates ~/.tmux/zarc.conf and adds a source-file line to tmux.conf.
func ConfigureTmux(tmuxDir, tmuxConfPath string) error {
	// Ensure ~/.tmux/ exists
	if err := os.MkdirAll(tmuxDir, 0755); err != nil {
		return fmt.Errorf("failed to create %s: %w", tmuxDir, err)
	}

	// Write zarc.conf
	zarcConfPath := filepath.Join(tmuxDir, "zarc.conf")
	if err := os.WriteFile(zarcConfPath, []byte(configs.TmuxConfig), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", zarcConfPath, err)
	}

	// Add source-file to tmux.conf
	sourceLine := fmt.Sprintf("source-file %s", zarcConfPath)

	existing, _ := os.ReadFile(tmuxConfPath)
	if strings.Contains(string(existing), "zarc.conf") {
		return nil // already configured
	}

	f, err := os.OpenFile(tmuxConfPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", tmuxConfPath, err)
	}
	defer f.Close()

	content := "\n# ─── zarc orchestrator ────────────────────────────────────────\n"
	content += sourceLine + "\n"

	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("failed to write to %s: %w", tmuxConfPath, err)
	}

	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/setup/ -v -run TestConfigureTmux
```

Expected: 3 PASS

- [ ] **Step 5: Commit**

```bash
git add internal/setup/tmux.go internal/setup/tmux_test.go
git commit -m "feat: add tmux configuration step for setup"
```

---

### Task 12: Setup — tpm Installation

**Files:**
- Create: `internal/setup/tpm.go`
- Create: `internal/setup/tpm_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/setup/tpm_test.go`:

```go
package setup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallTPM_DetectsExisting(t *testing.T) {
	tmpDir := t.TempDir()
	tpmDir := filepath.Join(tmpDir, "plugins", "tpm")
	os.MkdirAll(tpmDir, 0755)

	result, err := InstallTPM(filepath.Join(tmpDir, "plugins"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.AlreadyInstalled {
		t.Fatal("should detect existing tpm installation")
	}
}

func TestInstallTPM_FailsGracefullyWithoutGit(t *testing.T) {
	tmpDir := t.TempDir()
	pluginsDir := filepath.Join(tmpDir, "plugins")

	// This test will actually try to clone if git is available
	// We just verify it doesn't panic
	_, _ = InstallTPM(pluginsDir)
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/setup/ -v -run TestInstallTPM
```

Expected: FAIL — `InstallTPM` not defined.

- [ ] **Step 3: Write implementation**

Create `internal/setup/tpm.go`:

```go
package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// TPMResult describes what happened during tpm installation.
type TPMResult struct {
	AlreadyInstalled bool
	Cloned           bool
	PluginsInstalled bool
}

// InstallTPM clones tpm if not present and installs plugins.
func InstallTPM(pluginsDir string) (TPMResult, error) {
	tpmDir := filepath.Join(pluginsDir, "tpm")
	result := TPMResult{}

	// Check if tpm already exists
	if _, err := os.Stat(tpmDir); err == nil {
		result.AlreadyInstalled = true
	} else {
		// Clone tpm
		if err := os.MkdirAll(pluginsDir, 0755); err != nil {
			return result, fmt.Errorf("failed to create plugins dir: %w", err)
		}

		cmd := exec.Command("git", "clone", "https://github.com/tmux-plugins/tpm", tpmDir)
		if err := cmd.Run(); err != nil {
			return result, fmt.Errorf("failed to clone tpm: %w", err)
		}
		result.Cloned = true
	}

	// Install plugins
	installScript := filepath.Join(tpmDir, "bin", "install_plugins")
	if _, err := os.Stat(installScript); err == nil {
		cmd := exec.Command(installScript)
		if err := cmd.Run(); err != nil {
			// Non-fatal: plugins can be installed later via prefix + I
			return result, nil
		}
		result.PluginsInstalled = true
	}

	return result, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/setup/ -v -run TestInstallTPM
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/setup/tpm.go internal/setup/tpm_test.go
git commit -m "feat: add tpm installation step for setup"
```

---

### Task 13: Setup — CLAUDE.md Configuration

**Files:**
- Create: `internal/setup/claude.go`
- Create: `internal/setup/claude_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/setup/claude_test.go`:

```go
package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigureClaude_CreatesNewFile(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	claudeMD := filepath.Join(claudeDir, "CLAUDE.md")

	err := ConfigureClaude(claudeDir, claudeMD)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatalf("CLAUDE.md not created: %v", err)
	}
	if !strings.Contains(string(content), "Memória de Sessão") {
		t.Fatal("should contain memory section")
	}
}

func TestConfigureClaude_AppendsToExisting(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(claudeDir, 0755)
	claudeMD := filepath.Join(claudeDir, "CLAUDE.md")

	os.WriteFile(claudeMD, []byte("# My Config\n\nSome existing content.\n"), 0644)

	err := ConfigureClaude(claudeDir, claudeMD)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(claudeMD)
	s := string(content)
	if !strings.Contains(s, "My Config") {
		t.Fatal("should preserve existing content")
	}
	if !strings.Contains(s, "Memória de Sessão") {
		t.Fatal("should append memory section")
	}
}

func TestConfigureClaude_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(claudeDir, 0755)
	claudeMD := filepath.Join(claudeDir, "CLAUDE.md")

	ConfigureClaude(claudeDir, claudeMD)
	ConfigureClaude(claudeDir, claudeMD) // run again

	content, _ := os.ReadFile(claudeMD)
	count := strings.Count(string(content), "Memória de Sessão")
	if count != 1 {
		t.Fatalf("expected 1 memory section, got %d", count)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/setup/ -v -run TestConfigureClaude
```

Expected: FAIL — `ConfigureClaude` not defined.

- [ ] **Step 3: Write implementation**

Create `internal/setup/claude.go`:

```go
package setup

import (
	"fmt"
	"os"
	"strings"

	"github.com/zarc-tech/claude-orchestrator/configs"
)

// ConfigureClaude ensures ~/.claude/CLAUDE.md has the memory session section.
func ConfigureClaude(claudeDir, claudeMDPath string) error {
	// Ensure ~/.claude/ exists
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create %s: %w", claudeDir, err)
	}

	existing, _ := os.ReadFile(claudeMDPath)

	// Check if already configured
	if strings.Contains(string(existing), "Memória de Sessão") {
		return nil
	}

	// If file doesn't exist, create with just the memory section
	if len(existing) == 0 {
		content := "# Global Preferences\n" + configs.ClaudeMemorySection
		return os.WriteFile(claudeMDPath, []byte(content), 0644)
	}

	// Append to existing file
	f, err := os.OpenFile(claudeMDPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", claudeMDPath, err)
	}
	defer f.Close()

	if _, err := f.WriteString(configs.ClaudeMemorySection); err != nil {
		return fmt.Errorf("failed to write to %s: %w", claudeMDPath, err)
	}

	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/setup/ -v -run TestConfigureClaude
```

Expected: 3 PASS

- [ ] **Step 5: Commit**

```bash
git add internal/setup/claude.go internal/setup/claude_test.go
git commit -m "feat: add CLAUDE.md memory section configuration step"
```

---

### Task 14: Setup — Shell Alias Configuration

**Files:**
- Create: `internal/setup/shell.go`
- Create: `internal/setup/shell_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/setup/shell_test.go`:

```go
package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectShell_ReturnsShellName(t *testing.T) {
	shell := DetectShell()
	// Should be one of: fish, zsh, bash, unknown
	valid := map[string]bool{"fish": true, "zsh": true, "bash": true, "unknown": true}
	if !valid[shell] {
		t.Fatalf("unexpected shell: %s", shell)
	}
}

func TestConfigureShellAlias_Fish(t *testing.T) {
	tmpDir := t.TempDir()
	functionsDir := filepath.Join(tmpDir, ".config", "fish", "functions")

	err := configureShellFish(functionsDir, "/usr/local/bin/zarc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(functionsDir, "zarc.fish"))
	if err != nil {
		t.Fatal("zarc.fish not created")
	}
	if !strings.Contains(string(content), "/usr/local/bin/zarc") {
		t.Fatal("should contain zarc binary path")
	}
}

func TestConfigureShellAlias_FishIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	functionsDir := filepath.Join(tmpDir, ".config", "fish", "functions")

	configureShellFish(functionsDir, "/usr/local/bin/zarc")
	configureShellFish(functionsDir, "/usr/local/bin/zarc") // again

	content, _ := os.ReadFile(filepath.Join(functionsDir, "zarc.fish"))
	count := strings.Count(string(content), "function zarc")
	if count != 1 {
		t.Fatalf("expected 1 function definition, got %d", count)
	}
}

func TestConfigureShellAlias_Bash(t *testing.T) {
	tmpDir := t.TempDir()
	rcPath := filepath.Join(tmpDir, ".bashrc")
	os.WriteFile(rcPath, []byte("# existing\n"), 0644)

	err := configureShellRC(rcPath, "/usr/local/bin/zarc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(rcPath)
	s := string(content)
	if !strings.Contains(s, "alias zarc=") {
		t.Fatal("should contain alias")
	}
	if !strings.Contains(s, "existing") {
		t.Fatal("should preserve existing content")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/setup/ -v -run TestDetectShell -run TestConfigureShell
```

Expected: FAIL — functions not defined.

- [ ] **Step 3: Write implementation**

Create `internal/setup/shell.go`:

```go
package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DetectShell returns the current user's shell: "fish", "zsh", "bash", or "unknown".
func DetectShell() string {
	// Check SHELL env var
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "fish") {
		return "fish"
	}
	if strings.Contains(shell, "zsh") {
		return "zsh"
	}
	if strings.Contains(shell, "bash") {
		return "bash"
	}
	return "unknown"
}

// ConfigureShellAlias sets up the zarc alias/function for the detected shell.
// Returns a description of what was done, or empty string if skipped.
func ConfigureShellAlias(zarcBin string) (string, error) {
	// Check if zarc is already in PATH
	if path, err := exec.LookPath("zarc"); err == nil && path != "" {
		return "zarc já está no PATH", nil
	}

	shell := DetectShell()
	home, _ := os.UserHomeDir()

	switch shell {
	case "fish":
		dir := filepath.Join(home, ".config", "fish", "functions")
		if err := configureShellFish(dir, zarcBin); err != nil {
			return "", err
		}
		return "fish", nil

	case "zsh":
		rc := filepath.Join(home, ".zshrc")
		if err := configureShellRC(rc, zarcBin); err != nil {
			return "", err
		}
		return "zsh", nil

	case "bash":
		rc := filepath.Join(home, ".bashrc")
		if err := configureShellRC(rc, zarcBin); err != nil {
			return "", err
		}
		return "bash", nil

	default:
		return "", fmt.Errorf("shell não detectado — configure manualmente: alias zarc=\"%s\"", zarcBin)
	}
}

func configureShellFish(functionsDir, zarcBin string) error {
	if err := os.MkdirAll(functionsDir, 0755); err != nil {
		return err
	}

	fishFile := filepath.Join(functionsDir, "zarc.fish")

	// Idempotent check
	if _, err := os.Stat(fishFile); err == nil {
		return nil
	}

	content := fmt.Sprintf(`function zarc --description 'Claude Code + tmux session launcher'
  %s $argv
end
`, zarcBin)

	return os.WriteFile(fishFile, []byte(content), 0644)
}

func configureShellRC(rcPath, zarcBin string) error {
	existing, _ := os.ReadFile(rcPath)

	if strings.Contains(string(existing), "alias zarc=") {
		return nil // already configured
	}

	f, err := os.OpenFile(rcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	line := fmt.Sprintf("\n# zarc — Claude Code orchestrator\nalias zarc=\"%s\"\n", zarcBin)
	_, err = f.WriteString(line)
	return err
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/setup/ -v -run "TestDetectShell|TestConfigureShell"
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/setup/shell.go internal/setup/shell_test.go
git commit -m "feat: add shell alias configuration with fish/zsh/bash support"
```

---

### Task 15: Setup — Orchestrator

**Files:**
- Create: `internal/setup/setup.go`

- [ ] **Step 1: Write the setup orchestrator**

Create `internal/setup/setup.go`:

```go
package setup

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	checkMark = "\033[32m✓\033[0m"
	crossMark = "\033[31m✗\033[0m"
	warnMark  = "\033[33m⚠\033[0m"
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
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/setup/
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/setup/setup.go
git commit -m "feat: add setup orchestrator wiring all configuration steps"
```

---

### Task 16: Wire Setup into Cobra Command

**Files:**
- Modify: `cmd/zarc/main.go`

- [ ] **Step 1: Update the setup command in main.go**

Replace the `setupCmd` definition in `cmd/zarc/main.go`:

```go
	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Configure tmux, CLAUDE.md, and shell alias",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setup.Run()
		},
	}
```

Add import: `"github.com/zarc-tech/claude-orchestrator/internal/setup"`

- [ ] **Step 2: Verify it builds**

```bash
make build
```

Expected: binary compiles

- [ ] **Step 3: Commit**

```bash
git add cmd/zarc/main.go
git commit -m "feat: wire setup command into cobra CLI"
```

---

### Task 17: CI — GitHub Actions

**Files:**
- Create: `.github/workflows/ci.yml`
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: Create CI workflow**

Create `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [dev, main]
  pull_request:
    branches: [dev]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: "1.25"

      - name: Build
        run: go build ./...

      - name: Test
        run: go test ./... -v

      - name: Vet
        run: go vet ./...
```

- [ ] **Step 2: Create release workflow**

Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: "1.25"

      - uses: goreleaser/goreleaser-action@v6
        with:
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          TAP_GITHUB_TOKEN: ${{ secrets.TAP_GITHUB_TOKEN }}
```

- [ ] **Step 3: Commit**

```bash
git add .github/
git commit -m "ci: add GitHub Actions for CI and release"
```

---

### Task 18: GoReleaser Configuration

**Files:**
- Create: `.goreleaser.yml`

- [ ] **Step 1: Create goreleaser config**

Create `.goreleaser.yml`:

```yaml
version: 2

project_name: zarc

before:
  hooks:
    - go mod tidy
    - go test ./...

builds:
  - main: ./cmd/zarc
    binary: zarc
    ldflags:
      - -s -w -X main.version={{.Version}}
    goos:
      - darwin
      - linux
    goarch:
      - amd64
      - arm64

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

brews:
  - repository:
      owner: zarc-tech
      name: homebrew-tools
      token: "{{ .Env.TAP_GITHUB_TOKEN }}"
    directory: Formula
    homepage: "https://github.com/zarc-tech/claude-orchestrator"
    description: "Claude Code + tmux session launcher"
    license: "MIT"
    install: |
      bin.install "zarc"
    test: |
      system "#{bin}/zarc", "--version"

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^ci:"
```

- [ ] **Step 2: Verify goreleaser config is valid**

```bash
go install github.com/goreleaser/goreleaser/v2@latest && goreleaser check
```

If goreleaser is not installed, just verify the yaml is valid:
```bash
cat .goreleaser.yml | python3 -c "import sys, yaml; yaml.safe_load(sys.stdin)"
```

- [ ] **Step 3: Commit**

```bash
git add .goreleaser.yml
git commit -m "ci: add goreleaser config for multi-platform builds and homebrew tap"
```

---

### Task 19: README and .gitignore

**Files:**
- Create: `README.md`
- Create: `.gitignore`

- [ ] **Step 1: Create .gitignore**

Create `.gitignore`:

```
bin/
dist/
*.exe
*.dylib
*.so
.DS_Store
```

- [ ] **Step 2: Create README**

Create `README.md`:

```markdown
# zarc

Claude Code + tmux session launcher.

Gerenciador de sessões tmux com TUI interativo para Claude Code, com configuração automatizada do ambiente de desenvolvimento.

## Instalação

```bash
brew tap zarc-tech/tools
brew install zarc
zarc setup
```

## Uso

```bash
# Abrir o TUI (criar/gerenciar sessões)
zarc

# Configurar o ambiente (tmux, CLAUDE.md, alias)
zarc setup
```

## O que o `zarc setup` configura

1. **Verifica dependências** — tmux, Claude Code, git
2. **Configura tmux** — cria `~/.tmux/zarc.conf` com resurrect + continuum
3. **Instala tpm** — gerenciador de plugins do tmux
4. **Configura CLAUDE.md** — adiciona memória persistente por projeto
5. **Configura alias** — detecta fish/zsh/bash automaticamente

## Desenvolvimento

```bash
make build    # compila o binário
make test     # roda os testes
make run      # compila e executa
make lint     # go vet
```

## Release

```bash
git tag v1.0.0
git push origin v1.0.0
# GitHub Actions roda goreleaser automaticamente
```
```

- [ ] **Step 3: Commit**

```bash
git add .gitignore README.md
git commit -m "docs: add README and .gitignore"
```

---

### Task 20: Integration Test — Full Build and Smoke Test

**Files:** None (verification only)

- [ ] **Step 1: Clean build from scratch**

```bash
rm -rf bin/
make build
```

Expected: binary compiles at `bin/zarc`

- [ ] **Step 2: Verify version flag**

```bash
./bin/zarc --version
```

Expected: `zarc version dev`

- [ ] **Step 3: Verify setup help**

```bash
./bin/zarc setup --help
```

Expected: shows "Configure tmux, CLAUDE.md, and shell alias"

- [ ] **Step 4: Run all tests**

```bash
go test ./... -v
```

Expected: all tests PASS

- [ ] **Step 5: Verify go vet passes**

```bash
go vet ./...
```

Expected: no issues

- [ ] **Step 6: Final commit if any changes needed**

If any fixes were required during smoke testing, commit them:

```bash
git add -A
git commit -m "fix: address issues found during integration testing"
```
