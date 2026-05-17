package event

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/Digital-Shane/treeview/v2"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/tui/bubblon"
	"github.com/jerry-harm/nosmec/utils"
)

func extractParentID(event *nostr.Event) string {
	if event == nil {
		return ""
	}

	eTags := collectETags(event.Tags)
	if len(eTags) == 0 {
		return ""
	}

	// Phase 1: marked e tags (NIP-10 preferred)
	hasMarkers := false
	var rootTagValue string
	var replyTagValue string

	for _, tag := range eTags {
		if len(tag) < 4 || tag[3] == "" {
			continue
		}
		hasMarkers = true
		switch tag[3] {
		case "reply":
			replyTagValue = tag[1]
		case "root":
			rootTagValue = tag[1]
		}
	}

	if hasMarkers {
		if replyTagValue != "" {
			return replyTagValue
		}
		if rootTagValue != "" && rootTagValue != event.ID.Hex() {
			return rootTagValue
		}
		return ""
	}

	// Phase 2: positional e tags (deprecated, backward compat)
	// One e tag: ["e", <id>] → reply to that event
	// Two or more: first = root, last = reply
	if len(eTags) == 1 {
		return eTags[0][1]
	}
	if len(eTags) >= 2 {
		last := eTags[len(eTags)-1]
		if last[1] != event.ID.Hex() {
			return last[1]
		}
	}
	return ""
}

func collectETags(tags nostr.Tags) []nostr.Tag {
	var eTags []nostr.Tag
	for _, tag := range tags {
		if tag[0] == "e" && len(tag) >= 2 && nostr.IsValid32ByteHex(tag[1]) {
			eTags = append(eTags, tag)
		}
	}
	return eTags
}

func extractRootEvent(event *nostr.Event) (rootID nostr.ID, isRoot bool, err error) {
	if event == nil {
		return nostr.ID{}, false, errors.New("nil event")
	}

	// Collect all e tags (don't filter by hex validity yet — we need
	// to report invalid hex as errors rather than silently drop tags)
	var eTags []nostr.Tag
	for tag := range event.Tags.FindAll("e") {
		eTags = append(eTags, tag)
	}
	if len(eTags) == 0 {
		return event.ID, true, nil
	}

	// Phase 1: marked e tags (NIP-10 preferred)
	hasMarkers := false
	hasReplyMarker := false
	var rootTagValue string

	for _, tag := range eTags {
		if len(tag) < 4 || tag[3] == "" {
			continue
		}
		hasMarkers = true
		switch tag[3] {
		case "reply":
			hasReplyMarker = true
		case "root":
			rootTagValue = tag[1]
		}
	}

	if hasMarkers {
		if hasReplyMarker {
			if rootTagValue != "" {
				rootFromTag, err := nostr.IDFromHex(rootTagValue)
				if err != nil {
					return nostr.ID{}, false, err
				}
				return rootFromTag, false, nil
			}
			return event.ID, true, nil
		}
		if rootTagValue != "" && rootTagValue != event.ID.Hex() {
			rootFromTag, err := nostr.IDFromHex(rootTagValue)
			if err != nil {
				return nostr.ID{}, false, err
			}
			return rootFromTag, false, nil
		}
		return event.ID, true, nil
	}

	// Phase 2: positional e tags (deprecated, backward compat)
	// First valid e tag = root, last = reply
	validTags := collectETags(event.Tags)
	if len(validTags) >= 2 {
		rootTagValue = validTags[0][1]
		if rootTagValue != "" && rootTagValue != event.ID.Hex() {
			rootFromTag, err := nostr.IDFromHex(rootTagValue)
			if err != nil {
				return nostr.ID{}, false, err
			}
			return rootFromTag, false, nil
		}
	}
	return event.ID, true, nil
}

type NostrEventProvider struct{}

func (p *NostrEventProvider) ID(event nostr.Event) string {
	return event.ID.Hex()
}

func (p *NostrEventProvider) Name(event nostr.Event) string {
	content := event.Content
	if len(content) > 50 {
		content = content[:47] + "..."
	}
	pubkey := event.PubKey.Hex()[:8]
	return strings.TrimSpace(content) + " (" + pubkey + ")"
}

func (p *NostrEventProvider) ParentID(event nostr.Event) string {
	return extractParentID(&event)
}

type threadTreeView struct {
	event    *nostr.Event
	root     *nostr.Event
	app      *config.AppContext
	tuiModel *treeview.TuiTreeModel[nostr.Event]
	provider *NostrEventProvider
	styles   threadStyles
	keys     threadKeyMap
	ctrl     *bubblon.Controller
	width    int
	height   int

	currentEventID string
	searching      bool
	fetched        bool

	mu sync.Mutex
}

func threadKeyMapCustom() treeview.KeyMap {
	km := treeview.DefaultKeyMap()
	km.Quit = nil
	km.SearchStart = []string{"/"}
	return km
}

func newThreadTreeView(event *nostr.Event, app *config.AppContext, width, height int, ctrl *bubblon.Controller) *threadTreeView {
	m := &threadTreeView{
		event:          event,
		app:            app,
		styles:         newThreadStyles(),
		keys:           newThreadKeyMap(),
		ctrl:           ctrl,
		width:          width,
		height:         height,
		currentEventID: event.ID.Hex(),
		provider:       &NostrEventProvider{},
	}
	m.buildInitialTree()
	return m
}

func (m *threadTreeView) buildInitialTree() {
	events := []*nostr.Event{m.event}

	rootID, isRoot, _ := extractRootEvent(m.event)
	if isRoot {
		m.root = m.event
	} else {
		rootP := nostr.Event{ID: rootID, Content: "[...]", Kind: nostr.KindTextNote}
		events = append(events, &rootP)

		parentID := extractParentID(m.event)
		if parentID != "" && parentID != rootID.Hex() {
			pid, err := nostr.IDFromHex(parentID)
			if err == nil {
				parentP := nostr.Event{
					ID:      pid,
					Content: "[...]",
					Kind:    nostr.KindTextNote,
					Tags:    nostr.Tags{{"e", rootID.Hex(), "", "root"}},
				}
				events = append(events, &parentP)
			}
		}
	}

	tuiModel, _ := m.buildTuiModel(events)
	m.tuiModel = tuiModel
}

func (m *threadTreeView) Init() tea.Cmd {
	return m.fetchThread()
}

func (m *threadTreeView) fetchThread() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), m.app.QueryTimeout())
		defer cancel()

		var events []*nostr.Event

		rootID, isRoot, err := extractRootEvent(m.event)
		if err != nil {
			return threadTreeLoadedMsg{err: err}
		}

		if isRoot {
			m.mu.Lock()
			m.root = m.event
			m.mu.Unlock()
			events = append(events, m.event)
		} else {
			rootEvent, _ := m.fetchRootEvent(ctx, rootID)
			m.mu.Lock()
			m.root = rootEvent
			m.mu.Unlock()
			if rootEvent != nil {
				events = append(events, rootEvent)
			}
		}

		if m.root != nil {
			replyEvents := m.fetchThreadReplies(ctx, m.root.ID)
			events = append(events, replyEvents...)
		}

		m.mu.Lock()
		tuiModel, err := m.buildTuiModel(events)
		m.tuiModel = tuiModel
		m.fetched = true
		m.mu.Unlock()

		return threadTreeLoadedMsg{err: err}
	}
}

func (m *threadTreeView) fetchRootEvent(ctx context.Context, rootID nostr.ID) (*nostr.Event, []string) {
	relays := utils.GetQueryRelays(m.event, m.app)
	if len(relays) == 0 {
		return nil, nil
	}

	filter, err := utils.BuildNoteFilter(rootID.Hex())
	if err != nil {
		return nil, nil
	}

	result := m.app.Pool().QuerySingle(ctx, relays, filter, nostr.SubscriptionOptions{})
	if result == nil {
		return nil, nil
	}

	return &result.Event, relays
}

const maxThreadDepth = 10

func (m *threadTreeView) fetchThreadReplies(ctx context.Context, rootID nostr.ID) []*nostr.Event {
	relays := utils.GetQueryRelays(m.event, m.app)
	if len(relays) == 0 {
		return nil
	}

	var allEvents []*nostr.Event
	seen := map[string]bool{rootID.Hex(): true}
	queryIDs := []string{rootID.Hex()}

	for depth := 0; depth < maxThreadDepth && len(queryIDs) > 0; depth++ {
		filter := nostr.Filter{
			Kinds: []nostr.Kind{nostr.KindTextNote, nostr.KindComment},
			Tags:  nostr.TagMap{"e": queryIDs},
		}

		ctxQuery, cancel := context.WithTimeout(ctx, m.app.QueryTimeout())
		results := m.app.Pool().FetchMany(ctxQuery, relays, filter, nostr.SubscriptionOptions{})
		cancel()

		var nextIDs []string

		for relayEvent := range results {
			ev := relayEvent.Event
			if seen[ev.ID.Hex()] {
				continue
			}
			seen[ev.ID.Hex()] = true

			eventCopy := ev
			allEvents = append(allEvents, &eventCopy)
			nextIDs = append(nextIDs, ev.ID.Hex())
		}

		queryIDs = nextIDs
	}

	return allEvents
}

func (m *threadTreeView) buildTuiModel(events []*nostr.Event) (*treeview.TuiTreeModel[nostr.Event], error) {
	if len(events) == 0 {
		return nil, nil
	}

	var items []nostr.Event
	seen := make(map[string]bool)

	for _, e := range events {
		if seen[e.ID.Hex()] {
			continue
		}
		seen[e.ID.Hex()] = true
		items = append(items, *e)
	}

	for _, e := range events {
		parentID := extractParentID(e)
		if parentID != "" && !seen[parentID] {
			seen[parentID] = true
			id, err := nostr.IDFromHex(parentID)
			if err != nil {
				continue
			}
			items = append(items, nostr.Event{
				Content: "[...]",
				ID:      id,
				Kind:    nostr.KindTextNote,
			})
		}
	}

	tree, err := treeview.NewTreeFromFlatData(
		context.Background(),
		items,
		m.provider,
	)
	if err != nil {
		return nil, err
	}

	if _, err := tree.SetFocusedID(context.Background(), m.currentEventID); err != nil {
		if m.root != nil {
			tree.SetFocusedID(context.Background(), m.root.ID.Hex())
		}
	}

	// Expand all ancestors so the focused node is visible in the tree.
	// Without this, collapsed ancestors hide the current event.
	nodeToParent := make(map[string]string)
	for _, item := range items {
		nodeToParent[m.provider.ID(item)] = m.provider.ParentID(item)
	}
	id := m.currentEventID
	for {
		pid, ok := nodeToParent[id]
		if !ok || pid == "" {
			break
		}
		tree.SetExpanded(context.Background(), pid, true)
		id = pid
	}

	tuiModel := treeview.NewTuiTreeModel(tree,
		treeview.WithTuiWidth[nostr.Event](m.width),
		treeview.WithTuiHeight[nostr.Event](m.height-4),
		treeview.WithTuiKeyMap[nostr.Event](threadKeyMapCustom()),
		treeview.WithTuiDisableNavBar[nostr.Event](true),
		treeview.WithTuiAltScreen[nostr.Event](true),
	)

	return tuiModel, nil
}

type threadTreeLoadedMsg struct {
	err error
}

func (m *threadTreeView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case threadTreeLoadedMsg:
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.tuiModel != nil {
			_, cmd := m.tuiModel.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.KeyPressMsg:
		if msg.String() == "/" {
			m.searching = true
		}

		if key.Matches(msg.Key(), m.keys.quit) {
			if m.searching {
				m.searching = false
				if m.tuiModel != nil {
					_, cmd := m.tuiModel.Update(msg)
					return m, cmd
				}
				return m, nil
			}
			return m, func() tea.Msg { return bubblon.Close() }
		}

		if msg.String() == "enter" {
			if m.searching {
				m.searching = false
				if m.tuiModel != nil {
					_, cmd := m.tuiModel.Update(msg)
					return m, cmd
				}
				return m, nil
			}
			if m.tuiModel != nil {
				focused := m.tuiModel.GetFocusedNode()
				if focused != nil {
					eventPtr := focused.Data()
					if eventPtr != nil {
						ev := *eventPtr
						if ev.Content != "[...]" && ev.ID != [32]byte{} {
							eventView := New(&ev, m.app, m.width, m.height, "", m.ctrl)
							return m, bubblon.Open(eventView)
						}
					}
				}
			}
			return m, nil
		}

		if m.tuiModel != nil {
			_, cmd := m.tuiModel.Update(msg)
			return m, cmd
		}
	}

	if m.tuiModel != nil {
		_, cmd := m.tuiModel.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *threadTreeView) View() tea.View {
	var b strings.Builder

	b.WriteString(m.styles.title.Render("Thread"))
	if !m.fetched {
		b.WriteString(m.styles.placeholder.Render(" fetching..."))
	}
	b.WriteString("\n")

	if m.tuiModel != nil {
		b.WriteString("\n")
		tuiView := m.tuiModel.View()
		b.WriteString(tuiView.Content)
	} else {
		b.WriteString(m.styles.statusMessage.Render("  [no thread data]"))
		b.WriteString("\n")
	}

	b.WriteString(m.styles.helpStyle.Render("\n↑↓ navigate · →← expand/collapse · enter view · esc back"))

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

func truncateContent(content string, maxLen int) string {
	if len(content) > maxLen {
		return content[:maxLen-3] + "..."
	}
	return content
}

func NewThreadTreeView(event *nostr.Event, app *config.AppContext, width, height int, ctrl *bubblon.Controller) *threadTreeView {
	return newThreadTreeView(event, app, width, height, ctrl)
}
