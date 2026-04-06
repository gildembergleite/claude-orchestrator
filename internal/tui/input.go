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
	b.WriteString("\u2588\n") // cursor block

	if m.Err != "" {
		b.WriteString(errorStyle.Render(fmt.Sprintf("  \u26a0 %s", m.Err)))
		b.WriteString("\n")
	}

	return b.String()
}
