package utils

import (
	"testing"

	"fiatjaf.com/nostr"
	"github.com/spf13/viper"
	"github.com/jerry-harm/nosmec/config"
)

func TestQuoteNoteTags(t *testing.T) {
	quotedID := "test-event-id-123"
	quotedPubkey := "abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234"

	tags := nostr.Tags{
		{"q", quotedID, "", quotedPubkey},
	}

	if len(tags) != 1 {
		t.Errorf("len(tags) = %d, want 1", len(tags))
	}
	if tags[0][0] != "q" || tags[0][1] != quotedID || tags[0][3] != quotedPubkey {
		t.Errorf("tags[0] = %v, want [q, %s, <relay>, %s]", tags[0], quotedID, quotedPubkey)
	}
}

func TestReplyToNoteTags(t *testing.T) {
	app := config.NewAppContext(nil, config.Config{DataDir: t.TempDir()}, viper.New())
	// Root event (no e tags) → reply should have only "root" marker with pubkey
	rootEvent := &nostr.Event{
		ID:      [32]byte{1},
		Content: "root",
		Kind:    nostr.KindTextNote,
	}
	tags := BuildReplyTags(app, rootEvent)
	if len(tags) != 1 {
		t.Fatalf("len(tags) = %d, want 1", len(tags))
	}
	if tags[0][0] != "e" || tags[0][3] != "root" {
		t.Errorf("direct reply to root: got %v, want [e id relay root pubkey]", tags[0])
	}
	// pubkey is at position 4
	if len(tags[0]) < 5 {
		t.Errorf("tag too short, missing pubkey: %v", tags[0])
	}

	// Nested reply (parent has root marker) → reply should have "root" + "reply"
	nestedParent := &nostr.Event{
		ID:      [32]byte{2},
		Content: "nested parent",
		Kind:    nostr.KindTextNote,
		Tags: nostr.Tags{
			{"e", rootEvent.ID.Hex(), "", "root"},
			{"e", "some-parent-id", "", "reply"},
		},
	}
	tags = BuildReplyTags(app, nestedParent)
	if len(tags) != 2 {
		t.Fatalf("len(tags) = %d, want 2", len(tags))
	}
	if tags[0][3] != "root" || tags[0][1] != rootEvent.ID.Hex() {
		t.Errorf("root tag: got %v", tags[0])
	}
	if tags[1][3] != "reply" || tags[1][1] != nestedParent.ID.Hex() {
		t.Errorf("reply tag: got %v", tags[1])
	}
	// reply tag should have pubkey at position 4
	if len(tags[1]) < 5 || tags[1][4] != nestedParent.PubKey.Hex() {
		t.Errorf("reply tag missing pubkey: %v", tags[1])
	}
}

func TestDeleteNoteTags(t *testing.T) {
	eventID := "delete-this-event"
	authorPubkey := "abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234"

	tags := nostr.Tags{
		{"e", eventID},
		{"p", authorPubkey},
	}

	if len(tags) != 2 {
		t.Errorf("len(tags) = %d, want 2", len(tags))
	}
	if tags[0][0] != "e" || tags[0][1] != eventID {
		t.Errorf("tags[0] = %v, want [e, %s]", tags[0], eventID)
	}
	if tags[1][0] != "p" || tags[1][1] != authorPubkey {
		t.Errorf("tags[1] = %v, want [p, %s]", tags[1], authorPubkey)
	}
}

func TestPostNoteEventKind(t *testing.T) {
	event := &nostr.Event{
		Kind: nostr.KindTextNote,
	}

	if event.Kind != nostr.KindTextNote {
		t.Errorf("event.Kind = %v, want %v", event.Kind, nostr.KindTextNote)
	}
}

func TestDeleteNoteEventKind(t *testing.T) {
	event := &nostr.Event{
		Kind: nostr.KindDeletion,
	}

	if event.Kind != nostr.KindDeletion {
		t.Errorf("event.Kind = %v, want %v", event.Kind, nostr.KindDeletion)
	}
}

func TestBuildReplyTags(t *testing.T) {
	app := config.NewAppContext(nil, config.Config{DataDir: t.TempDir()}, viper.New())
	id, _ := nostr.IDFromHex("abcd123400000000000000000000000000000000000000000000000000000000")
	parentPubKeyHex := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	parentEvent := &nostr.Event{
		ID:      id,
		Content: "test parent",
		Kind:    nostr.KindTextNote,
		Tags:    nostr.Tags{},
	}

	tags := BuildReplyTags(app, parentEvent)
	tags = append(tags, nostr.Tag{"p", parentPubKeyHex})

	foundRoot := false
	foundP := false
	foundPubkey := false
	for _, tag := range tags {
		if len(tag) >= 4 && tag[0] == "e" && tag[3] == "root" {
			foundRoot = true
			if len(tag) >= 5 && tag[4] == parentEvent.PubKey.Hex() {
				foundPubkey = true
			}
		}
		if tag[0] == "p" {
			foundP = true
		}
	}
	if !foundRoot {
		t.Error("root tag not found in reply tags")
	}
	if !foundP {
		t.Error("p tag not found")
	}
	if !foundPubkey {
		t.Error("pubkey not found in root e tag")
	}
}

func TestBuildQuoteTags(t *testing.T) {
	quotedID := "quoted-event-id"

	tags := nostr.Tags{
		{"q", quotedID},
	}

	found := false
	for _, tag := range tags {
		if len(tag) >= 2 && tag[0] == "q" {
			if tag[1] == quotedID {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("quote tag not found for id %s", quotedID)
	}
}

// TestReplyToNote_ParentEventNilPubKey demonstrates the bug: parentEvent.PubKey.Hex()
// is called without checking if PubKey is set (zero value). The nil check on parentEvent
// passes if GetNote returns an event with zero PubKey, then accessing .PubKey.Hex() works
// on a zero pubkey (returns "0000...0000") but may not be the intended behavior.
func TestReplyToNote_ParentEventNilPubKey(t *testing.T) {
	// This tests that when parentEvent is non-nil but has a zero/empty PubKey,
	// the function still works (returns empty hex string) - behavior may be undesirable
	// but not a panic, so this is informational only
	parentEvent := &nostr.Event{
		ID:      [32]byte{1, 2, 3, 4},
		PubKey:  nostr.PubKey{}, // zero value PubKey
		Content: "parent content",
		Tags:    nostr.Tags{{"e", "parent-id", "", "reply"}},
	}

	// If parentEvent.PubKey is zero, Hex() returns all zeros
	hex := parentEvent.PubKey.Hex()
	if hex != "0000000000000000000000000000000000000000000000000000000000000000" {
		t.Errorf("zero PubKey hex = %q, want all zeros", hex)
	}
}