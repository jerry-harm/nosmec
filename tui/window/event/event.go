package event

import (
	"context"
	"fmt"
	"os/exec"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
	tea "charm.land/bubbletea/v2"
	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/logger"
	"github.com/jerry-harm/nosmec/tui/toolkit"
	"github.com/jerry-harm/nosmec/tui/windowmanager"
	"github.com/jerry-harm/nosmec/utils"
)

const (
	WindowID   = "event"
	helpLines  = 3
)

type CloseMsg struct{}

type EventLoadedMsg struct {
	Event *nostr.Event
}

type ProfileLoadedMsg struct {
	Name string
}

type EventView struct {
	event        *nostr.Event
	eventID      string
	app          *config.AppContext
	viewport     viewport.Model
	width        int
	height       int
	darkBG       bool
	styles       eventStyles
	tk           *toolkit.ToolKit
	authorName   string
	fetchedName  bool
	showRawJSON  bool
	loading      bool
	fetchedEvent bool
	help         help.Model
	keys         eventKeyMap

	windowManager *windowmanager.WindowManager
}

type eventKeyMap struct {
	reply   key.Binding
	quote   key.Binding
	delete  key.Binding
	follow  key.Binding
	open    key.Binding
	rawjson key.Binding
	quit    key.Binding
}

func (k eventKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.reply, k.quote, k.delete, k.follow, k.open, k.rawjson, k.quit}
}

func (k eventKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.reply, k.quote, k.delete, k.follow, k.open, k.rawjson, k.quit},
	}
}

var _ help.KeyMap = (*eventKeyMap)(nil)

func New(event *nostr.Event, app *config.AppContext, width, height int, authorName string, wm *windowmanager.WindowManager) *EventView {
	m := &EventView{
		event:         event,
		app:           app,
		width:         width,
		height:        height,
		darkBG:        false,
		tk:            toolkit.New(),
		authorName:    authorName,
		fetchedName:   authorName != "",
		fetchedEvent:  true,
		loading:       false,
		showRawJSON:   false,
		windowManager: wm,
	}
	m.initStyles()
	m.initViewport(width, height)
	m.initKeyBindings()
	m.initToolkit()
	return m
}

func NewFromID(eventID string, app *config.AppContext, width, height int, wm *windowmanager.WindowManager) *EventView {
	m := &EventView{
		eventID:       eventID,
		app:           app,
		width:         width,
		height:        height,
		darkBG:        false,
		tk:            toolkit.New(),
		fetchedName:   false,
		fetchedEvent:  false,
		loading:       true,
		showRawJSON:   false,
		windowManager: wm,
	}
	m.initStyles()
	m.initViewport(width, height)
	m.initKeyBindings()
	m.initToolkit()
	return m
}

func (m *EventView) initStyles() {
	m.styles = newStyles(m.darkBG)
}

func (m *EventView) initViewport(width, height int) {
	borderColor := lipgloss.Color("#25A065")
	if m.darkBG {
		borderColor = lipgloss.Color("#00875A")
	}

	m.viewport = viewport.New()
	m.viewport.SetWidth(width)
	m.viewport.SetHeight(height - helpLines)
	m.viewport.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		PaddingRight(2)
}

func (m *EventView) initKeyBindings() {
	m.keys = eventKeyMap{
		reply:   key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reply")),
		quote:   key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quote")),
		delete:  key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		follow:  key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "follow")),
		open:    key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open")),
		rawjson: key.NewBinding(key.WithKeys("j"), key.WithHelp("j", "json")),
		quit:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "close")),
	}
	m.help = help.New()
	m.help.ShowAll = false
}

func (m *EventView) initToolkit() {
	m.tk.KeymapAdd("reply", "reply", "r")
	m.tk.KeymapAdd("quote", "quote", "q")
	m.tk.KeymapAdd("delete", "delete", "d")
	m.tk.KeymapAdd("follow", "follow", "f")
	m.tk.KeymapAdd("open", "open in browser", "o")
	m.tk.KeymapAdd("rawjson", "raw json", "j")
	m.tk.KeymapAdd("quit", "close", "esc")

	m.tk.SetMsgHandling(WindowID, m.handleMsg)
	m.tk.Focus(WindowID)
}

func (m *EventView) ID() string {
	return WindowID
}

func (m *EventView) Init() tea.Cmd {
	logger.Debug("EventView.Init called", "fetchedName", m.fetchedName, "fetchedEvent", m.fetchedEvent, "loading", m.loading)

	if m.fetchedEvent && !m.fetchedName {
		return m.fetchProfileNameAsync()
	}

	if !m.fetchedEvent && m.eventID != "" {
		return m.fetchEventAsync()
	}

	return nil
}

func (m *EventView) fetchEventAsync() tea.Cmd {
	return func() tea.Msg {
		logger.Debug("fetchEventAsync starting", "eventID", m.eventID)
		event := utils.GetNoteAsync(context.Background(), m.eventID, &utils.GetOptions{App: m.app})
		logger.Debug("fetchEventAsync done", "event", event)
		return EventLoadedMsg{Event: event}
	}
}

func (m *EventView) fetchProfileNameAsync() tea.Cmd {
	return func() tea.Msg {
		logger.Debug("fetchProfileNameAsync starting")
		name := utils.GetProfileNameAsync(context.Background(), m.event.PubKey, &utils.GetOptions{App: m.app})
		logger.Debug("fetchProfileNameAsync done", "name", name)
		return ProfileLoadedMsg{Name: name}
	}
}

func (m *EventView) handleMsg(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		logger.Debug("handleMsg received key", "key", msg.String())
		switch msg.String() {
		case "r":
			return m.reply()
		case "q":
			return m.quote()
		case "d":
			return m.delete()
		case "f":
			return m.follow()
		case "o":
			return m.openInBrowser()
		case "j":
			m.showRawJSON = !m.showRawJSON
			logger.Debug("toggled showRawJSON", "showRawJSON", m.showRawJSON)
			return nil
		case "esc":
			logger.Debug("ESC pressed, sending CloseMsg")
			return func() tea.Msg { return CloseMsg{} }
		}
	}
	return nil
}

func (m *EventView) reply() tea.Cmd {
	if m.event == nil {
		return nil
	}
	if m.windowManager == nil {
		return nil
	}
	m.windowManager.PrepareReply(m.event)
	return m.windowManager.OpenCompose()
}

func (m *EventView) quote() tea.Cmd {
	if m.event == nil {
		return nil
	}
	if m.windowManager == nil {
		return nil
	}
	m.windowManager.PrepareQuote(m.event)
	return m.windowManager.OpenCompose()
}

func (m *EventView) delete() tea.Cmd {
	if m.event == nil {
		return nil
	}
	ctx := context.Background()
	_, err := utils.DeleteNote(ctx, m.app, m.event.ID.Hex())
	if err != nil {
		logger.Error("delete note failed", "error", err.Error())
	}
	return nil
}

func (m *EventView) follow() tea.Cmd {
	if m.event == nil {
		return nil
	}
	ctx := context.Background()
	pubkeyHex := m.event.PubKey.Hex()

	isFollowing := false
	for _, sub := range m.app.ListSubscriptions("user") {
		if sub.ID == pubkeyHex {
			isFollowing = true
			break
		}
	}

	if isFollowing {
		utils.UnfollowUser(ctx, m.app, pubkeyHex)
		logger.Debug("unfollowed user", "pubkey", pubkeyHex)
	} else {
		utils.FollowUser(ctx, m.app, pubkeyHex, "", "")
		logger.Debug("followed user", "pubkey", pubkeyHex)
	}
	return nil
}

func (m *EventView) openInBrowser() tea.Cmd {
	if m.event == nil {
		return nil
	}
	nostrURI := fmt.Sprintf("nostr:%s", m.event.ID.Hex())
	cmd := exec.Command("xdg-open", nostrURI)
	if err := cmd.Run(); err != nil {
		logger.Debug("open in browser failed, fallback to copy", "uri", nostrURI)
	}
	return nil
}

func (m *EventView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case CloseMsg:
		logger.Debug("CloseMsg received, quitting")
		return m, tea.Quit

	case EventLoadedMsg:
		logger.Debug("EventLoadedMsg received", "event", msg.Event)
		m.event = msg.Event
		m.loading = false
		m.fetchedEvent = true
		if m.event != nil {
			return m, m.fetchProfileNameAsync()
		}
		return m, nil

	case ProfileLoadedMsg:
		logger.Debug("ProfileLoadedMsg received", "name", msg.Name)
		if msg.Name != "" {
			m.authorName = msg.Name
		}
		m.fetchedName = true
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.SetWidth(msg.Width)
		m.viewport.SetHeight(msg.Height - helpLines)

	case tea.BackgroundColorMsg:
		m.darkBG = msg.IsDark()
		m.styles = newStyles(m.darkBG)

	case tea.KeyPressMsg:
		cmd = m.tk.HandleMsg(WindowID, msg)
	}

	var viewportCmd tea.Cmd
	m.viewport, viewportCmd = m.viewport.Update(msg)
	if cmd == nil {
		cmd = viewportCmd
	}
	return m, cmd
}

func (m *EventView) View() tea.View {
	header := m.renderHeader()
	if m.loading {
		header = "Loading..."
	}

	content := m.renderContent()
	fullContent := header + "\n\n" + content
	m.viewport.SetContent(fullContent)

	return tea.NewView(m.viewport.View() + "\n" + m.help.View(m.keys))
}

func (m *EventView) Close() bool {
	return true
}