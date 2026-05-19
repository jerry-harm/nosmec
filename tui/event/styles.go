package event

import (
	"charm.land/lipgloss/v2"
	"github.com/jerry-harm/nosmec/tui/theme"
)

type eventStyles struct {
	t             *theme.Theme
	container      lipgloss.Style
	header         lipgloss.Style
	viewport       lipgloss.Style
	footer         lipgloss.Style
	author         lipgloss.Style
	time           lipgloss.Style
	content        lipgloss.Style
	tags           lipgloss.Style
	confirm        lipgloss.Style
	communityAddr  lipgloss.Style
}

func newStyles(t *theme.Theme) eventStyles {
	return eventStyles{
		t: t,
		container: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(t.Border).
			Padding(1, 1).
			Width(78).
			Height(22),
		header: lipgloss.NewStyle().
			Foreground(t.TextBright).
			Bold(true),
		viewport: lipgloss.NewStyle().
			Margin(0, 0).
			Padding(0, 0),
		footer: lipgloss.NewStyle().
			Foreground(t.TextMutedDark),
		author: lipgloss.NewStyle().
			Foreground(t.AuthorText).
			Bold(true),
		time: lipgloss.NewStyle().
			Foreground(t.TextMutedAlt),
		content: lipgloss.NewStyle().
			Foreground(t.TextDark),
		tags: lipgloss.NewStyle().
			Foreground(t.TextMutedAlt),
		confirm: lipgloss.NewStyle().
			Foreground(t.ErrorAlt).
			Bold(true),
		communityAddr: lipgloss.NewStyle().
			Foreground(t.CommunityAddr).
			Bold(true),
	}
}