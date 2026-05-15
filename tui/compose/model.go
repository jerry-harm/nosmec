package compose

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"github.com/jerry-harm/nosmec/tui/bubblon"
	"charm.land/lipgloss/v2"
	tea "charm.land/bubbletea/v2"
	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/tui/window"
)

const ComposeWindowID = "compose"

type ComposeKind int

const (
	KindNote ComposeKind = iota
	KindReply
	KindQuote
	KindCommunity
)

type Tag = []string

type model struct {
	styles styles
	width  int
	height int
	keys   *keyMap
	help   help.Model

	app           *config.AppContext
	composeKind   ComposeKind
	parentEvent   *nostr.Event
	parentID      string
	quotedID      string
	communityAddr string
	isStandalone  bool

	kindInput    textinput.Model
	contentInput textarea.Model
	tagInput     textinput.Model

	tags              []Tag
	editingTagIndex   int
	editingItemIndex  int

	errMsg    string
	success   bool
	sending   bool
	statusMsg string
}

type keyMap struct {
	send          key.Binding
	quit          key.Binding
	kill          key.Binding
	nextField     key.Binding
	prevField     key.Binding
	addTag        key.Binding
	removeTag     key.Binding
	deselectTag   key.Binding
}

func newKeyMap() *keyMap {
	return &keyMap{
		send: key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("ctrl+p", "send"),
		),
		quit: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "quit"),
		),
		kill: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "kill"),
		),
		nextField: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next field"),
		),
		prevField: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev field"),
		),
		addTag: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "add tag"),
		),
		removeTag: key.NewBinding(
			key.WithKeys("backspace", "delete"),
			key.WithHelp("backspace", "delete tag"),
		),
		deselectTag: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "deselect"),
		),
	}
}

func (k *keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.send, k.quit, k.nextField}
}

func (k *keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.send, k.quit},
		{k.nextField, k.prevField},
		{k.addTag},
	}
}

type sendErrorMsg struct {
	err string
}

type sendSuccessMsg struct {
	eventID string
}

// CloseComposeMsg is sent when the user presses esc or sends successfully
// to notify the window manager to close the compose window (but preserve state)
type CloseComposeMsg struct{}

func NewNoteCompose(app *config.AppContext) *model {
	return newCompose(app, KindNote, nil, "", "")
}

// SetStandalone marks this compose as running in standalone mode (not under wm).
// In standalone mode, esc sends tea.Quit instead of bubblon.Close().
func (m *model) SetStandalone() {
	m.isStandalone = true
}

func NewModel(app *config.AppContext) *model {
	return newCompose(app, KindNote, nil, "", "")
}

func NewCommunityCompose(app *config.AppContext, communityAddr string) *model {
	return newCompose(app, KindCommunity, nil, communityAddr, "")
}

func newCompose(app *config.AppContext, kind ComposeKind, parentEvent *nostr.Event, communityAddr, quotedID string) *model {
	m := &model{
		app:           app,
		composeKind:   kind,
		parentEvent:   parentEvent,
		communityAddr: communityAddr,
		quotedID:      quotedID,
	}
	m.styles = newStyles()
	m.keys = newKeyMap()
	m.help = help.New()

	m.kindInput = textinput.New()
	m.kindInput.Placeholder = "Kind (default: 1)"
	m.kindInput.Focus()

	m.contentInput = textarea.New()
	m.contentInput.Placeholder = "Write your note..."
	m.contentInput.Prompt = "| "

	m.tagInput = textinput.New()
	m.tagInput.Placeholder = "e:eventId p:pubkey a:addr t:hashtag r:relay:purpose q:eventId"
	m.tagInput.SetStyles(textinput.Styles{
		Focused: textinput.StyleState{
			Placeholder: lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")),
		},
		Blurred: textinput.StyleState{
			Placeholder: lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")),
		},
	})

	return m
}

func (m *model) AddReply(parentEvent *nostr.Event) {
	m.composeKind = KindReply
	m.parentEvent = parentEvent
	m.parentID = parentEvent.ID.Hex()
	m.tags = append(m.tags,
		Tag{"e", parentEvent.ID.Hex()},
		Tag{"p", parentEvent.PubKey.Hex()},
	)
}

func (m *model) AddQuote(parentEvent *nostr.Event) {
	m.composeKind = KindQuote
	m.parentEvent = parentEvent
	m.quotedID = parentEvent.ID.Hex()
	m.tags = append(m.tags,
		Tag{"q", parentEvent.ID.Hex()},
	)
}

// ClearDraft resets all compose state after successful send.
func (m *model) ClearDraft() {
	m.contentInput.SetValue("")
	m.kindInput.SetValue("")
	m.tags = nil
	m.parentEvent = nil
	m.parentID = ""
	m.quotedID = ""
	m.communityAddr = ""
	m.composeKind = KindNote
	m.errMsg = ""
	m.success = false
}

var _ tea.Model = (*model)(nil)

func (m *model) Init() tea.Cmd {
	m.kindInput.Focus()
	return tea.Batch(
		tea.RequestBackgroundColor,
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
		m.contentInput.SetWidth(msg.Width - 4)
		m.contentInput.SetHeight(10)
		m.tagInput.SetWidth(msg.Width - 4)
		return m, nil

	case sendErrorMsg:
		m.errMsg = msg.err
		m.statusMsg = "Error: " + msg.err
		return m, tea.Batch(
			tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
				m.sending = false
				m.statusMsg = ""
				return nil
			}),
		)

	case sendSuccessMsg:
		m.success = true
		m.errMsg = ""
		m.statusMsg = "Posted successfully!"
		m.sending = false
		m.ClearDraft()
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.sending {
			return m, nil
		}

		if key.Matches(msg, m.keys.kill) {
			os.Exit(0)
		}

		if key.Matches(msg, m.keys.quit) {
			if m.editingTagIndex >= 0 {
				m.saveTagEdit()
				m.editingTagIndex = -1
				m.tagInput.SetValue("")
				return m, nil
			}
			if m.isStandalone {
				return m, tea.Quit
			}
			// Send bubblon close instead of tea.Quit to preserve draft state
			return m, func() tea.Msg { return bubblon.Close() }
		}

		if m.kindInput.Focused() {
			if key.Matches(msg, m.keys.nextField) || key.Matches(msg, m.keys.addTag) {
				m.kindInput.Blur()
				m.tagInput.Focus()
				if len(m.tags) > 0 {
					m.editingTagIndex = len(m.tags) - 1
					m.editingItemIndex = len(m.tags[m.editingTagIndex])
					m.tagInput.SetValue("")
				} else {
					m.editingTagIndex = -1
					m.editingItemIndex = -1
					m.tagInput.SetValue("")
				}
				return m, nil
			}
			if key.Matches(msg, m.keys.prevField) {
				m.kindInput.Blur()
				m.tagInput.Focus()
				if len(m.tags) > 0 {
					m.editingTagIndex = len(m.tags) - 1
					m.editingItemIndex = len(m.tags[m.editingTagIndex])
					m.tagInput.SetValue("")
				} else {
					m.editingTagIndex = -1
					m.editingItemIndex = -1
					m.tagInput.SetValue("")
				}
				return m, nil
			}
		}

		if m.contentInput.Focused() {
			if msg.String() == "tab" || key.Matches(msg, m.keys.nextField) {
				m.contentInput.Blur()
				m.kindInput.Focus()
				return m, nil
			}
			if msg.String() == "shift+tab" || key.Matches(msg, m.keys.prevField) {
				m.contentInput.Blur()
				m.tagInput.Focus()
				if len(m.tags) > 0 {
					m.editingTagIndex = len(m.tags) - 1
					m.editingItemIndex = len(m.tags[m.editingTagIndex])
					m.tagInput.SetValue("")
				} else {
					m.editingTagIndex = -1
					m.editingItemIndex = -1
					m.tagInput.SetValue("")
				}
				return m, nil
			}
			if msg.String() == "ctrl+p" {
				content := m.contentInput.Value()
				if content = strings.TrimSpace(content); content != "" {
					m.sending = true
					return m, m.sendContent(content)
				}
			}
		}

		if m.tagInput.Focused() {
			tagValue := m.tagInput.Value()

			if key.Matches(msg, m.keys.addTag) {
				if tagValue != "" {
					if m.editingTagIndex < 0 {
						m.tags = append(m.tags, Tag{tagValue})
					} else {
						tag := m.tags[m.editingTagIndex]
						if m.editingItemIndex >= len(tag) {
							m.tags[m.editingTagIndex] = append(tag, tagValue)
						} else {
							newTag := make([]string, len(tag)+1)
							copy(newTag, tag[:m.editingItemIndex])
							newTag[m.editingItemIndex] = tagValue
							copy(newTag[m.editingItemIndex+1:], tag[m.editingItemIndex:])
							m.tags[m.editingTagIndex] = newTag
						}
					}
				}
				m.tagInput.SetValue("")
				return m, nil
			}

			if msg.String() == "backspace" && tagValue == "" {
				if m.editingTagIndex < 0 {
					if len(m.tags) > 0 {
						m.editingTagIndex = len(m.tags) - 1
						m.editingItemIndex = len(m.tags[m.editingTagIndex])
					}
					return m, nil
				}
				tag := m.tags[m.editingTagIndex]
				if m.editingItemIndex >= len(tag) {
					if len(tag) > 0 {
						m.tags[m.editingTagIndex] = tag[:len(tag)-1]
					}
					m.editingItemIndex = len(m.tags[m.editingTagIndex])
				} else if m.editingItemIndex > 0 {
					newTag := append(tag[:m.editingItemIndex-1], tag[m.editingItemIndex:]...)
					m.tags[m.editingTagIndex] = newTag
					m.editingItemIndex--
				} else {
					if len(tag) == 1 {
						m.tags = append(m.tags[:m.editingTagIndex], m.tags[m.editingTagIndex+1:]...)
						if len(m.tags) > 0 {
							m.editingTagIndex = len(m.tags) - 1
							m.editingItemIndex = len(m.tags[m.editingTagIndex])
						} else {
							m.editingTagIndex = -1
							m.editingItemIndex = -1
						}
					} else {
						m.tags[m.editingTagIndex] = tag[1:]
						m.editingItemIndex = 0
					}
				}
				return m, nil
			}

			if key.Matches(msg, m.keys.nextField) || msg.String() == "tab" {
				if m.editingTagIndex < 0 {
					m.tagInput.Blur()
					m.contentInput.Focus()
				} else {
					m.tagInput.SetValue("")
					m.editingTagIndex = -1
					m.editingItemIndex = -1
				}
				return m, nil
			}

			if key.Matches(msg, m.keys.prevField) || msg.String() == "shift+tab" {
				if m.editingTagIndex < 0 {
					if len(m.tags) > 0 {
						m.editingTagIndex = len(m.tags) - 1
						m.editingItemIndex = len(m.tags[m.editingTagIndex])
					}
				} else if m.editingItemIndex > 0 {
					m.editingItemIndex--
				} else {
					if m.editingTagIndex > 0 {
						m.editingTagIndex--
						m.editingItemIndex = len(m.tags[m.editingTagIndex])
					}
				}
				return m, nil
			}
		}
	}

	kindModel, kindCmd := m.kindInput.Update(msg)
	m.kindInput = kindModel
	cmds = append(cmds, kindCmd)

	contentModel, contentCmd := m.contentInput.Update(msg)
	m.contentInput = contentModel
	cmds = append(cmds, contentCmd)

	tagModel, tagCmd := m.tagInput.Update(msg)
	m.tagInput = tagModel
	cmds = append(cmds, tagCmd)

	return m, tea.Batch(cmds...)
}

func (m *model) nextField() {
}

func (m *model) prevField() {
}

func (m *model) saveTagEdit() {
	if m.editingTagIndex >= 0 && m.editingTagIndex < len(m.tags) && m.editingItemIndex >= 0 && m.editingItemIndex < len(m.tags[m.editingTagIndex]) {
		tagValue := strings.TrimSpace(m.tagInput.Value())
		if tagValue != "" {
			m.tags[m.editingTagIndex][m.editingItemIndex] = tagValue
		}
	}
}

func (m *model) sendContent(content string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		secretKey, err := m.app.GetMySecretKey()
		if err != nil {
			return sendErrorMsg{err: err.Error()}
		}

		var tags nostr.Tags
		for _, t := range m.tags {
			tag := nostr.Tag(t)
			tags = append(tags, tag)
		}

		kind := m.parseKind()

		event := &nostr.Event{
			Kind:      kind,
			CreatedAt: nostr.Timestamp(time.Now().Unix()),
			Tags:      tags,
			Content:   content,
			PubKey:    secretKey.Public(),
		}

		if err := event.Sign(secretKey); err != nil {
			return sendErrorMsg{err: err.Error()}
		}

		writableRelays := m.app.AllWritableRelays()
		if len(writableRelays) > 0 {
			resultChan := m.app.Pool().PublishMany(ctx, writableRelays, *event)
			for result := range resultChan {
				if result.Error != nil {
					return sendErrorMsg{err: fmt.Errorf("failed to publish to %s: %w", result.RelayURL, result.Error).Error()}
				}
			}
		}

		return sendSuccessMsg{eventID: event.ID.Hex()}
	}
}

func (m *model) parseKind() nostr.Kind {
	kindStr := strings.TrimSpace(m.kindInput.Value())
	if kindStr == "" {
		switch m.composeKind {
		case KindReply:
			if m.parentEvent != nil {
				return m.parentEvent.Kind
			}
			return nostr.KindComment
		case KindQuote:
			return nostr.KindTextNote
		case KindCommunity:
			return nostr.KindComment
		default:
			return nostr.KindTextNote
		}
	}
	var kind int
	_, err := fmt.Sscanf(kindStr, "%d", &kind)
	if err != nil {
		return nostr.KindTextNote
	}
	return nostr.Kind(kind)
}



func (m *model) renderSendingOverlay() string {
	var msg string
	if m.statusMsg != "" {
		msg = m.statusMsg
	} else {
		msg = "Sending..."
	}

	var b strings.Builder
	b.WriteString(m.styles.header.Render(m.renderHeader()))
	b.WriteString("\n\n")
	b.WriteString(m.styles.statusText.Render(msg))
	b.WriteString("\n\n")

	return b.String()
}

func (m *model) View() tea.View {
	content := m.renderView()
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m *model) renderView() string {
	if m.sending {
		return m.renderSendingOverlay()
	}

	var b strings.Builder

	b.WriteString(m.styles.header.Render(m.renderHeader()))
	b.WriteString("\n\n")

	if m.errMsg != "" {
		b.WriteString(m.styles.errorMsg.Render("Error: " + m.errMsg))
		b.WriteString("\n\n")
	}

	if m.success {
		b.WriteString(m.styles.successMsg.Render("Posted successfully!"))
		b.WriteString("\n\n")
	}

	b.WriteString(m.styles.fieldLabel.Render("Kind: "))
	b.WriteString(m.kindInput.View())
	b.WriteString(" (default: 1)\n\n")

	b.WriteString(m.styles.fieldLabel.Render("Tags:"))
	b.WriteString("\n")
	for i, tag := range m.tags {
		if i == m.editingTagIndex && m.tagInput.Focused() {
			if m.editingItemIndex < len(tag) {
				for j := 0; j < len(tag); j++ {
					if j == m.editingItemIndex {
						if j > 0 {
							b.WriteString(fmt.Sprintf("  [%s] %s >%s\n", tag[0], strings.Join(tag[1:j], ", "), m.tagInput.View()))
						} else {
							b.WriteString(fmt.Sprintf("  [%s] >%s\n", tag[0], m.tagInput.View()))
						}
						break
					}
				}
			}
			if m.editingItemIndex >= len(tag) {
				if len(tag) > 0 {
					b.WriteString(fmt.Sprintf("  [%s] %s\n", tag[0], strings.Join(tag[1:], ", ")))
				}
				b.WriteString("  > ")
				b.WriteString(m.tagInput.View())
				b.WriteString("\n")
			}
		} else {
			if len(tag) > 0 {
				b.WriteString(fmt.Sprintf("  [%s] %s\n", tag[0], strings.Join(tag[1:], ", ")))
			}
		}
	}
	if m.tagInput.Focused() && m.editingTagIndex < 0 {
		b.WriteString("  > ")
		b.WriteString(m.tagInput.View())
		b.WriteString("\n")
	} else if !m.tagInput.Focused() {
		b.WriteString("  >\n")
	}
	b.WriteString(m.styles.fieldLabel.Render("Content:"))
	b.WriteString("\n")
	b.WriteString(m.styles.inputArea.Render(m.contentInput.View()))
	b.WriteString("\n\n")

	b.WriteString(m.help.View(m.keys))

	return b.String()
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

func PrepareReply(w window.Window, event *nostr.Event) {
	if wm, ok := w.(*model); ok {
		wm.AddReply(event)
	}
}

func PrepareQuote(w window.Window, event *nostr.Event) {
	if wm, ok := w.(*model); ok {
		wm.AddQuote(event)
	}
}