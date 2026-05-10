package event

import (
	"encoding/json"
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

	// Line 1: full pubkey
	line1 := fmt.Sprintf("PubKey: %s", m.styles.author.Render(pubkeyStr))

	// Line 2: @username | time | kind
	var namePart string
	if m.authorName != "" {
		namePart = "@" + m.authorName
	} else {
		namePart = "@" + pubkeyStr[:8] + "..."
	}
	line2 := fmt.Sprintf("%s | %s | %s",
		m.styles.author.Render(namePart),
		m.styles.time.Render(timeStr),
		kindStr)

	return m.styles.header.Render(line1+"\n"+line2)
}

func (m *EventView) renderContent() string {
	// Show loading state
	if m.loading || m.event == nil {
		if m.eventID != "" {
			return fmt.Sprintf("Loading event %s...", m.eventID)
		}
		return "Loading event..."
	}

	// Show raw JSON if toggled
	if m.showRawJSON {
		return m.renderRawJSON()
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

func (m *EventView) renderRawJSON() string {
	if m.event == nil {
		return "No event to display"
	}

	// Convert nostr.Tags to [][]string
	tagsCopy := make([][]string, len(m.event.Tags))
	for i, tag := range m.event.Tags {
		tagsCopy[i] = make([]string, len(tag))
		copy(tagsCopy[i], tag)
	}

	data := struct {
		ID        string     `json:"id"`
		PubKey    string     `json:"pubkey"`
		CreatedAt int64      `json:"created_at"`
		Kind      int        `json:"kind"`
		Tags      [][]string `json:"tags"`
		Content   string     `json:"content"`
		Signature string     `json:"sig"`
	}{
		ID:        m.event.ID.Hex(),
		PubKey:    m.event.PubKey.Hex(),
		CreatedAt: int64(m.event.CreatedAt),
		Kind:      int(m.event.Kind),
		Tags:      tagsCopy,
		Content:   m.event.Content,
		Signature: fmt.Sprintf("%x", m.event.Sig),
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error rendering JSON: %v", err)
	}

	return string(jsonBytes)
}