package event

import (
	"encoding/json"
	"fmt"
	"strings"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"github.com/jerry-harm/nosmec/tui/component/label"
	"github.com/jerry-harm/nosmec/tui/theme"
)

func (m *EventView) renderHeader() string {
	if m.event == nil {
		return m.styles.header.Render("Loading...")
	}
	e := m.event

	timeStr := e.CreatedAt.Time().Format("2006-01-02 15:04")
	kindStr := fmt.Sprintf("Kind: %d", e.Kind)

	npub := nip19.EncodeNpub(e.PubKey)

	// Line 1: full pubkey
	line1 := fmt.Sprintf("PubKey: %s", m.styles.author.Render(npub))

	// Line 2: @username | time | kind
	var namePart string
	pubkeyHex := e.PubKey.Hex()
	if m.authorName != "" {
		namePart = label.RenderLabel(pubkeyHex, m.authorName, label.StateResolved, theme.Default())
	} else {
		namePart = label.RenderLabel(pubkeyHex, "", label.StateLoading, theme.Default())
	}
	line2 := fmt.Sprintf("%s | %s | %s",
		namePart,
		m.styles.time.Render(timeStr),
		kindStr)

	lines := line1 + "\n" + line2

	// Line 3 (kind 34550 only): Community Address
	if e.Kind == nostr.KindCommunityDefinition {
		addr := fmt.Sprintf("34550:%s:%s", e.PubKey.Hex(), e.Tags.GetD())
		lines += "\n" + fmt.Sprintf("Community Address: %s", m.styles.communityAddr.Render(addr))
	}

	return m.styles.header.Render(lines)
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

	var tagLines []string
	for i, tag := range e.Tags {
		if len(tag) >= 1 {
			tagStr := strings.Join(tag, " ")
			tagLines = append(tagLines, fmt.Sprintf("[%d] %s", i, m.styles.tags.Render(tagStr)))
		}
	}

	content := e.Content

	var out string
	out += content
	out += "\n"

	if len(tagLines) > 0 {
		out += "\n--- Tags ---\n"
		for _, tagLine := range tagLines {
			out += tagLine + "\n"
		}
	}

	out += "\n--- Signature ---\n"
	nevent := nip19.EncodeNevent(e.ID, nil, e.PubKey)
	out += fmt.Sprintf("ID: %s\n", m.styles.tags.Render(nevent))
	out += fmt.Sprintf("Sig: %x\n", e.Sig)

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
		NPubKey   string     `json:"npub,omitempty"`
		CreatedAt int64      `json:"created_at"`
		Kind      int        `json:"kind"`
		Tags      [][]string `json:"tags"`
		Content   string     `json:"content"`
		Signature string     `json:"sig"`
		NEvent    string     `json:"nevent,omitempty"`
	}{
		ID:        m.event.ID.Hex(),
		PubKey:    m.event.PubKey.Hex(),
		NPubKey:   nip19.EncodeNpub(m.event.PubKey),
		CreatedAt: int64(m.event.CreatedAt),
		Kind:      int(m.event.Kind),
		Tags:      tagsCopy,
		Content:   m.event.Content,
		Signature: fmt.Sprintf("%x", m.event.Sig),
		NEvent:    nip19.EncodeNevent(m.event.ID, nil, m.event.PubKey),
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error rendering JSON: %v", err)
	}

	return string(jsonBytes)
}