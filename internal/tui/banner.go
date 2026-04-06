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
