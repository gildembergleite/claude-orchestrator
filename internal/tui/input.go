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
	TabComplete bool // enable directory tab completion
	completions []string
	compIndex   int
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
		case "tab":
			if m.TabComplete {
				m.handleTabComplete()
			}
		case "backspace":
			if len(m.Value) > 0 {
				m.Value = m.Value[:len(m.Value)-1]
			}
			m.Err = ""
			m.completions = nil
		default:
			if len(msg.String()) == 1 {
				m.Value += msg.String()
				m.Err = ""
				m.completions = nil
			}
		}
	}
	return m, nil
}

func (m *InputModel) handleTabComplete() {
	value := m.Value
	if value == "" {
		value = "~/"
		m.Value = value
	}

	// Expand ~ for lookup
	expanded := value
	if strings.HasPrefix(expanded, "~/") {
		home, _ := os.UserHomeDir()
		expanded = filepath.Join(home, expanded[2:])
	}

	if m.completions == nil {
		// Build completions list
		dir := expanded
		prefix := ""
		if info, err := os.Stat(expanded); err != nil || !info.IsDir() {
			dir = filepath.Dir(expanded)
			prefix = filepath.Base(expanded)
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}

		var matches []string
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			if strings.HasPrefix(e.Name(), ".") {
				continue
			}
			if prefix == "" || strings.HasPrefix(strings.ToLower(e.Name()), strings.ToLower(prefix)) {
				full := filepath.Join(dir, e.Name()) + "/"
				// Convert back to ~/
				home, _ := os.UserHomeDir()
				if strings.HasPrefix(full, home) {
					full = "~" + full[len(home):]
				}
				matches = append(matches, full)
			}
		}

		if len(matches) == 0 {
			return
		}
		m.completions = matches
		m.compIndex = 0
	} else {
		// Cycle through completions
		m.compIndex = (m.compIndex + 1) % len(m.completions)
	}

	m.Value = m.completions[m.compIndex]
	m.Err = ""
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
	b.WriteString("\u2588\n") // cursor block

	if m.Err != "" {
		b.WriteString(errorStyle.Render(fmt.Sprintf("  \u26a0 %s", m.Err)))
		b.WriteString("\n")
	}

	return b.String()
}
