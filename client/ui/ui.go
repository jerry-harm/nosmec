package ui

import tea "github.com/charmbracelet/bubbletea"

type basic struct {
}

func initalBasic() basic {
	return basic{}
}

func (b *basic) Init() tea.Cmd {
	return nil
}

func (b *basic) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

}

func (b *basic) View() string {

}
