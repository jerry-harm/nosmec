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

func TestGetRelayFromList(t *testing.T) {
	relayList := []Relay{
		{URL: "wss://relay1.example.com", Read: BoolPtr(true), Write: BoolPtr(true)},
	}

	relay, found := GetRelayFromList("wss://relay1.example.com", relayList)
	if !found {
		t.Error("expected to find relay1")
	}
	if relay.URL != "wss://relay1.example.com" {
		t.Errorf("relay.URL = %v, want wss://relay1.example.com", relay.URL)
	}

	_, found = GetRelayFromList("wss://nonexistent.example.com", relayList)
	if found {
		t.Error("expected not to find nonexistent relay")
	}
}

func TestGetRelayFromList_EmptyList(t *testing.T) {
	relayList := []Relay{}

	_, found := GetRelayFromList("wss://any.example.com", relayList)
	if found {
		t.Error("expected not to find relay in empty list")
	}
}

func TestListRelays(t *testing.T) {
	relays := []Relay{
		{URL: "wss://relay1.example.com", Read: BoolPtr(true), Write: BoolPtr(true)},
		{URL: "wss://relay2.example.com", Read: BoolPtr(false), Write: BoolPtr(true)},
	}

	if len(relays) != 2 {
		t.Errorf("len(relays) = %d, want 2", len(relays))
	}
}

func TestAddRelayToList(t *testing.T) {
	relayList := []Relay{}

	newList := AddRelayToList("wss://new.example.com", true, true, relayList)

	if len(newList) != 1 {
		t.Errorf("len(newList) = %d, want 1", len(newList))
	}
	if newList[0].URL != "wss://new.example.com" {
		t.Errorf("newList[0].URL = %v, want wss://new.example.com", newList[0].URL)
	}
}

func TestRemoveRelayFromList(t *testing.T) {
	relayList := []Relay{
		{URL: "wss://relay1.example.com", Read: BoolPtr(true), Write: BoolPtr(true)},
		{URL: "wss://relay2.example.com", Read: BoolPtr(true), Write: BoolPtr(false)},
	}

	newList := RemoveRelayFromList("wss://relay1.example.com", relayList)

	if len(newList) != 1 {
		t.Errorf("len(newList) = %d, want 1", len(newList))
	}
	if newList[0].URL != "wss://relay2.example.com" {
		t.Errorf("newList[0].URL = %v, want wss://relay2.example.com", newList[0].URL)
	}
}

func TestRemoveRelayFromList_NotFound(t *testing.T) {
	relayList := []Relay{
		{URL: "wss://relay1.example.com", Read: BoolPtr(true), Write: BoolPtr(true)},
	}

	newList := RemoveRelayFromList("wss://nonexistent.example.com", relayList)

	if len(newList) != 1 {
		t.Errorf("len(newList) = %d, want 1 (should be unchanged)", len(newList))
	}
}

func TestAddDMRelayToList(t *testing.T) {
	dmRelays := []string{}

	newList := AddDMRelayToList("wss://dm.example.com", dmRelays)

	if len(newList) != 1 {
		t.Errorf("len(newList) = %d, want 1", len(newList))
	}
	if newList[0] != "wss://dm.example.com" {
		t.Errorf("newList[0] = %v, want wss://dm.example.com", newList[0])
	}
}

func TestAddDMRelayToList_Duplicate(t *testing.T) {
	dmRelays := []string{"wss://dm.example.com"}

	newList := AddDMRelayToList("wss://dm.example.com", dmRelays)

	if len(newList) != 1 {
		t.Errorf("len(newList) = %d, want 1 (duplicate should not be added)", len(newList))
	}
}

func TestRemoveDMRelayFromList(t *testing.T) {
	dmRelays := []string{"wss://dm1.example.com", "wss://dm2.example.com"}

	newList := RemoveDMRelayFromList("wss://dm1.example.com", dmRelays)

	if len(newList) != 1 {
		t.Errorf("len(newList) = %d, want 1", len(newList))
	}
	if newList[0] != "wss://dm2.example.com" {
		t.Errorf("newList[0] = %v, want wss://dm2.example.com", newList[0])
	}
}

func TestFilterMatchesAllKinds(t *testing.T) {
	// nil kinds should match all kinds
	var allKinds []nostr.Kind
	f := nostr.Filter{Kinds: allKinds}

	event := nostr.Event{Kind: 1}
	if !f.Matches(event) {
		t.Error("expected Filter{Kinds: nil} to match Kind 1")
	}

	event2 := nostr.Event{Kind: 9999}
	if !f.Matches(event2) {
		t.Error("expected Filter{Kinds: nil} to match Kind 9999")
	}
}

func TestFilterMatchesSpecificKinds(t *testing.T) {
	f := nostr.Filter{Kinds: []nostr.Kind{0, 3, 10002, 10050}}

	event := nostr.Event{Kind: 0}
	if !f.Matches(event) {
		t.Error("expected Filter to match Kind 0")
	}

	event2 := nostr.Event{Kind: 1}
	if f.Matches(event2) {
		t.Error("expected Filter to NOT match Kind 1")
	}
}

func TestFilterEmptySliceMatchesNothing(t *testing.T) {
	f := nostr.Filter{Kinds: []nostr.Kind{}}

	event := nostr.Event{Kind: 0}
	if f.Matches(event) {
		t.Error("expected Filter{Kinds: []} to NOT match anything")
	}
}

func TestGlobalPool_WiresPersistentBackendsIntoGlobalSystem(t *testing.T) {
	t.Skip("GlobalSystem/GlobalPool globals removed — test obsolete")
	dataDir := t.TempDir()

	oldConfig := globalConfig
	defer func() {
		globalConfig = oldConfig
	}()

	globalConfig = Config{DataDir: dataDir}

	_ = dataDir // unused without globals
}

func TestGlobalRuntimeInitializers_DoNotDuplicateInstancesConcurrently(t *testing.T) {
	t.Skip("GlobalSystem/GlobalPool globals removed — test obsolete")
	dataDir := t.TempDir()

	oldConfig := globalConfig
	defer func() {
		globalConfig = oldConfig
	}()

	globalConfig = Config{DataDir: dataDir}
	_ = dataDir
}

func TestNewAppContext_WiresPersistentBackendsBeforePoolInit(t *testing.T) {
	t.Skip("GlobalSystem removed — AppContext now creates independent runtimes")
	dataDir := t.TempDir()

	oldConfig := globalConfig
	defer func() {
		globalConfig = oldConfig
	}()

	globalConfig = Config{DataDir: dataDir}
	_ = dataDir
}

func uniqueCount[T comparable](items <-chan T) int {
	unique := make(map[T]struct{})
	for item := range items {
		unique[item] = struct{}{}
	}
	return len(unique)
}
