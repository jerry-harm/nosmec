package event

import "charm.land/lipgloss/v2"

type eventStyles struct {
	container lipgloss.Style
	header    lipgloss.Style
	viewport  lipgloss.Style
	footer    lipgloss.Style
	author    lipgloss.Style
	time      lipgloss.Style
	content   lipgloss.Style
	tags      lipgloss.Style
}

func newStyles(darkBG bool) eventStyles {
	lightDark := lipgloss.LightDark(darkBG)

	borderColor := lipgloss.Color("#25A065")
	if darkBG {
		borderColor = lipgloss.Color("#00875A")
	}

	return eventStyles{
		container: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(borderColor).
			Padding(1, 1).
			Width(78).
			Height(22),
		header: lipgloss.NewStyle().
			Foreground(lightDark(lipgloss.Color("#00FF00"), lipgloss.Color("#00875A"))).
			Bold(true),
		viewport: lipgloss.NewStyle().
			Margin(0, 0).
			Padding(0, 0),
		footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B6B6B")),
		author: lipgloss.NewStyle().
			Foreground(lightDark(lipgloss.Color("#00AA00"), lipgloss.Color("#008800"))).
			Bold(true),
		time: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")),
		content: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#333333")),
		tags: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")),
	}
}