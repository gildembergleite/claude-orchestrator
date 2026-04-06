package tui

import "github.com/charmbracelet/lipgloss"

var bannerMain = `  ██████╗██╗      █████╗ ██╗   ██╗██████╗ ███████╗
 ██╔════╝██║     ██╔══██╗██║   ██║██╔══██╗██╔════╝
 ██║     ██║     ███████║██║   ██║██║  ██║█████╗
 ██║     ██║     ██╔══██║██║   ██║██║  ██║██╔══╝
 ╚██████╗███████╗██║  ██║╚██████╔╝██████╔╝███████╗
  ╚═════╝╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝`

var bannerSub = ` █▀█ █▀█ █▀▀  █ █ █▀▀ █▀  ▀█▀ █▀█ ▄▀█  ▀█▀ █▀█ █▀█
 █▄█ █▀▄ █▄▄  █▀█ ██▄ ▄█   █  █▀▄ █▀█   █  █▄█ █▀▄`

var (
	bannerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E07A2F")). // claude orange
			Bold(true)

	bannerSubStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E07A2F"))

	companyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E07A2F")).
			Faint(true)

	subtitleStyle = lipgloss.NewStyle().
			Faint(true)
)

// RenderBanner returns the styled Claude Orchestrator banner with zarc branding.
func RenderBanner() string {
	main := bannerStyle.Render(bannerMain)
	sub := bannerSubStyle.Render(bannerSub)
	company := companyStyle.Render("                                          by zarc")
	subtitle := subtitleStyle.Render(" Persistência de sessões Claude Code")
	return main + "\n" + sub + "\n" + company + "\n" + subtitle + "\n"
}
