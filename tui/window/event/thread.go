package event

import (
	"context"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/tui/bubblon"
	"github.com/jerry-harm/nosmec/utils"
)

type threadView struct {
	event      *nostr.Event
	parent    *nostr.Event
	replies   []*nostr.Event
	app       *config.AppContext
	styles    threadStyles
	keys      threadKeyMap
	ctrl      *bubblon.Controller
	width     int
	height    int
}

type threadStyles struct {
	title         lipgloss.Style
	header        lipgloss.Style
	statusMessage lipgloss.Style
	helpStyle     lipgloss.Style
}

func newThreadStyles() threadStyles {
	return threadStyles{
		title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1),
		header: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true),
		statusMessage: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")),
		helpStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")),
	}
}

type threadKeyMap struct {
	quit key.Binding
}

func newThreadKeyMap() threadKeyMap {
	return threadKeyMap{
		quit: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	}
}

func NewThreadView(event *nostr.Event, app *config.AppContext, width, height int, ctrl *bubblon.Controller) *threadView {
	m := &threadView{
		event:   event,
		app:     app,
		styles:  newThreadStyles(),
		keys:    newThreadKeyMap(),
		ctrl:    ctrl,
		width:   width,
		height:  height,
	}
	return m
}

func (m *threadView) Init() tea.Cmd {
	return m.fetchThread()
}

func (m *threadView) fetchThread() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), m.app.QueryTimeout())
		defer cancel()

		opts := &utils.GetOptions{App: m.app}

		if m.event != nil {
			m.parent = utils.GetParentEvent(ctx, m.event, opts)
		}

		if m.parent != nil {
			m.replies = getRepliesToEvent(ctx, m.parent, opts)
		}

		return threadLoadedMsg{}
	}
}

type threadLoadedMsg struct{}

func (m *threadView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case threadLoadedMsg:
		return m, nil

	case tea.KeyPressMsg:
		kmsg := msg.(tea.KeyPressMsg)
		if key.Matches(kmsg.Key(), m.keys.quit) {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *threadView) View() tea.View {
	var b strings.Builder

	b.WriteString(m.styles.title.Render("Thread"))
	b.WriteString("\n\n")

	if m.parent != nil {
		b.WriteString(m.styles.header.Render("Parent:"))
		b.WriteString("\n")
		b.WriteString(m.renderEvent(m.parent, "  "))
		b.WriteString("\n\n")
	}

	if len(m.replies) > 0 {
		b.WriteString(m.styles.header.Render("Replies:"))
		b.WriteString("\n")
		for _, reply := range m.replies {
			b.WriteString(m.renderEvent(reply, "  "))
			b.WriteString("\n")
		}
	}

	if m.parent == nil && len(m.replies) == 0 {
		b.WriteString(m.styles.statusMessage.Render("No thread data"))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.styles.helpStyle.Render("esc: back"))

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

func (m *threadView) renderEvent(event *nostr.Event, indent string) string {
	var b strings.Builder
	content := event.Content
	if len(content) > 100 {
		content = content[:100] + "..."
	}
	b.WriteString(indent)
	b.WriteString(content)
	b.WriteString("\n")
	return b.String()
}

func getRepliesToEvent(ctx context.Context, parent *nostr.Event, opts *utils.GetOptions) []*nostr.Event {
	if parent == nil || opts == nil || opts.App == nil {
		return nil
	}

	filter := nostr.Filter{
		Kinds: []nostr.Kind{nostr.KindTextNote, nostr.KindComment},
		Tags:  nostr.TagMap{"e": []string{parent.ID.Hex()}},
		Limit: 50,
	}

	relays := opts.App.AllReadableRelays()
	if len(relays) == 0 {
		return nil
	}

	ctxQuery, cancel := context.WithTimeout(ctx, opts.App.QueryTimeout())
	defer cancel()

	result := opts.App.Pool().QuerySingle(ctxQuery, relays, filter, nostr.SubscriptionOptions{})
	if result == nil {
		return nil
	}

	return []*nostr.Event{&result.Event}
}