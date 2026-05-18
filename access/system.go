package access

import (
	"fiatjaf.com/nostr"
	sdk_hints "fiatjaf.com/nostr/sdk/hints"
	"fiatjaf.com/nostr/sdk/kvstore"
)

// GlobalSystem holds the singleton System instance, set early during
// initialization so that event→relay tracking works through EventMiddleware
// before AppContext is fully constructed.
var GlobalSystem *System

// System is the central access layer, inspired by nostr/sdk's System module.
// It holds the Pool, HintsDB, KVStore, and per-category RelayStreams.
//
// Embed in AppContext for progressive migration — existing callers continue
// to work through AppContext methods while new code can use System directly.
type System struct {
	Pool  *nostr.Pool
	Hints sdk_hints.HintsDB
	Store kvstore.KVStore // event→relay persistence

	ReadRelays     *RelayStream
	WriteRelays    *RelayStream
	DMRelays       *RelayStream
	SearchRelays   *RelayStream
	FallbackRelays *RelayStream

	localRelayURL string
}

// NewSystem creates a System with the given dependencies and relay URL lists.
//
// Each *RelayStream is built from the corresponding slice. Streams are
// ready-to-use round-robin iterators.
func NewSystem(
	pool *nostr.Pool,
	hints sdk_hints.HintsDB,
	kvStore kvstore.KVStore,
	readRelays []string,
	writeRelays []string,
	dmRelays []string,
	searchRelays []string,
	fallbackRelays []string,
	localRelayURL string,
) *System {
	return &System{
		Pool:           pool,
		Hints:          hints,
		Store:          kvStore,
		ReadRelays:     NewRelayStream(readRelays...),
		WriteRelays:    NewRelayStream(writeRelays...),
		DMRelays:       NewRelayStream(dmRelays...),
		SearchRelays:   NewRelayStream(searchRelays...),
		FallbackRelays: NewRelayStream(fallbackRelays...),
		localRelayURL:  localRelayURL,
	}
}

// TrackEventRelay persists an event→relay association.
// First-write-wins: if the event already has a relay recorded, the new one
// is silently ignored. Uses atomic Update to avoid races.
func (s *System) TrackEventRelay(eventID, relayURL string) error {
	if s.Store == nil {
		return nil
	}
	key := []byte(eventID)
	val := []byte(relayURL)
	return s.Store.Update(key, func(existing []byte) ([]byte, error) {
		if existing != nil {
			return existing, nil // first write wins
		}
		return val, nil
	})
}

// GetEventRelay returns the relay URL an event was fetched from.
// Returns "" if never tracked.
func (s *System) GetEventRelay(eventID string) string {
	if s.Store == nil {
		return ""
	}
	val, err := s.Store.Get([]byte(eventID))
	if err != nil || val == nil {
		return ""
	}
	return string(val)
}

// Close releases resources held by the System.
func (s *System) Close() error {
	if s.Store != nil {
		return s.Store.Close()
	}
	return nil
}
