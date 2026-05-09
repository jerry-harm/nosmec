package timeline

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/jerry-harm/nosmec/utils"
)

func newItemDelegate(keys *delegateKeyMap, styles *styles) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		if i, ok := m.SelectedItem().(item); ok {
			switch msg := msg.(type) {
			case tea.KeyPressMsg:
				switch {
				case key.Matches(msg, keys.open):
					return func() tea.Msg {
						return showDetailMsg{event: i.event}
					}
				}
			}
		}
		return nil
	}

	help := []key.Binding{keys.open}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}

type delegateKeyMap struct {
	open key.Binding
}

func (d delegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		d.open,
	}
}

func (d delegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.open,
		},
	}
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		open: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "view"),
		),
	}
}

type showDetailMsg struct {
	event utils.TimelineEvent
}

type closeDetailMsg struct{}