package event

import (
	"context"
	"fmt"
	"strings"

	"github.com/jerry-harm/nosmec/utils"
)

func (m *EventView) renderHeader() string {
	e := m.event

	author := e.PubKey.Hex()[:8]
	if profileName := utils.GetProfileName(context.Background(), e.PubKey, &utils.GetOptions{App: m.app}); profileName != "" {
		author = profileName
	}

	timeStr := e.CreatedAt.Time().Format("2006-01-02 15:04")
	kindStr := fmt.Sprintf("Kind: %d", e.Kind)

	return fmt.Sprintf("@%s | %s | %s", m.styles.author.Render(author), m.styles.time.Render(timeStr), kindStr)
}

func (m *EventView) renderContent() string {
	e := m.event

	var tagParts []string
	for _, tag := range e.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "t":
				tagParts = append(tagParts, m.styles.tags.Render("#"+tag[1]))
			case "p":
				tagParts = append(tagParts, m.styles.tags.Render("@"+tag[1][:8]))
			case "e":
				tagParts = append(tagParts, m.styles.tags.Render("→"+tag[1][:8]))
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

	out += fmt.Sprintf("\nID: %s", m.styles.tags.Render(e.ID.Hex()[:16]+"..."))

	return out
}