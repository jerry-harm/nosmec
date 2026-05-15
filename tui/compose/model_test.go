package compose

import (
	"testing"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbletea/v2"
)

func TestTagAdd_NewTag(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.tagInput = textinput.New()
	m.tags = []Tag{{"e", "event1"}}
	m.editingTagIndex = -1
	m.editingItemIndex = -1
	m.tagInput.Focus()
	m.tagInput.SetValue("newtag")

	msg := tea.KeyPressMsg{Text: "enter"}
	m.Update(msg)

	if len(m.tags) != 2 {
		t.Errorf("len(m.tags) = %d, want 2", len(m.tags))
	}
	if m.tags[1][0] != "newtag" {
		t.Errorf("m.tags[1][0] = %q, want 'newtag'", m.tags[1][0])
	}
}

func TestTagAdd_AppendToExisting(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.tagInput = textinput.New()
	m.tags = []Tag{{"e", "event1"}}
	m.editingTagIndex = 0
	m.editingItemIndex = 2
	m.tagInput.Focus()
	m.tagInput.SetValue("relay1")

	msg := tea.KeyPressMsg{Text: "enter"}
	m.Update(msg)

	if len(m.tags[0]) != 3 {
		t.Errorf("len(m.tags[0]) = %d, want 3", len(m.tags[0]))
	}
	if m.tags[0][2] != "relay1" {
		t.Errorf("m.tags[0][2] = %q, want 'relay1'", m.tags[0][2])
	}
}

func TestTagAdd_InsertAtMiddle(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.tagInput = textinput.New()
	m.tags = []Tag{{"e", "a", "c"}}
	m.editingTagIndex = 0
	m.editingItemIndex = 1
	m.tagInput.Focus()
	m.tagInput.SetValue("b")

	msg := tea.KeyPressMsg{Text: "enter"}
	m.Update(msg)

	if len(m.tags[0]) != 4 {
		t.Errorf("len(m.tags[0]) = %d, want 4", len(m.tags[0]))
	}
	if m.tags[0][0] != "e" || m.tags[0][1] != "b" || m.tags[0][2] != "a" || m.tags[0][3] != "c" {
		t.Errorf("m.tags[0] = %v, want [e b a c]", m.tags[0])
	}
}

func TestTagBackspace_AtAppendPosition(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.tagInput = textinput.New()
	m.tags = []Tag{{"e", "a", "b"}}
	m.editingTagIndex = 0
	m.editingItemIndex = 3
	m.tagInput.Focus()
	m.tagInput.SetValue("")

	msg := tea.KeyPressMsg{Text: "backspace"}
	m.Update(msg)

	if len(m.tags[0]) != 2 {
		t.Errorf("len(m.tags[0]) = %d, want 2", len(m.tags[0]))
	}
	if m.tags[0][0] != "e" || m.tags[0][1] != "a" {
		t.Errorf("m.tags[0] = %v, want [e a]", m.tags[0])
	}
	if m.editingItemIndex != 2 {
		t.Errorf("m.editingItemIndex = %d, want 2", m.editingItemIndex)
	}
}

func TestTagBackspace_AtMiddleItem(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.tagInput = textinput.New()
	m.tags = []Tag{{"e", "a", "b", "c"}}
	m.editingTagIndex = 0
	m.editingItemIndex = 2
	m.tagInput.Focus()
	m.tagInput.SetValue("")

	msg := tea.KeyPressMsg{Text: "backspace"}
	m.Update(msg)

	if len(m.tags[0]) != 3 {
		t.Errorf("len(m.tags[0]) = %d, want 3", len(m.tags[0]))
	}
	if m.tags[0][0] != "e" || m.tags[0][1] != "a" || m.tags[0][2] != "c" {
		t.Errorf("m.tags[0] = %v, want [e a c]", m.tags[0])
	}
	if m.editingItemIndex != 1 {
		t.Errorf("m.editingItemIndex = %d, want 1", m.editingItemIndex)
	}
}

func TestTagBackspace_AtFirstItemDeletesTag(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.tagInput = textinput.New()
	m.tags = []Tag{{"e", "event1"}, {"p", "pubkey1"}}
	m.editingTagIndex = 0
	m.editingItemIndex = 0
	m.tagInput.Focus()
	m.tagInput.SetValue("")

	msg := tea.KeyPressMsg{Text: "backspace"}
	m.Update(msg)

	if len(m.tags) != 2 {
		t.Errorf("len(m.tags) = %d, want 2", len(m.tags))
	}
	if m.tags[0][0] != "event1" {
		t.Errorf("m.tags[0] = %v, want [event1]", m.tags[0])
	}
	if m.editingItemIndex != 0 {
		t.Errorf("m.editingItemIndex = %d, want 0", m.editingItemIndex)
	}
}

func TestTagTab_EmptySlotGoesToContent(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.tagInput = textinput.New()
	m.contentInput = textarea.New()
	m.kindInput = textinput.New()
	m.tags = []Tag{{"e", "event1"}}
	m.editingTagIndex = -1
	m.editingItemIndex = -1
	m.tagInput.Focus()
	m.tagInput.SetValue("")

	msg := tea.KeyPressMsg{Text: "tab"}
	m.Update(msg)

	if m.tagInput.Focused() {
		t.Errorf("tagInput should be blurred")
	}
	if !m.contentInput.Focused() {
		t.Errorf("contentInput should be focused")
	}
	if m.editingTagIndex != -1 {
		t.Errorf("m.editingTagIndex = %d, want -1", m.editingTagIndex)
	}
}

func TestTagShiftTab_EmptySlotGoesToPreviousTag(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.tagInput = textinput.New()
	m.contentInput = textarea.New()
	m.kindInput = textinput.New()
	m.tags = []Tag{{"e", "event1"}, {"p", "pubkey1"}}
	m.editingTagIndex = 1
	m.editingItemIndex = 0
	m.tagInput.Focus()
	m.tagInput.SetValue("")

	msg := tea.KeyPressMsg{Text: "shift+tab"}
	m.Update(msg)

	if m.editingTagIndex != 0 {
		t.Errorf("m.editingTagIndex = %d, want 0", m.editingTagIndex)
	}
	if m.editingItemIndex != 2 {
		t.Errorf("m.editingItemIndex = %d, want 2", m.editingItemIndex)
	}
}
