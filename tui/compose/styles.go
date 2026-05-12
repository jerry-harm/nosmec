package compose

import (
	"charm.land/lipgloss/v2"
)

type styles struct {
	header         lipgloss.Style
	formArea       lipgloss.Style
	inputArea      lipgloss.Style
	errorMsg       lipgloss.Style
	successMsg     lipgloss.Style
	help           lipgloss.Style
	button         lipgloss.Style
	buttonActive   lipgloss.Style
	fieldLabel     lipgloss.Style
	fieldValue     lipgloss.Style
	tagList        lipgloss.Style
	tagItem        lipgloss.Style
	tagItemSelected lipgloss.Style
}

func newStyles() styles {
	return styles{
		header: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#25A065")).
			Bold(true).
			Padding(0, 1),
		formArea: lipgloss.NewStyle().
			Padding(1, 0),
		inputArea: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Padding(0, 1),
		errorMsg: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4444")).
			Bold(true),
		successMsg: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true),
		help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")),
		button: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")),
		buttonActive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true),
		fieldLabel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")),
		fieldValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")),
		tagList: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")),
		tagItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00AAFF")),
		tagItemSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#25A065")).
			Bold(true).
			Padding(0, 1),
	}
}