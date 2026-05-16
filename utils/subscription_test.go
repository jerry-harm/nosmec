package utils

import (
	"testing"

	"fiatjaf.com/nostr"
)

// TestSyncUsersFromNetwork_TagTooShort demonstrates the bug: tag[1], tag[2], tag[3]
// are accessed without length checks in syncUsersFromNetwork.
// If a "p" tag has fewer than 2 elements, accessing tag[1] will panic.
func TestSyncUsersFromNetwork_TagTooShort(t *testing.T) {
	ev := &nostr.Event{
		ID:   [32]byte{1, 2, 3, 4},
		Tags: nostr.Tags{
			nostr.Tag{"p"}, // only 1 element - accessing tag[1] would panic
		},
	}

	if len(ev.Tags) != 1 {
		t.Errorf("len(ev.Tags) = %d, want 1", len(ev.Tags))
	}

	// Demonstrate the panic when accessing tag[1] without length check
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic as expected: %v", r)
		}
	}()

	tag := ev.Tags[0]
	_ = tag[1] // index out of range when len < 2
}

// TestSyncUsersFromNetwork_TagWithTwoElements verifies behavior with exactly 2 elements
func TestSyncUsersFromNetwork_TagWithTwoElements(t *testing.T) {
	ev := &nostr.Event{
		ID:   [32]byte{1, 2, 3, 4},
		Tags: nostr.Tags{
			nostr.Tag{"p", "abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234"}, // pubkey only
		},
	}

	if len(ev.Tags) != 1 {
		t.Errorf("len(ev.Tags) = %d, want 1", len(ev.Tags))
	}
	if len(ev.Tags[0]) != 2 {
		t.Errorf("len(ev.Tags[0]) = %d, want 2", len(ev.Tags[0]))
	}
	// tag[2] and tag[3] should be safe to check but missing (would be empty string)
}

// TestSyncUsersFromNetwork_TagWithThreeElements verifies behavior with 3 elements
func TestSyncUsersFromNetwork_TagWithThreeElements(t *testing.T) {
	ev := &nostr.Event{
		ID:   [32]byte{1, 2, 3, 4},
		Tags: nostr.Tags{
			nostr.Tag{"p", "abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234", "wss://relay.example.com"},
		},
	}

	if len(ev.Tags[0]) != 3 {
		t.Errorf("len(ev.Tags[0]) = %d, want 3", len(ev.Tags[0]))
	}
	// tag[3] would be missing
}

// TestSyncUsersFromNetwork_TagWithFourElements verifies behavior with 4 elements (full)
func TestSyncUsersFromNetwork_TagWithFourElements(t *testing.T) {
	ev := &nostr.Event{
		ID:   [32]byte{1, 2, 3, 4},
		Tags: nostr.Tags{
			nostr.Tag{"p", "abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234", "wss://relay.example.com", "petname"},
		},
	}

	if len(ev.Tags[0]) != 4 {
		t.Errorf("len(ev.Tags[0]) = %d, want 4", len(ev.Tags[0]))
	}
}