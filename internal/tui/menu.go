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
