package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gildembergleite/claude-orchestrator/internal/tmux"
)

const (
	dirColumnWidth  = 36
	idleColumnWidth = 10
)

type state int

const (
	stateMainMenu state = iota
	stateNewSessionDir
	stateNewSessionName
	stateSubMenu
	stateConfirmKill
)

// AppModel is the top-level bubbletea model.
type AppModel struct {
	state      state
	tmuxBin    string
	claudeBin  string
	menu       MenuModel
	subMenu    MenuModel
	dirBrowser DirBrowserModel
	nameInput  InputModel
	sessions   []string
	selected   string // selected session name
	err        error
	quitting   bool

	// results to act on after tea.Program exits
	Action      string // "attach", "new", "kill", ""
	SessionDir  string
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
	tmux.CleanupOrphans()

	sessions, _ := tmux.ListSessions(m.tmuxBin)
	m.sessions = sessions

	nameWidth := 8
	for _, name := range sessions {
		if len(name) > nameWidth {
			nameWidth = len(name)
		}
	}

	items := []MenuItem{{Label: "[+] Nova sessão", ID: "new"}}
	now := time.Now()
	for _, name := range sessions {
		label := padRight(name, nameWidth)
		if sess, ok, _ := tmux.GetSession(name); ok {
			dir := truncateLeft(collapsePath(sess.Dir), dirColumnWidth)
			idle := formatIdle(now.Sub(sess.LastAttachedAt))
			label = fmt.Sprintf("%s  %s  %s", label, padRight(dir, dirColumnWidth), padLeft(idle, idleColumnWidth))
		}
		items = append(items, MenuItem{Label: label, ID: "session:" + name})
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
			cwd, _ := os.Getwd()
			m.dirBrowser = NewDirBrowser(cwd)
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
			tmux.KillSession(m.tmuxBin, m.selected)
			tmux.UnregisterSession(m.selected)
			m.loadMainMenu()
			return m, nil
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
		b.WriteString(m.dirBrowser.View())
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
