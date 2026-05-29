package config

import (
	"testing"
)

func TestRelayFilter_matchRelay(t *testing.T) {
	tests := []struct {
		name        string
		filter      RelayFilter
		relay       Relay
		shouldMatch bool
	}{
		{
			name:        "nil filter matches all",
			filter:      RelayFilter{},
			relay:       Relay{URL: "wss://example.com", Read: BoolPtr(true), Write: BoolPtr(true)},
			shouldMatch: true,
		},
		{
			name:        "filter read=true matches relay read=true",
			filter:      RelayFilter{Read: BoolPtr(true)},
			relay:       Relay{URL: "wss://example.com", Read: BoolPtr(true), Write: BoolPtr(false)},
			shouldMatch: true,
		},
		{
			name:        "filter read=true does not match relay read=false",
			filter:      RelayFilter{Read: BoolPtr(true)},
			relay:       Relay{URL: "wss://example.com", Read: BoolPtr(false), Write: BoolPtr(true)},
			shouldMatch: false,
		},
		{
			name:        "filter write=true matches relay write=true",
			filter:      RelayFilter{Write: BoolPtr(true)},
			relay:       Relay{URL: "wss://example.com", Read: BoolPtr(true), Write: BoolPtr(true)},
			shouldMatch: true,
		},
		{
			name:        "filter write=true does not match relay write=false",
			filter:      RelayFilter{Write: BoolPtr(true)},
			relay:       Relay{URL: "wss://example.com", Read: BoolPtr(true), Write: BoolPtr(false)},
			shouldMatch: false,
		},
		{
			name:        "both read and write filter matches",
			filter:      RelayFilter{Read: BoolPtr(true), Write: BoolPtr(true)},
			relay:       Relay{URL: "wss://example.com", Read: BoolPtr(true), Write: BoolPtr(true)},
			shouldMatch: true,
		},
		{
			name:        "mismatched read and write filter",
			filter:      RelayFilter{Read: BoolPtr(true), Write: BoolPtr(true)},
			relay:       Relay{URL: "wss://example.com", Read: BoolPtr(false), Write: BoolPtr(true)},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.matchRelay(tt.relay)
			if got != tt.shouldMatch {
				t.Errorf("RelayFilter.matchRelay() = %v, want %v", got, tt.shouldMatch)
			}
		})
	}
}

func TestGetWritableRelaysFromList(t *testing.T) {
	relayList := []Relay{
		{URL: "wss://relay1.example.com", Read: BoolPtr(true), Write: BoolPtr(true)},
		{URL: "wss://relay2.example.com", Read: BoolPtr(true), Write: BoolPtr(false)},
		{URL: "wss://relay3.example.com", Read: BoolPtr(false), Write: BoolPtr(true)},
	}

	got := GetWritableRelaysFromList(relayList)
	if len(got) != 2 {
		t.Errorf("GetWritableRelaysFromList() returned %d relays, want 2", len(got))
	}
}

func TestGetReadableRelaysFromList(t *testing.T) {
	relayList := []Relay{
		{URL: "wss://relay1.example.com", Read: BoolPtr(true), Write: BoolPtr(true)},
		{URL: "wss://relay2.example.com", Read: BoolPtr(true), Write: BoolPtr(false)},
		{URL: "wss://relay3.example.com", Read: BoolPtr(false), Write: BoolPtr(true)},
	}

	got := GetReadableRelaysFromList(relayList)
	if len(got) != 2 {
		t.Errorf("GetReadableRelaysFromList() returned %d relays, want 2", len(got))
	}
}