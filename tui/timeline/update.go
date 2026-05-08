package timeline

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbletea/v2"
	"github.com/jerry-harm/nosmec/tui/common"
)

func (m *TimelineModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, common.Keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, common.Keys.Refresh):
			m.refresh()
			return m, tea.Batch(m.spinner.Tick, m.fetchTimeline())
		}
		m.list, cmd = m.list.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil

	case fetchMsg:
		m.loading = false
		if len(msg.events) > 0 {
			items := make([]eventItem, len(msg.events))
			var cmds []tea.Cmd
			for i, e := range msg.events {
				items[i] = eventItem{
					event: e,
					kind:  detectEventKind(e),
					width: m.width,
				}
			}
			m.items = items
			m.list.SetItems(listItemsToItems(items))

			var pubkeys []string
			seen := make(map[string]bool)
			for _, item := range items {
				if !seen[item.event.Event.PubKey.Hex()] {
					seen[item.event.Event.PubKey.Hex()] = true
					pubkeys = append(pubkeys, item.event.Event.PubKey.Hex())
				}
			}

			if len(pubkeys) > 0 {
				for _, pk := range pubkeys {
					cmds = append(cmds, m.fetchUsername(pk))
				}
			}
			return m, tea.Batch(cmds...)
		}
		m.err = nil
		return m, nil

	case usernameMsg:
		for i, item := range m.items {
			if item.event.Event.PubKey.Hex() == msg.pubkey {
				item.authorName = msg.name
				m.items[i] = item
				m.list.SetItem(i, item)
				break
			}
		}
		return m, nil

	case errorMsg:
		m.loading = false
		m.err = msg.err
		return m, nil

	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func listItemsToItems(items []eventItem) []list.Item {
	result := make([]list.Item, len(items))
	for i := range items {
		result[i] = items[i]
	}
	return result
}
