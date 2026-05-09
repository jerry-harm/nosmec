package timeline

import (
	"context"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/tui/window/event"
	"github.com/jerry-harm/nosmec/utils"
)

type eventKind int

const (
	kindNote eventKind = iota
	kindReply
	kindQuote
	kindRepost
	kindCommunity
)

type item struct {
	event      utils.TimelineEvent
	authorName string
	kind       eventKind
}

func (i item) Title() string       { return formatItemTitle(i) }
func (i item) Description() string { return formatItemDescription(i.event.Event.Content) }
func (i item) FilterValue() string { return i.event.Event.Content }

type styles struct {
	app           lipgloss.Style
	title         lipgloss.Style
	statusMessage lipgloss.Style
	itemTitle     lipgloss.Style
	itemDesc      lipgloss.Style
	itemSelected  lipgloss.Style
	detailBox     lipgloss.Style
	detailHeader  lipgloss.Style
	detailContent lipgloss.Style
	helpStyle     lipgloss.Style
}

func newStyles(darkBG bool) styles {
	lightDark := lipgloss.LightDark(darkBG)

	return styles{
		app: lipgloss.NewStyle().
			Padding(1, 2),
		title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1),
		statusMessage: lipgloss.NewStyle().
			Foreground(lightDark(lipgloss.Color("#04B575"), lipgloss.Color("#04B575"))),
		itemTitle: lipgloss.NewStyle().
			Foreground(lightDark(lipgloss.Color("#00FF00"), lipgloss.Color("#00875A"))).
			Bold(true),
		itemDesc: lipgloss.NewStyle().
			Foreground(lightDark(lipgloss.Color("#AAAAAA"), lipgloss.Color("#7A7A7A"))),
		itemSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")).
			Bold(true),
		detailBox: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#25A065")).
			Padding(1, 1),
		detailHeader: lipgloss.NewStyle().
			Foreground(lightDark(lipgloss.Color("#00FF00"), lipgloss.Color("#00875A"))).
			Bold(true),
		detailContent: lipgloss.NewStyle().
			Foreground(lightDark(lipgloss.Color("#FFFFFF"), lipgloss.Color("#1A1A1A"))),
		helpStyle: lipgloss.NewStyle().
			Foreground(lightDark(lipgloss.Color("#9A9A9A"), lipgloss.Color("#6B6B6B"))),
	}
}

type model struct {
	styles    styles
	darkBG    bool
	width     int
	height    int
	list      list.Model
	keys      *listKeyMap
	delegateKeys *delegateKeyMap

	app      *config.AppContext
	filter   string
	hashtags []string
	limit    int

	showingDetail bool
	detailEvent   *utils.TimelineEvent
}

type listKeyMap struct {
	refresh         key.Binding
	quit            key.Binding
	toggleSpinner   key.Binding
	toggleTitleBar  key.Binding
	toggleStatusBar key.Binding
	togglePagination key.Binding
	toggleHelpMenu  key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		toggleSpinner: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "toggle spinner"),
		),
		toggleTitleBar: key.NewBinding(
			key.WithKeys("T"),
			key.WithHelp("T", "toggle title"),
		),
		toggleStatusBar: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "toggle status"),
		),
		togglePagination: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "toggle pagination"),
		),
		toggleHelpMenu: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "toggle help"),
		),
	}
}

type fetchMsg struct {
	events []utils.TimelineEvent
}

type errorMsg struct {
	err error
}

func NewModel(app *config.AppContext, filter string, hashtags []string, limit int) *model {
	m := &model{
		app:      app,
		filter:   filter,
		hashtags: hashtags,
		limit:    limit,
	}
	m.styles = newStyles(false)
	m.keys = newListKeyMap()
	m.delegateKeys = newDelegateKeyMap()

	delegate := newItemDelegate(m.delegateKeys, &m.styles)
	groceryList := list.New(nil, delegate, 0, 0)
	groceryList.Title = "Timeline"
	groceryList.Styles.Title = m.styles.title

	groceryList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			m.keys.refresh,
			m.keys.toggleSpinner,
			m.keys.toggleTitleBar,
			m.keys.toggleStatusBar,
			m.keys.togglePagination,
			m.keys.toggleHelpMenu,
		}
	}

	m.list = groceryList

	return m
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		tea.RequestBackgroundColor,
		m.list.StartSpinner(),
		m.fetchTimeline(),
	)
}

func (m *model) updateListProperties() {
	h, v := m.styles.app.GetFrameSize()
	m.list.SetSize(m.width-h, m.height-v)

	m.styles = newStyles(m.darkBG)
	m.list.Styles.Title = m.styles.title
}

func (m *model) fetchTimeline() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		opts := &utils.GetOptions{App: m.app}

		var events []utils.TimelineEvent
		var err error

		switch m.filter {
		case "global":
			nostrEvents, e := utils.GetGlobalTimeline(ctx, m.limit, opts)
			if e != nil {
				err = e
			} else {
				for _, e := range nostrEvents {
					events = append(events, utils.TimelineEvent{Event: e})
				}
			}
		case "mine":
			nostrEvents, e := utils.GetMyTimeline(ctx, m.limit, opts)
			if e != nil {
				err = e
			} else {
				for _, e := range nostrEvents {
					events = append(events, utils.TimelineEvent{Event: e})
				}
			}
		default:
			events, err = utils.GetFollowedTimeline(ctx, m.limit, m.hashtags, opts)
		}

		if err != nil {
			return errorMsg{err: err}
		}

		return fetchMsg{events: events}
	}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.darkBG = msg.IsDark()
		m.updateListProperties()
		return m, nil

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.updateListProperties()
		return m, nil

	case fetchMsg:
		items := make([]list.Item, 0, len(msg.events))
		for _, e := range msg.events {
			kind := detectEventKind(e)
			authorName := utils.GetProfileName(context.Background(), e.Event.PubKey, &utils.GetOptions{App: m.app})
			items = append(items, item{
				event:      e,
				authorName: authorName,
				kind:       kind,
			})
		}
		m.list.SetItems(items)
		m.list.StopSpinner()
		return m, nil

	case errorMsg:
		statusCmd := m.list.NewStatusMessage(m.styles.statusMessage.Render("Error: " + msg.err.Error()))
		m.list.StopSpinner()
		return m, statusCmd

	case showDetailMsg:
		m.showingDetail = true
		m.detailEvent = &msg.event
		return m, nil

	case closeDetailMsg:
		m.showingDetail = false
		m.detailEvent = nil
		return m, nil
	}

	if m.showingDetail {
		return m.updateDetail(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.refresh):
			cmds = append(cmds, m.list.StartSpinner())
			cmds = append(cmds, m.fetchTimeline())

		case key.Matches(msg, m.keys.toggleSpinner):
			cmd := m.list.ToggleSpinner()
			return m, cmd

		case key.Matches(msg, m.keys.quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.toggleTitleBar):
			v := !m.list.ShowTitle()
			m.list.SetShowTitle(v)
			m.list.SetShowFilter(v)
			m.list.SetFilteringEnabled(v)
			return m, nil

		case key.Matches(msg, m.keys.toggleStatusBar):
			m.list.SetShowStatusBar(!m.list.ShowStatusBar())
			return m, nil

		case key.Matches(msg, m.keys.togglePagination):
			m.list.SetShowPagination(!m.list.ShowPagination())
			return m, nil

		case key.Matches(msg, m.keys.toggleHelpMenu):
			m.list.SetShowHelp(!m.list.ShowHelp())
			return m, nil
		}
	}

	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) updateDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyPressMsg:
		return m, func() tea.Msg { return closeDetailMsg{} }
	}
	return m, nil
}

func (m *model) View() tea.View {
	if m.showingDetail && m.detailEvent != nil {
		ev := event.New(&m.detailEvent.Event, m.app, m.width, m.height)
		return ev.View()
	}
	v := tea.NewView(m.styles.app.Render(m.list.View()))
	v.AltScreen = true
	return v
}

func detectEventKind(e utils.TimelineEvent) eventKind {
	ev := e.Event
	if ev.Kind == 6 || ev.Kind == 16 {
		return kindRepost
	}
	if ev.Kind == 1111 {
		return kindCommunity
	}
	if ev.Kind == 1 {
		for _, tag := range ev.Tags {
			if len(tag) >= 4 && tag[0] == "e" && tag[3] == "reply" {
				return kindReply
			}
			if len(tag) >= 2 && tag[0] == "q" {
				return kindQuote
			}
		}
	}
	return kindNote
}

func formatItemTitle(i item) string {
	author := i.authorName
	if author == "" {
		author = i.event.Event.PubKey.Hex()[:8]
	}

	prefix := "@" + author

	switch i.kind {
	case kindReply:
		prefix += " [Reply]"
	case kindQuote:
		prefix += " [Quote]"
	case kindRepost:
		prefix += " [Repost]"
	case kindCommunity:
		prefix += " [Community]"
	default:
		prefix += " [Note]"
	}

	return prefix
}

func formatItemDescription(content string) string {
	if content == "" {
		return ""
	}
	wrapped := lipgloss.Wrap(content, 60, " \t\n")
	lines := strings.Split(wrapped, "\n")
	if len(lines) > 3 {
		lines = lines[:3]
		lines[len(lines)-1] = lines[len(lines)-1] + "..."
	}
	return strings.Join(lines, "\n")
}