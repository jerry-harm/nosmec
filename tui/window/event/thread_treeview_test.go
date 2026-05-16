package event

import (
	"testing"

	"fiatjaf.com/nostr"
)

func TestExtractParentID_RootEvent(t *testing.T) {
	// Event with no e tags = root (no parent)
	event := &nostr.Event{
		Content: "root note",
		Tags:    nostr.Tags{},
	}
	parentID := extractParentID(event)
	if parentID != "" {
		t.Errorf("root event should have empty parent ID, got %q", parentID)
	}
}

func TestExtractParentID_RootMarker(t *testing.T) {
	// Event with "root" marker = root (no parent)
	event := &nostr.Event{
		Content: "note with root marker",
		Tags: nostr.Tags{
			nostr.Tag{"e", "abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234", "", "root"},
		},
	}
	parentID := extractParentID(event)
	if parentID != "" {
		t.Errorf("event with root marker should have empty parent ID, got %q", parentID)
	}
}

func TestExtractParentID_ReplyMarker(t *testing.T) {
	// Event with "reply" marker - parent is the e tag value
	event := &nostr.Event{
		Content: "reply note",
		Tags: nostr.Tags{
			nostr.Tag{"e", "parentID1234parentID1234parentID1234parentID1234parentID1234parent", "", "reply"},
		},
	}
	parentID := extractParentID(event)
	expected := "parentID1234parentID1234parentID1234parentID1234parentID1234parent"
	if parentID != expected {
		t.Errorf("reply event should have parent ID %q, got %q", expected, parentID)
	}
}

func TestExtractParentID_ReplyMarkerWithRelay(t *testing.T) {
	// Event with "reply" marker and relay hint
	event := &nostr.Event{
		Content: "reply note with relay",
		Tags: nostr.Tags{
			nostr.Tag{"e", "parentID1234parentID1234parentID1234parentID1234parentID1234parent", "wss://relay.example.com", "reply"},
		},
	}
	parentID := extractParentID(event)
	expected := "parentID1234parentID1234parentID1234parentID1234parentID1234parent"
	if parentID != expected {
		t.Errorf("reply event should have parent ID %q, got %q", expected, parentID)
	}
}

func TestExtractParentID_MultipleETags(t *testing.T) {
	// Multiple e tags - should pick the first one with "reply" marker
	event := &nostr.Event{
		Content: "multi e-tag note",
		Tags: nostr.Tags{
			nostr.Tag{"e", "firstParent1234firstParent1234firstParent1234firstParent1234firstP", "", "reply"},
			nostr.Tag{"e", "secondParent1234secondParent1234secondParent1234secondParent12", "", "root"},
		},
	}
	parentID := extractParentID(event)
	expected := "firstParent1234firstParent1234firstParent1234firstParent1234firstP"
	if parentID != expected {
		t.Errorf("should use first reply marker parent ID, got %q", parentID)
	}
}

func TestExtractParentID_NoMarker(t *testing.T) {
	// e tag without marker - treat as root (no parent per NIP-10)
	event := &nostr.Event{
		Content: "note with e tag but no marker",
		Tags: nostr.Tags{
			nostr.Tag{"e", "someID1234someID1234someID1234someID1234someID1234someID1234"},
		},
	}
	parentID := extractParentID(event)
	if parentID != "" {
		t.Errorf("e tag without marker should be treated as root, got %q", parentID)
	}
}

func TestExtractRootEvent_NilEvent(t *testing.T) {
	event := (*nostr.Event)(nil)
	rootID, isRoot, err := extractRootEvent(event)
	if err == nil {
		t.Errorf("nil event should return error")
	}
	if rootID != (nostr.ID{}) {
		t.Errorf("nil event should have zero root ID")
	}
	if isRoot {
		t.Errorf("nil event should not be root")
	}
}

func TestExtractRootEvent_NoETags(t *testing.T) {
	// No e tags = root
	event := &nostr.Event{
		Content: "root note",
		Tags:    nostr.Tags{},
	}
	rootID, isRoot, err := extractRootEvent(event)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !isRoot {
		t.Errorf("event with no e tags should be root")
	}
	if rootID != event.ID {
		t.Errorf("root ID should equal event ID")
	}
}

func TestExtractRootEvent_RootMarker(t *testing.T) {
	// Root marker = event IS the root
	event := &nostr.Event{
		Content: "root note with marker",
		Tags: nostr.Tags{
			nostr.Tag{"e", "abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234", "", "root"},
		},
	}
	rootID, isRoot, err := extractRootEvent(event)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !isRoot {
		t.Errorf("event with root marker should be root")
	}
	if rootID != event.ID {
		t.Errorf("root ID should equal event ID")
	}
}

func TestExtractRootEvent_ReplyMarker(t *testing.T) {
	// Reply marker = event is NOT root, find parent with root marker
	event := &nostr.Event{
		Content: "reply note",
		Tags: nostr.Tags{
			nostr.Tag{"e", "1111111111111111111111111111111111111111111111111111111111111111", "", "reply"},
			nostr.Tag{"e", "2222222222222222222222222222222222222222222222222222222222222222", "", "root"},
		},
	}
	rootID, isRoot, err := extractRootEvent(event)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if isRoot {
		t.Errorf("reply event should not be root")
	}
	expectedRootID, _ := nostr.IDFromHex("2222222222222222222222222222222222222222222222222222222222222222")
	if rootID != expectedRootID {
		t.Errorf("root ID should be extracted from root marker, got %v", rootID)
	}
}

func TestExtractRootEvent_ReplyNoRoot(t *testing.T) {
	// Reply marker but no root marker - treat event as root
	event := &nostr.Event{
		Content: "reply without root marker",
		Tags: nostr.Tags{
			nostr.Tag{"e", "parentID1234parentID1234parentID1234parentID1234parentID1234parent", "", "reply"},
		},
	}
	rootID, isRoot, err := extractRootEvent(event)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !isRoot {
		t.Errorf("reply without root marker should be treated as root")
	}
	if rootID != event.ID {
		t.Errorf("root ID should equal event ID")
	}
}

func TestNostrEventProvider_ID(t *testing.T) {
	p := &NostrEventProvider{}
	event := nostr.Event{
		ID: [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
	}
	id := p.ID(event)
	expected := "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"
	if id != expected {
		t.Errorf("expected ID %q, got %q", expected, id)
	}
}

func TestNostrEventProvider_Name(t *testing.T) {
	p := &NostrEventProvider{}
	event := nostr.Event{
		Content: "This is a long content that should be truncated to fit the display",
		PubKey:  [32]byte{1, 2, 3, 4, 5, 6, 7, 8},
	}
	name := p.Name(event)
	// Name should be truncated and include short pubkey
	if len(name) == 0 {
		t.Errorf("expected non-empty name")
	}
}

func TestNostrEventProvider_ParentID_Root(t *testing.T) {
	p := &NostrEventProvider{}
	event := nostr.Event{
		Content: "root note",
		Tags:    nostr.Tags{},
	}
	parentID := p.ParentID(event)
	if parentID != "" {
		t.Errorf("root event should have empty parent ID, got %q", parentID)
	}
}

func TestNostrEventProvider_ParentID_Reply(t *testing.T) {
	p := &NostrEventProvider{}
	event := nostr.Event{
		Content: "reply note",
		Tags: nostr.Tags{
			nostr.Tag{"e", "parentID1234parentID1234parentID1234parentID1234parentID1234parent", "", "reply"},
		},
	}
	parentID := p.ParentID(event)
	expected := "parentID1234parentID1234parentID1234parentID1234parentID1234parent"
	if parentID != expected {
		t.Errorf("expected parent ID %q, got %q", expected, parentID)
	}
}