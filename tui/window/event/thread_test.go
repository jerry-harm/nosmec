package event

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"fiatjaf.com/nostr"
)

func TestThreadView_EmptyReplies(t *testing.T) {
	m := &threadView{
		replies: []*nostr.Event{},
		styles:  newThreadStyles(),
		keys:    newThreadKeyMap(),
	}

	result := m.View()
	if result.Content == "" {
		t.Errorf("expected non-empty view for empty replies")
	}
}

func TestThreadView_NilReplies(t *testing.T) {
	m := &threadView{
		replies: nil,
		styles:  newThreadStyles(),
		keys:    newThreadKeyMap(),
	}

	result := m.View()
	if result.Content == "" {
		t.Errorf("expected non-empty view for nil replies")
	}
}

func TestThreadView_WithParentAndEmptyReplies(t *testing.T) {
	parent := &nostr.Event{
		Content: "parent content",
	}
	m := &threadView{
		parent:  parent,
		replies: []*nostr.Event{},
		styles:  newThreadStyles(),
		keys:    newThreadKeyMap(),
	}

	result := m.View()
	if result.Content == "" {
		t.Errorf("expected non-empty view with parent but empty replies")
	}
}

func TestThreadView_WithParentAndNilReplies(t *testing.T) {
	parent := &nostr.Event{
		Content: "parent content",
	}
	m := &threadView{
		parent:  parent,
		replies: nil,
		styles:  newThreadStyles(),
		keys:    newThreadKeyMap(),
	}

	result := m.View()
	if result.Content == "" {
		t.Errorf("expected non-empty view with parent but nil replies")
	}
}

func TestThreadView_NoParentNoReplies(t *testing.T) {
	m := &threadView{
		parent:  nil,
		replies: nil,
		styles:  newThreadStyles(),
		keys:    newThreadKeyMap(),
	}

	result := m.View()
	if result.Content == "" {
		t.Errorf("expected non-empty view with no parent and no replies")
	}
}

func TestNewThreadView(t *testing.T) {
	event := &nostr.Event{
		Content: "test event",
	}

	m := &threadView{
		event:   event,
		styles:  newThreadStyles(),
		keys:    newThreadKeyMap(),
		width:   80,
		height:  24,
	}

	if m.event != event {
		t.Errorf("event not set correctly")
	}
	if m.styles.header.String() == "" {
		t.Errorf("styles.header should not be empty")
	}
}

func TestUpdate_ThreadLoadedMsg(t *testing.T) {
	m := &threadView{
		replies: []*nostr.Event{},
	}

	msg := threadLoadedMsg{}
	_, cmd := m.Update(msg)

	if cmd != nil {
		t.Errorf("expected nil command for threadLoadedMsg")
	}
}

func TestUpdate_QuitKey(t *testing.T) {
	m := &threadView{
		replies: []*nostr.Event{},
		keys:    newThreadKeyMap(),
	}

	msg := tea.KeyPressMsg{Text: "esc"}
	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Errorf("expected non-nil command for quit key")
	}
}