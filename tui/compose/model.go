package compose

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"github.com/jerry-harm/nosmec/tui/component/bubblon"
	"charm.land/lipgloss/v2"
	tea "charm.land/bubbletea/v2"
	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip10"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/utils"
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
	spinner      spinner.Model

	tags         []Tag
	editingIndex int // -2=not in tag mode, -1=empty slot, >=0=editing tags[editingIndex]

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
m.styles = newStyles(m.app.Theme())
	m.keys = newKeyMap()
	m.help = help.New()

	m.kindInput = textinput.New()
	m.kindInput.Placeholder = "Kind (default: 1)"
	m.kindInput.Focus()

	m.contentInput = textarea.New()
	m.contentInput.Placeholder = "Write your note..."
	m.contentInput.Prompt = "| "

	m.tagInput = textinput.New()
	m.tagInput.Placeholder = `["tag1","tag2"]`
	m.tagInput.SetStyles(textinput.Styles{
		Focused: textinput.StyleState{
			Placeholder: lipgloss.NewStyle().Foreground(m.styles.t.InputPlaceholder),
		},
		Blurred: textinput.StyleState{
			Placeholder: lipgloss.NewStyle().Foreground(m.styles.t.InputPlaceholder),
		},
	})

	m.spinner = spinner.New(spinner.WithSpinner(spinner.Dot))
	m.spinner.Style = lipgloss.NewStyle().Foreground(m.styles.t.Spinner)

	return m
}

// AddReplyFromTarget sets up the compose model for replying to an event using
// the pre-computed ReplyTarget from utils.DetermineReplyTarget.
func (m *model) AddReplyFromTarget(ctx context.Context, app *config.AppContext, parentEvent *nostr.Event, target utils.ReplyTarget) {
	m.composeKind = KindReply
	m.parentEvent = parentEvent
	m.parentID = parentEvent.ID.Hex()

	for _, t := range target.RootTags {
		m.tags = append(m.tags, Tag(t))
	}
	for _, t := range target.ParentTags {
		m.tags = append(m.tags, Tag(t))
	}
	m.kindInput.SetValue(strconv.Itoa(int(target.ReplyKind)))
}

func (m *model) AddReply(ctx context.Context, app *config.AppContext, parentEvent *nostr.Event) {
	ptr := nip10.GetThreadRoot(parentEvent.Tags)
	rootID := parentEvent.ID
	isRoot := true
	if ptr != nil {
		if ep, ok := ptr.(nostr.EventPointer); ok {
			rootID = ep.ID
			isRoot = false
		}
	}
	var rootPubKey string
	if !isRoot && rootID != parentEvent.ID {
		ext := app.System()
		if rootEvent := ext.FetchNote(ctx, rootID.Hex(), app.QueryTimeoutms()); rootEvent != nil {
			rootPubKey = rootEvent.PubKey.Hex()
		}
	}
	m.AddReplyWithRoot(parentEvent, rootPubKey)
	m.kindInput.SetValue(strconv.Itoa(int(parentEvent.Kind)))
}

func (m *model) AddReplyWithRoot(parentEvent *nostr.Event, rootPubKey string) {
	m.composeKind = KindReply
	m.parentEvent = parentEvent
	m.parentID = parentEvent.ID.Hex()

	tags := utils.BuildReplyTagsWithRoot(m.app, parentEvent, rootPubKey)
	for _, t := range tags {
		m.tags = append(m.tags, Tag(t))
	}
	m.tags = append(m.tags, Tag{"p", parentEvent.PubKey.Hex()})
}

func (m *model) AddQuoteFromTarget(parentEvent *nostr.Event, target utils.ReplyTarget) {
	m.composeKind = KindQuote
	m.parentEvent = parentEvent
	m.quotedID = parentEvent.ID.Hex()
	relay := m.app.GetEventRelay(parentEvent.ID.Hex())
	m.tags = append(m.tags,
		Tag{"q", parentEvent.ID.Hex(), relay, parentEvent.PubKey.Hex()},
	)
	// For non-kind:1 events, also add the root tags for context
	if target.ReplyKind != nostr.KindTextNote {
		for _, t := range target.RootTags {
			m.tags = append(m.tags, Tag(t))
		}
	}
}

func (m *model) AddQuote(parentEvent *nostr.Event) {
	m.composeKind = KindQuote
	m.parentEvent = parentEvent
	m.quotedID = parentEvent.ID.Hex()
	relay := m.app.GetEventRelay(parentEvent.ID.Hex())
	m.tags = append(m.tags,
		Tag{"q", parentEvent.ID.Hex(), relay, parentEvent.PubKey.Hex()},
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
		m.styles = newStyles(m.app.Theme())
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.contentInput.SetWidth(msg.Width - 4)
		m.tagInput.SetWidth(msg.Width - 4)
		return m, nil

	case spinner.TickMsg:
		newSpinner, cmd := m.spinner.Update(msg)
		m.spinner = newSpinner
		cmds = append(cmds, cmd)

	case sendErrorMsg:
		m.errMsg = msg.err
		m.statusMsg = "Failed: " + msg.err
		m.sending = false
		// Stay on compose page so user can retry — don't close
		return m, nil

	case sendSuccessMsg:
		m.success = true
		m.errMsg = ""
		m.statusMsg = "Posted successfully!"
		m.sending = false
		m.ClearDraft()
		if m.isStandalone {
			return m, tea.Quit
		}
		return m, func() tea.Msg { return bubblon.Close() }
	}

	if m.sending {
		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, m.keys.kill) {
			os.Exit(0)
		}

		if key.Matches(msg, m.keys.quit) {
			if m.editingIndex >= 0 {
				m.saveTagEdit()
				m.editingIndex = -2
				m.tagInput.SetValue("")
			}
			if m.isStandalone {
				return m, tea.Quit
			}
			return m, func() tea.Msg { return bubblon.Close() }
		}

		if m.kindInput.Focused() {
			if key.Matches(msg, m.keys.nextField) || key.Matches(msg, m.keys.addTag) {
				m.kindInput.Blur()
				m.tagInput.Focus()
				if len(m.tags) > 0 {
					m.editingIndex = 0
					m.tagInput.SetValue(tagToListString(m.tags[m.editingIndex]))
				} else {
					m.editingIndex = -1
					m.tagInput.SetValue("")
				}
				return m, nil
			}
			if key.Matches(msg, m.keys.prevField) {
				m.kindInput.Blur()
				m.contentInput.Focus()
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
					m.editingIndex = len(m.tags) - 1
					m.tagInput.SetValue(tagToListString(m.tags[m.editingIndex]))
				} else {
					m.editingIndex = -1
					m.tagInput.SetValue("")
				}
				return m, nil
			}
			if msg.String() == "ctrl+p" {
				content := m.contentInput.Value()
				if content = strings.TrimSpace(content); content != "" {
					m.sending = true
					spinnerTick := func() tea.Msg { return m.spinner.Tick() }
					return m, tea.Batch(
						spinnerTick,
						m.sendContent(content),
					)
				}
			}
		}

		if m.tagInput.Focused() {
			tagValue := m.tagInput.Value()

			if key.Matches(msg, m.keys.addTag) {
				if tagValue != "" {
					if tag, err := parseTagListInput(tagValue); err == nil {
						if m.editingIndex < 0 {
							m.tags = append(m.tags, tag)
						} else {
							m.tags[m.editingIndex] = tag
						}
					}
					m.tagInput.SetValue("")
					m.editingIndex = -1
				} else {
					// Empty input: blur to contentInput
					m.tagInput.Blur()
					m.contentInput.Focus()
					m.editingIndex = -2
				}
				return m, nil
			}

			if msg.String() == "backspace" && tagValue == "" {
				if m.editingIndex < 0 {
					if len(m.tags) > 0 {
						m.editingIndex = len(m.tags) - 1
						m.tagInput.SetValue(tagToListString(m.tags[m.editingIndex]))
					}
				} else {
					m.tags = append(m.tags[:m.editingIndex], m.tags[m.editingIndex+1:]...)
					if len(m.tags) == 0 {
						m.editingIndex = -1
					} else {
						m.editingIndex = len(m.tags) - 1
					}
				}
				return m, nil
			}

			if key.Matches(msg, m.keys.nextField) || msg.String() == "tab" {
				if m.editingIndex < 0 {
					// editingIndex == -1: empty slot → go to contentInput
					m.tagInput.Blur()
					m.contentInput.Focus()
					m.editingIndex = -2
				} else if m.editingIndex < len(m.tags)-1 {
					// editingIndex points to a tag, advance to next
					m.editingIndex++
					m.tagInput.SetValue(tagToListString(m.tags[m.editingIndex]))
				} else {
					// editingIndex is last tag → go to empty slot (-1), stay in tagInput
					m.tagInput.SetValue("")
					m.editingIndex = -1
				}
				return m, nil
			}

			if key.Matches(msg, m.keys.prevField) || msg.String() == "shift+tab" {
				if m.editingIndex < 0 {
					if len(m.tags) > 0 {
						m.editingIndex = len(m.tags) - 1
						m.tagInput.SetValue(tagToListString(m.tags[m.editingIndex]))
					} else {
						// empty slot with no tags: go to kindInput
						m.tagInput.Blur()
						m.kindInput.Focus()
						m.editingIndex = -2
					}
				} else if m.editingIndex > 0 {
					m.editingIndex--
					m.tagInput.SetValue(tagToListString(m.tags[m.editingIndex]))
				} else {
					// editingIndex == 0: go to kindInput
					m.tagInput.Blur()
					m.kindInput.Focus()
					m.editingIndex = -2
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
	if m.editingIndex >= 0 && m.editingIndex < len(m.tags) {
		tagValue := strings.TrimSpace(m.tagInput.Value())
		if tagValue != "" {
			m.tags[m.editingIndex] = Tag{tagValue}
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

		var failedRelays []string
		var hasSuccess bool

		writableRelays := m.app.AllWritableRelays()
		if len(writableRelays) > 0 {
			resultChan := m.app.Pool().PublishMany(ctx, writableRelays, *event)
			for result := range resultChan {
				if result.Error != nil {
					failedRelays = append(failedRelays, result.RelayURL)
				} else {
					hasSuccess = true
				}
			}
		} else {
			return sendErrorMsg{err: "no writable relays configured"}
		}

		if hasSuccess {
			return sendSuccessMsg{eventID: event.ID.Hex()}
		}

		return sendErrorMsg{err: strings.Join(failedRelays, ", ")}
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
	var b strings.Builder
	b.WriteString(m.styles.header.Render(m.renderHeader()))
	b.WriteString("\n\n")
	b.WriteString(m.styles.statusText.Render("Sending..."))
	b.WriteString(" ")
	b.WriteString(m.spinner.View())
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
		if i == m.editingIndex && m.tagInput.Focused() {
			b.WriteString("  > ")
			b.WriteString(m.tagInput.View())
			b.WriteString("\n")
		} else {
			if len(tag) > 0 {
				b.WriteString(fmt.Sprintf("  [%s] %s\n", tag[0], strings.Join(tag[1:], ", ")))
			}
		}
	}
	if m.tagInput.Focused() && m.editingIndex == -1 {
		b.WriteString("  > ")
		b.WriteString(m.tagInput.View())
		b.WriteString("\n")
	} else if !m.tagInput.Focused() {
		b.WriteString("  >\n")
	}
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

func tagToListString(tag Tag) string {
	if len(tag) == 0 {
		return "[]"
	}
	b, _ := json.Marshal(tag)
	return string(b)
}

func parseTagListInput(s string) (Tag, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty input")
	}
	var tag Tag
	if err := json.Unmarshal([]byte(s), &tag); err != nil {
		return nil, err
	}
	return tag, nil
}