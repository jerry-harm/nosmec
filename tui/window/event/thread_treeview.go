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

// extractParentID extracts the parent event ID from NIP-10 e tags.
// For nested replies: uses the "reply" marker tag.
// For direct replies to root: falls back to the "root" marker tag (no "reply" marker per NIP-10).
// Returns empty string if event is a root (no e tags, or only self-referencing root tag).
func extractParentID(event *nostr.Event) string {
	if event == nil {
		return ""
	}

	var rootTagValue string
	var replyTagValue string

	for _, tag := range event.Tags {
		if len(tag) < 4 || tag[0] != "e" || !nostr.IsValid32ByteHex(tag[1]) {
			continue
		}
		switch tag[3] {
		case "reply":
			replyTagValue = tag[1]
		case "root":
			rootTagValue = tag[1]
		}
	}

	if replyTagValue != "" {
		return replyTagValue
	}
	// Direct reply: only "root" marker, root IS the parent
	if rootTagValue != "" && rootTagValue != event.ID.Hex() {
		return rootTagValue
	}
	return ""
}

// extractRootEvent identifies the root event per NIP-10.
// Returns root ID, whether the given event IS the root, and any error.
func extractRootEvent(event *nostr.Event) (rootID nostr.ID, isRoot bool, err error) {
	if event == nil {
		return nostr.ID{}, false, errors.New("nil event")
	}

	var eTags []nostr.Tag
	for tag := range event.Tags.FindAll("e") {
		eTags = append(eTags, tag)
	}

	if len(eTags) == 0 {
		return event.ID, true, nil
	}

	hasReplyMarker := false
	var rootTagValue string

	for _, tag := range eTags {
		if len(tag) < 4 {
			continue
		}
		switch tag[3] {
		case "reply":
			hasReplyMarker = true
		case "root":
			rootTagValue = tag[1]
		}
	}

	if hasReplyMarker {
		// Nested reply: has "reply" marker → not root, find root from "root" marker
		if rootTagValue != "" {
			rootFromTag, err := nostr.IDFromHex(rootTagValue)
			if err != nil {
				return nostr.ID{}, false, err
			}
			return rootFromTag, false, nil
		}
		// Has "reply" but no "root" marker — treat event as root (legacy)
		return event.ID, true, nil
	}

	if rootTagValue != "" && rootTagValue != event.ID.Hex() {
		// Direct reply: only "root" marker pointing to a different event → NOT root
		rootFromTag, err := nostr.IDFromHex(rootTagValue)
		if err != nil {
			return nostr.ID{}, false, err
		}
		return rootFromTag, false, nil
	}

	// Has e tags but no markers, or root tag points to self → treat as root
	return event.ID, true, nil
}

// NostrEventProvider implements treeview.FlatDataProvider for nostr.Event
type NostrEventProvider struct{}

func (p *NostrEventProvider) ID(event nostr.Event) string {
	return event.ID.Hex()
}

func (p *NostrEventProvider) Name(event nostr.Event) string {
	// Truncate content to 50 chars and append short pubkey
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

// threadTreeView is the tree-based thread view model
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

	// Loading states
	loading   bool
	loadError error

	// Current event tracking for highlighting
	currentEventID string

	// Track treeview search mode so esc can cancel search vs close thread
	searching bool

	mu sync.Mutex
}

// threadKeyMapCustom returns a TuiTreeModel keymap that works inside a bubblon bubble.
// - Esc is removed from Quit → handled locally as bubblon.Close()
// - Enter is removed from Toggle/SearchStart → handled locally for event detail
// - SearchStart uses "/", SearchCancel uses "esc" (default, no conflict)
func threadKeyMapCustom() treeview.KeyMap {
	km := treeview.DefaultKeyMap()
	km.Quit = nil
	km.Toggle = nil
	km.SearchStart = []string{"/"}
	return km
}

func newThreadTreeView(event *nostr.Event, app *config.AppContext, width, height int, ctrl *bubblon.Controller) *threadTreeView {
	return &threadTreeView{
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
}

func (m *threadTreeView) Init() tea.Cmd {
	return m.fetchThread()
}

func (m *threadTreeView) fetchThread() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), m.app.QueryTimeout())
		defer cancel()

		m.mu.Lock()
		m.loading = true
		m.mu.Unlock()

		var events []*nostr.Event

		// Step 1: Identify root per NIP-10
		rootID, isRoot, err := extractRootEvent(m.event)
		if err != nil {
			m.mu.Lock()
			m.loadError = err
			m.loading = false
			m.mu.Unlock()
			return threadTreeLoadedMsg{err: err}
		}

		if isRoot {
			m.mu.Lock()
			m.root = m.event
			m.mu.Unlock()
			events = append(events, m.event)
		} else {
			// Fetch root event using relay hints from current event's e tags
			rootEvent, rootRelays := m.fetchRootEvent(ctx, rootID)
			m.mu.Lock()
			m.root = rootEvent
			m.mu.Unlock()
			if rootEvent != nil {
				events = append(events, rootEvent)
			}
			_ = rootRelays // available for future relay hint collection
		}

		// Step 2: Fetch all replies recursively (multi-level)
		if m.root != nil {
			replyEvents := m.fetchThreadReplies(ctx, m.root.ID)
			events = append(events, replyEvents...)
		}

		// Step 3: Build tree from flat data, then wrap in TuiTreeModel
		tuiModel, err := m.buildTuiModel(events)

		m.mu.Lock()
		m.tuiModel = tuiModel
		m.loading = false
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

	// Try with priority relays
	result := m.app.Pool().QuerySingle(ctx, relays, filter, nostr.SubscriptionOptions{})
	if result == nil {
		return nil, nil
	}

	return &result.Event, relays
}

const maxThreadDepth = 10

// fetchThreadReplies recursively fetches all replies in a thread tree.
// Starts from root and works outward level by level, each level querying
// #e against the previous level's event IDs. Stops when no new events
// are found or maxThreadDepth is reached.
func (m *threadTreeView) fetchThreadReplies(ctx context.Context, rootID nostr.ID) []*nostr.Event {
	relays := utils.GetQueryRelays(m.event, m.app)
	if len(relays) == 0 {
		return nil
	}

	var allEvents []*nostr.Event
	seen := map[string]bool{rootID.Hex(): true}

	// The set of IDs to query #e against in each iteration
	queryIDs := []string{rootID.Hex()}

	for depth := 0; depth < maxThreadDepth && len(queryIDs) > 0; depth++ {
		filter := nostr.Filter{
			Kinds: []nostr.Kind{nostr.KindTextNote, nostr.KindComment},
			Tags:  nostr.TagMap{"e": queryIDs},
		}

		ctxQuery, cancel := context.WithTimeout(ctx, m.app.QueryTimeout())
		results := m.app.Pool().FetchMany(ctxQuery, relays, filter, nostr.SubscriptionOptions{})
		cancel()

		var batch []*nostr.Event
		var nextIDs []string

		for relayEvent := range results {
			ev := relayEvent.Event
			if seen[ev.ID.Hex()] {
				continue
			}
			seen[ev.ID.Hex()] = true

			eventCopy := ev
			batch = append(batch, &eventCopy)
			allEvents = append(allEvents, &eventCopy)

			// Collect event IDs for the next level of querying
			nextIDs = append(nextIDs, ev.ID.Hex())
		}

		queryIDs = nextIDs
	}

	return allEvents
}

// buildTuiModel builds a treeview.Tree from flat events, then wraps it in a
// TuiTreeModel with proper keyboard navigation, custom keymap, and the current
// event pre-focused.
func (m *threadTreeView) buildTuiModel(events []*nostr.Event) (*treeview.TuiTreeModel[nostr.Event], error) {
	if len(events) == 0 {
		return nil, nil
	}

	var items []nostr.Event
	seen := make(map[string]bool)

	// Pass 1: add real events, deduplicating by ID
	for _, e := range events {
		if seen[e.ID.Hex()] {
			continue
		}
		seen[e.ID.Hex()] = true
		items = append(items, *e)
	}

	// Pass 2: add placeholder nodes for missing parents
	for _, e := range events {
		parentID := extractParentID(e)
		if parentID != "" && !seen[parentID] {
			seen[parentID] = true
			id, err := nostr.IDFromHex(parentID)
			if err != nil {
				continue
			}
			placeholder := nostr.Event{
				Content: "[loading...]",
				ID:      id,
				Kind:    nostr.KindTextNote,
			}
			items = append(items, placeholder)
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

	// Focus the current event so the user sees where they are in the thread.
	// Note: SetFocusedID triggers the focus policy which may scroll the viewport
	// to bring the node into view, but the actual visual highlight depends on
	// the NodeProvider used by TuiTreeModel's renderer.
	if _, err := tree.SetFocusedID(context.Background(), m.currentEventID); err != nil {
		// If the current event isn't in the tree (e.g. it's a reply to a parent
		// that hasn't loaded yet), focus the root instead.
		_ = err // non-fatal
		if m.root != nil {
			tree.SetFocusedID(context.Background(), m.root.ID.Hex())
		}
	}

	// Wrap in TuiTreeModel for keyboard navigation and viewport management
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
		if msg.err != nil {
			m.mu.Lock()
			m.loadError = msg.err
			m.mu.Unlock()
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Propagate resize to the TuiTreeModel's viewport
		if m.tuiModel != nil {
			_, cmd := m.tuiModel.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.KeyPressMsg:
		// Track search mode: "/" enters search, esc/enter exit it
		if msg.Text == "/" {
			m.searching = true
		}

		// Delegate to TuiTreeModel first (handles search, nav, expand/collapse)
		var tuiCmd tea.Cmd
		if m.tuiModel != nil {
			_, tuiCmd = m.tuiModel.Update(msg)
		}

		// Esc: cancel search if active, otherwise close thread
		if key.Matches(msg.Key(), m.keys.quit) {
			if m.searching {
				m.searching = false
				return m, tuiCmd
			}
			return m, func() tea.Msg { return bubblon.Close() }
		}

		// Enter while searching = accept search, not open event detail
		if m.searching && msg.Text == "enter" {
			m.searching = false
			return m, tuiCmd
		}

		// Enter: open event detail for the focused tree node (not in search mode)
		if msg.Text == "enter" && m.tuiModel != nil {
			focused := m.tuiModel.GetFocusedNode()
			if focused != nil {
				eventPtr := focused.Data()
				if eventPtr != nil {
					ev := *eventPtr
					if ev.Content != "[loading...]" && ev.ID != [32]byte{} {
						eventView := New(&ev, m.app, m.width, m.height, "", m.ctrl)
						return m, bubblon.Open(eventView)
					}
				}
			}
		}

		// Delegate remaining keyboard input to TuiTreeModel
		if m.tuiModel != nil {
			_, cmd := m.tuiModel.Update(msg)
			return m, cmd
		}
	}

	// Let the TuiTreeModel handle any other messages (background color, etc.)
	if m.tuiModel != nil {
		_, cmd := m.tuiModel.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *threadTreeView) View() tea.View {
	var b strings.Builder

	b.WriteString(m.styles.title.Render("Thread"))
	b.WriteString("\n")

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.loading {
		b.WriteString(m.styles.placeholder.Render("  [loading thread...]"))
		b.WriteString("\n")
		return tea.NewView(b.String())
	}

	if m.loadError != nil {
		b.WriteString(m.styles.statusMessage.Render("  [error: " + m.loadError.Error() + "]"))
		b.WriteString("\n")
	}

	if m.tuiModel != nil {
		// Use the TuiTreeModel's viewport-based rendering which includes
		// focus indicator, expand/collapse markers, and scroll support.
		b.WriteString("\n")
		tuiView := m.tuiModel.View()
		b.WriteString(tuiView.Content)
	} else if m.event != nil {
		// Fallback: show at least the current event when no tree is available
		b.WriteString("\n")
		b.WriteString(m.styles.currentEvent.Render("> " + truncateContent(m.event.Content, 60) + " (" + m.event.PubKey.Hex()[:8] + ")"))
		b.WriteString("\n")
	} else {
		b.WriteString(m.styles.statusMessage.Render("  [no thread data]"))
		b.WriteString("\n")
	}

	b.WriteString(m.styles.helpStyle.Render("\n↑↓ navigate · →← expand/collapse · enter search · esc back"))

	return tea.NewView(b.String())
}

func truncateContent(content string, maxLen int) string {
	if len(content) > maxLen {
		return content[:maxLen-3] + "..."
	}
	return content
}

// NewThreadTreeView creates a new tree-based thread view
func NewThreadTreeView(event *nostr.Event, app *config.AppContext, width, height int, ctrl *bubblon.Controller) *threadTreeView {
	return newThreadTreeView(event, app, width, height, ctrl)
}
