package compose

import (
	"charm.land/lipgloss/v2"
	"github.com/jerry-harm/nosmec/tui/theme"
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
	sendingOverlay  lipgloss.Style
	statusText      lipgloss.Style
}

func newStyles(t *theme.Theme) styles {
	return styles{
		header: lipgloss.NewStyle().
			Foreground(t.Text).
			Background(t.TitleBg).
			Bold(true).
			Padding(0, 1),
		formArea: lipgloss.NewStyle().
			Padding(1, 0),
		inputArea: lipgloss.NewStyle().
			Foreground(t.TextBright).
			Padding(0, 1),
		errorMsg: lipgloss.NewStyle().
			Foreground(t.Error).
			Bold(true),
		successMsg: lipgloss.NewStyle().
			Foreground(t.TextBright).
			Bold(true),
		help: lipgloss.NewStyle().
			Foreground(t.TextMutedAlt),
		button: lipgloss.NewStyle().
			Foreground(t.TextMutedAlt),
		buttonActive: lipgloss.NewStyle().
			Foreground(t.TextBright).
			Bold(true),
		fieldLabel: lipgloss.NewStyle().
			Foreground(t.TextMutedAlt),
		fieldValue: lipgloss.NewStyle().
			Foreground(t.Text),
		tagList: lipgloss.NewStyle().
			Foreground(t.TextMutedAlt),
		tagItem: lipgloss.NewStyle().
			Foreground(t.TagColor),
		tagItemSelected: lipgloss.NewStyle().
			Foreground(t.Text).
			Background(t.TitleBg).
			Bold(true).
			Padding(0, 1),
		sendingOverlay: lipgloss.NewStyle().
			Foreground(t.Text).
			Background(t.OverlayBg).
			Padding(2, 4),
		statusText: lipgloss.NewStyle().
			Foreground(t.TextBright).
			Bold(true),
	}
}