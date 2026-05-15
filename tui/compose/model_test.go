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
	if m.tags[0][0] != "e" || m.tags[0][1] != parent.ID.Hex() {
		t.Errorf("tags[0] = %v, want e-tag with parent ID", m.tags[0])
	}
	if m.tags[1][0] != "p" || m.tags[1][1] != parent.PubKey.Hex() {
		t.Errorf("tags[1] = %v, want p-tag with parent PubKey", m.tags[1])
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
	if m.tags[0][0] != "q" || m.tags[0][1] != parent.ID.Hex() {
		t.Errorf("tags[0] = %v, want q-tag with quoted ID", m.tags[0])
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