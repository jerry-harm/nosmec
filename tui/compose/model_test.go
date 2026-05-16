package compose

import (
	"testing"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbletea/v2"
)

func TestTagAdd_NewTagFromJSON(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.tagInput = textinput.New()
	m.tags = []Tag{}
	m.editingIndex = -1
	m.tagInput.Focus()
	m.tagInput.SetValue(`["event123"]`)

	msg := tea.KeyPressMsg{Text: "enter"}
	m.Update(msg)

	if len(m.tags) != 1 {
		t.Errorf("len(m.tags) = %d, want 1", len(m.tags))
	}
	if len(m.tags[0]) != 1 || m.tags[0][0] != "event123" {
		t.Errorf("m.tags[0] = %v, want [event123]", m.tags[0])
	}
	if m.editingIndex != -1 {
		t.Errorf("m.editingIndex = %d, want -1", m.editingIndex)
	}
}

func TestTagAdd_NewTagMultiValue(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.tagInput = textinput.New()
	m.tags = []Tag{}
	m.editingIndex = -1
	m.tagInput.Focus()
	m.tagInput.SetValue(`["pubkey1","pubkey2"]`)

	msg := tea.KeyPressMsg{Text: "enter"}
	m.Update(msg)

	if len(m.tags) != 1 {
		t.Errorf("len(m.tags) = %d, want 1", len(m.tags))
	}
	if len(m.tags[0]) != 2 || m.tags[0][0] != "pubkey1" || m.tags[0][1] != "pubkey2" {
		t.Errorf("m.tags[0] = %v, want [pubkey1, pubkey2]", m.tags[0])
	}
}

func TestTagAdd_ReplaceExisting(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.tagInput = textinput.New()
	m.tags = []Tag{{"e", "event1"}}
	m.editingIndex = 0
	m.tagInput.Focus()
	m.tagInput.SetValue(`["replaced"]`)

	msg := tea.KeyPressMsg{Text: "enter"}
	m.Update(msg)

	if len(m.tags) != 1 {
		t.Errorf("len(m.tags) = %d, want 1", len(m.tags))
	}
	if m.tags[0][0] != "replaced" {
		t.Errorf("m.tags[0][0] = %q, want 'replaced'", m.tags[0][0])
	}
}

func TestTagEnter_EmptyBlursToContent(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.tagInput = textinput.New()
	m.contentInput = textarea.New()
	m.tags = []Tag{{"e", "event1"}}
	m.editingIndex = -1
	m.tagInput.Focus()
	m.tagInput.SetValue("")

	msg := tea.KeyPressMsg{Text: "enter"}
	m.Update(msg)

	if m.tagInput.Focused() {
		t.Errorf("tagInput should be blurred")
	}
	if !m.contentInput.Focused() {
		t.Errorf("contentInput should be focused")
	}
	if m.editingIndex != -2 {
		t.Errorf("m.editingIndex = %d, want -2", m.editingIndex)
	}
}

func TestTagBackspace_EmptyGoesToLastTag(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.tagInput = textinput.New()
	m.tags = []Tag{{"e", "event1"}, {"p", "pubkey1"}}
	m.editingIndex = -1
	m.tagInput.Focus()
	m.tagInput.SetValue("")

	msg := tea.KeyPressMsg{Text: "backspace"}
	m.Update(msg)

	if m.editingIndex != 1 {
		t.Errorf("m.editingIndex = %d, want 1", m.editingIndex)
	}
	if m.tagInput.Value() != `["p","pubkey1"]` {
		t.Errorf("m.tagInput.Value() = %q, want %q", m.tagInput.Value(), `["p","pubkey1"]`)
	}
}

func TestTagBackspace_DeletesTag(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.tagInput = textinput.New()
	m.tags = []Tag{{"e", "event1"}, {"p", "pubkey1"}}
	m.editingIndex = 0
	m.tagInput.Focus()
	m.tagInput.SetValue("")

	msg := tea.KeyPressMsg{Text: "backspace"}
	m.Update(msg)

	if len(m.tags) != 1 {
		t.Errorf("len(m.tags) = %d, want 1", len(m.tags))
	}
	if m.editingIndex != 0 {
		t.Errorf("m.editingIndex = %d, want 0", m.editingIndex)
	}
}

func TestTagTab_EmptySlotGoesToContent(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.tagInput = textinput.New()
	m.contentInput = textarea.New()
	m.kindInput = textinput.New()
	m.tags = []Tag{{"e", "event1"}}
	m.editingIndex = -1
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
	if m.editingIndex != -2 {
		t.Errorf("m.editingIndex = %d, want -2", m.editingIndex)
	}
}

func TestTagTab_EditingTagAdvances(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.tagInput = textinput.New()
	m.tags = []Tag{{"e", "event1"}, {"p", "pubkey1"}}
	m.editingIndex = 0
	m.tagInput.Focus()
	m.tagInput.SetValue(`["e","event1"]`)

	msg := tea.KeyPressMsg{Text: "tab"}
	m.Update(msg)

	if m.editingIndex != 1 {
		t.Errorf("m.editingIndex = %d, want 1", m.editingIndex)
	}
	if m.tagInput.Value() != `["p","pubkey1"]` {
		t.Errorf("m.tagInput.Value() = %q, want %q", m.tagInput.Value(), `["p","pubkey1"]`)
	}
}

func TestTagTab_LastTagGoesToEmptySlot(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.tagInput = textinput.New()
	m.contentInput = textarea.New()
	m.tags = []Tag{{"e", "event1"}}
	m.editingIndex = 0
	m.tagInput.Focus()
	m.tagInput.SetValue(`["e","event1"]`)

	msg := tea.KeyPressMsg{Text: "tab"}
	m.Update(msg)

	if m.editingIndex != -1 {
		t.Errorf("m.editingIndex = %d, want -1", m.editingIndex)
	}
	if m.tagInput.Value() != "" {
		t.Errorf("m.tagInput.Value() = %q, want empty", m.tagInput.Value())
	}
	if !m.tagInput.Focused() {
		t.Errorf("tagInput should still be focused")
	}
}

func TestKindTabGoesToFirstTag(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.kindInput = textinput.New()
	m.tagInput = textinput.New()
	m.tags = []Tag{{"e", "event1"}, {"p", "pubkey1"}}
	m.kindInput.Focus()

	msg := tea.KeyPressMsg{Text: "tab"}
	m.Update(msg)

	if !m.tagInput.Focused() {
		t.Errorf("tagInput should be focused")
	}
	if m.editingIndex != 0 {
		t.Errorf("m.editingIndex = %d, want 0 (first tag)", m.editingIndex)
	}
	if m.tagInput.Value() != `["e","event1"]` {
		t.Errorf("m.tagInput.Value() = %q, want %q", m.tagInput.Value(), `["e","event1"]`)
	}
}

func TestTagShiftTab_EmptyGoesToLastTag(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.tagInput = textinput.New()
	m.kindInput = textinput.New()
	m.tags = []Tag{{"e", "event1"}, {"p", "pubkey1"}}
	m.editingIndex = -1
	m.tagInput.Focus()
	m.tagInput.SetValue("")

	msg := tea.KeyPressMsg{Text: "shift+tab"}
	m.Update(msg)

	if m.editingIndex != 1 {
		t.Errorf("m.editingIndex = %d, want 1", m.editingIndex)
	}
	if m.tagInput.Value() != `["p","pubkey1"]` {
		t.Errorf("m.tagInput.Value() = %q, want %q", m.tagInput.Value(), `["p","pubkey1"]`)
	}
}

func TestTagShiftTab_EmptyWithNoTagsGoesToKind(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.tagInput = textinput.New()
	m.kindInput = textinput.New()
	m.tags = []Tag{}
	m.editingIndex = -1
	m.tagInput.Focus()
	m.tagInput.SetValue("")

	msg := tea.KeyPressMsg{Text: "shift+tab"}
	m.Update(msg)

	if !m.kindInput.Focused() {
		t.Errorf("kindInput should be focused")
	}
	if m.editingIndex != -2 {
		t.Errorf("m.editingIndex = %d, want -2", m.editingIndex)
	}
}

func TestTagShiftTab_EditingTagRetreats(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.tagInput = textinput.New()
	m.tags = []Tag{{"e", "event1"}, {"p", "pubkey1"}}
	m.editingIndex = 1
	m.tagInput.Focus()
	m.tagInput.SetValue(`["p","pubkey1"]`)

	msg := tea.KeyPressMsg{Text: "shift+tab"}
	m.Update(msg)

	if m.editingIndex != 0 {
		t.Errorf("m.editingIndex = %d, want 0", m.editingIndex)
	}
	if m.tagInput.Value() != `["e","event1"]` {
		t.Errorf("m.tagInput.Value() = %q, want %q", m.tagInput.Value(), `["e","event1"]`)
	}
}

func TestTagShiftTab_FirstTagGoesToKind(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.tagInput = textinput.New()
	m.kindInput = textinput.New()
	m.tags = []Tag{{"e", "event1"}}
	m.editingIndex = 0
	m.tagInput.Focus()
	m.tagInput.SetValue(`["e","event1"]`)

	msg := tea.KeyPressMsg{Text: "shift+tab"}
	m.Update(msg)

	if !m.kindInput.Focused() {
		t.Errorf("kindInput should be focused")
	}
	if m.editingIndex != -2 {
		t.Errorf("m.editingIndex = %d, want -2", m.editingIndex)
	}
}

func TestTagAdd_WithType(t *testing.T) {
	m := &model{}
	m.keys = newKeyMap()
	m.tagInput = textinput.New()
	m.tags = []Tag{}
	m.editingIndex = -1
	m.tagInput.Focus()
	m.tagInput.SetValue(`["e","event123","relay1"]`)

	msg := tea.KeyPressMsg{Text: "enter"}
	m.Update(msg)

	if len(m.tags) != 1 {
		t.Errorf("len(m.tags) = %d, want 1", len(m.tags))
	}
	if m.tags[0][0] != "e" || m.tags[0][1] != "event123" || m.tags[0][2] != "relay1" {
		t.Errorf("m.tags[0] = %v, want [e, event123, relay1]", m.tags[0])
	}
}