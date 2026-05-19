package list

import (
	"context"
	"fmt"
	"os"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"fiatjaf.com/nostr/sdk"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/tui/theme"
	"github.com/jerry-harm/nosmec/utils"
)

type conversationItem struct {
	pubKey    string
	name      string
	latestMsg string
	latestAt  nostr.Timestamp
	fromMe    bool
}

func (c conversationItem) Title() string {
	name := c.name
	if name == "" {
		name = c.pubKey[:16] + "..."
	}
	return name
}

func (c conversationItem) Description() string {
	prefix := "←"
	if c.fromMe {
		prefix = "→"
	}
	content := c.latestMsg
	if len(content) > 50 {
		content = content[:50] + "..."
	}
	return fmt.Sprintf("%s %s", prefix, content)
}

func (c conversationItem) FilterValue() string { return c.name }

type model struct {
	styles styles
	list   list.Model
	keys   *keyMap
	app    *config.AppContext
	items  []conversationItem
	errMsg string
	loaded bool
}

type styles struct {
	app           lipgloss.Style
	title         lipgloss.Style
	statusMessage lipgloss.Style
	itemTitle     lipgloss.Style
	itemDesc      lipgloss.Style
	itemSelected  lipgloss.Style
	helpStyle     lipgloss.Style
}

func newStyles(t *theme.Theme) styles {
	return styles{
		app: lipgloss.NewStyle().Padding(1, 2),
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
		helpStyle: lipgloss.NewStyle().
			Foreground(t.TextMuted),
	}
}

type keyMap struct {
	quit key.Binding
	kill key.Binding
}

func newKeyMap() *keyMap {
	return &keyMap{
		quit: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		kill: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "kill"),
		),
	}
}

func NewModel(app *config.AppContext) *model {
	m := &model{app: app}
	m.styles = newStyles(theme.DefaultTheme(false))
	m.keys = newKeyMap()

	delegate := list.NewDefaultDelegate()
	m.list = list.New(nil, delegate, 0, 0)
	m.list.Title = "DM Conversations"
	m.list.Styles.Title = m.styles.title

	return m
}

func (m *model) Init() tea.Cmd {
	return m.loadConversations()
}

func (m *model) loadConversations() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), m.app.QueryTimeout())
		defer cancel()

		conversations, err := utils.ListDMConversations(ctx, m.app, 50)
		if err != nil {
			return errMsg{err: err.Error()}
		}

		var items []conversationItem
		for _, conv := range conversations {
			item := conversationItem{
				pubKey:    conv.PubKey,
				latestMsg: conv.LatestDM.Content,
				latestAt:  conv.LatestAt,
				fromMe:    conv.LatestDM.FromMe,
			}

			if pk, err := nostr.PubKeyFromHex(conv.PubKey); err == nil {
				pm := m.app.System().FetchProfileMetadata(ctx, pk)
				if pm.Event != nil {
					if meta, err := sdk.ParseMetadata(*pm.Event); err == nil && meta.Name != "" {
						item.name = meta.Name
					}
				}
				if item.name == "" {
					item.name = nip19.EncodeNpub(pk)[:16] + "..."
				}
			}

			items = append(items, item)
		}

		return loadedMsg{items: items}
	}
}

type errMsg struct {
	err string
}

type loadedMsg struct {
	items []conversationItem
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadedMsg:
		m.items = msg.items
		listItems := make([]list.Item, len(m.items))
		for i, item := range m.items {
			listItems[i] = item
		}
		m.list.SetItems(listItems)
		m.loaded = true
		if len(m.items) == 0 {
			return m, m.list.NewStatusMessage(m.styles.statusMessage.Render("No conversations"))
		}
		return m, m.list.NewStatusMessage(m.styles.statusMessage.Render(""))

	case errMsg:
		m.errMsg = msg.err
		return m, m.list.NewStatusMessage(m.styles.statusMessage.Render("Error: "+msg.err))

	case tea.KeyMsg:
		if key.Matches(msg, m.keys.quit) {
			return m, tea.Quit
		}
		if key.Matches(msg, m.keys.kill) {
			os.Exit(0)
		}
		if msg.String() == "enter" {
			if m.loaded && m.list.SelectedItem() != nil {
				selected := m.items[m.list.Index()]
				pk, _ := nostr.PubKeyFromHex(selected.pubKey)
				npubStr := nip19.EncodeNpub(pk)
				return m, func() tea.Msg {
					return openDM{npubOrHex: npubStr}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

type openDM struct {
	npubOrHex string
}

func (m *model) View() tea.View {
	v := tea.NewView(m.styles.app.Render(m.list.View()))
	v.AltScreen = true
	return v
}

func RunDMList(app *config.AppContext) error {
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	m := NewModel(app)
	_, err := tea.NewProgram(m).Run()
	return err
}