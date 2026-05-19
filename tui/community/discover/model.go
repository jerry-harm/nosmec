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
	"github.com/jerry-harm/nosmec/tui/component/bubblon"
	"github.com/jerry-harm/nosmec/tui/event"
	"github.com/jerry-harm/nosmec/tui/timeline"
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
	loaded bool

	ctrl   *bubblon.Controller
	width  int
	height int
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

func (s styles) setupListDelegate(delegate *list.DefaultDelegate) {
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00")).
		Bold(true)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00"))
	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00"))
	delegate.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA"))
}

type keyMap struct {
	quit       key.Binding
	kill       key.Binding
	refresh    key.Binding
	eventDetail key.Binding
	open       key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.open, k.eventDetail, k.refresh, k.quit, k.kill}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.open, k.eventDetail, k.refresh, k.quit, k.kill},
	}
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
		refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		eventDetail: key.NewBinding(
			key.WithKeys("ctrl+e"),
			key.WithHelp("ctrl+e", "event"),
		),
		open: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "timeline"),
		),
	}
}

func NewModel(app *config.AppContext) *model {
	m := &model{app: app}
	m.styles = newStyles()
	m.keys = newKeyMap()

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#FFFF00")).
		BorderForeground(lipgloss.Color("#25A065"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#DDDDDD"))

	m.list = list.New(nil, delegate, 80, 20)
	m.list.Title = "Community Discovery"
	m.list.Styles.Title = m.styles.title
	m.list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{m.keys.refresh, m.keys.open, m.keys.eventDetail}
	}

	return m
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		m.list.StartSpinner(),
		m.loadCommunities(),
	)
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
		m.list.StopSpinner()
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
		m.list.StopSpinner()
		return m, m.list.NewStatusMessage(m.styles.statusMessage.Render("Error: " + msg.err))

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height)

	case bubblon.Closed:
		// A child view (event detail or timeline) was closed by the controller.
		// The controller already popped the child; we just need to resume rendering.
		return m, nil

	case tea.KeyMsg:
		if key.Matches(msg, m.keys.quit) {
			return m, tea.Quit
		}
		if key.Matches(msg, m.keys.kill) {
			os.Exit(0)
		}
		if key.Matches(msg, m.keys.refresh) {
			m.loaded = false
			m.items = nil
			m.list.SetItems(nil)
			return m, tea.Batch(m.list.StartSpinner(), m.loadCommunities())
		}
		if key.Matches(msg, m.keys.eventDetail) {
			if !m.loaded || m.list.SelectedItem() == nil {
				return m, nil
			}
			selected := m.items[m.list.Index()]
			if selected.def.Event == nil {
				return m, nil
			}
			ev := event.New(selected.def.Event, m.app, m.width, m.height, "", m.ctrl)
			return m, bubblon.Open(ev)
		}
		if key.Matches(msg, m.keys.open) {
			if !m.loaded || m.list.SelectedItem() == nil {
				return m, nil
			}
			selected := m.items[m.list.Index()]
			if len(selected.def.Moderators) == 0 {
				return m, nil
			}
			communityAddr := fmt.Sprintf("%d:%s:%s",
				34550,
				selected.def.Moderators[0].Hex(),
				selected.def.ID,
			)
			tlModel := timeline.NewModel(m.app, "community", nil, 10, communityAddr)
			tlModel.SetBubblonController(m.ctrl)
			tlModel.InjectSize(m.width, m.height)
			return m, bubblon.Open(tlModel)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
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
	ctrl, err := bubblon.New(m)
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	m.ctrl = &ctrl
	_, err = tea.NewProgram(ctrl).Run()
	if err != nil {
		fmt.Println("Error running community discover:", err)
		os.Exit(1)
	}
	return nil
}
