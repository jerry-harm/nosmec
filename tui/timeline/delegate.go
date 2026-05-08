package timeline

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

type timelineDelegate struct {
	list.DefaultDelegate
	width  int
	height int
}

func newTimelineDelegate(width, height int) timelineDelegate {
	d := list.NewDefaultDelegate()
	d.ShowDescription = true
	d.SetHeight(height)
	d.SetSpacing(0)

	return timelineDelegate{
		DefaultDelegate: d,
		width:          width,
		height:         height,
	}
}

func (d timelineDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	var title, desc string

	if i, ok := item.(eventItem); ok {
		title = i.Title()
		desc = i.Description()
	} else {
		return
	}

	if m.Width() <= 0 {
		return
	}

	textwidth := m.Width() - 4

	title = ansi.Truncate(title, textwidth, ellipsis)

	if d.ShowDescription {
		var lines []string
		for i, line := range strings.Split(desc, "\n") {
			if i >= d.height-1 {
				break
			}
			lines = append(lines, ansi.Truncate(line, textwidth, ellipsis))
		}
		desc = strings.Join(lines, "\n")
	}

	isSelected := index == m.Index()

	if isSelected {
		title = SelectedTitleStyle.Render(title)
		desc = SelectedDescStyle.Render(desc)
	} else {
		title = NormalTitleStyle.Render(title)
		desc = NormalDescStyle.Render(desc)
	}

	if d.ShowDescription {
		fmt.Fprintf(w, "%s\n%s", title, desc)
		return
	}
	fmt.Fprintf(w, "%s", title)
}

var (
	ellipsis = "…"

	NormalTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")).
				Bold(true).
				Padding(0, 0, 0, 2)

	NormalDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#AAAAAA"))

	SelectedTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")).
				Bold(true).
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(lipgloss.Color("#25A065")).
				Padding(0, 0, 0, 1)

	SelectedDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(lipgloss.Color("#25A065")).
				Padding(0, 0, 0, 1)
)
