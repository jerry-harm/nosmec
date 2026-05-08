package timeline

import (
	"fmt"

	"charm.land/bubbletea/v2"
)

func (m *TimelineModel) View() tea.View {
	if m.width == 0 {
		m.width = 80
	}
	if m.height == 0 {
		m.height = 24
	}

	itemCount := len(m.list.Items())
	var content string

	if m.loading && itemCount == 0 {
		content = m.renderWithSpinner("Loading events...")
	} else if m.err != nil && itemCount == 0 {
		content = m.renderError()
	} else if itemCount == 0 {
		content = m.renderEmpty()
	} else {
		content = m.list.View()
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m *TimelineModel) renderWithSpinner(text string) string {
	return fmt.Sprintf("\n\n   %s %s\n\n", m.spinner.View(), SpinnerStyle.Render(text))
}

func (m *TimelineModel) renderError() string {
	return fmt.Sprintf("\n\n%40s Error: %v\n\n",
		ErrorStyle.Render("!"),
		m.err,
	)
}

func (m *TimelineModel) renderEmpty() string {
	return fmt.Sprintf("\n\n%40s No events found.\n\n",
		EmptyStyle.Render("•"),
	)
}
