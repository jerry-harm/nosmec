package compose

import (
	"testing"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

func TestUpdate_KindInputTabNavigatesToTagInput(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.kindInput = textinput.New()
	m.tagInput = textinput.New()
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
	m.keys = newKeyMap()
	m.contentInput = textarea.New()
	m.kindInput = textinput.New()
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
	m.keys = newKeyMap()
	m.tagInput = textinput.New()
	m.tagInput.Focus()
	m.tagInput.SetValue("e:abc123")
	m.tags = []Tag{}
	m.editingTagIndex = -1
	m.editingItemIndex = -1

	msg := tea.KeyPressMsg{Text: "enter"}
	m.Update(msg)

	if len(m.tags) != 1 {
		t.Errorf("len(m.tags) = %d, want 1", len(m.tags))
	}
	if len(m.tags[0]) != 1 || m.tags[0][0] != "e:abc123" {
		t.Errorf("tags[0] = %v, want [e:abc123]", m.tags[0])
	}
}

func TestUpdate_WindowSizeMsg(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.contentInput = textarea.New()
	m.kindInput = textinput.New()
	m.tagInput = textinput.New()
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
	m.keys = newKeyMap()
	m.kindInput = textinput.New()
	m.tagInput = textinput.New()
	m.isStandalone = true
	m.editingTagIndex = -1

	msg := tea.KeyPressMsg{Text: "esc"}
	got, cmd := m.Update(msg)
	if _, ok := got.(*model); !ok {
		t.Errorf("Update did not return *model")
	}
	if cmd == nil {
		t.Errorf("cmd should not be nil for standalone quit")
	}
}