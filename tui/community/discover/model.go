package discover

import (
	"context"
	"fmt"
	"os"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/utils"
)

type communityItem struct {
	def utils.CommunityDefinition
}

func (c communityItem) Title() string {
	if c.def.Name != "" {
		return c.def.Name
	}
	return c.def.ID
}

func (c communityItem) Description() string {
	desc := c.def.Description
	if len(desc) > 60 {
		desc = desc[:57] + "..."
	}
	if desc == "" {
		mods := len(c.def.Moderators)
		return fmt.Sprintf("%d moderator(s)", mods)
	}
	return fmt.Sprintf("%s  (%d mods)", desc, len(c.def.Moderators))
}

func (c communityItem) FilterValue() string { return c.def.Name + c.def.ID }

type model struct {
	styles styles
	list   list.Model
	keys   *keyMap
	app    *config.AppContext
	items  []communityItem
	errMsg string
	loaded bool
}

type styles struct {
	app           lipgloss.Style
	title         lipgloss.Style
	statusMessage lipgloss.Style
	helpStyle     lipgloss.Style
}

func newStyles() styles {
	return styles{
		app: lipgloss.NewStyle().Padding(1, 2),
		title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1),
		statusMessage: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")),
		helpStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")),
	}
}

type keyMap struct {
	quit key.Binding
	kill key.Binding
}

func newKeyMap() *keyMap {
	return &keyMap{
		quit: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc/q", "back"),
		),
		kill: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "kill"),
		),
	}
}

func NewModel(app *config.AppContext) *model {
	m := &model{app: app}
	m.styles = newStyles()
	m.keys = newKeyMap()

	delegate := list.NewDefaultDelegate()
	m.list = list.New(nil, delegate, 0, 0)
	m.list.Title = "Community Discovery"
	m.list.Styles.Title = m.styles.title

	return m
}

func (m *model) Init() tea.Cmd {
	return m.loadCommunities()
}

func (m *model) loadCommunities() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), m.app.QueryTimeout())
		defer cancel()

		communities, err := utils.FetchCommunityEvents(ctx, m.app)
		if err != nil {
			return errMsg{err: err.Error()}
		}

		var items []communityItem
		for _, def := range communities {
			items = append(items, communityItem{def: def})
		}

		return loadedMsg{items: items}
	}
}

type errMsg struct {
	err string
}

type loadedMsg struct {
	items []communityItem
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
			return m, m.list.NewStatusMessage(m.styles.statusMessage.Render("No communities found"))
		}
		return m, m.list.NewStatusMessage(m.styles.statusMessage.Render(fmt.Sprintf("%d communities", len(m.items))))

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
				addr := fmt.Sprintf("%d:%s:%s",
					34550,
					selected.def.Moderators[0].Hex(),
					selected.def.ID,
				)
				return m, func() tea.Msg {
					return openCommunity{addr: addr, name: selected.def.Name}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

type openCommunity struct {
	addr string
	name string
}

func (m *model) View() tea.View {
	v := tea.NewView(m.styles.app.Render(m.list.View()))
	v.AltScreen = true
	return v
}

func RunCommunityDiscover(app *config.AppContext) error {
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
