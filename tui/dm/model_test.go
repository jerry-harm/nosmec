package dm

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func TestHandleMessage_EmptyMessages(t *testing.T) {
	m := &model{
		messages: []message{},
	}

	msg := newMessageMsg{
		content:   "hello",
		fromMe:    false,
		timestamp: time.Now(),
		npub:      "npub1...",
	}

	// handleMessage appends to messages and returns nil
	cmd := m.handleMessage(msg)
	if cmd != nil {
		t.Errorf("handleMessage should return nil command, got %T", cmd)
	}

	if len(m.messages) != 1 {
		t.Errorf("len(m.messages) = %d, want 1", len(m.messages))
	}
}

func TestHandleMessage_MultipleMessages(t *testing.T) {
	m := &model{
		messages: []message{},
	}

	for i := 0; i < 5; i++ {
		msg := newMessageMsg{
			content:   "hello",
			fromMe:    i%2 == 0,
			timestamp: time.Now(),
			npub:      "npub1...",
		}
		m.handleMessage(msg)
	}

	if len(m.messages) != 5 {
		t.Errorf("len(m.messages) = %d, want 5", len(m.messages))
	}
}

func TestRenderMessages_EmptyList(t *testing.T) {
	m := &model{
		messages: []message{},
	}

	result := m.renderMessages()
	if result == "" {
		t.Errorf("expected non-empty render output for empty messages")
	}
}

func TestUpdate_WindowSizeWithEmptyMessages(t *testing.T) {
	m := &model{
		messages: []message{},
	}

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	_, _ = m.Update(msg)

	if len(m.messages) != 0 {
		t.Errorf("len(m.messages) = %d, want 0", len(m.messages))
	}
}