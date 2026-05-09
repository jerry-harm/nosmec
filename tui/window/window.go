package window

import tea "charm.land/bubbletea/v2"

type Window interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	View() tea.View
	ID() string
}