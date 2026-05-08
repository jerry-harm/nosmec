package timeline

import "charm.land/lipgloss/v2"

type styles struct {
	app           lipgloss.Style
	title         lipgloss.Style
	statusMessage lipgloss.Style
	helpStyle     lipgloss.Style
	itemTitle     lipgloss.Style
	itemDesc      lipgloss.Style
	detailBox     lipgloss.Style
	detailHeader  lipgloss.Style
	detailContent lipgloss.Style
	detailFooter  lipgloss.Style
}

func newStyles(darkBG bool) styles {
	lightDark := lipgloss.LightDark(darkBG)

	return styles{
		app: lipgloss.NewStyle().
			Padding(1, 2),
		title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1),
		statusMessage: lipgloss.NewStyle().
			Foreground(lightDark(lipgloss.Color("#04B575"), lipgloss.Color("#059C4B"))),
		helpStyle: lipgloss.NewStyle().
			Foreground(lightDark(lipgloss.Color("#9A9A9A"), lipgloss.Color("#6B6B6B"))),
		itemTitle: lipgloss.NewStyle().
			Foreground(lightDark(lipgloss.Color("#00FF00"), lipgloss.Color("#00875A"))).
			Bold(true),
		itemDesc: lipgloss.NewStyle().
			Foreground(lightDark(lipgloss.Color("#AAAAAA"), lipgloss.Color("#7A7A7A"))),
		detailBox: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lightDark(lipgloss.Color("#25A065"), lipgloss.Color("#25A065"))).
			Padding(1, 1),
		detailHeader: lipgloss.NewStyle().
			Foreground(lightDark(lipgloss.Color("#00FF00"), lipgloss.Color("#00875A"))).
			Bold(true),
		detailContent: lipgloss.NewStyle().
			Foreground(lightDark(lipgloss.Color("#FFFFFF"), lipgloss.Color("#1A1A1A"))),
		detailFooter: lipgloss.NewStyle().
			Foreground(lightDark(lipgloss.Color("#666666"), lipgloss.Color("#888888"))),
	}
}