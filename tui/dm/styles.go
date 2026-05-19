package dm

import (
	"charm.land/lipgloss/v2"
	"github.com/jerry-harm/nosmec/tui/theme"
)

type styles struct {
	header      lipgloss.Style
	messageArea lipgloss.Style
	inputArea   lipgloss.Style
	message     lipgloss.Style
	mine        lipgloss.Style
	theirs      lipgloss.Style
	timestamp   lipgloss.Style
	npub        lipgloss.Style
	errorMsg    lipgloss.Style
	help        lipgloss.Style
}

func newStyles(t *theme.Theme) styles {
	return styles{
		header: lipgloss.NewStyle().
			Foreground(t.Text).
			Background(t.TitleBg).
			Bold(true).
			Padding(0, 1),
		messageArea: lipgloss.NewStyle().
			Padding(1, 0),
		inputArea: lipgloss.NewStyle().
			Foreground(t.TextBright).
			Padding(0, 1),
		mine: lipgloss.NewStyle().
			Foreground(t.TagColor),
		theirs: lipgloss.NewStyle().
			Foreground(t.Text),
		timestamp: lipgloss.NewStyle().
			Foreground(t.TextMutedAlt),
		npub: lipgloss.NewStyle().
			Foreground(t.TextMutedAlt),
		errorMsg: lipgloss.NewStyle().
			Foreground(t.Error).
			Bold(true),
		help: lipgloss.NewStyle().
			Foreground(t.TextMutedAlt),
	}
}