package event

import (
	"fmt"
	"strings"
)

func (m *EventView) renderHeader() string {
	if m.event == nil {
		return m.styles.header.Render("Loading...")
	}
	e := m.event

	timeStr := e.CreatedAt.Time().Format("2006-01-02 15:04")
	kindStr := fmt.Sprintf("Kind: %d", e.Kind)

	pubkeyStr := e.PubKey.Hex()
	if m.authorName != "" && m.authorName != pubkeyStr[:8] {
		return fmt.Sprintf("@%s (%s…) | %s | %s",
			m.styles.author.Render(m.authorName),
			m.styles.author.Render(pubkeyStr[:8]),
			m.styles.time.Render(timeStr),
			kindStr)
	}
	return fmt.Sprintf("@%s… | %s | %s",
		m.styles.author.Render(pubkeyStr[:8]),
		m.styles.time.Render(timeStr),
		kindStr)
}

func (m *EventView) renderContent() string {
	if m.event == nil {
		return "Loading event..."
	}
	e := m.event

	var tagParts []string
	for _, tag := range e.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "t":
				tagParts = append(tagParts, m.styles.tags.Render("#"+tag[1]))
			case "p":
				tagParts = append(tagParts, m.styles.tags.Render("@"+tag[1]))
			case "e":
				tagParts = append(tagParts, m.styles.tags.Render("→"+tag[1]))
			case "r":
				tagParts = append(tagParts, m.styles.tags.Render(tag[1]))
			}
		}
	}

	content := e.Content
	if m.glamour != nil {
		rendered, err := m.glamour.Render(content)
		if err == nil {
			content = rendered
		}
	}

	var out string
	out += content
	out += "\n"
	out += "\n"

	if len(tagParts) > 0 {
		out += strings.Join(tagParts, " ")
	}

	out += fmt.Sprintf("\nID: %s", m.styles.tags.Render(e.ID.Hex()))

	return out
}