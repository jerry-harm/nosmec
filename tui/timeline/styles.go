package timeline

import "charm.land/lipgloss/v2"

var (
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF00"))

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00"))

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4444"))

	EmptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	DelegateItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF"))

	DelegateTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00AAFF"))

	DelegateDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#AAAAAA"))

	DelegateSelectedPrefix = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00"))

	DelegateNormalPrefix = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666"))

	CommunityStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00")).
			Bold(true)

	NoteStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00AAFF"))

	AuthorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA"))

	TimeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))

	ContentStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DDDDDD"))

	EventIDStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))

	ProfileNameStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	NpubStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA"))

	ReplyToStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))

	QuoteStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))
)
