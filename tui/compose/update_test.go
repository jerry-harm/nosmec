package compose

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestUpdate_KindInputTabNavigatesToTagInput(t *testing.T) {
	m := &model{}
	m.kindInput.Focus()

	msg := tea.KeyPressMsg{Text: "tab"}
	got, _ := m.Update(msg)
	if _, ok := got.(*model); !ok {
		t.Errorf("Update did not return *model")
	}
	if !m.tagInput.Focused() {
		t.Errorf("tagInput should be focused after tab on kindInput")
	}
}

func TestUpdate_TabFromContentInputNavigatesToKindInput(t *testing.T) {
	m := &model{}
	m.contentInput.Focus()

	msg := tea.KeyPressMsg{Text: "tab"}
	got, _ := m.Update(msg)
	if _, ok := got.(*model); !ok {
		t.Errorf("Update did not return *model")
	}
	if !m.kindInput.Focused() {
		t.Errorf("kindInput should be focused after tab on contentInput")
	}
}

func TestUpdate_AddTag(t *testing.T) {
	m := &model{}
	m.tagInput.Focus()
	m.tagInput.SetValue("e:abc123")

	msg := tea.KeyPressMsg{Text: "enter"}
	got, _ := m.Update(msg)
	if _, ok := got.(*model); !ok {
		t.Errorf("Update did not return *model")
	}
	if len(m.tags) != 1 {
		t.Errorf("len(m.tags) = %d, want 1", len(m.tags))
	}
	if m.tags[0][0] != "e" || m.tags[0][1] != "abc123" {
		t.Errorf("tags[0] = %v, want e-tag with abc123", m.tags[0])
	}
}

func TestUpdate_WindowSizeMsg(t *testing.T) {
	m := &model{}
	m.width = 80
	m.height = 24

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	got, _ := m.Update(msg)
	if _, ok := got.(*model); !ok {
		t.Errorf("Update did not return *model")
	}
	if m.width != 120 {
		t.Errorf("width = %d, want 120", m.width)
	}
	if m.height != 40 {
		t.Errorf("height = %d, want 40", m.height)
	}
}

func TestUpdate_Quit(t *testing.T) {
	m := &model{}
	m.isStandalone = true

	msg := tea.KeyPressMsg{Code: tea.KeyEscape}
	got, cmd := m.Update(msg)
	if _, ok := got.(*model); !ok {
		t.Errorf("Update did not return *model")
	}
	if cmd == nil {
		t.Errorf("cmd should not be nil for standalone quit")
	}
}