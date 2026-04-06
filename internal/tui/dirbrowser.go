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
	dirConfirmStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2")) // green
	dirNormalStyle   = lipgloss.NewStyle().Faint(true)
	dirPathStyle     = lipgloss.NewStyle().Bold(true)
	dirHintStyle     = lipgloss.NewStyle().Faint(true)
)

const maxVisible = 10

const selectEntry = "__select__"

// DirBrowserModel is a visual directory navigator with drill-down.
type DirBrowserModel struct {
	currentDir string
	entries    []string // visible directory names (first is selectEntry)
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
	m.cursor = 0 // first open: cursor on [Selecionar] for quick Enter
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
	m.offset = 0

	// First entry is always the "select this dir" option
	m.entries = append(m.entries, selectEntry)

	// Add ../ unless at filesystem root
	parent := filepath.Dir(m.currentDir)
	if parent != m.currentDir {
		m.entries = append(m.entries, "..")
	}

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

	// Start cursor on ".." if it exists, otherwise on first dir
	if len(m.entries) > 1 && m.entries[1] == ".." {
		m.cursor = 1
	} else {
		m.cursor = 0
	}
}

func (m DirBrowserModel) filteredEntries() []string {
	if m.filter == "" {
		return m.entries
	}
	lower := strings.ToLower(m.filter)
	// Always keep selectEntry and .. in filtered results
	var result []string
	for _, e := range m.entries {
		if e == selectEntry || e == ".." {
			result = append(result, e)
			continue
		}
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
			if m.cursor >= len(filtered) {
				return m, nil
			}
			selected := filtered[m.cursor]
			switch selected {
			case selectEntry:
				m.done = true
				return m, tea.Quit
			case "..":
				m.goUp()
				return m, nil
			default:
				m.enterSelected(selected)
				return m, nil
			}
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

	b.WriteString(dirPathStyle.Render("Diretorio: " + m.displayPath()))
	if m.filter != "" {
		b.WriteString("/" + m.filter)
	}
	b.WriteString("\n")
	b.WriteString(dirHintStyle.Render("  enter selecionar | delete voltar | esc cancelar | digite para filtrar"))
	b.WriteString("\n\n")

	filtered := m.filteredEntries()

	end := m.offset + maxVisible
	if end > len(filtered) {
		end = len(filtered)
	}

	if m.offset > 0 {
		b.WriteString(dirHintStyle.Render("  ... mais itens"))
		b.WriteString("\n")
	}

	for i := m.offset; i < end; i++ {
		entry := filtered[i]
		if entry == selectEntry {
			label := fmt.Sprintf("[Selecionar: %s]", m.displayPath())
			if i == m.cursor {
				b.WriteString(dirConfirmStyle.Render(fmt.Sprintf("  > %s", label)))
			} else {
				b.WriteString(dirNormalStyle.Render(fmt.Sprintf("    %s", label)))
			}
		} else {
			name := entry
			if name != ".." {
				name += "/"
			}
			if i == m.cursor {
				b.WriteString(dirSelectedStyle.Render(fmt.Sprintf("  > %s", name)))
			} else {
				b.WriteString(dirNormalStyle.Render(fmt.Sprintf("    %s", name)))
			}
		}
		b.WriteString("\n")
	}

	if end < len(filtered) {
		b.WriteString(dirHintStyle.Render("  ... mais itens"))
		b.WriteString("\n")
	}

	return b.String()
}
