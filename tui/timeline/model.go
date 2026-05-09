package timeline

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/tui/window/event"
	"github.com/jerry-harm/nosmec/tui/windowmanager"
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

	windowManager *windowmanager.WindowManager

	// Infinite scroll state
	isLoadingMore bool
	hasMoreOld    bool
	seenEventIDs  map[nostr.ID]bool

	// Subscription for real-time updates
	subCh        chan nostr.RelayEvent
	subCtx       context.Context
	subCancel     context.CancelFunc
	subStarted    bool
	newestSince   nostr.Timestamp  // Timestamp of newest event for subscription
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

type namesMsg struct {
	names map[string]string
}

type loadMoreMsg struct {
	events []utils.TimelineEvent
	isNew  bool // true if loading newer (prepend), false if loading older (append)
}

type loadMoreErrorMsg struct {
	err error
	isNew bool
}

type newEventMsg struct {
	event utils.TimelineEvent
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
	m.windowManager = windowmanager.New()
	m.seenEventIDs = make(map[nostr.ID]bool)
	m.hasMoreOld = true
	m.isLoadingMore = false

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
			nostrEvents, e := utils.GetGlobalTimeline(ctx, m.limit, 0, opts)
			if e != nil {
				err = e
			} else {
				for _, e := range nostrEvents {
					events = append(events, utils.TimelineEvent{Event: e})
				}
			}
		case "mine":
			nostrEvents, e := utils.GetMyTimeline(ctx, m.limit, 0, opts)
			if e != nil {
				err = e
			} else {
				for _, e := range nostrEvents {
					events = append(events, utils.TimelineEvent{Event: e})
				}
			}
		default:
			events, err = utils.GetFollowedTimeline(ctx, m.limit, 0, m.hashtags, opts)
		}

		if err != nil {
			return errorMsg{err: err}
		}
		return fetchMsg{events: events}
	}
}

func (m *model) fetchProfileNames(pubkeys []string) tea.Cmd {
	return func() tea.Msg {
		names := make(map[string]string)
		var wg sync.WaitGroup

		for _, pk := range pubkeys {
			wg.Add(1)
			go func(pubkeyStr string) {
				defer wg.Done()
				var pubKey nostr.PubKey
				if err := pubKey.UnmarshalJSON([]byte("\"" + pubkeyStr + "\"")); err == nil {
					if name := utils.GetProfileName(context.Background(), pubKey, &utils.GetOptions{App: m.app}); name != "" {
						names[pubkeyStr] = name
					}
				}
			}(pk)
		}

		wg.Wait()
		return namesMsg{names: names}
	}
}

func (m *model) fetchMoreOld() tea.Cmd {
	return func() tea.Msg {
		if m.isLoadingMore {
			return nil
		}
		m.isLoadingMore = true

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		opts := &utils.GetOptions{App: m.app}

		// Get the oldest event's timestamp for pagination
		items := m.list.Items()
		if len(items) == 0 {
			m.isLoadingMore = false
			return loadMoreErrorMsg{err: fmt.Errorf("no events to paginate"), isNew: false}
		}
		oldestItem := items[len(items)-1]
		var oldestEvent nostr.Event
		if it, ok := oldestItem.(item); ok {
			oldestEvent = it.event.Event
		} else {
			m.isLoadingMore = false
			return loadMoreErrorMsg{err: fmt.Errorf("invalid item type"), isNew: false}
		}

		until := oldestEvent.CreatedAt - 1 // Use timestamp just before the oldest

		var events []utils.TimelineEvent
		var err error

		switch m.filter {
		case "global":
			nostrEvents, e := utils.GetGlobalTimeline(ctx, m.limit, until, opts)
			if e != nil {
				err = e
			} else {
				for _, e := range nostrEvents {
					events = append(events, utils.TimelineEvent{Event: e})
				}
			}
		case "mine":
			nostrEvents, e := utils.GetMyTimeline(ctx, m.limit, until, opts)
			if e != nil {
				err = e
			} else {
				for _, e := range nostrEvents {
					events = append(events, utils.TimelineEvent{Event: e})
				}
			}
		default:
			events, err = utils.GetFollowedTimeline(ctx, m.limit, until, m.hashtags, opts)
		}

		m.isLoadingMore = false

		if err != nil {
			return loadMoreErrorMsg{err: err, isNew: false}
		}

		// If we got fewer events than limit, we've reached the end
		if len(events) < m.limit {
			m.hasMoreOld = false
		}

		return loadMoreMsg{events: events, isNew: false}
	}
}

func (m *model) startSubscription(since nostr.Timestamp) tea.Cmd {
	return func() tea.Msg {
		// Cancel any existing subscription
		if m.subCancel != nil {
			m.subCancel()
		}

		ctx, cancel := context.WithCancel(context.Background())
		m.subCtx = ctx
		m.subCancel = cancel

		relays := m.app.Config().KnownRelays
		if len(relays) == 0 {
			relays = m.app.AllReadableRelays()
		}
		privateRelays := m.app.PrivateRelays()
		relays = append(relays, privateRelays...)

		// Build filter based on timeline type
		var filter nostr.Filter
		switch m.filter {
		case "global":
			filter = nostr.Filter{
				Kinds: []nostr.Kind{nostr.KindTextNote},
				Since: since,
				Limit: 100,
			}
		case "mine":
			pubKey, err := m.app.GetMyPubKey()
			if err != nil {
				return errorMsg{err: err}
			}
			filter = nostr.Filter{
				Kinds:   []nostr.Kind{nostr.KindTextNote},
				Authors: []nostr.PubKey{pubKey},
				Since:   since,
				Limit:   100,
			}
		default:
			// For followed timeline, we need authors and communities
			subs := m.app.ListSubscriptions("")
			var authors []nostr.PubKey
			var communityAddrs []string

			for _, sub := range subs {
				switch sub.Type {
				case "user":
					pk, err := utils.ResolveAliasToPubKey(m.app, sub.ID)
					if err == nil {
						authors = append(authors, pk)
					}
				case "community":
					communityAddrs = append(communityAddrs, sub.ID)
				}
			}

			kinds := []nostr.Kind{nostr.KindTextNote, nostr.KindComment}
			filter = nostr.Filter{
				Kinds: kinds,
				Since: since,
				Limit: 100,
			}
			if len(authors) > 0 {
				filter.Authors = authors
			}
			if len(communityAddrs) > 0 {
				filter.Tags = nostr.TagMap{"a": communityAddrs}
			}
		}

		m.subCh = m.app.Pool().SubscribeMany(ctx, relays, filter, nostr.SubscriptionOptions{Label: "timeline"})
		m.subStarted = true

		// Return nil - subscription channel will be processed in Update via pollSubscription
		return nil
	}
}

// pollSubscription is called from Update to check for new subscription events
func (m *model) pollSubscription() tea.Cmd {
	return func() tea.Msg {
		if m.subCh == nil || m.subCtx == nil {
			return nil
		}

		select {
		case <-m.subCtx.Done():
			return nil
		case relayEvent, ok := <-m.subCh:
			if !ok {
				m.subCh = nil
				return nil
			}

			event := relayEvent.Event

			// Deduplicate by event ID
			if m.seenEventIDs[event.ID] {
				// Continue polling via timer
				return tea.Tick(time.Millisecond*100, func(time.Time) tea.Msg {
					return pollSubMsg{}
				})
			}
			m.seenEventIDs[event.ID] = true

			return newEventMsg{event: utils.TimelineEvent{Event: event}}
		default:
			// No event ready, continue polling via timer
			return tea.Tick(time.Millisecond*100, func(time.Time) tea.Msg {
				return pollSubMsg{}
			})
		}
	}
}

type pollSubMsg struct{}

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
		// Forward resize to WindowManager
		m.windowManager.ResizeAll(msg.Width, msg.Height)
		return m, nil

	case fetchMsg:
		m.list.StopSpinner()
		items := make([]list.Item, 0, len(msg.events))
		pubkeys := make([]string, 0, len(msg.events))
		seenPubkeys := make(map[string]bool)

		// Clear seen IDs on fresh fetch and reset subscription state
		m.seenEventIDs = make(map[nostr.ID]bool)
		m.hasMoreOld = true
		m.subStarted = false
		// Cancel any existing subscription
		if m.subCancel != nil {
			m.subCancel()
			m.subCancel = nil
			m.subCh = nil
		}

		// Track newest event timestamp for subscription
		var newestTimestamp nostr.Timestamp
		if len(msg.events) > 0 {
			newestTimestamp = msg.events[0].Event.CreatedAt
		}

		for _, e := range msg.events {
			// Track seen IDs for deduplication
			m.seenEventIDs[e.Event.ID] = true

			kind := detectEventKind(e)
			pubkeyStr := e.Event.PubKey.Hex()
			authorName := pubkeyStr[:8] // placeholder

			if !seenPubkeys[pubkeyStr] {
				seenPubkeys[pubkeyStr] = true
				pubkeys = append(pubkeys, pubkeyStr)
			}

			items = append(items, item{
				event:      e,
				authorName: authorName,
				kind:       kind,
			})
		}
		m.list.SetItems(items)

		// Start subscription after initial fetch
		if newestTimestamp > 0 {
			m.newestSince = newestTimestamp
		}

		// Fetch profile names asynchronously
		if len(pubkeys) > 0 {
			return m, m.fetchProfileNames(pubkeys)
		}
		// No pubkeys to fetch, start subscription immediately
		if m.newestSince > 0 && !m.subStarted {
			cmd := m.startSubscription(m.newestSince)
			// Return batch of commands: start subscription and begin polling
			return m, tea.Batch(cmd, m.pollSubscription())
		}
		return m, nil

	case namesMsg:
		// Update author names in items and refresh the list
		currentItems := m.list.Items()
		for i, listItem := range currentItems {
			if it, ok := listItem.(item); ok {
				pubkeyStr := it.event.Event.PubKey.Hex()
				if name, ok := msg.names[pubkeyStr]; ok {
					it.authorName = name
					currentItems[i] = it
				}
			}
		}
		m.list.SetItems(currentItems)

		// Start subscription after profile names are loaded
		if m.newestSince > 0 && !m.subStarted {
			cmd := m.startSubscription(m.newestSince)
			// Return batch of commands: start subscription and begin polling
			return m, tea.Batch(cmd, m.pollSubscription())
		}
		return m, nil

	case errorMsg:
		statusCmd := m.list.NewStatusMessage(m.styles.statusMessage.Render("Error: " + msg.err.Error()))
		m.list.StopSpinner()
		return m, statusCmd

	case showDetailMsg:
		// Create EventView and open it in WindowManager
		ev := event.New(&msg.event.Event, m.app, m.width, m.height)
		m.windowManager.Open(ev)
		return m, nil

	case closeDetailMsg:
		// Close the event window
		m.windowManager.Close(event.WindowID)
		return m, nil

	case event.CloseMsg:
		// Close event window when ESC is pressed in EventView
		m.windowManager.Close(event.WindowID)
		return m, nil

	case loadMoreMsg:
		m.list.StopSpinner()
		if len(msg.events) == 0 {
			return m, nil
		}

		currentItems := m.list.Items()
		var newItems []list.Item
		pubkeys := make([]string, 0, len(msg.events))
		seenPubkeys := make(map[string]bool)

		for _, e := range msg.events {
			// Deduplicate by event ID
			if m.seenEventIDs[e.Event.ID] {
				continue
			}
			m.seenEventIDs[e.Event.ID] = true

			kind := detectEventKind(e)
			pubkeyStr := e.Event.PubKey.Hex()
			authorName := pubkeyStr[:8] // placeholder

			if !seenPubkeys[pubkeyStr] {
				seenPubkeys[pubkeyStr] = true
				pubkeys = append(pubkeys, pubkeyStr)
			}

			newItems = append(newItems, item{
				event:      e,
				authorName: authorName,
				kind:       kind,
			})
		}

		if msg.isNew {
			// Prepend newer events
			currentItems = append(newItems, currentItems...)
		} else {
			// Append older events
			currentItems = append(currentItems, newItems...)
		}

		m.list.SetItems(currentItems)

		// Fetch profile names for new items
		if len(pubkeys) > 0 {
			return m, m.fetchProfileNames(pubkeys)
		}
		return m, nil

	case loadMoreErrorMsg:
		m.list.StopSpinner()
		if !msg.isNew {
			m.hasMoreOld = false
		}
		statusCmd := m.list.NewStatusMessage(m.styles.statusMessage.Render("加载失败: " + msg.err.Error()))
		return m, statusCmd

	case newEventMsg:
		// New event from subscription - prepend to list
		kind := detectEventKind(msg.event)
		pubkeyStr := msg.event.Event.PubKey.Hex()

		newItem := item{
			event:      msg.event,
			authorName: pubkeyStr[:8], // placeholder
			kind:       kind,
		}

		currentItems := m.list.Items()
		currentItems = append([]list.Item{newItem}, currentItems...)
		m.list.SetItems(currentItems)

		// Fetch profile name for new item and continue polling
		return m, tea.Batch(
			m.fetchProfileNames([]string{pubkeyStr}),
			m.pollSubscription(),
		)

	case pollSubMsg:
		// Continue polling for subscription events
		if m.subStarted && m.subCh != nil {
			return m, m.pollSubscription()
		}
		return m, nil
	}

	// If we have open windows, route keys to the focused window
	if m.windowManager.WindowCount() > 0 {
		switch msg.(type) {
		case tea.KeyMsg:
			_, cmd := m.windowManager.UpdateFocused(msg)
			return m, cmd
		}
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
			// Cancel subscription before quitting
			if m.subCancel != nil {
				m.subCancel()
			}
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

	// Infinite scroll: detect when approaching bottom and load more
	if !m.isLoadingMore && m.hasMoreOld && m.list.FilterState() != list.Filtering {
		paginator := m.list.Paginator
		// Trigger load more when we're on the second-to-last page
		if !paginator.OnLastPage() && paginator.Page >= paginator.TotalPages-2 {
			cmds = append(cmds, m.fetchMoreOld())
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *model) View() tea.View {
	// If we have open windows, render them via WindowManager
	if m.windowManager.WindowCount() > 0 {
		v := tea.NewView(m.windowManager.View())
		v.AltScreen = true
		return v
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