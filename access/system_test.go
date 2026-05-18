package access

import (
	"os"
	"testing"

	"fiatjaf.com/nostr/sdk/kvstore"
	bboltKv "fiatjaf.com/nostr/sdk/kvstore/bbolt"
)

func newTempKVStore(t *testing.T) (kvstore.KVStore, func()) {
	t.Helper()
	path, err := os.CreateTemp("", "nosmec-test-kv-*.db")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	name := path.Name()
	path.Close()
	os.Remove(name) // bbolt.Open creates the file

	store, err := bboltKv.NewStore(name)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return store, func() {
		store.Close()
		os.Remove(name)
	}
}

func TestSystem_TrackEventRelay_FirstWriteWins(t *testing.T) {
	store, cleanup := newTempKVStore(t)
	defer cleanup()

	sys := NewSystem(nil, nil, store, nil, nil, nil, nil, nil, "")

	const eventID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	const relay1 = "wss://first.example.com"
	const relay2 = "wss://second.example.com"

	// First write should succeed
	if err := sys.TrackEventRelay(eventID, relay1); err != nil {
		t.Fatalf("TrackEventRelay first: %v", err)
	}

	// Second write should be ignored (first-write-wins)
	if err := sys.TrackEventRelay(eventID, relay2); err != nil {
		t.Fatalf("TrackEventRelay second: %v", err)
	}

	got := sys.GetEventRelay(eventID)
	if got != relay1 {
		t.Errorf("GetEventRelay() = %q, want %q", got, relay1)
	}
}

func TestSystem_GetEventRelay_NotFound(t *testing.T) {
	store, cleanup := newTempKVStore(t)
	defer cleanup()

	sys := NewSystem(nil, nil, store, nil, nil, nil, nil, nil, "")

	got := sys.GetEventRelay("nonexistent")
	if got != "" {
		t.Errorf("GetEventRelay() = %q, want empty", got)
	}
}

func TestSystem_TrackEventRelay_NilStore(t *testing.T) {
	sys := NewSystem(nil, nil, nil, nil, nil, nil, nil, nil, "")
	// Should not panic
	if err := sys.TrackEventRelay("id", "wss://example.com"); err != nil {
		t.Errorf("TrackEventRelay with nil store: %v", err)
	}
}

func TestSystem_GetEventRelay_NilStore(t *testing.T) {
	sys := NewSystem(nil, nil, nil, nil, nil, nil, nil, nil, "")
	if got := sys.GetEventRelay("id"); got != "" {
		t.Errorf("GetEventRelay with nil store = %q, want empty", got)
	}
}

func TestSystem_Close(t *testing.T) {
	store, cleanup := newTempKVStore(t)
	defer cleanup()

	sys := NewSystem(nil, nil, store, nil, nil, nil, nil, nil, "")

	if err := sys.TrackEventRelay("id", "wss://example.com"); err != nil {
		t.Fatalf("TrackEventRelay: %v", err)
	}

	got := sys.GetEventRelay("id")
	if got != "wss://example.com" {
		t.Errorf("GetEventRelay() = %q, want %q", got, "wss://example.com")
	}

	// Close should not error
	if err := sys.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestNewSystem_RelayStreams(t *testing.T) {
	sys := NewSystem(
		nil, nil, nil,
		[]string{"read1", "read2"},
		[]string{"write1"},
		[]string{"dm1", "dm2", "dm3"},
		[]string{},
		[]string{"fallback1"},
		"ws://localhost:8989",
	)

	// ReadRelays
	if sys.ReadRelays == nil {
		t.Fatal("ReadRelays is nil")
	}
	if got := sys.ReadRelays.Next(); got != "read1" {
		t.Errorf("ReadRelays.Next() = %q, want %q", got, "read1")
	}

	// WriteRelays
	if got := sys.WriteRelays.Next(); got != "write1" {
		t.Errorf("WriteRelays.Next() = %q, want %q", got, "write1")
	}

	// DMRelays
	if got := sys.DMRelays.Next(); got != "dm1" {
		t.Errorf("DMRelays.Next() = %q, want %q", got, "dm1")
	}

	// Empty SearchRelays
	if got := sys.SearchRelays.Next(); got != "" {
		t.Errorf("SearchRelays.Next() = %q, want empty", got)
	}

	// FallbackRelays
	if got := sys.FallbackRelays.Next(); got != "fallback1" {
		t.Errorf("FallbackRelays.Next() = %q, want %q", got, "fallback1")
	}
}
