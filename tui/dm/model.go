package dm

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/keyer"
	"fiatjaf.com/nostr/nip19"
	"fiatjaf.com/nostr/nip59"
	sdk "github.com/jerry-harm/nosmec/nostr_sdk"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/utils"
)

const (
	chatPanelName = "dm-thread"
)

type message struct {
	content   string
	fromMe    bool
	timestamp time.Time
	npub      string
}

type model struct {
	styles  styles
	width   int
	height  int
	viewport viewport.Model
	ta      textinput.Model
	keys    *keyMap

	app             *config.AppContext
	recipientPubKey  nostr.PubKey
	recipientNpub    string
	recipientName    string
	messages         []message
	errMsg           string

	subCh     chan nostr.Event
	subCancel context.CancelFunc
}

type keyMap struct {
	send  key.Binding
	quit  key.Binding
	kill  key.Binding
	scroll key.Binding
}

func newKeyMap() *keyMap {
	return &keyMap{
		send: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "send"),
		),
		quit: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "quit"),
		),
		kill: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "kill"),
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

type newMessageMsg struct {
	content   string
	fromMe    bool
	timestamp time.Time
	npub      string
}

type sendErrorMsg struct {
	err string
}

type profileNameMsg struct {
	name string
}

func NewModel(app *config.AppContext, recipientPubKey nostr.PubKey) *model {
	m := &model{
		app:            app,
		recipientPubKey: recipientPubKey,
		recipientNpub:  nip19.EncodeNpub(recipientPubKey),
	}
	m.styles = newStyles(app.Theme())
	m.keys = newKeyMap()
	m.viewport = viewport.New()
	m.viewport.SetYOffset(0)
	m.ta = textinput.New()
	m.ta.Placeholder = "Type a message..."
	m.ta.Focus()

	return m
}

func (m *model) Init() tea.Cmd {
	m.viewport = viewport.New()
	m.viewport.SetWidth(80)
	m.viewport.SetHeight(20)
	return tea.Batch(
		tea.RequestBackgroundColor,
		m.startSubscription(),
		m.fetchRecipientProfileNameAsync(),
	)
}

func (m *model) fetchRecipientProfileNameAsync() tea.Cmd {
	return func() tea.Msg {
		pm := m.app.System().FetchProfileMetadata(context.Background(), m.recipientPubKey)
		name := ""
		if pm.Event != nil {
			if meta, err := sdk.ParseMetadata(*pm.Event); err == nil && meta.Name != "" {
				name = meta.Name
			}
		}
		return profileNameMsg{name: name}
	}
}

func (m *model) handleMessage(msg newMessageMsg) tea.Cmd {
	m.messages = append(m.messages, message{
		content:   msg.content,
		fromMe:    msg.fromMe,
		timestamp: msg.timestamp,
		npub:      msg.npub,
	})
	m.viewport.SetContent(m.renderMessages())
	return nil
}

func (m *model) startSubscription() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		m.subCancel = cancel

		ourPubKey, _ := m.app.GetMyPubKey()
		relays := m.app.ListDMRelays()
		if len(relays) == 0 {
			relays = m.app.ReadableRelays()
		}
		if len(relays) == 0 {
			relays = m.app.AllReadableRelays()
		}

		filter := nostr.Filter{
			Kinds: []nostr.Kind{nostr.KindGiftWrap},
			Tags:  nostr.TagMap{"p": []string{ourPubKey.Hex(), m.recipientPubKey.Hex()}},
			Limit: 300,
		}

		subCh := m.app.Pool().SubscribeMany(ctx, relays, filter, nostr.SubscriptionOptions{Label: "dm-tui"})
		m.subCh = make(chan nostr.Event, 100)

		go func() {
			for relayEvent := range subCh {
				m.subCh <- relayEvent.Event
			}
			close(m.subCh)
		}()

		return nil
	}
}

func (m *model) pollSubscription() tea.Cmd {
	return func() tea.Msg {
		if m.subCh == nil {
			return tea.Tick(time.Millisecond*500, func(time.Time) tea.Msg {
				return pollMsg{}
			})
		}

		select {
		case event, ok := <-m.subCh:
			if !ok {
				m.subCh = nil
				return tea.Tick(time.Millisecond*500, func(time.Time) tea.Msg {
					return pollMsg{}
				})
			}

			secretKey, err := m.app.GetMySecretKey()
			if err != nil {
				return nil
			}

			kr := keyer.NewPlainKeySigner(secretKey)
			ourPubKey, _ := kr.GetPublicKey(context.Background())

			rumor, err := nip59.GiftUnwrap(
				event,
				func(otherpubkey nostr.PubKey, ciphertext string) (string, error) {
					return kr.Decrypt(context.Background(), ciphertext, otherpubkey)
				},
			)
			if err != nil {
				return nil
			}

			fromMe := rumor.PubKey == ourPubKey
			var npubStr string
			if fromMe {
				npubStr = nip19.EncodeNpub(ourPubKey)[:16] + "..."
			} else {
				npubStr = m.recipientNpub[:16] + "..."
			}

			return newMessageMsg{
				content:   rumor.Content,
				fromMe:    fromMe,
				timestamp: rumor.CreatedAt.Time(),
				npub:      npubStr,
			}
		default:
			return tea.Tick(time.Millisecond*500, func(time.Time) tea.Msg {
				return pollMsg{}
			})
		}
	}
}

type pollMsg struct{}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
m.styles = newStyles(m.app.Theme())
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
		m.viewport.SetContent(m.renderMessages())
		m.ta.SetWidth(msg.Width - 4)
		return m, nil

	case newMessageMsg:
		m.messages = append(m.messages, message{
			content:   msg.content,
			fromMe:    msg.fromMe,
			timestamp: msg.timestamp,
			npub:      msg.npub,
		})
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, m.pollSubscription()

	case sendErrorMsg:
		m.errMsg = msg.err
		return m, nil

	case profileNameMsg:
		m.recipientName = msg.name
		return m, nil

	case sendMsg:
		content := strings.TrimSpace(msg.content)
		if content == "" {
			return m, nil
		}
		m.errMsg = ""
		return m, m.sendDM(content)

	case pollMsg:
		if m.subCh != nil || m.subCancel != nil {
			return m, m.pollSubscription()
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, m.keys.kill) {
			os.Exit(0)
		}

		if m.ta.Focused() {
			switch {
			case key.Matches(msg, m.keys.send):
				content := m.ta.Value()
				m.ta.SetValue("")
				if content = strings.TrimSpace(content); content != "" {
					cmds = append(cmds, m.sendDM(content))
				}
			case key.Matches(msg, m.keys.quit):
				if m.subCancel != nil {
					m.subCancel()
				}
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

func (m *model) sendDM(content string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), m.app.QueryTimeout())
		defer cancel()

		err := utils.SendDM(ctx, m.app, m.recipientPubKey, content)
		if err != nil {
			return sendErrorMsg{err: err.Error()}
		}

		return nil
	}
}

func (m *model) View() tea.View {
	var b strings.Builder

	headerTitle := "DM: " + m.recipientNpub[:32] + "..."
	if m.recipientName != "" {
		headerTitle = "DM: " + m.recipientName
	}
	b.WriteString(m.styles.header.Render(headerTitle))
	b.WriteString("\n")

	if m.errMsg != "" {
		b.WriteString(m.styles.errorMsg.Render("Error: "+m.errMsg))
		b.WriteString("\n")
	}

	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	b.WriteString(m.styles.inputArea.Render(m.ta.View()))
	b.WriteString("\n")

	b.WriteString(m.styles.help.Render("Enter: send | q/ctrl+c: quit | pgup/pgdown: scroll"))

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

func (m *model) renderMessages() string {
	if len(m.messages) == 0 {
		return "No messages yet. Send the first message!"
	}

	var b strings.Builder
	for _, msg := range m.messages {
		npubStyle := m.styles.theirs
		if msg.fromMe {
			npubStyle = m.styles.mine
		}

		timestamp := msg.timestamp.Format("2006-01-02 15:04")
		b.WriteString(fmt.Sprintf("[%s] %s: %s\n",
			m.styles.timestamp.Render(timestamp),
			npubStyle.Render(msg.npub),
			m.styles.theirs.Render(msg.content),
		))
	}
	return b.String()
}

func (m *model) ID() string {
	return chatPanelName
}