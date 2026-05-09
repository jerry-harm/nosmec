package event

import (
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/glamour"
	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
)

const (
	WindowID = "event"
)

type EventView struct {
	event    *nostr.Event
	app      *config.AppContext
	viewport viewport.Model
	glamour  *glamour.TermRenderer
	width    int
	height   int
	darkBG   bool
	styles   eventStyles
}

func New(event *nostr.Event, app *config.AppContext, width, height int) *EventView {
	m := &EventView{
		event:  event,
		app:    app,
		width:  width,
		height: height,
		darkBG: false,
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

	return m
}

func (m *EventView) ID() string {
	return WindowID
}

func (m *EventView) Init() tea.Cmd {
	return nil
}

func (m *EventView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
	}

	var cmd tea.Cmd
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
			m.styles.footer.Render("esc: close"),
	)
	v.AltScreen = true
	return v
}

func (m *EventView) Close() bool {
	return true
}