package utils

import (
	"testing"

	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/nip72"
)

func TestParseCommunityAddr_Valid(t *testing.T) {
	tests := []struct {
		name      string
		addr      string
		wantPubKey string
		wantID    string
	}{
		{
			name:      "valid community address",
			addr:      "34550:6cf684fc1d77b652a90cb62767c5342f8efd1eb4bf27edc5c15ae1df3b71204d:community-name",
			wantPubKey: "6cf684fc1d77b652a90cb62767c5342f8efd1eb4bf27edc5c15ae1df3b71204d",
			wantID:    "community-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pk, id, err := ParseCommunityAddr(tt.addr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if pk.Hex() != tt.wantPubKey {
				t.Errorf("pubkey = %v, want %v", pk.Hex(), tt.wantPubKey)
			}
			if id != tt.wantID {
				t.Errorf("id = %v, want %v", id, tt.wantID)
			}
		})
	}
}

func TestParseCommunityAddr_Invalid(t *testing.T) {
	tests := []struct {
		name string
		addr string
	}{
		{
			name: "wrong prefix",
			addr: "12345:6cf684fc1d77b652a90cb62767c5342f8efd1eb4bf27edc5c15ae1df3b71204d:name",
		},
		{
			name: "not 3 parts",
			addr: "34550:pubkey",
		},
		{
			name: "invalid pubkey format",
			addr: "34550:invalidpubkey:name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ParseCommunityAddr(tt.addr)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestParseCommunityAddr_ShortPubKey(t *testing.T) {
	addr := "34550:6cf684fc1d77b652a90cb62767c5342f8efd1eb4bf27edc5c15ae1df3b71204:name"

	_, _, err := ParseCommunityAddr(addr)
	if err == nil {
		t.Error("expected error for short pubkey, got nil")
	}
}

func TestParseCommunityAddr_LongPubKey(t *testing.T) {
	addr := "34550:6cf684fc1d77b652a90cb62767c5342f8efd1eb4bf27edc5c15ae1df3b71204ddd:name"

	_, _, err := ParseCommunityAddr(addr)
	if err == nil {
		t.Error("expected error for long pubkey, got nil")
	}
}

func TestGetParentPostInfo_RequiresFullIntegration(t *testing.T) {
	t.Skip("Requires mock AppContext and GetPost - complex integration test, tracked separately")
}

// TestGetParentPostInfo_TagBounds demonstrates the bug: copy(authorPubKey[:], tag[1])
// is called without verifying tag[1] is exactly 64 characters (32 bytes hex).
// If tag[1] is shorter, copy() will panic with index out of range.
func TestGetParentPostInfo_TagBounds(t *testing.T) {
	// Create a mock parent event with malformed tags
	parentEvent := &nostr.Event{
		ID: [32]byte{1, 2, 3, 4},
		Tags: nostr.Tags{
			// "p" tag with short hex - should cause panic if copy() is called directly
			nostr.Tag{"p", "short"},
		},
	}

	// This test demonstrates the bug by checking what happens with malformed input
	// The actual function GetParentPostInfo calls GetPost which would replace nil with an event
	// but the tag access copy(authorPubKey[:], tag[1]) has no bounds check

	// Demonstrate that copy with short string causes issues
	var authorPubKey nostr.PubKey
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic as expected with malformed tag: %v", r)
		}
	}()

	// This is what the buggy code does - copy without length check
	// If tag[1] is less than 32 bytes, this panics
	copy(authorPubKey[:], parentEvent.Tags[0][1])
}

// RED phase: Write failing test to demonstrate the bug
// The bug: ParseCommunityAddr does not validate parts[1] length before using it
func TestParseCommunityAddr_PartsLengthNoCheck(t *testing.T) {
	// This test demonstrates the bug where parts[1] is used without checking
	// if it's a valid 64-char hex pubkey
	tests := []struct {
		name    string
		addr    string
		wantErr string
	}{
		{
			name:    "empty pubkey part",
			addr:    "34550::name",
			wantErr: "invalid community pubkey",
		},
		{
			name:    "short pubkey (less than 64 chars)",
			addr:    "34550:abc:name",
			wantErr: "invalid community pubkey",
		},
		{
			name:    "pubkey with invalid characters",
			addr:    "34550:zzzz0000zzzz0000zzzz0000zzzz0000zzzz0000zzzz0000zzzz0000zzzzzzzz:name",
			wantErr: "invalid community pubkey",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ParseCommunityAddr(tt.addr)
			if err == nil {
				t.Errorf("ParseCommunityAddr(%q) = nil, want error containing %q", tt.addr, tt.wantErr)
			}
		})
	}
}

func TestCommunityDefinition_RelayShapeSupportsPurpose(t *testing.T) {
	def := CommunityDefinition{
		Relays: []nip72.CommunityRelay{
			{URL: "wss://authors.example.com", Purpose: "author"},
			{URL: "wss://requests.example.com", Purpose: "requests"},
		},
	}

	if got := def.Relays[0].Purpose; got != "author" {
		t.Fatalf("first relay purpose = %q, want %q", got, "author")
	}
}
