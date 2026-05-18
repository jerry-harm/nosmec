package label

import (
	"context"
	"time"

	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/sdk"
	"github.com/jerry-harm/nosmec/config"
)

type Config struct {
	Pubkey string
	App    *config.AppContext
}

type State int

const (
	StateLoading  State = iota
	StateResolved
	StateError
)

type Model struct {
	pubkey  string
	app     *config.AppContext
	name    string
	state   State
	focused bool
}

func New(cfg Config) *Model {
	return &Model{
		pubkey: cfg.Pubkey,
		app:    cfg.App,
		state:  StateLoading,
	}
}

func (m *Model) Init() tea.Cmd {
	return m.fetchProfile
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case nameResolvedMsg:
		m.name = msg.name
		if msg.name != "" {
			m.state = StateResolved
		} else {
			m.state = StateError
		}
		return m, nil
	}
	return m, nil
}

func (m *Model) fetchProfile() tea.Msg {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var pubKey nostr.PubKey
	if err := pubKey.UnmarshalJSON([]byte(`"` + m.pubkey + `"`)); err != nil {
		return nameResolvedMsg{name: ""}
	}

	pm := m.app.System().FetchProfileMetadata(ctx, pubKey)
	name := ""
	if pm.Event != nil {
		if meta, err := sdk.ParseMetadata(*pm.Event); err == nil && meta.Name != "" {
			name = meta.Name
		}
	}
	return nameResolvedMsg{name: name}
}

func (m *Model) View() tea.View {
	text := m.renderText()
	style := m.effectiveStyle()

	v := tea.NewView(style.Render(text))
	v.OnMouse = func(msg tea.MouseMsg) tea.Cmd {
		click, ok := msg.(tea.MouseClickMsg)
		if !ok {
			return nil
		}
		mouse := click.Mouse()
		if mouse.Button == tea.MouseLeft {
			return func() tea.Msg {
				return LabelClickedMsg{Pubkey: m.pubkey}
			}
		}
		return nil
	}

	return v
}

func (m *Model) effectiveStyle() lipgloss.Style {
	if m.focused {
		return focusedStyle
	}
	switch m.state {
	case StateLoading:
		return loadingStyle
	case StateResolved:
		return normalStyle
	case StateError:
		return errorStyle
	}
	return loadingStyle
}

func (m *Model) renderText() string {
	switch m.state {
	case StateLoading:
		return truncateNpub(m.pubkey)
	case StateResolved:
		return "@" + m.name
	case StateError:
		return truncateNpub(m.pubkey)
	}
	return truncateNpub(m.pubkey)
}

func (m *Model) Focus() {
	m.focused = true
}

func (m *Model) Blur() {
	m.focused = false
}

func (m *Model) IsFocused() bool {
	return m.focused
}

func truncateNpub(pubkey string) string {
	if len(pubkey) >= 16 {
		return "@" + pubkey[:8] + "..."
	}
	return "@" + pubkey
}

func RenderLabel(pubkey, name string, state State) string {
	var text string
	switch state {
	case StateResolved:
		text = "@" + name
	default:
		text = truncateNpub(pubkey)
	}

	style := renderStyleForState(state)
	return style.Render(text)
}

func renderStyleForState(s State) lipgloss.Style {
	switch s {
	case StateLoading:
		return loadingStyle
	case StateResolved:
		return normalStyle
	case StateError:
		return errorStyle
	}
	return loadingStyle
}

type nameResolvedMsg struct {
	name string
}

type LabelClickedMsg struct {
	Pubkey string
}

var (
	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Background(lipgloss.Color("#004400")).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Background(lipgloss.Color("#220000")).
			Padding(0, 1)

	focusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#008800")).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#00FF00")).
			Padding(0, 1)
)