package dm

import (
	"charm.land/lipgloss/v2"
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

func newStyles() styles {
	return styles{
		header: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#25A065")).
			Bold(true).
			Padding(0, 1),
		messageArea: lipgloss.NewStyle().
			Padding(1, 0),
		inputArea: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Padding(0, 1),
		mine: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00AAFF")),
		theirs: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")),
		timestamp: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")),
		npub: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")),
		errorMsg: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4444")).
			Bold(true),
		help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")),
	}
}