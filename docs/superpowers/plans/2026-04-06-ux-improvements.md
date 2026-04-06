# UX Improvements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add custom alias setup, visual directory browser, and default directory to zarc CLI.

**Architecture:** Three independent features touching setup (alias choice) and TUI (dir browser + default dir). The dir browser is a new Bubble Tea component (`DirBrowserModel`) that replaces `InputModel` for directory selection. The alias feature adds an interactive prompt to the setup flow.

**Tech Stack:** Go, Bubble Tea, lipgloss, Cobra

---

## File Structure

| File | Action | Responsibility |
|------|--------|----------------|
| `internal/setup/shell.go` | Modify | Accept `[]string` aliases, create multiple aliases per shell |
| `internal/setup/shell_test.go` | Modify | Tests for multi-alias configuration |
| `internal/setup/setup.go` | Modify | Prompt user for alias choice before shell config |
| `internal/tui/dirbrowser.go` | Create | `DirBrowserModel` — visual dir navigator with drill-down |
| `internal/tui/dirbrowser_test.go` | Create | Unit tests for dir browser |
| `internal/tui/model.go` | Modify | Wire `DirBrowserModel` into `stateNewSessionDir`, pass `os.Getwd()` |

---

### Task 1: Refactor shell.go to accept multiple aliases

**Files:**
- Modify: `internal/setup/shell.go`
- Modify: `internal/setup/shell_test.go`

- [ ] **Step 1: Write failing tests for multi-alias fish config**

Add to `internal/setup/shell_test.go`:

```go
func TestConfigureShellFish_MultipleAliases(t *testing.T) {
	tmpDir := t.TempDir()
	functionsDir := filepath.Join(tmpDir, ".config", "fish", "functions")

	err := configureShellFishAliases(functionsDir, "/usr/local/bin/zarc", []string{"zarc", "claude"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, name := range []string{"zarc", "claude"} {
		content, err := os.ReadFile(filepath.Join(functionsDir, name+".fish"))
		if err != nil {
			t.Fatalf("%s.fish not created: %v", name, err)
		}
		if !strings.Contains(string(content), "/usr/local/bin/zarc") {
			t.Fatalf("%s.fish should contain zarc binary path", name)
		}
		if !strings.Contains(string(content), "function "+name) {
			t.Fatalf("%s.fish should contain function %s", name, name)
		}
	}
}

func TestConfigureShellRC_MultipleAliases(t *testing.T) {
	tmpDir := t.TempDir()
	rcPath := filepath.Join(tmpDir, ".bashrc")
	os.WriteFile(rcPath, []byte("# existing\n"), 0644)

	err := configureShellRCAliases(rcPath, "/usr/local/bin/zarc", []string{"zarc", "claude"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(rcPath)
	s := string(content)
	if !strings.Contains(s, `alias zarc=`) {
		t.Fatal("should contain zarc alias")
	}
	if !strings.Contains(s, `alias claude=`) {
		t.Fatal("should contain claude alias")
	}
}

func TestConfigureShellRC_CustomAlias(t *testing.T) {
	tmpDir := t.TempDir()
	rcPath := filepath.Join(tmpDir, ".zshrc")
	os.WriteFile(rcPath, []byte(""), 0644)

	err := configureShellRCAliases(rcPath, "/usr/local/bin/zarc", []string{"meu-cli"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(rcPath)
	if !strings.Contains(string(content), `alias meu-cli=`) {
		t.Fatal("should contain custom alias")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/gildembergleite/workspace/zarc-claude-orchestrator && go test ./internal/setup/ -run "TestConfigureShellFish_MultipleAliases|TestConfigureShellRC_MultipleAliases|TestConfigureShellRC_CustomAlias" -v`

Expected: FAIL — `configureShellFishAliases` and `configureShellRCAliases` undefined.

- [ ] **Step 3: Implement multi-alias functions in shell.go**

Replace the internals of `internal/setup/shell.go`. Keep `DetectShell()` unchanged. Replace `ConfigureShellAlias`, `configureShellFish`, and `configureShellRC` with:

```go
// ConfigureShellAliases sets up shell aliases for the given names.
// Each name in aliases will point to zarcBin.
func ConfigureShellAliases(zarcBin string, aliases []string) (string, error) {
	shell := DetectShell()
	home, _ := os.UserHomeDir()

	switch shell {
	case "fish":
		dir := filepath.Join(home, ".config", "fish", "functions")
		if err := configureShellFishAliases(dir, zarcBin, aliases); err != nil {
			return "", err
		}
		return "fish", nil

	case "zsh":
		rc := filepath.Join(home, ".zshrc")
		if err := configureShellRCAliases(rc, zarcBin, aliases); err != nil {
			return "", err
		}
		return "zsh", nil

	case "bash":
		rc := filepath.Join(home, ".bashrc")
		if err := configureShellRCAliases(rc, zarcBin, aliases); err != nil {
			return "", err
		}
		return "bash", nil

	default:
		return "", fmt.Errorf("shell não detectado — configure manualmente: alias zarc=\"%s\"", zarcBin)
	}
}

func configureShellFishAliases(functionsDir, zarcBin string, aliases []string) error {
	if err := os.MkdirAll(functionsDir, 0755); err != nil {
		return err
	}

	for _, name := range aliases {
		fishFile := filepath.Join(functionsDir, name+".fish")
		if _, err := os.Stat(fishFile); err == nil {
			continue // already exists
		}
		content := fmt.Sprintf("function %s --description 'Claude Code + tmux session launcher'\n  %s $argv\nend\n", name, zarcBin)
		if err := os.WriteFile(fishFile, []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

func configureShellRCAliases(rcPath, zarcBin string, aliases []string) error {
	existing, _ := os.ReadFile(rcPath)
	content := string(existing)

	f, err := os.OpenFile(rcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, name := range aliases {
		marker := fmt.Sprintf("alias %s=", name)
		if strings.Contains(content, marker) {
			continue // already configured
		}
		line := fmt.Sprintf("\n# %s — Claude Code orchestrator\nalias %s=\"%s\"\n", name, name, zarcBin)
		if _, err := f.WriteString(line); err != nil {
			return err
		}
	}
	return nil
}
```

Also keep backward-compatible `ConfigureShellAlias` as a thin wrapper:

```go
// ConfigureShellAlias sets up a single zarc alias (backward compat).
func ConfigureShellAlias(zarcBin string) (string, error) {
	return ConfigureShellAliases(zarcBin, []string{"zarc"})
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/gildembergleite/workspace/zarc-claude-orchestrator && go test ./internal/setup/ -v`

Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add internal/setup/shell.go internal/setup/shell_test.go
git commit -m "feat: support multiple shell aliases in setup"
```

---

### Task 2: Add alias choice prompt to setup flow

**Files:**
- Modify: `internal/setup/setup.go`

- [ ] **Step 1: Add interactive alias prompt to setup.go**

Replace step 5 in the `Run()` function in `internal/setup/setup.go`:

```go
	// Step 5: Configure shell alias
	zarcBin, _ := os.Executable()
	aliases := promptAliasChoice()
	shellResult, err := ConfigureShellAliases(zarcBin, aliases)
	if err != nil {
		fmt.Printf(" %s Alias — %v\n", warnMark, err)
	} else {
		fmt.Printf(" %s Alias configurado (%s): %s\n", checkMark, shellResult, strings.Join(aliases, ", "))
	}
```

Add the `promptAliasChoice` function and required imports (`bufio`, `strings`):

```go
func promptAliasChoice() []string {
	fmt.Println()
	fmt.Println(" Como deseja chamar o CLI?")
	fmt.Println("   1) zarc + claude (dois aliases)")
	fmt.Println("   2) Somente zarc")
	fmt.Println("   3) Nome personalizado")
	fmt.Print("   Escolha [1/2/3]: ")

	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	choice := strings.TrimSpace(line)

	switch choice {
	case "1":
		return []string{"zarc", "claude"}
	case "3":
		fmt.Print("   Nome do alias: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)
		if name == "" {
			return []string{"zarc"}
		}
		return []string{name}
	default:
		return []string{"zarc"}
	}
}
```

- [ ] **Step 2: Build and verify manually**

Run: `cd /Users/gildembergleite/workspace/zarc-claude-orchestrator && go build ./cmd/zarc/`

Expected: BUILD SUCCESS

- [ ] **Step 3: Run all setup tests**

Run: `cd /Users/gildembergleite/workspace/zarc-claude-orchestrator && go test ./internal/setup/ -v`

Expected: ALL PASS

- [ ] **Step 4: Commit**

```bash
git add internal/setup/setup.go
git commit -m "feat: add alias choice prompt to setup flow"
```

---

### Task 3: Create DirBrowserModel component

**Files:**
- Create: `internal/tui/dirbrowser.go`
- Create: `internal/tui/dirbrowser_test.go`

- [ ] **Step 1: Write failing tests for DirBrowserModel**

Create `internal/tui/dirbrowser_test.go`:

```go
package tui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewDirBrowser_UsesInitialDir(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, "subdir-a"), 0755)
	os.MkdirAll(filepath.Join(tmp, "subdir-b"), 0755)

	db := NewDirBrowser(tmp)
	if db.currentDir != tmp {
		t.Fatalf("expected currentDir=%s, got %s", tmp, db.currentDir)
	}
	if len(db.entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(db.entries))
	}
}

func TestNewDirBrowser_HidesHiddenDirs(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, ".hidden"), 0755)
	os.MkdirAll(filepath.Join(tmp, "visible"), 0755)

	db := NewDirBrowser(tmp)
	if len(db.entries) != 1 {
		t.Fatalf("expected 1 visible entry, got %d", len(db.entries))
	}
	if db.entries[0] != "visible" {
		t.Fatalf("expected 'visible', got '%s'", db.entries[0])
	}
}

func TestDirBrowser_FilterByPrefix(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, "alpha"), 0755)
	os.MkdirAll(filepath.Join(tmp, "beta"), 0755)
	os.MkdirAll(filepath.Join(tmp, "alpha-two"), 0755)

	db := NewDirBrowser(tmp)
	db.filter = "al"
	filtered := db.filteredEntries()
	if len(filtered) != 2 {
		t.Fatalf("expected 2 filtered entries, got %d", len(filtered))
	}
}

func TestDirBrowser_DrillDown(t *testing.T) {
	tmp := t.TempDir()
	child := filepath.Join(tmp, "child")
	os.MkdirAll(filepath.Join(child, "grandchild"), 0755)

	db := NewDirBrowser(tmp)
	db.enterSelected("child")

	if db.currentDir != child {
		t.Fatalf("expected currentDir=%s, got %s", child, db.currentDir)
	}
	if len(db.entries) != 1 {
		t.Fatalf("expected 1 entry (grandchild), got %d", len(db.entries))
	}
}

func TestDirBrowser_GoUp(t *testing.T) {
	tmp := t.TempDir()
	child := filepath.Join(tmp, "child")
	os.MkdirAll(child, 0755)

	db := NewDirBrowser(child)
	db.goUp()

	if db.currentDir != tmp {
		t.Fatalf("expected currentDir=%s, got %s", tmp, db.currentDir)
	}
}

func TestDirBrowser_Result(t *testing.T) {
	tmp := t.TempDir()
	db := NewDirBrowser(tmp)
	if db.Result() != tmp {
		t.Fatalf("expected Result()=%s, got %s", tmp, db.Result())
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/gildembergleite/workspace/zarc-claude-orchestrator && go test ./internal/tui/ -run "TestNewDirBrowser|TestDirBrowser" -v`

Expected: FAIL — `NewDirBrowser` undefined.

- [ ] **Step 3: Implement DirBrowserModel**

Create `internal/tui/dirbrowser.go`:

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
	dirSelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	dirNormalStyle   = lipgloss.NewStyle().Faint(true)
	dirPathStyle     = lipgloss.NewStyle().Bold(true)
	dirHintStyle     = lipgloss.NewStyle().Faint(true)
)

const maxVisible = 10

// DirBrowserModel is a visual directory navigator with drill-down.
type DirBrowserModel struct {
	currentDir string
	entries    []string // visible directory names
	filter     string   // typed prefix filter
	cursor     int
	offset     int // scroll offset
	done       bool
	quitting   bool
}

// NewDirBrowser creates a dir browser starting at initialDir.
func NewDirBrowser(initialDir string) DirBrowserModel {
	m := DirBrowserModel{
		currentDir: initialDir,
	}
	m.loadEntries()
	return m
}

// Done returns true when the user confirmed the directory.
func (m DirBrowserModel) Done() bool {
	return m.done
}

// IsQuitting returns true if user pressed Esc/Ctrl+C.
func (m DirBrowserModel) IsQuitting() bool {
	return m.quitting
}

// Result returns the confirmed directory path.
func (m DirBrowserModel) Result() string {
	return m.currentDir
}

func (m *DirBrowserModel) loadEntries() {
	m.entries = nil
	m.cursor = 0
	m.offset = 0

	dirEntries, err := os.ReadDir(m.currentDir)
	if err != nil {
		return
	}

	for _, e := range dirEntries {
		if !e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		m.entries = append(m.entries, e.Name())
	}
}

func (m DirBrowserModel) filteredEntries() []string {
	if m.filter == "" {
		return m.entries
	}
	var result []string
	lower := strings.ToLower(m.filter)
	for _, e := range m.entries {
		if strings.HasPrefix(strings.ToLower(e), lower) {
			result = append(result, e)
		}
	}
	return result
}

func (m *DirBrowserModel) enterSelected(name string) {
	next := filepath.Join(m.currentDir, name)
	info, err := os.Stat(next)
	if err != nil || !info.IsDir() {
		return
	}
	m.currentDir = next
	m.filter = ""
	m.loadEntries()
}

func (m *DirBrowserModel) goUp() {
	parent := filepath.Dir(m.currentDir)
	if parent == m.currentDir {
		return // at root
	}
	m.currentDir = parent
	m.filter = ""
	m.loadEntries()
}

func (m DirBrowserModel) displayPath() string {
	home, _ := os.UserHomeDir()
	if strings.HasPrefix(m.currentDir, home) {
		return "~" + m.currentDir[len(home):]
	}
	return m.currentDir
}

func (m DirBrowserModel) Init() tea.Cmd {
	return nil
}

func (m DirBrowserModel) Update(msg tea.Msg) (DirBrowserModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			m.quitting = true
			return m, nil
		case "enter":
			filtered := m.filteredEntries()
			if len(filtered) > 0 && m.cursor < len(filtered) {
				m.enterSelected(filtered[m.cursor])
				return m, nil
			}
			// No selection or empty list — confirm current dir
			m.done = true
			return m, tea.Quit
		case "ctrl+d":
			m.done = true
			return m, tea.Quit
		case "backspace":
			if m.filter != "" {
				m.filter = m.filter[:len(m.filter)-1]
				m.cursor = 0
				m.offset = 0
			} else {
				m.goUp()
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.offset {
					m.offset = m.cursor
				}
			}
		case "down", "j":
			filtered := m.filteredEntries()
			if m.cursor < len(filtered)-1 {
				m.cursor++
				if m.cursor >= m.offset+maxVisible {
					m.offset = m.cursor - maxVisible + 1
				}
			}
		case "left":
			m.goUp()
		case "right":
			filtered := m.filteredEntries()
			if len(filtered) > 0 && m.cursor < len(filtered) {
				m.enterSelected(filtered[m.cursor])
			}
		default:
			if len(msg.String()) == 1 {
				m.filter += msg.String()
				m.cursor = 0
				m.offset = 0
			}
		}
	}
	return m, nil
}

func (m DirBrowserModel) View() string {
	var b strings.Builder

	b.WriteString(dirPathStyle.Render("Diretório: " + m.displayPath()))
	if m.filter != "" {
		b.WriteString("/" + m.filter)
	}
	b.WriteString("\n")
	b.WriteString(dirHintStyle.Render("  arrows navigate | enter drill-down/confirm | esc cancel | type to filter"))
	b.WriteString("\n\n")

	filtered := m.filteredEntries()

	if len(filtered) == 0 {
		b.WriteString(dirNormalStyle.Render("  (sem subdiretórios)"))
		b.WriteString("\n")
	} else {
		end := m.offset + maxVisible
		if end > len(filtered) {
			end = len(filtered)
		}

		if m.offset > 0 {
			b.WriteString(dirHintStyle.Render("  ↑ mais itens"))
			b.WriteString("\n")
		}

		for i := m.offset; i < end; i++ {
			name := filtered[i] + "/"
			if i == m.cursor {
				b.WriteString(dirSelectedStyle.Render(fmt.Sprintf("  > %s", name)))
			} else {
				b.WriteString(dirNormalStyle.Render(fmt.Sprintf("    %s", name)))
			}
			b.WriteString("\n")
		}

		if end < len(filtered) {
			b.WriteString(dirHintStyle.Render("  ↓ mais itens"))
			b.WriteString("\n")
		}
	}

	return b.String()
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/gildembergleite/workspace/zarc-claude-orchestrator && go test ./internal/tui/ -run "TestNewDirBrowser|TestDirBrowser" -v`

Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add internal/tui/dirbrowser.go internal/tui/dirbrowser_test.go
git commit -m "feat: add DirBrowserModel component with visual navigation"
```

---

### Task 4: Wire DirBrowserModel into AppModel with default dir

**Files:**
- Modify: `internal/tui/model.go`

- [ ] **Step 1: Replace dirInput with dirBrowser in model.go**

In `internal/tui/model.go`, make these changes:

1. Replace the `dirInput InputModel` field with `dirBrowser DirBrowserModel` in `AppModel`:

```go
type AppModel struct {
	state      state
	tmuxBin    string
	claudeBin  string
	menu       MenuModel
	subMenu    MenuModel
	dirBrowser DirBrowserModel
	nameInput  InputModel
	sessions   []tmux.Session
	selected   string
	err        error
	quitting   bool

	Action      string
	SessionDir  string
	SessionName string
}
```

2. In `updateMainMenu`, replace the `InputModel` creation block (when `chosen.ID == "new"`) with:

```go
		if chosen.ID == "new" {
			cwd, _ := os.Getwd()
			m.dirBrowser = NewDirBrowser(cwd)
			m.state = stateNewSessionDir
			return m, nil
		}
```

3. Replace `updateDirInput` method entirely:

```go
func (m AppModel) updateDirInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.dirBrowser, cmd = m.dirBrowser.Update(msg)

	if m.dirBrowser.IsQuitting() {
		m.loadMainMenu()
		return m, nil
	}

	if m.dirBrowser.Done() {
		m.SessionDir = m.dirBrowser.Result()
		defaultName := filepath.Base(m.SessionDir)
		m.nameInput = NewInput("Nome da sessão", defaultName, nil)
		m.state = stateNewSessionName
		return m, nil
	}

	return m, cmd
}
```

4. In `View()`, replace the `stateNewSessionDir` case:

```go
	case stateNewSessionDir:
		b.WriteString(m.dirBrowser.View())
```

- [ ] **Step 2: Build to verify compilation**

Run: `cd /Users/gildembergleite/workspace/zarc-claude-orchestrator && go build ./cmd/zarc/`

Expected: BUILD SUCCESS

- [ ] **Step 3: Run all TUI tests**

Run: `cd /Users/gildembergleite/workspace/zarc-claude-orchestrator && go test ./internal/tui/ -v`

Expected: ALL PASS

- [ ] **Step 4: Run full test suite**

Run: `cd /Users/gildembergleite/workspace/zarc-claude-orchestrator && go test ./... -v`

Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add internal/tui/model.go
git commit -m "feat: wire DirBrowserModel with default pwd into session creation"
```

---

### Task 5: Clean up unused InputModel tab-complete code

**Files:**
- Modify: `internal/tui/input.go`

- [ ] **Step 1: Remove tab-complete fields and logic from InputModel**

In `internal/tui/input.go`, remove the `TabComplete`, `completions`, `compIndex` fields from `InputModel` struct. Remove the `handleTabComplete` method entirely. Remove the `"tab"` case in `Update`. Remove the `completions = nil` lines in `backspace` and `default` cases. Remove the `"path/filepath"` import if no longer used.

The cleaned `InputModel` struct:

```go
type InputModel struct {
	Prompt      string
	Placeholder string
	Value       string
	Err         string
	Validate    func(string) error
	done        bool
	quitting    bool
}
```

The cleaned `Update` method `tea.KeyMsg` switch:

```go
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
```

- [ ] **Step 2: Run all tests**

Run: `cd /Users/gildembergleite/workspace/zarc-claude-orchestrator && go test ./... -v`

Expected: ALL PASS

- [ ] **Step 3: Commit**

```bash
git add internal/tui/input.go
git commit -m "refactor: remove tab-complete from InputModel (replaced by DirBrowser)"
```
