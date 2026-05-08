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
			name: "nil filter matches all",
			filter: RelayFilter{},
			relay: Relay{
				URL:   "wss://example.com",
				Read:  BoolPtr(true),
				Write: BoolPtr(true),
			},
			shouldMatch: true,
		},
		{
			name: "filter read=true matches relay read=true",
			filter: RelayFilter{Read: BoolPtr(true)},
			relay: Relay{
				URL:   "wss://example.com",
				Read:  BoolPtr(true),
				Write: BoolPtr(false),
			},
			shouldMatch: true,
		},
		{
			name: "filter read=true does not match relay read=false",
			filter: RelayFilter{Read: BoolPtr(true)},
			relay: Relay{
				URL:   "wss://example.com",
				Read:  BoolPtr(false),
				Write: BoolPtr(true),
			},
			shouldMatch: false,
		},
		{
			name: "filter write=true matches relay write=true",
			filter: RelayFilter{Write: BoolPtr(true)},
			relay: Relay{
				URL:   "wss://example.com",
				Read:  BoolPtr(true),
				Write: BoolPtr(true),
			},
			shouldMatch: true,
		},
		{
			name: "filter write=false does not match relay write=true",
			filter: RelayFilter{Write: BoolPtr(false)},
			relay: Relay{
				URL:   "wss://example.com",
				Read:  BoolPtr(true),
				Write: BoolPtr(true),
			},
			shouldMatch: false,
		},
		{
			name: "both read and write filter",
			filter: RelayFilter{Read: BoolPtr(true), Write: BoolPtr(true)},
			relay: Relay{
				URL:   "wss://example.com",
				Read:  BoolPtr(true),
				Write: BoolPtr(true),
			},
			shouldMatch: true,
		},
		{
			name: "both filters but relay write=false",
			filter: RelayFilter{Read: BoolPtr(true), Write: BoolPtr(true)},
			relay: Relay{
				URL:   "wss://example.com",
				Read:  BoolPtr(true),
				Write: BoolPtr(false),
			},
			shouldMatch: false,
		},
		{
			name: "relay with nil read defaults to true",
			filter: RelayFilter{Read: BoolPtr(false)},
			relay: Relay{
				URL:   "wss://example.com",
				Read:  nil,
				Write: BoolPtr(true),
			},
			shouldMatch: false,
		},
		{
			name: "relay with nil write defaults to false",
			filter: RelayFilter{Write: BoolPtr(true)},
			relay: Relay{
				URL:   "wss://example.com",
				Read:  BoolPtr(true),
				Write: nil,
			},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.matchRelay(tt.relay)
			if got != tt.shouldMatch {
				t.Errorf("RelayFilter{Read:%v, Write:%v}.matchRelay(%+v) = %v, want %v",
					tt.filter.Read, tt.filter.Write, tt.relay, got, tt.shouldMatch)
			}
		})
	}
}

func TestRelayFilter_Matches(t *testing.T) {
	relays := []Relay{
		{URL: "wss://read1.com", Read: BoolPtr(true), Write: BoolPtr(false)},
		{URL: "wss://write1.com", Read: BoolPtr(false), Write: BoolPtr(true)},
		{URL: "wss://readwrite.com", Read: BoolPtr(true), Write: BoolPtr(true)},
		{URL: "wss://none.com", Read: BoolPtr(false), Write: BoolPtr(false)},
	}

	tests := []struct {
		name           string
		filter         RelayFilter
		expectedURLs   []string
	}{
		{
			name:   "get writable relays",
			filter: RelayFilter{Write: BoolPtr(true)},
			expectedURLs: []string{"wss://write1.com", "wss://readwrite.com"},
		},
		{
			name:   "get readable relays",
			filter: RelayFilter{Read: BoolPtr(true)},
			expectedURLs: []string{"wss://read1.com", "wss://readwrite.com"},
		},
		{
			name:   "get read-write relays",
			filter: RelayFilter{Read: BoolPtr(true), Write: BoolPtr(true)},
			expectedURLs: []string{"wss://readwrite.com"},
		},
		{
			name:   "no matches",
			filter: RelayFilter{Read: BoolPtr(false), Write: BoolPtr(false)},
			expectedURLs: []string{"wss://none.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.Matches(relays)
			if len(got) != len(tt.expectedURLs) {
				t.Errorf("Matches() returned %v, want %v", got, tt.expectedURLs)
				return
			}
			for i, url := range got {
				if url != tt.expectedURLs[i] {
					t.Errorf("Matches()[%d] = %v, want %v", i, url, tt.expectedURLs[i])
				}
			}
		})
	}
}

func TestBoolPtr(t *testing.T) {
	tests := []struct {
		name     string
		input    bool
		expected bool
	}{
		{name: "true", input: true, expected: true},
		{name: "false", input: false, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BoolPtr(tt.input)
			if *got != tt.expected {
				t.Errorf("BoolPtr(%v) = %v, want %v", tt.input, *got, tt.expected)
			}
		})
	}
}

func TestGetWritableRelaysFromList(t *testing.T) {
	relays := []Relay{
		{URL: "wss://rw.com", Read: BoolPtr(true), Write: BoolPtr(true)},
		{URL: "wss://ro.com", Read: BoolPtr(true), Write: BoolPtr(false)},
		{URL: "wss://wo.com", Read: BoolPtr(false), Write: BoolPtr(true)},
	}

	got := GetWritableRelaysFromList(relays)
	if len(got) != 2 {
		t.Errorf("GetWritableRelaysFromList() returned %d, want 2", len(got))
	}
}

func TestGetReadableRelaysFromList(t *testing.T) {
	relays := []Relay{
		{URL: "wss://rw.com", Read: BoolPtr(true), Write: BoolPtr(true)},
		{URL: "wss://ro.com", Read: BoolPtr(true), Write: BoolPtr(false)},
		{URL: "wss://wo.com", Read: BoolPtr(false), Write: BoolPtr(true)},
	}

	got := GetReadableRelaysFromList(relays)
	if len(got) != 2 {
		t.Errorf("GetReadableRelaysFromList() returned %d, want 2", len(got))
	}
}
