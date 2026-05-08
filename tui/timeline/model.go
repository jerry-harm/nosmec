package timeline

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
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
	width      int
}

func (i item) Title() string {
	return formatItemTitle(i)
}

func (i item) Description() string {
	return formatItemDescription(i.event.Event.Content, i.width)
}

func (i item) FilterValue() string {
	return i.event.Event.Content
}

type model struct {
	styles   styles
	darkBG   bool
	width    int
	height   int
	list     list.Model
	keys     *keyMap
	delegateKeys *delegateKeyMap

	app      *config.AppContext
	filter   string
	hashtags []string
	limit    int

	showingDetail bool
	detailEvent   *utils.TimelineEvent
}

type keyMap struct {
	refresh         key.Binding
	quit            key.Binding
	toggleTitleBar  key.Binding
	toggleStatusBar key.Binding
	togglePagination key.Binding
	toggleHelpMenu  key.Binding
	closeDetail     key.Binding
}

func newKeyMap() *keyMap {
	return &keyMap{
		refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
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
		closeDetail: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "close"),
		),
	}
}

type fetchMsg struct {
	events []utils.TimelineEvent
}

type errorMsg struct {
	err error
}

type usernameMsg struct {
	pubkey string
	name   string
}

func NewModel(app *config.AppContext, filter string, hashtags []string, limit int) *model {
	m := model{
		app:      app,
		filter:   filter,
		hashtags: hashtags,
		limit:    limit,
	}
	m.styles = newStyles(false)
	m.keys = newKeyMap()
	m.delegateKeys = newDelegateKeyMap()

	delegate := newItemDelegate(m.delegateKeys, &m.styles)
	l := list.New(nil, delegate, 0, 0)
	l.SetShowTitle(true)
	l.SetShowHelp(true)
	l.SetShowStatusBar(true)
	l.SetShowPagination(true)
	l.SetFilteringEnabled(true)
	l.Title = "Timeline"
	l.Styles.Title = m.styles.title
	m.list = l

	m.list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			m.keys.refresh,
			m.keys.toggleTitleBar,
			m.keys.toggleStatusBar,
			m.keys.togglePagination,
			m.keys.toggleHelpMenu,
		}
	}

	return &m
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		tea.RequestBackgroundColor,
		m.fetchTimeline(),
	)
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

func (m *model) updateListProperties() {
	h, v := m.styles.app.GetFrameSize()
	if !m.showingDetail {
		m.list.SetSize(m.width-h, m.height-v)
	}
	m.styles = newStyles(m.darkBG)
	m.list.Styles.Title = m.styles.title
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.darkBG = msg.IsDark()
		m.updateListProperties()
		return m, nil

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.updateListProperties()
		return m, nil

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

	return m.updateList(msg)
}

func (m *model) updateDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyPressMsg:
		return m, func() tea.Msg { return closeDetailMsg{} }
	case tea.WindowSizeMsg:
		m.width, m.height = msg.(tea.WindowSizeMsg).Width, msg.(tea.WindowSizeMsg).Height
	}
	return m, nil
}

func (m *model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.refresh):
			return m, m.fetchTimeline()

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
	return m, cmd
}

func (m *model) View() tea.View {
	if m.showingDetail {
		v := tea.NewView(m.viewDetail())
		v.AltScreen = true
		return v
	}
	v := tea.NewView(m.styles.app.Render(m.list.View()))
	v.AltScreen = true
	return v
}

func (m *model) viewDetail() string {
	if m.detailEvent == nil {
		return ""
	}

	e := m.detailEvent.Event
	author := e.PubKey.Hex()[:8]
	if profileName := utils.GetProfileName(context.Background(), e.PubKey, &utils.GetOptions{App: m.app}); profileName != "" {
		author = profileName
	}

	timeStr := e.CreatedAt.Time().Format("2006-01-02 15:04")
	kindStr := fmt.Sprintf("Kind: %d", e.Kind)
	idStr := "ID: " + e.ID.Hex()[:16] + "..."

	content := e.Content
	lines := strings.Split(content, "\n")
	maxWidth := m.width - 10
	if maxWidth < 20 {
		maxWidth = 60
	}
	var wrappedLines []string
	for _, line := range lines {
		wrapped := lipgloss.Wrap(line, maxWidth, " \t")
		wrappedLines = append(wrappedLines, strings.Split(wrapped, "\n")...)
	}
	content = strings.Join(wrappedLines, "\n")

	var tagParts []string
	for _, tag := range e.Tags {
		if len(tag) >= 2 {
			if tag[0] == "t" || tag[0] == "p" || tag[0] == "e" {
				tagParts = append(tagParts, "#"+tag[1])
			}
		}
	}
	tagStr := strings.Join(tagParts, " ")

	header := fmt.Sprintf("@%s | %s | %s\n%s | %s", author, timeStr, kindStr, idStr, tagStr)
	helpLine := "esc: close | r: refresh"

	headerText := m.styles.detailHeader.Render(header)
	contentText := m.styles.detailContent.Render(content)
	helpText := m.styles.helpStyle.Render(helpLine)

	box := m.styles.detailBox.Render(headerText + "\n\n" + contentText)

	return m.styles.app.Render(box + "\n" + helpText)
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

func formatItemDescription(content string, width int) string {
	if width < 20 {
		width = 60
	}
	wrapped := lipgloss.Wrap(content, width-10, " \t\n")
	lines := strings.Split(wrapped, "\n")
	if len(lines) > 3 {
		lines = lines[:3]
		lines[len(lines)-1] = lines[len(lines)-1] + "..."
	}
	return strings.Join(lines, "\n")
}

func (m *model) fetchUsername(pubkeyHex string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var pk nostr.PubKey
		copy(pk[:], []byte(pubkeyHex))
		name := utils.GetProfileName(ctx, pk, &utils.GetOptions{App: m.app})
		return usernameMsg{pubkey: pubkeyHex, name: name}
	}
}