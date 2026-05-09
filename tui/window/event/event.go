package event

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/glamour"
	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/tui/toolkit"
	"github.com/jerry-harm/nosmec/utils"
)

const (
	WindowID = "event"
)

// CloseMsg is sent when the event view should be closed.
type CloseMsg struct{}

type EventView struct {
	event       *nostr.Event
	app         *config.AppContext
	viewport    viewport.Model
	glamour     *glamour.TermRenderer
	width       int
	height      int
	darkBG      bool
	styles      eventStyles
	tk          *toolkit.ToolKit
	authorName  string
	fetchedName bool
	help        help.Model
}

func New(event *nostr.Event, app *config.AppContext, width, height int) *EventView {
	helpModel := help.New()
	helpModel.ShowAll = false

	m := &EventView{
		event:       event,
		app:         app,
		width:       width,
		height:      height,
		darkBG:      false,
		tk:          toolkit.New(),
		authorName:  event.PubKey.Hex()[:8], // placeholder until name loads
		fetchedName: false,
		help:        helpModel,
	}
	m.styles = newStyles(m.darkBG)
	m.viewport = viewport.New(
		viewport.WithWidth(width-4),
		viewport.WithHeight(height-6),
	)

	gl, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width-8),
	)
	m.glamour = gl

	// Initialize help component
	m.help = help.New()
	m.help.ShowAll = false

	// Register keymaps
	m.tk.KeymapAdd("reply", "reply", "r")
	m.tk.KeymapAdd("quote", "quote", "q")
	m.tk.KeymapAdd("delete", "delete", "d")
	m.tk.KeymapAdd("follow", "follow", "f")
	m.tk.KeymapAdd("open", "open in browser", "o")
	m.tk.KeymapAdd("quit", "close", "esc")

	// Set up message handler
	m.tk.SetMsgHandling(WindowID, m.handleMsg)
	m.tk.Focus(WindowID)

	return m
}

func (m *EventView) ID() string {
	return WindowID
}

type helpKeyMap struct {
	reply  key.Binding
	quote  key.Binding
	delete key.Binding
	follow key.Binding
	open   key.Binding
	quit   key.Binding
}

func (h helpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{h.reply, h.quote, h.delete, h.follow, h.open, h.quit}
}

func (h helpKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{h.reply, h.quote, h.delete, h.follow, h.open, h.quit},
	}
}

type ProfileLoadedMsg struct {
	Name string
}

func (m *EventView) Init() tea.Cmd {
	return m.fetchProfileName()
}

func (m *EventView) fetchProfileName() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		name := utils.GetProfileName(ctx, m.event.PubKey, &utils.GetOptions{App: m.app})
		return ProfileLoadedMsg{Name: name}
	}
}

// handleMsg processes key messages for the EventView.
func (m *EventView) handleMsg(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			// Reply to the note
			return m.reply()
		case "q":
			// Quote/repost the note
			return m.quote()
		case "d":
			// Delete the note
			return m.delete()
		case "f":
			// Follow/unfollow the author
			return m.follow()
		case "o":
			// Open in browser
			return m.openInBrowser()
		case "esc":
			// Close the event window
			return func() tea.Msg { return CloseMsg{} }
		}
	}
	return nil
}

func (m *EventView) reply() tea.Cmd {
	// TODO: Open reply input UI
	// For now, just log the intent
	fmt.Println("reply to:", m.event.ID.Hex())
	return nil
}

func (m *EventView) quote() tea.Cmd {
	// TODO: Open quote input UI
	fmt.Println("quote:", m.event.ID.Hex())
	return nil
}

func (m *EventView) delete() tea.Cmd {
	if m.event == nil {
		return nil
	}
	ctx := context.Background()
	_, err := utils.DeleteNote(ctx, m.app, m.event.ID.Hex())
	if err != nil {
		fmt.Println("delete error:", err)
	}
	return nil
}

func (m *EventView) follow() tea.Cmd {
	if m.event == nil {
		return nil
	}
	ctx := context.Background()
	pubkeyHex := m.event.PubKey.Hex()

	// Check if already following by checking subscriptions
	isFollowing := false
	for _, sub := range m.app.ListSubscriptions("user") {
		if sub.ID == pubkeyHex {
			isFollowing = true
			break
		}
	}

	if isFollowing {
		utils.UnfollowUser(ctx, m.app, pubkeyHex)
		fmt.Println("unfollowed:", pubkeyHex)
	} else {
		utils.FollowUser(ctx, m.app, pubkeyHex, "", "")
		fmt.Println("followed:", pubkeyHex)
	}
	return nil
}

func (m *EventView) openInBrowser() tea.Cmd {
	if m.event == nil {
		return nil
	}
	// Build a nostr: URI
	nostrURI := fmt.Sprintf("nostr:%s", m.event.ID.Hex())
	// Try to open using xdg-open on Linux, open on macOS
	var cmd *exec.Cmd
	cmd = exec.Command("xdg-open", nostrURI)
	if err := cmd.Run(); err != nil {
		// Fallback: just print
		fmt.Println("open:", nostrURI)
	}
	return nil
}

func (m *EventView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case ProfileLoadedMsg:
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

		gl, _ := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(m.width-8),
		)
		m.glamour = gl

	case tea.BackgroundColorMsg:
		m.darkBG = msg.IsDark()
		m.styles = newStyles(m.darkBG)

		gl, _ := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(m.width-8),
		)
		m.glamour = gl

	case tea.KeyMsg:
		// Let the handler process the key
		cmd = m.tk.HandleMsg(WindowID, msg)
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *EventView) View() tea.View {
	content := m.renderContent()
	m.viewport.SetContent(content)

	v := tea.NewView(
		m.styles.container.Render(m.styles.header.Render(m.renderHeader())) +
			"\n" +
			m.viewport.View() +
			"\n" +
			m.styles.footer.Render(m.help.View(m.helpKeyMap())),
	)
	v.AltScreen = true
	return v
}

func (m *EventView) helpKeyMap() help.KeyMap {
	return helpKeyMap{
		reply:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reply")),
		quote:  key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quote")),
		delete: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		follow: key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "follow")),
		open:   key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open")),
		quit:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "close")),
	}
}

var _ help.KeyMap = (*helpKeyMap)(nil)

func (m *EventView) renderFooterHelp() string {
	helps := m.tk.KeymapHelpStrings()
	if len(helps) == 0 {
		return ""
	}
	joined := strings.Join(helps, " · ")
	return joined
}

func (m *EventView) Close() bool {
	return true
}
