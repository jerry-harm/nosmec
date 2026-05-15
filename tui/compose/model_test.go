package compose

import (
	"testing"

	"charm.land/bubbles/v2/textinput"
	"fiatjaf.com/nostr"
)

func TestParseKind_Default(t *testing.T) {
	m := &model{composeKind: KindNote}
	got := m.parseKind()
	if got != 1 {
		t.Errorf("KindNote default = %d, want 1", got)
	}
}

func TestParseKind_Reply(t *testing.T) {
	m := &model{composeKind: KindReply}
	got := m.parseKind()
	if got != 1111 {
		t.Errorf("KindReply with nil parent = %d, want 1111 (nostr.KindComment)", got)
	}
}

func TestParseKind_ReplyWithParent(t *testing.T) {
	parentEvent := &nostr.Event{Kind: 1}
	m := &model{composeKind: KindReply, parentEvent: parentEvent}
	got := m.parseKind()
	if got != 1 {
		t.Errorf("KindReply with parent Kind=1 = %d, want 1", got)
	}
}

func TestParseKind_Community(t *testing.T) {
	m := &model{composeKind: KindCommunity}
	got := m.parseKind()
	if got != 1111 {
		t.Errorf("KindCommunity = %d, want 1111 (nostr.KindComment)", got)
	}
}

func TestParseKind_Explicit(t *testing.T) {
	m := &model{kindInput: textinput.Model{}}
	m.kindInput.SetValue("3")

	got := m.parseKind()
	if got != 3 {
		t.Errorf("parseKind() = %d, want 3", got)
	}
}

func TestParseKind_ExplicitInvalid(t *testing.T) {
	m := &model{kindInput: textinput.Model{}}
	m.kindInput.SetValue("abc")

	got := m.parseKind()
	if got != 1 {
		t.Errorf("parseKind() with invalid = %d, want 1 (default)", got)
	}
}

func TestParseTagInput_Hashtag(t *testing.T) {
	m := &model{}
	tag := m.parseTagInput("nostr")
	if tag.Type != "t" {
		t.Errorf("tag.Type = %q, want 't'", tag.Type)
	}
	if len(tag.Values) != 1 || tag.Values[0] != "nostr" {
		t.Errorf("tag.Values = %v, want ['nostr']", tag.Values)
	}
}

func TestParseTagInput_Eevent(t *testing.T) {
	m := &model{}
	tag := m.parseTagInput("e:abc123")
	if tag.Type != "e" {
		t.Errorf("tag.Type = %q, want 'e'", tag.Type)
	}
	if len(tag.Values) != 1 || tag.Values[0] != "abc123" {
		t.Errorf("tag.Values = %v, want ['abc123']", tag.Values)
	}
}

func TestParseTagInput_Pubkey(t *testing.T) {
	m := &model{}
	tag := m.parseTagInput("p:def456")
	if tag.Type != "p" {
		t.Errorf("tag.Type = %q, want 'p'", tag.Type)
	}
	if len(tag.Values) != 1 || tag.Values[0] != "def456" {
		t.Errorf("tag.Values = %v, want ['def456']", tag.Values)
	}
}

func TestParseTagInput_Relay(t *testing.T) {
	m := &model{}
	tag := m.parseTagInput("r:relay.example")
	if tag.Type != "r" {
		t.Errorf("tag.Type = %q, want 'r'", tag.Type)
	}
	if len(tag.Values) != 1 {
		t.Errorf("len(tag.Values) = %d, want 1", len(tag.Values))
	}
	if tag.Values[0] != "relay.example" {
		t.Errorf("tag.Values[0] = %q, want relay URL", tag.Values[0])
	}
}

func TestParseTagInput_NoColon(t *testing.T) {
	m := &model{}
	tag := m.parseTagInput("justtext")
	if tag.Type != "t" {
		t.Errorf("tag.Type = %q, want 't' (default)", tag.Type)
	}
	if len(tag.Values) != 1 || tag.Values[0] != "justtext" {
		t.Errorf("tag.Values = %v, want ['justtext']", tag.Values)
	}
}

func TestParseTagInput_UnknownType(t *testing.T) {
	m := &model{}
	tag := m.parseTagInput("x:abc")
	if tag.Type != "t" {
		t.Errorf("tag.Type = %q, want 't' (fallback)", tag.Type)
	}
}

func TestParseTagInput_EmptyInput(t *testing.T) {
	m := &model{}
	tag := m.parseTagInput("")
	if tag.Type != "" || len(tag.Values) != 0 {
		t.Errorf("empty input should return empty TagValue, got %+v", tag)
	}
}

func TestParseTagInput_WhitespaceOnly(t *testing.T) {
	m := &model{}
	tag := m.parseTagInput("   ")
	if tag.Type != "" || len(tag.Values) != 0 {
		t.Errorf("whitespace only should return empty TagValue, got %+v", tag)
	}
}

func TestParseTagInput_RelayWithColon(t *testing.T) {
	m := &model{}
	tag := m.parseTagInput("r:relay.example")
	if tag.Type != "r" {
		t.Errorf("tag.Type = %q, want 'r'", tag.Type)
	}
	if len(tag.Values) != 1 {
		t.Errorf("len(tag.Values) = %d, want 1", len(tag.Values))
	}
}

func TestParseTagInput_RelayWithColonsSplits(t *testing.T) {
	m := &model{}
	tag := m.parseTagInput("r:wss://relay.example")
	if tag.Type != "r" {
		t.Errorf("tag.Type = %q, want 'r'", tag.Type)
	}
	if len(tag.Values) != 2 {
		t.Errorf("len(tag.Values) = %d, want 2 (:// splits into separate parts)", len(tag.Values))
	}
}

func TestParseTagInput_LeadingWhitespace(t *testing.T) {
	m := &model{}
	tag := m.parseTagInput("  p:abc123  ")
	if tag.Type != "p" {
		t.Errorf("tag.Type = %q, want 'p'", tag.Type)
	}
	if len(tag.Values) != 1 || tag.Values[0] != "abc123" {
		t.Errorf("tag.Values = %v, want ['abc123']", tag.Values)
	}
}

func TestParseTagInput_ValueWithComma(t *testing.T) {
	m := &model{}
	tag := m.parseTagInput("p:abc,def")
	if tag.Type != "p" {
		t.Errorf("tag.Type = %q, want 'p'", tag.Type)
	}
	if len(tag.Values) != 1 {
		t.Errorf("len(tag.Values) = %d, want 1 (comma is not a separator)", len(tag.Values))
	}
	if tag.Values[0] != "abc,def" {
		t.Errorf("tag.Values[0] = %q, want 'abc,def'", tag.Values[0])
	}
}

func TestParseTagInput_MultipleEmptyValues(t *testing.T) {
	m := &model{}
	tag := m.parseTagInput("p::")
	if tag.Type != "t" {
		t.Errorf("all empty values should fallback to 't', got Type=%q", tag.Type)
	}
}

func TestTagToString_Empty(t *testing.T) {
	m := &model{}
	tag := TagValue{Type: "", Values: nil}
	got := m.tagToString(tag)
	if got != ":" {
		t.Errorf("tagToString = %q, want ':'", got)
	}
}

func TestTagToString_SingleValue(t *testing.T) {
	m := &model{}
	tag := TagValue{Type: "t", Values: []string{"nostr"}}
	got := m.tagToString(tag)
	if got != "t:nostr" {
		t.Errorf("tagToString = %q, want 't:nostr'", got)
	}
}

func TestTagToString_MultipleValues(t *testing.T) {
	m := &model{}
	tag := TagValue{Type: "e", Values: []string{"abc123", "def456"}}
	got := m.tagToString(tag)
	if got != "e:abc123:def456" {
		t.Errorf("tagToString = %q, want 'e:abc123:def456'", got)
	}
}

func TestAddReply(t *testing.T) {
	parent := &nostr.Event{
		ID:    [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
		PubKey: [32]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
		Kind: 1,
	}
	m := &model{}
	m.AddReply(parent)

	if m.composeKind != KindReply {
		t.Errorf("composeKind = %v, want KindReply", m.composeKind)
	}
	if m.parentEvent != parent {
		t.Errorf("parentEvent not set correctly")
	}
	if m.parentID != parent.ID.Hex() {
		t.Errorf("parentID = %q, want %q", m.parentID, parent.ID.Hex())
	}
	if len(m.tags) != 2 {
		t.Fatalf("len(m.tags) = %d, want 2", len(m.tags))
	}
	if m.tags[0].Type != "e" || m.tags[0].Values[0] != parent.ID.Hex() {
		t.Errorf("tags[0] = %+v, want e-tag with parent ID", m.tags[0])
	}
	if m.tags[1].Type != "p" || m.tags[1].Values[0] != parent.PubKey.Hex() {
		t.Errorf("tags[1] = %+v, want p-tag with parent PubKey", m.tags[1])
	}
}

func TestAddQuote(t *testing.T) {
	parent := &nostr.Event{
		ID:    [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
		PubKey: [32]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
		Kind: 1,
	}
	m := &model{}
	m.AddQuote(parent)

	if m.composeKind != KindQuote {
		t.Errorf("composeKind = %v, want KindQuote", m.composeKind)
	}
	if m.parentEvent != parent {
		t.Errorf("parentEvent not set correctly")
	}
	if m.quotedID != parent.ID.Hex() {
		t.Errorf("quotedID = %q, want %q", m.quotedID, parent.ID.Hex())
	}
	if len(m.tags) != 1 {
		t.Fatalf("len(m.tags) = %d, want 1", len(m.tags))
	}
	if m.tags[0].Type != "q" || m.tags[0].Values[0] != parent.ID.Hex() {
		t.Errorf("tags[0] = %+v, want q-tag with quoted ID", m.tags[0])
	}
}

// Note: ClearDraft calls contentInput.SetValue("") which panics on
// uninitialized textarea (nil viewport). Requires full tea.App
// initialization. Covered by AddReply+Update integration tests in the future.

func TestRenderHeader(t *testing.T) {
	tests := []struct {
		name string
		kind ComposeKind
		addr string
		want string
	}{
		{"KindNote", KindNote, "", "New Note"},
		{"KindReply", KindReply, "", "Reply"},
		{"KindQuote", KindQuote, "", "Quote"},
		{"KindCommunity", KindCommunity, "cool.nfx", "Community: cool.nfx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &model{composeKind: tt.kind, communityAddr: tt.addr}
			got := m.renderHeader()
			if got != tt.want {
				t.Errorf("renderHeader() = %q, want %q", got, tt.want)
			}
		})
	}
}