package utils

import (
	"testing"

	"fiatjaf.com/nostr"
)

func TestBuildNoteFilter(t *testing.T) {
	tests := []struct {
		name    string
		noteID  string
		wantErr bool
	}{
		{
			name:    "valid 64-char hex",
			noteID:  "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456",
			wantErr: false,
		},
		{
			name:    "empty string",
			noteID:  "",
			wantErr: true,
		},
		{
			name:    "too short",
			noteID:  "abc123",
			wantErr: true,
		},
		{
			name:    "invalid hex chars",
			noteID:  "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := BuildNoteFilter(tt.noteID)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildNoteFilter(%q) error = %v, wantErr %v", tt.noteID, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(filter.IDs) != 1 {
					t.Errorf("filter.IDs = %v, want 1 ID", filter.IDs)
				}
				if filter.Limit != 1 {
					t.Errorf("filter.Limit = %d, want 1", filter.Limit)
				}
			}
		})
	}
}

func TestBuildNoteFilter_ID(t *testing.T) {
	noteID := "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456"
	filter, err := BuildNoteFilter(noteID)
	if err != nil {
		t.Fatalf("BuildNoteFilter failed: %v", err)
	}
	if len(filter.IDs) != 1 {
		t.Fatalf("len(filter.IDs) = %d, want 1", len(filter.IDs))
	}
	gotHex := filter.IDs[0].Hex()
	if gotHex != noteID {
		t.Errorf("filter.IDs[0].Hex() = %q, want %q", gotHex, noteID)
	}
}

func pubKeyFromHex(t *testing.T, hex string) nostr.PubKey {
	pk, err := nostr.PubKeyFromHex(hex)
	if err != nil {
		t.Fatalf("nostr.PubKeyFromHex(%q) failed: %v", hex, err)
	}
	return pk
}

func TestBuildProfileFilter(t *testing.T) {
	pubKey := pubKeyFromHex(t, "6cf684fc1d77b652a90cb62767c5342f8efd1eb4bf27edc5c15ae1df3b71204d")

	filter := BuildProfileFilter(pubKey)

	if len(filter.Kinds) != 1 || filter.Kinds[0] != nostr.KindProfileMetadata {
		t.Errorf("filter.Kinds = %v, want [KindProfileMetadata]", filter.Kinds)
	}
	if len(filter.Authors) != 1 {
		t.Errorf("len(filter.Authors) = %d, want 1", len(filter.Authors))
	}
	if filter.Limit != 1 {
		t.Errorf("filter.Limit = %d, want 1", filter.Limit)
	}
}

func TestBuildProfilesFilter(t *testing.T) {
	pubKeys := []nostr.PubKey{
		pubKeyFromHex(t, "6cf684fc1d77b652a90cb62767c5342f8efd1eb4bf27edc5c15ae1df3b71204d"),
		pubKeyFromHex(t, "6cf684fc1d77b652a90cb62767c5342f8efd1eb4bf27edc5c15ae1df3b71204e"),
	}

	filter := BuildProfilesFilter(pubKeys)

	if len(filter.Kinds) != 1 || filter.Kinds[0] != nostr.KindProfileMetadata {
		t.Errorf("filter.Kinds = %v, want [KindProfileMetadata]", filter.Kinds)
	}
	if len(filter.Authors) != 2 {
		t.Errorf("len(filter.Authors) = %d, want 2", len(filter.Authors))
	}
	if filter.Limit != 1 {
		t.Errorf("filter.Limit = %d, want 1", filter.Limit)
	}
}

func TestBuildTimelineFilter(t *testing.T) {
	pubKey := pubKeyFromHex(t, "6cf684fc1d77b652a90cb62767c5342f8efd1eb4bf27edc5c15ae1df3b71204d")
	filter := BuildTimelineFilter(pubKey, 20, 0)

	if len(filter.Kinds) != 1 || filter.Kinds[0] != nostr.KindTextNote {
		t.Errorf("filter.Kinds = %v, want [KindTextNote]", filter.Kinds)
	}
	if len(filter.Authors) != 1 || filter.Authors[0] != pubKey {
		t.Errorf("filter.Authors = %v, want [%v]", filter.Authors, pubKey)
	}
	if filter.Limit != 20 {
		t.Errorf("filter.Limit = %d, want 20", filter.Limit)
	}
	if filter.Until != 0 {
		t.Errorf("filter.Until = %d, want 0", filter.Until)
	}
}

func TestBuildTimelineFilter_WithUntil(t *testing.T) {
	pubKey := pubKeyFromHex(t, "6cf684fc1d77b652a90cb62767c5342f8efd1eb4bf27edc5c15ae1df3b71204d")
	until := nostr.Timestamp(1700000000)
	filter := BuildTimelineFilter(pubKey, 20, until)

	if filter.Until != until {
		t.Errorf("filter.Until = %d, want %d", filter.Until, until)
	}
}

func TestBuildGlobalTimelineFilter(t *testing.T) {
	filter := BuildGlobalTimelineFilter(50, 0)

	if len(filter.Kinds) != 1 || filter.Kinds[0] != nostr.KindTextNote {
		t.Errorf("filter.Kinds = %v, want [KindTextNote]", filter.Kinds)
	}
	if len(filter.Authors) != 0 {
		t.Errorf("len(filter.Authors) = %d, want 0", len(filter.Authors))
	}
	if filter.Limit != 50 {
		t.Errorf("filter.Limit = %d, want 50", filter.Limit)
	}
}

func TestBuildFollowedTimelineFilter(t *testing.T) {
	pubKeys := []nostr.PubKey{
		pubKeyFromHex(t, "6cf684fc1d77b652a90cb62767c5342f8efd1eb4bf27edc5c15ae1df3b71204d"),
	}
	communityAddrs := []string{"345yyd5ksfk6qklh19pp13zrp9z7vy2myk8qww4hlytnmksgcyx5lqy2hxzqq"}

	filter := BuildFollowedTimelineFilter(pubKeys, communityAddrs, nil, 30, 0)

	if len(filter.Kinds) != 2 {
		t.Errorf("len(filter.Kinds) = %d, want 2", len(filter.Kinds))
	}
	if filter.Kinds[0] != nostr.KindTextNote || filter.Kinds[1] != nostr.KindComment {
		t.Errorf("filter.Kinds = %v, want [KindTextNote, KindComment]", filter.Kinds)
	}
	if len(filter.Authors) != 1 {
		t.Errorf("len(filter.Authors) = %d, want 1", len(filter.Authors))
	}
	if filter.Tags == nil {
		t.Errorf("filter.Tags is nil, want non-nil")
	}
	if len(filter.Tags["a"]) != 1 || filter.Tags["a"][0] != communityAddrs[0] {
		t.Errorf("filter.Tags['a'] = %v, want %v", filter.Tags["a"], communityAddrs)
	}
	if filter.Limit != 90 {
		t.Errorf("filter.Limit = %d, want 90 (limit*3)", filter.Limit)
	}
}

func TestBuildFollowedTimelineFilter_NoAuthors(t *testing.T) {
	filter := BuildFollowedTimelineFilter(nil, nil, nil, 10, 0)

	if len(filter.Kinds) != 2 {
		t.Errorf("len(filter.Kinds) = %d, want 2", len(filter.Kinds))
	}
	if filter.Authors != nil {
		t.Errorf("filter.Authors = %v, want nil", filter.Authors)
	}
	if filter.Tags != nil {
		t.Errorf("filter.Tags = %v, want nil", filter.Tags)
	}
}

func TestBuildParentEventFilter(t *testing.T) {
	parentID := "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456"
	filter, err := BuildParentEventFilter(parentID)
	if err != nil {
		t.Fatalf("BuildParentEventFilter failed: %v", err)
	}
	if len(filter.IDs) != 1 {
		t.Errorf("len(filter.IDs) = %d, want 1", len(filter.IDs))
	}
	if filter.Limit != 1 {
		t.Errorf("filter.Limit = %d, want 1", filter.Limit)
	}
}

func TestBuildParentEventFilter_InvalidID(t *testing.T) {
	_, err := BuildParentEventFilter("invalid")
	if err == nil {
		t.Errorf("BuildParentEventFilter(%q) error = nil, wantErr true", "invalid")
	}
}