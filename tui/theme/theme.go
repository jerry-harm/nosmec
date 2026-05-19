package theme

import (
	"charm.land/lipgloss/v2"
	"image/color"
)

type Theme struct {
	Primary       color.Color
	PrimaryDark   color.Color
	TextBright    color.Color
	TextBrightAlt color.Color
	Text          color.Color
	TextDark      color.Color
	TextMuted     color.Color
	TextMutedDark color.Color
	TextMutedAlt  color.Color
	Selection     color.Color
	StatusText    color.Color
	AuthorText    color.Color
	AuthorTextAlt color.Color
	Error         color.Color
	ErrorAlt      color.Color
	TagColor      color.Color
	CommunityAddr color.Color
	OverlayBg     color.Color
	TitleText     color.Color
	TitleBg       color.Color
	Border        color.Color
	BorderDark    color.Color
	ViewportBorder    color.Color
	ViewportBorderDark color.Color
	InputPlaceholder  color.Color
	Spinner            color.Color
}

var defaultLight = Theme{
	Primary:       lipgloss.Color("#25A065"),
	PrimaryDark:   lipgloss.Color("#00875A"),
	TextBright:    lipgloss.Color("#00FF00"),
	TextBrightAlt: lipgloss.Color("#00875A"),
	Text:          lipgloss.Color("#FFFFFF"),
	TextDark:      lipgloss.Color("#333333"),
	TextMuted:     lipgloss.Color("#AAAAAA"),
	TextMutedDark: lipgloss.Color("#6B6B6B"),
	TextMutedAlt:  lipgloss.Color("#888888"),
	Selection:     lipgloss.Color("#FFFF00"),
	StatusText:    lipgloss.Color("#04B575"),
	AuthorText:    lipgloss.Color("#00AA00"),
	AuthorTextAlt: lipgloss.Color("#008800"),
	Error:         lipgloss.Color("#FF4444"),
	ErrorAlt:      lipgloss.Color("#FF6B6B"),
	TagColor:      lipgloss.Color("#00AAFF"),
	CommunityAddr: lipgloss.Color("#FFD700"),
	OverlayBg:     lipgloss.Color("#333333"),
	TitleText:     lipgloss.Color("#FFFDF5"),
	TitleBg:       lipgloss.Color("#25A065"),
	Border:             lipgloss.Color("#25A065"),
	BorderDark:         lipgloss.Color("#00875A"),
	ViewportBorder:     lipgloss.Color("#25A065"),
	ViewportBorderDark: lipgloss.Color("#00875A"),
	InputPlaceholder:   lipgloss.Color("#666666"),
	Spinner:            lipgloss.Color("#00FF00"),
}

var defaultDark = Theme{
	Primary:       lipgloss.Color("#00875A"),
	PrimaryDark:   lipgloss.Color("#00875A"),
	TextBright:    lipgloss.Color("#00875A"),
	TextBrightAlt: lipgloss.Color("#00875A"),
	Text:          lipgloss.Color("#FFFFFF"),
	TextDark:      lipgloss.Color("#333333"),
	TextMuted:     lipgloss.Color("#6B6B6B"),
	TextMutedDark: lipgloss.Color("#6B6B6B"),
	TextMutedAlt:  lipgloss.Color("#888888"),
	Selection:     lipgloss.Color("#FFFF00"),
	StatusText:    lipgloss.Color("#04B575"),
	AuthorText:    lipgloss.Color("#008800"),
	AuthorTextAlt: lipgloss.Color("#008800"),
	Error:         lipgloss.Color("#FF4444"),
	ErrorAlt:      lipgloss.Color("#FF6B6B"),
	TagColor:      lipgloss.Color("#00AAFF"),
	CommunityAddr: lipgloss.Color("#FFD700"),
	OverlayBg:     lipgloss.Color("#333333"),
	TitleText:     lipgloss.Color("#FFFDF5"),
	TitleBg:       lipgloss.Color("#00875A"),
	Border:             lipgloss.Color("#00875A"),
	BorderDark:         lipgloss.Color("#00875A"),
	ViewportBorder:     lipgloss.Color("#00875A"),
	ViewportBorderDark: lipgloss.Color("#00875A"),
	InputPlaceholder:   lipgloss.Color("#666666"),
	Spinner:            lipgloss.Color("#00FF00"),
}

func DefaultTheme(darkBG bool) *Theme {
	if darkBG {
		return &defaultDark
	}
	return &defaultLight
}

func Default() *Theme {
	return &defaultLight
}