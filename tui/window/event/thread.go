package event

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/tui/bubblon"
	"github.com/jerry-harm/nosmec/utils"
)

type threadView struct {
	event   *nostr.Event
	root    *nostr.Event
	parent  *nostr.Event
	replies []*nostr.Event
	app     *config.AppContext
	styles  threadStyles
	keys    threadKeyMap
	ctrl    *bubblon.Controller
	width   int
	height  int

	// Loading states
	loadingRoot    bool
	loadingParent  bool
	loadingReplies bool

	// Error states
	rootNotFound   bool
	parentNotFound bool

	// Mutex for thread data
	mu sync.Mutex
}

type threadStyles struct {
	title         lipgloss.Style
	header        lipgloss.Style
	statusMessage lipgloss.Style
	helpStyle     lipgloss.Style
	currentEvent  lipgloss.Style
	placeholder   lipgloss.Style
	rootEvent     lipgloss.Style
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
		currentEvent: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")),
		placeholder: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")),
		rootEvent: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true),
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
		event:  event,
		app:    app,
		styles: newThreadStyles(),
		keys:   newThreadKeyMap(),
		ctrl:   ctrl,
		width:  width,
		height: height,
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

		// Step 1: Identify root event per NIP-10
		_, isRoot, _ := utils.FindRootEvent(m.event)
		if isRoot {
			// Current event is the root
			m.mu.Lock()
			m.root = m.event
			m.mu.Unlock()
		} else {
			// Current event is a reply - fetch parent first
			m.mu.Lock()
			m.loadingParent = true
			m.mu.Unlock()

			parent := utils.GetParentEvent(ctx, m.event, opts)
			m.mu.Lock()
			m.parent = parent
			m.loadingParent = false
			if parent == nil {
				m.parentNotFound = true
			}
			// Parent becomes the root if it has "root" marker, otherwise keep looking
			if parent != nil {
				rootID, _, _ := utils.FindRootEvent(parent)
				if rootID != [32]byte{} {
					rootFilter, _ := utils.BuildNoteFilter(rootID.Hex())
					relays := m.app.AllReadableRelays()
					if len(relays) > 0 {
						ctxRoot, cancelRoot := context.WithTimeout(ctx, m.app.QueryTimeout())
						defer cancelRoot()
						result := m.app.Pool().QuerySingle(ctxRoot, relays, rootFilter, nostr.SubscriptionOptions{})
						if result != nil {
							m.root = &result.Event
						} else {
							m.rootNotFound = true
						}
					}
				} else {
					// Parent is also a reply - use it as root for display
					m.root = parent
				}
			}
			m.mu.Unlock()
		}

		// Step 2: Query all replies to root
		if m.root != nil || (m.event != nil && isRoot) {
			m.mu.Lock()
			m.loadingReplies = true
			m.mu.Unlock()

			var rootID nostr.ID
			if m.root != nil {
				rootID = m.root.ID
			} else {
				rootID = m.event.ID
			}

			replies := utils.QueryRepliesToRoot(ctx, rootID, opts)
			m.mu.Lock()
			m.replies = replies
			m.loadingReplies = false
			m.mu.Unlock()
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
			return m, func() tea.Msg { return bubblon.Close() }
		}
	}

	return m, nil
}

func (m *threadView) View() tea.View {
	var b strings.Builder

	b.WriteString(m.styles.title.Render("Thread"))
	b.WriteString("\n\n")

	m.mu.Lock()
	defer m.mu.Unlock()

	// Root event section
	if m.loadingRoot {
		b.WriteString(m.styles.placeholder.Render("[loading root event...]"))
		b.WriteString("\n\n")
	} else if m.rootNotFound {
		b.WriteString(m.styles.statusMessage.Render("[root event not found]"))
		b.WriteString("\n\n")
	} else if m.root != nil {
		b.WriteString(m.styles.header.Render("Thread root:"))
		b.WriteString("\n")
		b.WriteString(m.renderEvent(m.root, "  ", m.root.ID == m.event.ID))
		b.WriteString("\n")
	} else if m.event != nil {
		// Current event IS the root
		b.WriteString(m.styles.header.Render("Thread root (this event):"))
		b.WriteString("\n")
		b.WriteString(m.renderEvent(m.event, "  ", true))
		b.WriteString("\n")
	}

	// Parent event section
	if m.loadingParent {
		b.WriteString(m.styles.placeholder.Render("[loading parent event...]"))
		b.WriteString("\n\n")
	} else if m.parentNotFound {
		b.WriteString(m.styles.statusMessage.Render("[parent event not found]"))
		b.WriteString("\n\n")
	} else if m.parent != nil {
		b.WriteString(m.styles.header.Render("Parent:"))
		b.WriteString("\n")
		b.WriteString(m.renderEvent(m.parent, "  ", m.event != nil && m.parent.ID == m.event.ID))
		b.WriteString("\n\n")
	}

	// Replies section
	if m.loadingReplies {
		b.WriteString(m.styles.placeholder.Render("[loading replies...]"))
		b.WriteString("\n")
	} else if len(m.replies) > 0 {
		b.WriteString(m.styles.header.Render("Replies:"))
		b.WriteString(" (")
		b.WriteString(m.styles.statusMessage.Render(fmt.Sprintf("%d", len(m.replies))))
		b.WriteString(")\n")
		for _, reply := range m.replies {
			isCurrent := m.event != nil && reply.ID == m.event.ID
			b.WriteString(m.renderEvent(reply, "  ", isCurrent))
			b.WriteString("\n")
		}
	} else {
		b.WriteString(m.styles.statusMessage.Render("[no replies yet]"))
		b.WriteString("\n")
	}

	// Current event indicator
	b.WriteString("\n")
	b.WriteString(m.styles.helpStyle.Render("Current event: "))
	b.WriteString(m.styles.currentEvent.Render(m.eventIDPreview()))
	b.WriteString("\n")

	// Help
	b.WriteString(m.styles.helpStyle.Render("esc: back"))

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

func (m *threadView) renderEvent(event *nostr.Event, indent string, isCurrent bool) string {
	var b strings.Builder
	content := event.Content
	if len(content) > 100 {
		content = content[:100] + "..."
	}

	if isCurrent {
		b.WriteString(m.styles.currentEvent.Render(">"))
	} else {
		b.WriteString(" ")
	}
	b.WriteString(indent)

	// Show short ID
	shortID := event.ID.Hex()[:8]
	b.WriteString(m.styles.header.Render(shortID))
	b.WriteString(" ")

	if isCurrent {
		b.WriteString(m.styles.currentEvent.Render(content))
	} else {
		b.WriteString(content)
	}
	b.WriteString("\n")
	return b.String()
}

func (m *threadView) eventIDPreview() string {
	if m.event == nil {
		return "unknown"
	}
	return m.event.ID.Hex()[:8]
}
