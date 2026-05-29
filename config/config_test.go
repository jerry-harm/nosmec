package config

import (
	"testing"

	"fiatjaf.com/nostr"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(
		m,
		goleak.IgnoreTopFunction("github.com/blevesearch/bleve_index_api.AnalysisWorker"),
		goleak.IgnoreTopFunction("github.com/blevesearch/bleve/v2/index/scorch.(*Scorch).introducerLoop"),
		goleak.IgnoreTopFunction("github.com/blevesearch/bleve/v2/index/scorch.(*Scorch).persisterLoop"),
		goleak.IgnoreTopFunction("github.com/blevesearch/bleve/v2/index/scorch.(*Scorch).mergerLoop"),
		goleak.IgnoreTopFunction("github.com/dgraph-io/ristretto/v2.(*defaultPolicy[...]).processItems"),
		goleak.IgnoreTopFunction("github.com/dgraph-io/ristretto/v2.(*Cache[...]).processItems"),
		goleak.IgnoreAnyFunction("fiatjaf.com/nostr.(*Pool).startPenaltyBox.func1"),
	)
}

func TestGetRelay(t *testing.T) {
	relayList := []Relay{
		{URL: "wss://relay1.example.com", Read: BoolPtr(true), Write: BoolPtr(true)},
		{URL: "wss://relay2.example.com", Read: BoolPtr(true), Write: BoolPtr(false)},
	}

	tests := []struct {
		name      string
		url       string
		wantFound bool
		wantURL   string
	}{
		{name: "found relay1", url: "wss://relay1.example.com", wantFound: true, wantURL: "wss://relay1.example.com"},
		{name: "found relay2", url: "wss://relay2.example.com", wantFound: true, wantURL: "wss://relay2.example.com"},
		{name: "not found", url: "wss://nonexistent.example.com", wantFound: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			relay, found := GetRelayFromList(tt.url, relayList)
			if found != tt.wantFound {
				t.Errorf("GetRelay(%q) found = %v, want %v", tt.url, found, tt.wantFound)
			}
			if found && relay.URL != tt.wantURL {
				t.Errorf("GetRelay(%q) relay.URL = %v, want %v", tt.url, relay.URL, tt.wantURL)
			}
		})
	}
}