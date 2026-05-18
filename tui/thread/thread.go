package thread

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/Digital-Shane/treeview/v2"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip10"
	"fiatjaf.com/nostr/sdk"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/logger"
	"github.com/jerry-harm/nosmec/sdkplus"
	"github.com/jerry-harm/nosmec/tui/component/bubblon"
	"github.com/jerry-harm/nosmec/tui/component/label"
	"github.com/jerry-harm/nosmec/utils"
)

type styles struct {
	title         lipgloss.Style
	statusMessage lipgloss.Style
	helpStyle     lipgloss.Style
	currentEvent  lipgloss.Style
	placeholder   lipgloss.Style
}

func newStyles() styles {
	return styles{
		title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1),
		statusMessage: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")),
		helpStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")),
		currentEvent: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")),
		placeholder: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")),
	}
}

type keyMap struct {
	quit    key.Binding
	refresh key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		quit:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	}
}

func extractParentID(event *nostr.Event) string {
	if event == nil {
		return ""
	}
	ptr := nip10.GetImmediateParent(event.Tags)
	if ptr == nil {
		return ""
	}
	return ptr.AsTagReference()
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
	ptr := nip10.GetThreadRoot(event.Tags)
	if ptr == nil {
		return event.ID, true, nil
	}
	id, err := nostr.IDFromHex(ptr.AsTagReference())
	if err != nil {
		return nostr.ID{}, false, err
	}
	return id, id == event.ID, nil
}

var globalNameCache = make(map[string]string)
var globalNameCacheMu sync.RWMutex

type eventProvider struct{}

func (p *eventProvider) ID(event nostr.Event) string {
	return event.ID.Hex()
}

func (p *eventProvider) Name(event nostr.Event) string {
	content := event.Content
	if len(content) > 50 {
		content = content[:47] + "..."
	}
	pubkey := event.PubKey.Hex()
	globalNameCacheMu.RLock()
	name, ok := globalNameCache[pubkey]
	globalNameCacheMu.RUnlock()
	if ok && name != "" {
		labelStr := label.RenderLabel(pubkey, name, label.StateResolved)
		return strings.TrimSpace(content) + " (" + labelStr + ")"
	}
	labelStr := label.RenderLabel(pubkey, "", label.StateLoading)
	return strings.TrimSpace(content) + " (" + labelStr + ")"
}

func (p *eventProvider) ParentID(event nostr.Event) string {
	return extractParentID(&event)
}

// Model is the tree-based thread view.
type Model struct {
	event    *nostr.Event
	root     *nostr.Event
	app      *config.AppContext
	tuiModel *treeview.TuiTreeModel[nostr.Event]
	provider *eventProvider
	styles   styles
	keys     keyMap
	ctrl     *bubblon.Controller
	width    int
	height   int

	currentEventID string
	searching      bool
	fetched        bool

	nameCache map[string]string

	newEventView func(*nostr.Event) tea.Model

	mu sync.Mutex
}

func keyMapCustom() treeview.KeyMap {
	km := treeview.DefaultKeyMap()
	km.Quit = nil
	km.SearchStart = []string{"/"}
	return km
}

// New creates a new thread view for the given event.
// newEventView is called when the user presses enter on an event node to
// create a detail view for that event.
func New(event *nostr.Event, app *config.AppContext, width, height int, ctrl *bubblon.Controller, newEventView func(*nostr.Event) tea.Model) *Model {
	m := &Model{
		event:          event,
		app:            app,
		styles:         newStyles(),
		keys:           newKeyMap(),
		ctrl:           ctrl,
		width:          width,
		height:         height,
		currentEventID: event.ID.Hex(),
		provider:       &eventProvider{},
		newEventView:   newEventView,
	}
	m.buildInitialTree()
	return m
}

func (m *Model) buildInitialTree() {
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

func (m *Model) Init() tea.Cmd {
	return m.fetchThread()
}

func (m *Model) fetchThread() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), m.app.QueryTimeout())
		defer cancel()

		var events []*nostr.Event

		rootID, isRoot, err := extractRootEvent(m.event)
		if err != nil {
			logger.Error("thread: extractRootEvent failed", "error", err)
			return loadedMsg{err: err}
		}
		logger.Debug("thread: fetchThread start", "eventID", m.event.ID.Hex()[:8],
			"rootID", rootID.Hex()[:8], "isRoot", isRoot,
			"eTags", len(collectETags(m.event.Tags)))

		events = append(events, m.event)

		var parentChain []*nostr.Event
		if isRoot {
			m.mu.Lock()
			m.root = m.event
			m.mu.Unlock()
			logger.Debug("thread: event IS root")
		} else {
			rootEvent, rootRelays := m.fetchRootEvent(ctx, rootID)
			m.mu.Lock()
			m.root = rootEvent
			m.mu.Unlock()
			logger.Debug("thread: fetchRootEvent result", "found", rootEvent != nil,
				"relays", len(rootRelays))
			if rootEvent != nil {
				events = append(events, rootEvent)
			}

			// Walk up the reply chain fetching each parent by ID.
			// #e=Root may not find parents that lack the "root" marker.
			parentChain = m.fetchParentChain(ctx)
			events = append(events, parentChain...)
		}

		replyEvents := m.fetchThreadReplies(ctx, rootID)
		logger.Debug("thread: fetchThreadReplies result", "count", len(replyEvents))
		events = append(events, replyEvents...)

		logger.Debug("thread: total events for tree", "count", len(events))

		m.mu.Lock()
		tuiModel, err := m.buildTuiModel(events)
		if tuiModel != nil {
			m.tuiModel = tuiModel
		}
		m.fetched = true
		m.mu.Unlock()
		logger.Debug("thread: fetchThread done", "tuiModel", tuiModel != nil, "err", err)

		return loadedMsg{err: err}
	}
}

func (m *Model) fetchRootEvent(ctx context.Context, rootID nostr.ID) (*nostr.Event, []string) {
	logger.Debug("thread: fetchRootEvent", "rootID", rootID.Hex()[:8])
	return m.fetchEventByID(ctx, rootID)
}

func (m *Model) fetchParentChain(ctx context.Context) []*nostr.Event {
	var chain []*nostr.Event
	seen := map[string]bool{m.event.ID.Hex(): true}
	current := m.event

	for depth := 0; depth < maxThreadDepth; depth++ {
		parentID := extractParentID(current)
		if parentID == "" || seen[parentID] {
			break
		}
		seen[parentID] = true

		pid, err := nostr.IDFromHex(parentID)
		if err != nil {
			break
		}

		parent, _ := m.fetchEventByID(ctx, pid)
		if parent == nil {
			break
		}

		logger.Debug("thread: fetchParentChain found", "depth", depth,
			"eventID", parent.ID.Hex()[:8])
		chain = append(chain, parent)
		current = parent

		if _, isRoot, _ := extractRootEvent(parent); isRoot {
			break
		}
	}

	return chain
}

func (m *Model) fetchEventByID(ctx context.Context, id nostr.ID) (*nostr.Event, []string) {
	relays := utils.GetQueryRelays(m.event, m.app)
	if len(relays) == 0 {
		relays = m.app.AllReadableRelays()
	}
	if len(relays) == 0 {
		return nil, nil
	}

	filter, err := utils.BuildNoteFilter(id.Hex())
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
const queryBatchSize = 50

func (m *Model) fetchThreadReplies(ctx context.Context, rootID nostr.ID) []*nostr.Event {
	relays := utils.GetQueryRelays(m.event, m.app)
	logger.Debug("thread: fetchThreadReplies", "rootID", rootID.Hex()[:8], "relays", len(relays))
	if len(relays) == 0 {
		relays = m.app.AllReadableRelays()
		logger.Debug("thread: fetchThreadReplies fallback to AllReadableRelays", "relays", len(relays))
	}
	if len(relays) == 0 {
		logger.Debug("thread: fetchThreadReplies no relays")
		return nil
	}

	var allEvents []*nostr.Event
	seen := map[string]bool{rootID.Hex(): true}
	queryIDs := []string{rootID.Hex()}

	for depth := 0; depth < maxThreadDepth && len(queryIDs) > 0; depth++ {
		logger.Debug("thread: fetchThreadReplies depth", "depth", depth, "queryIDs", len(queryIDs))

		var nextIDs []string

		for start := 0; start < len(queryIDs); start += queryBatchSize {
			end := start + queryBatchSize
			if end > len(queryIDs) {
				end = len(queryIDs)
			}
			batch := queryIDs[start:end]

			filter := nostr.Filter{
				Kinds: []nostr.Kind{nostr.KindTextNote, nostr.KindComment},
				Tags:  nostr.TagMap{"e": batch},
			}

			ctxQuery, cancel := context.WithTimeout(ctx, m.app.QueryTimeout())
			results := m.app.Pool().FetchMany(ctxQuery, relays, filter, nostr.SubscriptionOptions{})

			var batchCount int
			for relayEvent := range results {
				batchCount++
				ev := relayEvent.Event
				if seen[ev.ID.Hex()] {
					continue
				}
				seen[ev.ID.Hex()] = true

				eventCopy := ev
				allEvents = append(allEvents, &eventCopy)
				nextIDs = append(nextIDs, ev.ID.Hex())
			}
			cancel()
			logger.Debug("thread: fetchThreadReplies batch", "depth", depth, "batchIDs", len(batch), "found", batchCount)
		}

		logger.Debug("thread: fetchThreadReplies depth done", "depth", depth, "newIDs", len(nextIDs))
		queryIDs = nextIDs
	}

	logger.Debug("thread: fetchThreadReplies done", "total", len(allEvents))
	return allEvents
}

func (m *Model) buildTuiModel(events []*nostr.Event) (*treeview.TuiTreeModel[nostr.Event], error) {
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

	if m.nameCache == nil {
		m.nameCache = make(map[string]string)
	}
	m.fetchProfileNames(items)

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

	tree.SetExpanded(context.Background(), m.currentEventID, true)

	tuiModel := treeview.NewTuiTreeModel(tree,
		treeview.WithTuiWidth[nostr.Event](m.width),
		treeview.WithTuiHeight[nostr.Event](m.height-4),
		treeview.WithTuiKeyMap[nostr.Event](keyMapCustom()),
		treeview.WithTuiDisableNavBar[nostr.Event](true),
		treeview.WithTuiAltScreen[nostr.Event](true),
	)

	return tuiModel, nil
}

type loadedMsg struct {
	err error
}

type namesMsg struct {
	names map[string]string
}

func (m *Model) fetchProfileNames(items []nostr.Event) {
	var pubkeys []string
	for _, e := range items {
		pk := e.PubKey.Hex()
		if _, ok := m.nameCache[pk]; !ok {
			pubkeys = append(pubkeys, pk)
		}
	}
	if len(pubkeys) == 0 {
		return
	}

	go func() {
		var pubKeys []nostr.PubKey
		pkToHex := make(map[nostr.PubKey]string)
		for _, pk := range pubkeys {
			var pubKey nostr.PubKey
			if err := pubKey.UnmarshalJSON([]byte("\"" + pk + "\"")); err == nil {
				pubKeys = append(pubKeys, pubKey)
				pkToHex[pubKey] = pk
			}
		}

		if len(pubKeys) == 0 {
			return
		}

		ext := sdkplus.Wrap(m.app.System())
		profiles := ext.FetchProfilesBatch(context.Background(), pubKeys)
		names := make(map[nostr.PubKey]string)
		for pk, event := range profiles {
			if meta, err := sdk.ParseMetadata(*event); err == nil && meta.Name != "" {
				names[pk] = meta.Name
			}
		}

		m.mu.Lock()
		for pk, name := range names {
			hex := pkToHex[pk]
			m.nameCache[hex] = name
		}
		m.mu.Unlock()
		globalNameCacheMu.Lock()
		for pk, name := range names {
			hex := pkToHex[pk]
			globalNameCache[hex] = name
		}
		globalNameCacheMu.Unlock()
	}()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadedMsg:
		return m, nil

	case namesMsg:
		m.mu.Lock()
		for pk, name := range msg.names {
			m.nameCache[pk] = name
		}
		m.mu.Unlock()
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
						if ev.Content != "[...]" && ev.ID != [32]byte{} && m.newEventView != nil {
							eventView := m.newEventView(&ev)
							return m, bubblon.Open(eventView)
						}
					}
				}
			}
			return m, nil
		}

		if key.Matches(msg.Key(), m.keys.refresh) {
			m.fetched = false
			return m, m.fetchThread()
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

func (m *Model) View() tea.View {
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

	b.WriteString(m.styles.helpStyle.Render("\n↑↓ navigate · →← expand/collapse · enter view · r refresh · esc back"))

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
