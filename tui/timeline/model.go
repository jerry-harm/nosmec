package timeline

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jerry-harm/nosmec/tui/component/bubblon"
	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/logger"
	"github.com/jerry-harm/nosmec/nostr_sdk"
	"github.com/jerry-harm/nosmec/tui/theme"
	"github.com/jerry-harm/nosmec/tui/event"
	"github.com/jerry-harm/nosmec/tui/component/label"
	"github.com/jerry-harm/nosmec/utils"
)

type TimelineEvent struct {
	Event       nostr.Event
	CommunityID string
	IsCommunity bool
}

type eventKind int

const (
	kindNote eventKind = iota
	kindReply
	kindQuote
	kindRepost
	kindCommunity
)

type item struct {
	event      TimelineEvent
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

func newStyles(t *theme.Theme) styles {
	return styles{
		app: lipgloss.NewStyle().
			Padding(1, 2),
		title: lipgloss.NewStyle().
			Foreground(t.TitleText).
			Background(t.TitleBg).
			Padding(0, 1),
		statusMessage: lipgloss.NewStyle().
			Foreground(t.StatusText),
		itemTitle: lipgloss.NewStyle().
			Foreground(t.TextBright).
			Bold(true),
		itemDesc: lipgloss.NewStyle().
			Foreground(t.TextMuted),
		itemSelected: lipgloss.NewStyle().
			Foreground(t.Selection).
			Bold(true),
		detailBox: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(t.Border).
			Padding(1, 1),
		detailHeader: lipgloss.NewStyle().
			Foreground(t.TextBright).
			Bold(true),
		detailContent: lipgloss.NewStyle().
			Foreground(t.Text),
		helpStyle: lipgloss.NewStyle().
			Foreground(t.TextMuted),
	}
}

type model struct {
	styles       styles
	darkBG       bool
	width        int
	height       int
	list         list.Model
	keys         *listKeyMap
	delegateKeys *delegateKeyMap

	app           *config.AppContext
	filter        string
	hashtags      []string
	limit         int
	communityAddr string

	ctrl *bubblon.Controller

	// Infinite scroll state
	isLoadingMore bool
	hasMoreOld    bool
	seenEventIDs  map[nostr.ID]bool

	// Subscription for real-time updates
	subCh       chan nostr.RelayEvent
	subCtx      context.Context
	subCancel   context.CancelFunc
	subStarted  bool
	newestSince nostr.Timestamp // Timestamp of newest event for subscription
	lastRefresh time.Time       // Last refresh timestamp for rate limiting
}

type listKeyMap struct {
	kill             key.Binding
	refresh          key.Binding
	quit             key.Binding
	toggleSpinner    key.Binding
	toggleTitleBar   key.Binding
	toggleStatusBar  key.Binding
	togglePagination key.Binding
	toggleHelpMenu   key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		quit: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "quit"),
		),
		kill: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "kill"),
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
	events []TimelineEvent
}

type errorMsg struct {
	err error
}

type namesMsg struct {
	names map[string]string
}

type loadMoreMsg struct {
	events []TimelineEvent
	isNew  bool // true if loading newer (prepend), false if loading older (append)
}

type loadMoreErrorMsg struct {
	err   error
	isNew bool
}

type newEventMsg struct {
	event TimelineEvent
}

func NewModel(app *config.AppContext, filter string, hashtags []string, limit int, communityAddr string) *model {
	m := &model{
		app:           app,
		filter:        filter,
		hashtags:      hashtags,
		limit:         limit,
		communityAddr: communityAddr,
	}
	m.styles = newStyles(app.Theme())
	m.keys = newListKeyMap()
	m.delegateKeys = newDelegateKeyMap()
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

// SetBubblonController stores the bubblon controller so the timeline
// model can navigate back to the parent when used as a child view.
func (m *model) SetBubblonController(ctrl *bubblon.Controller) {
	m.ctrl = ctrl
}

// InjectSize is called by discover model before pushing timeline onto bubblon stack
// to ensure timeline has correct dimensions when running as a child view.
func (m *model) InjectSize(width, height int) {
	m.width = width
	m.height = height
	m.updateListProperties()
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

	var t *theme.Theme
	if m.app != nil {
		t = m.app.Theme()
	} else {
		t = theme.DefaultTheme(m.darkBG)
	}
	m.styles = newStyles(t)
	m.list.Styles.Title = m.styles.title
}

func (m *model) fetchTimeline() tea.Cmd {
	return func() tea.Msg {
		if time.Since(m.lastRefresh) < 2*time.Second {
			return nil
		}
		m.lastRefresh = time.Now()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		ext := m.app.System()
		var rawEvents []nostr.Event
		var err error

		switch m.filter {
		case "global":
			rawEvents, err = ext.FetchGlobalTimelinePage(ctx, m.limit, 0)
		case "mine":
			pubKey, pkErr := m.app.GetMyPubKey()
			if pkErr != nil {
				return errorMsg{err: pkErr}
			}
			rawEvents, err = ext.FetchMyTimelinePage(ctx, pubKey, m.limit, 0)
		case "community":
			if _, _, parseErr := utils.ParseCommunityAddr(m.communityAddr); parseErr != nil {
				return fetchMsg{events: nil}
			}

			rawEvents, err = ext.FetchFollowedTimelinePage(ctx, nil, []string{m.communityAddr}, m.limit, 0)
		default:
			subs := m.app.ListSubscriptions("")
			var authors []nostr.PubKey
			var communityAddrs []string
			for _, sub := range subs {
				switch sub.Type {
				case "user":
					pk, aliasErr := utils.ResolveAliasToPubKey(m.app, sub.ID)
					if aliasErr == nil {
						authors = append(authors, pk)
					}
				case "community":
					communityAddrs = append(communityAddrs, sub.ID)
				case "hashtag":
					m.hashtags = append(m.hashtags, sub.ID)
				}
			}
			rawEvents, err = ext.FetchFollowedTimelinePage(ctx, authors, communityAddrs, m.limit*3, 0)
			if err == nil && len(rawEvents) > 0 {
				seen := make(map[nostr.ID]bool)
				filtered := rawEvents[:0]
				for _, ev := range rawEvents {
					if seen[ev.ID] {
						continue
					}
					seen[ev.ID] = true
					if len(m.hashtags) > 0 {
						hasMatch := false
						for _, tag := range ev.Tags {
							if len(tag) >= 2 && tag[0] == "t" {
								for _, h := range m.hashtags {
									if tag[1] == h {
										hasMatch = true
										break
									}
								}
							}
							if hasMatch {
								break
							}
						}
						if !hasMatch {
							continue
						}
					}
					filtered = append(filtered, ev)
				}
				rawEvents = filtered
			}
		}
		if err != nil {
			return errorMsg{err: err}
		}

		events := make([]TimelineEvent, 0, len(rawEvents))
		for i := range rawEvents {
			events = append(events, TimelineEvent{Event: rawEvents[i]})
		}
		return fetchMsg{events: events}
	}
}

func (m *model) fetchProfileNames(pubkeys []string) tea.Cmd {
	return func() tea.Msg {
		if len(pubkeys) == 0 {
			return namesMsg{names: nil}
		}

		var pubKeys []nostr.PubKey
		pkToStr := make(map[string]nostr.PubKey)
		for _, pk := range pubkeys {
			var pubKey nostr.PubKey
			if err := pubKey.UnmarshalJSON([]byte("\"" + pk + "\"")); err == nil {
				pubKeys = append(pubKeys, pubKey)
				pkToStr[pk] = pubKey
			}
		}

		if len(pubKeys) == 0 {
			return namesMsg{names: nil}
		}

		ext := m.app.System()
		profiles := ext.FetchProfilesBatch(context.Background(), pubKeys)
		result := make(map[string]string)
		for pkStr, pk := range pkToStr {
			if evt, ok := profiles[pk]; ok {
				if meta, err := nostr_sdk.ParseMetadata(*evt); err == nil && meta.Name != "" {
					result[pkStr] = meta.Name
				}
			}
		}
		return namesMsg{names: result}
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

		until := oldestEvent.CreatedAt - 1

		ext := m.app.System()
		var rawEvents []nostr.Event
		var err error

		switch m.filter {
		case "global":
			rawEvents, err = ext.FetchGlobalTimelinePage(ctx, m.limit, until)
		case "mine":
			pubKey, pkErr := m.app.GetMyPubKey()
			if pkErr != nil {
				m.isLoadingMore = false
				return loadMoreErrorMsg{err: pkErr, isNew: false}
			}
			rawEvents, err = ext.FetchMyTimelinePage(ctx, pubKey, m.limit, until)
		case "community":
			if _, _, parseErr := utils.ParseCommunityAddr(m.communityAddr); parseErr != nil {
				m.isLoadingMore = false
				return loadMoreErrorMsg{err: parseErr, isNew: false}
			}
			rawEvents, err = ext.FetchFollowedTimelinePage(ctx, nil, []string{m.communityAddr}, m.limit, until)
		default:
			subs := m.app.ListSubscriptions("")
			var authors []nostr.PubKey
			var communityAddrs []string
			for _, sub := range subs {
				switch sub.Type {
				case "user":
					pk, aliasErr := utils.ResolveAliasToPubKey(m.app, sub.ID)
					if aliasErr == nil {
						authors = append(authors, pk)
					}
				case "community":
					communityAddrs = append(communityAddrs, sub.ID)
				case "hashtag":
					m.hashtags = append(m.hashtags, sub.ID)
				}
			}
			rawEvents, err = ext.FetchFollowedTimelinePage(ctx, authors, communityAddrs, m.limit*3, until)
			if err == nil && len(rawEvents) > 0 {
				seen := make(map[nostr.ID]bool)
				filtered := rawEvents[:0]
				for _, ev := range rawEvents {
					if seen[ev.ID] {
						continue
					}
					seen[ev.ID] = true
					if len(m.hashtags) > 0 {
						hasMatch := false
						for _, tag := range ev.Tags {
							if len(tag) >= 2 && tag[0] == "t" {
								for _, h := range m.hashtags {
									if tag[1] == h {
										hasMatch = true
										break
									}
								}
							}
							if hasMatch {
								break
							}
						}
						if !hasMatch {
							continue
						}
					}
					filtered = append(filtered, ev)
				}
				rawEvents = filtered
			}
		}
		if err != nil {
			m.isLoadingMore = false
			return loadMoreErrorMsg{err: err, isNew: false}
		}

		m.isLoadingMore = false

		if len(rawEvents) == 0 {
			m.hasMoreOld = false
			return loadMoreErrorMsg{err: fmt.Errorf("no more events"), isNew: false}
		}

		events := make([]TimelineEvent, 0, len(rawEvents))
		for i := range rawEvents {
			events = append(events, TimelineEvent{Event: rawEvents[i]})
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
		case "community":
			// For community timeline, subscribe to the specific community
			// m.communityAddr is already in format "34550:pubkey:communityid"
			kinds := []nostr.Kind{nostr.KindTextNote, nostr.KindComment}
			filter = nostr.Filter{
				Kinds: kinds,
				Since: since,
				Limit: 100,
			}
			if m.communityAddr != "" {
				filter.Tags = nostr.TagMap{"a": []string{m.communityAddr}}
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

			return newEventMsg{event: TimelineEvent{Event: event}}
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
			npubStr := nip19.EncodeNpub(e.Event.PubKey)
			authorName := npubStr[:16] // placeholder truncated

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
		logger.Debug("showDetailMsg received")
		logger.Debug("about to call event.New")
		ev := event.New(&msg.event.Event, m.app, m.width, m.height, msg.authorName, m.ctrl)
		logger.Debug("event.New returned")
		logger.Debug("EventView created, about to Open")
		logger.Debug("EventView opened, returning")
		return m, bubblon.Open(ev)

	case closeDetailMsg:
		return m, nil

	case event.ProfileLoadedMsg:
		// Forward profile name result to the event view via controller
		_, cmd := m.ctrl.Update(msg)
		return m, cmd

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
			npubStr := nip19.EncodeNpub(e.Event.PubKey)
			authorName := npubStr[:16] // placeholder truncated

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
		npubStr := nip19.EncodeNpub(msg.event.Event.PubKey)

		newItem := item{
			event:      msg.event,
			authorName: npubStr[:16], // placeholder truncated
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

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		if key.Matches(msg, m.keys.kill) {
			os.Exit(0)
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
			if m.ctrl.Models() > 0 {
				return m, func() tea.Msg { return bubblon.Close() }
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
	// Always render the list. The bubblon root controller handles View delegation
	// to child models (EventView, compose, etc.) when they are stacked on top.
	v := tea.NewView(m.styles.app.Render(m.list.View()))
	v.AltScreen = true
	return v
}

func detectEventKind(e TimelineEvent) eventKind {
	ev := e.Event
	if ev.Kind == 6 || ev.Kind == 16 {
		return kindRepost
	}
	if ev.Kind == 1111 {
		return kindCommunity
	}
	if ev.Kind == 1 {
		for _, tag := range ev.Tags {
			if tag[0] == "q" && len(tag) >= 2 {
				return kindQuote
			}
			if tag[0] == "e" && len(tag) >= 2 {
				return kindReply
			}
		}
	}
	return kindNote
}

func formatItemTitle(i item) string {
	pubkey := i.event.Event.PubKey.Hex()
	author := i.authorName

	var prefix string
	if author == "" {
		prefix = label.RenderLabel(pubkey, "", label.StateLoading, theme.Default())
	} else {
		prefix = label.RenderLabel(pubkey, author, label.StateResolved, theme.Default())
	}

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
