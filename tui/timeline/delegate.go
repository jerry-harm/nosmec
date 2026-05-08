package timeline

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/jerry-harm/nosmec/utils"
)

type delegateKeyMap struct {
	open key.Binding
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		open: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "view"),
		),
	}
}

func (d delegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{d.open}
}

func (d delegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{d.open}}
}

func newItemDelegate(keys *delegateKeyMap, sty *styles) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		if kpm, ok := msg.(tea.KeyPressMsg); ok {
			if key.Matches(kpm, keys.open) {
				if i, ok := m.SelectedItem().(item); ok {
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

type showDetailMsg struct {
	event utils.TimelineEvent
}

type closeDetailMsg struct{}