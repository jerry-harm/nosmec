package config

import (
	"testing"
)

func TestHintsDB_Record(t *testing.T) {
	h := NewHintsDB()

	h.Record("pubkey1", "wss://relay1.example.com", HintEventFetched)
	h.Record("pubkey1", "wss://relay2.example.com", HintFromTag)
	h.Record("pubkey2", "wss://relay1.example.com", HintInRelayList)

	if h.Size() != 2 {
		t.Errorf("Size() = %d, want 2", h.Size())
	}
}

func TestHintsDB_TopN(t *testing.T) {
	h := NewHintsDB()

	// Record multiple hints for same pubkey on different relays
	h.Record("pubkey1", "wss://relay-a.example.com", HintEventFetched) // 700 pts
	h.Record("pubkey1", "wss://relay-b.example.com", HintFromTag)       // 20 pts
	h.Record("pubkey1", "wss://relay-a.example.com", HintInRelayList)  // +350 = 1050 total

	results := h.TopN("pubkey1", 5)
	if len(results) < 2 {
		t.Errorf("TopN() returned %d results, want at least 2", len(results))
	}
	// relay-a should be first (higher score)
	if len(results) >= 2 && results[0] != "wss://relay-a.example.com" {
		t.Errorf("TopN()[0] = %q, want wss://relay-a.example.com (highest score)", results[0])
	}
}

func TestHintsDB_TopN_Empty(t *testing.T) {
	h := NewHintsDB()
	results := h.TopN("nonexistent", 5)
	if len(results) != 0 {
		t.Errorf("TopN() = %v, want empty", results)
	}
}

func TestHintsDB_TopN_Limit(t *testing.T) {
	h := NewHintsDB()

	h.Record("pubkey1", "wss://r1.example.com", HintEventFetched)
	h.Record("pubkey1", "wss://r2.example.com", HintEventFetched)
	h.Record("pubkey1", "wss://r3.example.com", HintEventFetched)
	h.Record("pubkey1", "wss://r4.example.com", HintEventFetched)
	h.Record("pubkey1", "wss://r5.example.com", HintEventFetched)

	results := h.TopN("pubkey1", 3)
	if len(results) != 3 {
		t.Errorf("TopN(n=3) returned %d results, want 3", len(results))
	}
}

func TestHintsDB_Record_EmptyInput(t *testing.T) {
	h := NewHintsDB()
	h.Record("", "wss://relay.example.com", HintEventFetched)
	h.Record("pubkey1", "", HintEventFetched)

	if h.Size() != 0 {
		t.Errorf("Size() = %d, want 0 (empty inputs skipped)", h.Size())
	}
}
