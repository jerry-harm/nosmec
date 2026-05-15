package utils

import (
	"testing"
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