package event

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/logger"
	"github.com/jerry-harm/nosmec/tui/toolkit"
	"github.com/jerry-harm/nosmec/utils"
)

const (
	WindowID = "event"
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
	glamour      *glamour.TermRenderer
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
}

func New(event *nostr.Event, app *config.AppContext, width, height int, authorName string) *EventView {
	m := &EventView{
		event:        event,
		app:          app,
		width:        width,
		height:       height,
		darkBG:       false,
		tk:           toolkit.New(),
		authorName:   authorName,
		fetchedName:  authorName != "",
		fetchedEvent: true,
		loading:      false,
		showRawJSON:  false,
	}
	m.styles = newStyles(m.darkBG)
	m.viewport = viewport.New(
		viewport.WithWidth(width-4),
		viewport.WithHeight(height-6),
	)
	m.glamour = nil

	m.tk.KeymapAdd("reply", "reply", "r")
	m.tk.KeymapAdd("quote", "quote", "q")
	m.tk.KeymapAdd("delete", "delete", "d")
	m.tk.KeymapAdd("follow", "follow", "f")
	m.tk.KeymapAdd("open", "open in browser", "o")
	m.tk.KeymapAdd("rawjson", "raw json", "j")
	m.tk.KeymapAdd("quit", "close", "esc")

	m.tk.SetMsgHandling(WindowID, m.handleMsg)
	m.tk.Focus(WindowID)

	return m
}

func NewFromID(eventID string, app *config.AppContext, width, height int) *EventView {
	m := &EventView{
		eventID:      eventID,
		app:          app,
		width:        width,
		height:       height,
		darkBG:       false,
		tk:           toolkit.New(),
		fetchedName:  false,
		fetchedEvent: false,
		loading:      true,
		showRawJSON:  false,
	}
	m.styles = newStyles(m.darkBG)
	m.viewport = viewport.New(
		viewport.WithWidth(width-4),
		viewport.WithHeight(height-6),
	)
	m.glamour = nil

	m.tk.KeymapAdd("reply", "reply", "r")
	m.tk.KeymapAdd("quote", "quote", "q")
	m.tk.KeymapAdd("delete", "delete", "d")
	m.tk.KeymapAdd("follow", "follow", "f")
	m.tk.KeymapAdd("open", "open in browser", "o")
	m.tk.KeymapAdd("rawjson", "raw json", "j")
	m.tk.KeymapAdd("quit", "close", "esc")

	m.tk.SetMsgHandling(WindowID, m.handleMsg)
	m.tk.Focus(WindowID)

	return m
}

func (m *EventView) ID() string {
	return WindowID
}

func (m *EventView) Init() tea.Cmd {
	logger.Debug("EventView.Init called", "fetchedName", m.fetchedName, "fetchedEvent", m.fetchedEvent, "loading", m.loading)

	if m.fetchedEvent && !m.fetchedName {
		return m.fetchProfileName()
	}

	if !m.fetchedEvent && m.eventID != "" {
		return m.fetchEvent()
	}

	return nil
}

func (m *EventView) fetchEvent() tea.Cmd {
	return func() tea.Msg {
		logger.Debug("fetchEvent starting", "eventID", m.eventID)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		event := utils.GetNote(ctx, m.eventID, &utils.GetOptions{App: m.app})
		logger.Debug("fetchEvent done", "event", event)
		return EventLoadedMsg{Event: event}
	}
}

func (m *EventView) fetchProfileName() tea.Cmd {
	return func() tea.Msg {
		logger.Debug("fetchProfileName starting")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		name := utils.GetProfileName(ctx, m.event.PubKey, &utils.GetOptions{App: m.app})
		logger.Debug("fetchProfileName done", "name", name)
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
	logger.Debug("reply not implemented", "eventID", m.event.ID.Hex())
	return nil
}

func (m *EventView) quote() tea.Cmd {
	if m.event == nil {
		return nil
	}
	logger.Debug("quote not implemented", "eventID", m.event.ID.Hex())
	return nil
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
	case EventLoadedMsg:
		logger.Debug("EventLoadedMsg received", "event", msg.Event)
		m.event = msg.Event
		m.loading = false
		m.fetchedEvent = true
		if m.event != nil {
			return m, m.fetchProfileName()
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
		m.viewport.SetWidth(m.width - 4)
		m.viewport.SetHeight(m.height - 6)

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
	if m.glamour == nil {
		renderer, _ := glamour.NewTermRenderer(
			glamour.WithStyles(styles.DarkStyleConfig),
			glamour.WithWordWrap(m.width-8),
		)
		m.glamour = renderer
	}
	content := m.renderContent()
	m.viewport.SetContent(content)

	header := m.renderHeader()
	if m.loading {
		header = m.styles.header.Render("Loading...")
	}

	helpView := m.renderFooterHelp()

	v := tea.NewView(
		m.styles.container.Render(header) +
			"\n" +
			m.viewport.View() +
			"\n" +
			m.styles.footer.Render(helpView),
	)
	v.AltScreen = true
	return v
}

func (m *EventView) renderFooterHelp() string {
	helps := m.tk.KeymapHelpStrings()
	if len(helps) == 0 {
		return ""
	}
	return strings.Join(helps, " · ")
}

func (m *EventView) Close() bool {
	return true
}