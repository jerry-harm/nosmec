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
// Returns empty string if event is a root (no "reply" marker) or if parent hex is invalid.
func extractParentID(event *nostr.Event) string {
	if event == nil {
		return ""
	}

	for _, tag := range event.Tags {
		if len(tag) >= 4 && tag[0] == "e" && tag[3] == "reply" && nostr.IsValid32ByteHex(tag[1]) {
			return tag[1]
		}
	}
	return ""
}

// extractRootEvent identifies the root event per NIP-10.
// Returns root ID, whether the given event IS the root, and any error.
func extractRootEvent(event *nostr.Event) (rootID nostr.ID, isRoot bool, err error) {
	if event == nil {
		return nostr.ID{}, false, errors.New("nil event")
	}

	// Collect all e tags
	var eTags []nostr.Tag
	for tag := range event.Tags.FindAll("e") {
		eTags = append(eTags, tag)
	}

	// No e tags - this event IS the root (original note)
	if len(eTags) == 0 {
		return event.ID, true, nil
	}

	// Check if this event has a "reply" marker (it's a reply to something)
	hasReplyMarker := false
	for _, tag := range eTags {
		if len(tag) >= 4 && tag[3] == "reply" {
			hasReplyMarker = true
			break
		}
	}

	// If this event has "reply" marker, it's NOT the root - find root from "root" tags
	if hasReplyMarker {
		for _, tag := range eTags {
			if len(tag) >= 4 && tag[3] == "root" {
				rootFromTag, err := nostr.IDFromHex(tag[1])
				if err != nil {
					return nostr.ID{}, false, err
				}
				return rootFromTag, false, nil
			}
		}
		// Has "reply" but no "root" marker - treat event as root
		return event.ID, true, nil
	}

	// No "reply" marker - check for "root" marker
	for _, tag := range eTags {
		if len(tag) >= 4 && tag[3] == "root" {
			// "root" marker means this event IS the root
			return event.ID, true, nil
		}
	}

	// Has e tags but no markers - treat as root per NIP-10
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

	mu sync.Mutex
}

// threadKeyMapCustom returns a TuiTreeModel keymap that works inside a bubblon bubble.
// Esc is removed from the Quit binding so we can handle it ourselves for Close(),
// since the default tea.Quit would exit the entire application.
func threadKeyMapCustom() treeview.KeyMap {
	km := treeview.DefaultKeyMap()
	// Remove "esc" from Quit — we handle esc as bubblon.Close() in our Update
	km.Quit = nil
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

		// Step 2: Fetch all direct replies to root
		if m.root != nil {
			replyEvents := m.fetchRepliesToRoot(ctx, m.root.ID)
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
	// Get relay hints from current event's e tags - use them first
	relayHints := utils.ExtractRelayHints(m.event)

	relays := relayHints
	if len(relays) == 0 {
		relays = m.app.AllReadableRelays()
	}
	if len(relays) == 0 {
		return nil, nil
	}

	filter, err := utils.BuildNoteFilter(rootID.Hex())
	if err != nil {
		return nil, nil
	}

	result := m.app.Pool().QuerySingle(ctx, relays, filter, nostr.SubscriptionOptions{})
	if result == nil {
		// Fallback to all readable relays if e-tag relay hints failed
		fallback := m.app.AllReadableRelays()
		if len(fallback) > 0 && len(relayHints) > 0 {
			result = m.app.Pool().QuerySingle(ctx, fallback, filter, nostr.SubscriptionOptions{})
		}
		if result == nil {
			return nil, nil
		}
	}

	return &result.Event, relays
}

func (m *threadTreeView) fetchRepliesToRoot(ctx context.Context, rootID nostr.ID) []*nostr.Event {
	relays := m.app.AllReadableRelays()
	if len(relays) == 0 {
		return nil
	}

	filter := nostr.Filter{
		Kinds: []nostr.Kind{nostr.KindTextNote, nostr.KindComment},
		Tags:  nostr.TagMap{"e": []string{rootID.Hex()}},
		Limit: 100,
	}

	ctxQuery, cancel := context.WithTimeout(ctx, m.app.QueryTimeout())
	defer cancel()

	results := m.app.Pool().FetchMany(ctxQuery, relays, filter, nostr.SubscriptionOptions{})

	var events []*nostr.Event
	for relayEvent := range results {
		events = append(events, &relayEvent.Event)
	}

	return events
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
		treeview.WithTuiHeight[nostr.Event](m.height-4), // leave room for title + help
		treeview.WithTuiKeyMap[nostr.Event](threadKeyMapCustom()),
		treeview.WithTuiDisableNavBar[nostr.Event](true), // we render our own help bar
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
		// Esc always means "close the bubble" (not quit the app)
		if key.Matches(msg.Key(), m.keys.quit) {
			return m, func() tea.Msg { return bubblon.Close() }
		}

		// Delegate all other keyboard input to the TuiTreeModel for
		// Up/Down/Left/Right/Enter navigation and search
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
