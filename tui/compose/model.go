package compose

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/utils"
)

const (
	ComposeWindowID = "compose"
)

type ComposeKind int

const (
	KindNote ComposeKind = iota
	KindReply
	KindQuote
	KindCommunity
)

type model struct {
	styles  styles
	width   int
	height  int
	viewport viewport.Model
	ta      textarea.Model
	keys    *keyMap

	app           *config.AppContext
	composeKind   ComposeKind
	parentEvent   *nostr.Event
	parentID      string
	quotedID      string
	communityAddr string
	errMsg        string
	success       bool
}

type keyMap struct {
	send   key.Binding
	quit   key.Binding
	scroll key.Binding
}

func newKeyMap() *keyMap {
	return &keyMap{
		send: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "send"),
		),
		quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c", "esc"),
			key.WithHelp("q", "quit"),
		),
		scroll: key.NewBinding(
			key.WithKeys("pgup", "pgdown"),
			key.WithHelp("pgup/pgdown", "scroll"),
		),
	}
}

type sendMsg struct {
	content string
}

type sendErrorMsg struct {
	err string
}

type sendSuccessMsg struct {
	eventID string
}

func NewNoteCompose(app *config.AppContext) *model {
	return newCompose(app, KindNote, nil)
}

func NewReplyCompose(app *config.AppContext, parentEvent *nostr.Event) *model {
	m := newCompose(app, KindReply, parentEvent)
	m.parentID = parentEvent.ID.Hex()
	return m
}

func NewQuoteCompose(app *config.AppContext, parentEvent *nostr.Event) *model {
	m := newCompose(app, KindQuote, parentEvent)
	m.quotedID = parentEvent.ID.Hex()
	return m
}

func NewCommunityCompose(app *config.AppContext, communityAddr string) *model {
	m := newCompose(app, KindCommunity, nil)
	m.communityAddr = communityAddr
	return m
}

func newCompose(app *config.AppContext, kind ComposeKind, parentEvent *nostr.Event) *model {
	m := &model{
		app:         app,
		composeKind: kind,
		parentEvent: parentEvent,
	}
	m.styles = newStyles()
	m.keys = newKeyMap()
	m.viewport = viewport.New()
	m.ta = textarea.New()
	m.ta.Placeholder = "Write your note..."
	m.ta.Focus()

	m.updatePlaceholder()

	return m
}

func (m *model) updatePlaceholder() {
	switch m.composeKind {
	case KindReply:
		if m.parentEvent != nil {
			m.ta.Placeholder = fmt.Sprintf("Replying to %s...", nip19.EncodeNpub(m.parentEvent.PubKey)[:16]+"...")
		}
	case KindQuote:
		if m.parentEvent != nil {
			content := m.parentEvent.Content
			if len(content) > 30 {
				content = content[:30] + "..."
			}
			m.ta.Placeholder = fmt.Sprintf("Quoting: %s", content)
		}
	case KindCommunity:
		m.ta.Placeholder = fmt.Sprintf("Posting to %s...", m.communityAddr)
	default:
		m.ta.Placeholder = "Write your note..."
	}
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		tea.RequestBackgroundColor,
		func() tea.Msg {
			m.viewport.SetWidth(80)
			m.viewport.SetHeight(20)
			return tea.WindowSizeMsg{Width: 80, Height: 30}
		},
	)
}

func (m *model) ID() string {
	return ComposeWindowID
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.styles = newStyles()
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerHeight := 3
		inputHeight := 6
		viewportHeight := msg.Height - headerHeight - inputHeight - 2
		m.viewport = viewport.New()
		m.viewport.SetWidth(msg.Width - 4)
		m.viewport.SetHeight(viewportHeight)
		m.viewport.SetContent(m.renderHeader())
		m.ta.SetWidth(msg.Width - 4)
		return m, nil

	case sendErrorMsg:
		m.errMsg = msg.err
		return m, nil

	case sendSuccessMsg:
		m.success = true
		m.errMsg = ""
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.ta.Focused() {
			switch {
			case key.Matches(msg, m.keys.send):
				content := m.ta.Value()
				if content = strings.TrimSpace(content); content != "" {
					cmd := m.sendContent(content)
					return m, cmd
				}
			case key.Matches(msg, m.keys.quit):
				return m, tea.Quit
			}
		}

		switch {
		case key.Matches(msg, m.keys.scroll):
			if key.Matches(msg, key.NewBinding(key.WithKeys("pgup"))) {
				m.viewport.ScrollUp(10)
			} else {
				m.viewport.ScrollDown(10)
			}
		}
	}

	taModel, cmd := m.ta.Update(msg)
	m.ta = taModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) sendContent(content string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), m.app.QueryTimeout())
		defer cancel()

		var err error
		var event *nostr.Event

		switch m.composeKind {
		case KindNote:
			event, err = utils.PostNote(ctx, m.app, content)
		case KindReply:
			if m.parentID == "" && m.parentEvent != nil {
				m.parentID = m.parentEvent.ID.Hex()
			}
			event, err = utils.ReplyToNote(ctx, m.app, m.parentID, content)
		case KindQuote:
			if m.quotedID == "" && m.parentEvent != nil {
				m.quotedID = m.parentEvent.ID.Hex()
			}
			event, err = utils.QuoteNote(ctx, m.app, m.quotedID, content)
		case KindCommunity:
			event, err = m.postCommunity(ctx, content)
		}

		if err != nil {
			return sendErrorMsg{err: err.Error()}
		}

		return sendSuccessMsg{eventID: event.ID.Hex()}
	}
}

func (m *model) postCommunity(ctx context.Context, content string) (*nostr.Event, error) {
	secretKey, err := m.app.GetMySecretKey()
	if err != nil {
		return nil, err
	}

	tags := nostr.Tags{
		{"a", m.communityAddr},
	}

	event := &nostr.Event{
		Kind:      nostr.KindCommunityDefinition,
		CreatedAt: nostr.Timestamp(time.Now().Unix()),
		Tags:      tags,
		Content:   content,
		PubKey:    secretKey.Public(),
	}

	if err := event.Sign(secretKey); err != nil {
		return nil, err
	}

	writableRelays := m.app.AllWritableRelays()
	if len(writableRelays) > 0 {
		resultChan := m.app.Pool().PublishMany(ctx, writableRelays, *event)
		for result := range resultChan {
			if result.Error != nil {
				return nil, fmt.Errorf("failed to publish to %s: %w", result.RelayURL, result.Error)
			}
		}
	}

	return event, nil
}

func (m *model) View() tea.View {
	var b strings.Builder

	b.WriteString(m.styles.header.Render(m.renderHeader()))
	b.WriteString("\n")

	if m.errMsg != "" {
		b.WriteString(m.styles.errorMsg.Render("Error: " + m.errMsg))
		b.WriteString("\n")
	}

	if m.success {
		b.WriteString(m.styles.header.Render("Posted successfully!"))
		b.WriteString("\n")
	}

	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	b.WriteString(m.styles.inputArea.Render(m.ta.View()))
	b.WriteString("\n")

	b.WriteString(m.styles.help.Render("Enter: send | q/ctrl+c: quit | pgup/pgdown: scroll"))

	return tea.NewView(b.String())
}

func (m *model) renderHeader() string {
	switch m.composeKind {
	case KindReply:
		return "Reply"
	case KindQuote:
		return "Quote"
	case KindCommunity:
		return fmt.Sprintf("Community: %s", m.communityAddr)
	default:
		return "New Note"
	}
}

func (m *model) Close() bool {
	return true
}