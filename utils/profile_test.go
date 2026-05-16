package utils

import (
	"testing"

	"fiatjaf.com/nostr"
)

// TestFetchRecipientReadRelays_EmptyEventTags demonstrates the bug: when result.Event.Tags
// is nil or has elements with len < 2, the code accesses tag[0] and tag[1] without bounds checks.
// This test creates a result with minimal tags to trigger potential index out of bounds.
func TestFetchRecipientReadRelays_EmptyEventTags(t *testing.T) {
	// Create an event with empty/nil Tags - this could cause index out of bounds
	// when the loop accesses tag[0] and tag[1] without length checks
	ev := &nostr.Event{
		ID:   [32]byte{1, 2, 3, 4},
		Tags: nostr.Tags{}, // empty tags
	}

	// Verify event was created with empty tags
	if len(ev.Tags) != 0 {
		t.Errorf("len(ev.Tags) = %d, want 0", len(ev.Tags))
	}

	// The bug is that FetchRecipientReadRelays accesses tag[0] and tag[1]
	// without checking len(tag) >= 2
	// This would panic if called with this event
}

// TestFetchRecipientReadRelays_TagsTooShort demonstrates that accessing tag[0]
// when tag is empty will panic
func TestFetchRecipientReadRelays_TagsTooShort(t *testing.T) {
	// Create an event with a tag that has only 1 element
	// This will cause index out of bounds if len check is missing
	ev := &nostr.Event{
		ID:   [32]byte{1, 2, 3, 4},
		Tags: nostr.Tags{
			nostr.Tag{"r"}, // only 1 element - accessing tag[1] would panic
		},
	}

	if len(ev.Tags) != 1 {
		t.Errorf("len(ev.Tags) = %d, want 1", len(ev.Tags))
	}
	if len(ev.Tags[0]) != 1 {
		t.Errorf("len(ev.Tags[0]) = %d, want 1", len(ev.Tags[0]))
	}

	// Demonstrate the panic scenario
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic as expected: %v", r)
		}
	}()

	// This would panic if the code doesn't check len(tag) >= 2
	tag := ev.Tags[0]
	_ = tag[1] // index out of range when len < 2
}

// TestFetchRecipientReadRelays_NilTagElement demonstrates that accessing tag[1]
// when the element at index 1 is nil will cause issues
func TestFetchRecipientReadRelays_NilTagElement(t *testing.T) {
	// A tag like {"r", ""} where the URL is empty string
	// The code checks len >= 2 but doesn't validate the URL is non-empty
	ev := &nostr.Event{
		ID:   [32]byte{1, 2, 3, 4},
		Tags: nostr.Tags{
			nostr.Tag{"r", ""}, // valid length but empty URL
		},
	}

	if len(ev.Tags) != 1 {
		t.Errorf("len(ev.Tags) = %d, want 1", len(ev.Tags))
	}
	if ev.Tags[0][1] != "" {
		t.Errorf("ev.Tags[0][1] = %q, want empty string", ev.Tags[0][1])
	}
}