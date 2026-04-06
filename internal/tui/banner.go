package tui

import "github.com/charmbracelet/lipgloss"

var bannerMain = `  ██████╗██╗      █████╗ ██╗   ██╗██████╗ ███████╗
 ██╔════╝██║     ██╔══██╗██║   ██║██╔══██╗██╔════╝
 ██║     ██║     ███████║██║   ██║██║  ██║█████╗
 ██║     ██║     ██╔══██║██║   ██║██║  ██║██╔══╝
 ╚██████╗███████╗██║  ██║╚██████╔╝██████╔╝███████╗
  ╚═════╝╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝`

var bannerSub = ` █▀█ █▀█ █▀▀ █ █ █▀▀ █▀ ▀█▀ █▀█ ▄▀█ ▀█▀ █▀█ █▀█
 █▄█ █▀▄ █▄▄ █▀█ ██▄ ▄█  █  █▀▄ █▀█  █  █▄█ █▀▄`

var (
	bannerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E07A2F")). // claude orange
			Bold(true)

	bannerSubStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E07A2F"))

	subtitleStyle = lipgloss.NewStyle().
			Faint(true)
)

// RenderBanner returns the styled Claude Orchestrator banner.
func RenderBanner() string {
	main := bannerStyle.Render(bannerMain)
	sub := bannerSubStyle.Render(bannerSub)
	subtitle := subtitleStyle.Render(" Gerenciamento de persistência de sessões")
	return "\n" + main + "\n" + sub + "\n\n" + subtitle + "\n"
}
