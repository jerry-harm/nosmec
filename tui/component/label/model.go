package label

import (
	"charm.land/lipgloss/v2"
	"github.com/jerry-harm/nosmec/tui/theme"
)

type State int

const (
	StateLoading  State = iota
	StateResolved
	StateError
)

type LabelClickedMsg struct {
	Pubkey string
}

func truncateNpub(pubkey string) string {
	if len(pubkey) >= 16 {
		return "@" + pubkey[:8] + "..."
	}
	return "@" + pubkey
}

func RenderLabel(pubkey, name string, state State, t *theme.Theme) string {
	var text string
	var style lipgloss.Style

	switch state {
	case StateLoading:
		text = truncateNpub(pubkey)
		style = lipgloss.NewStyle().Foreground(t.TextMutedAlt)
	case StateResolved:
		text = "@" + name
		style = lipgloss.NewStyle().
			Foreground(t.TextBright).
			Background(t.AuthorTextAlt).
			Padding(0, 1)
	case StateError:
		text = truncateNpub(pubkey)
		style = lipgloss.NewStyle().
			Foreground(t.TextMutedAlt).
			Background(t.ErrorAlt).
			Padding(0, 1)
	default:
		text = truncateNpub(pubkey)
		style = lipgloss.NewStyle().Foreground(t.TextMutedAlt)
	}

	return style.Render(text)
}