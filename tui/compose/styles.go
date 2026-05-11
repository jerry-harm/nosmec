package compose

import (
	"charm.land/lipgloss/v2"
)

type styles struct {
	header      lipgloss.Style
	messageArea lipgloss.Style
	inputArea   lipgloss.Style
	errorMsg    lipgloss.Style
	help        lipgloss.Style
	mine        lipgloss.Style
	theirs      lipgloss.Style
}

func newStyles() styles {
	return styles{
		header: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#25A065")).
			Bold(true).
			Padding(0, 1),
		inputArea: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Padding(0, 1),
		errorMsg: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4444")).
			Bold(true),
		help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")),
		mine: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00AAFF")),
		theirs: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")),
	}
}