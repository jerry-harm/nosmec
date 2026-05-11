package utils

import (
	"testing"

	"fiatjaf.com/nostr"
)

func TestQuoteNoteTags(t *testing.T) {
	quotedID := "test-event-id-123"

	tags := nostr.Tags{
		{"q", quotedID},
	}

	if len(tags) != 1 {
		t.Errorf("len(tags) = %d, want 1", len(tags))
	}
	if tags[0][0] != "q" || tags[0][1] != quotedID {
		t.Errorf("tags[0] = %v, want [q, %s]", tags[0], quotedID)
	}
}

func TestReplyToNoteTags(t *testing.T) {
	parentID := "parent-event-id"
	parentPubKey := nostr.PubKey{}

	tags := nostr.Tags{
		{"e", parentID, "", "reply"},
		{"p", parentPubKey.Hex()},
	}

	if len(tags) != 2 {
		t.Errorf("len(tags) = %d, want 2", len(tags))
	}

	if tags[0][0] != "e" || tags[0][1] != parentID {
		t.Errorf("tags[0] = %v, want [e, %s]", tags[0], parentID)
	}
	if tags[0][3] != "reply" {
		t.Errorf("tags[0][3] = %q, want %q", tags[0][3], "reply")
	}

	if tags[1][0] != "p" {
		t.Errorf("tags[1][0] = %q, want %q", tags[1][0], "p")
	}
}

func TestDeleteNoteTags(t *testing.T) {
	eventID := "delete-this-event"

	tags := nostr.Tags{
		{"e", eventID},
	}

	if len(tags) != 1 {
		t.Errorf("len(tags) = %d, want 1", len(tags))
	}
	if tags[0][0] != "e" || tags[0][1] != eventID {
		t.Errorf("tags[0] = %v, want [e, %s]", tags[0], eventID)
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
	parentID := "abcd1234"
	parentPubKeyHex := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"

	tags := nostr.Tags{
		{"e", parentID, "", "reply"},
		{"p", parentPubKeyHex},
	}

	foundReply := false
	foundP := false
	for _, tag := range tags {
		if len(tag) >= 2 {
			if tag[0] == "e" && len(tag) >= 4 && tag[3] == "reply" {
				foundReply = true
			}
			if tag[0] == "p" {
				foundP = true
			}
		}
	}
	if !foundReply {
		t.Error("reply tag not found")
	}
	if !foundP {
		t.Error("p tag not found")
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