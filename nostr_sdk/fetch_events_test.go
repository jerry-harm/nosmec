package nostr_sdk

import (
	"context"
	"testing"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/eventstore/slicestore"
	"github.com/stretchr/testify/require"
)

func mustPubKey(t *testing.T, hex string) nostr.PubKey {
	t.Helper()
	pk, err := nostr.PubKeyFromHexCheap(hex)
	require.NoError(t, err)
	return pk
}

func newFetchEventsTestSystem(t *testing.T) *System {
	t.Helper()
	sys := NewSystem()
	store := &slicestore.SliceStore{}
	require.NoError(t, store.Init())
	sys.Store = store
	return sys
}

func TestFetchEventsByFilter_UsesLocalStoreAndSortsNewestFirst(t *testing.T) {
	sys := newFetchEventsTestSystem(t)
	author := mustPubKey(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	first := nostr.Event{
		ID:        mustID(t, "1111111111111111111111111111111111111111111111111111111111111111"),
		PubKey:    author,
		Kind:      nostr.KindCommunityDefinition,
		CreatedAt: 20,
		Tags:      nostr.Tags{{"d", "cats"}},
	}
	second := nostr.Event{
		ID:        mustID(t, "2222222222222222222222222222222222222222222222222222222222222222"),
		PubKey:    author,
		Kind:      nostr.KindCommunityDefinition,
		CreatedAt: 30,
		Tags:      nostr.Tags{{"d", "dogs"}},
	}
	require.NoError(t, sys.Store.SaveEvent(first))
	require.NoError(t, sys.Store.SaveEvent(second))

	events, err := sys.FetchEventsByFilter(
		context.Background(),
		nostr.Filter{Kinds: []nostr.Kind{nostr.KindCommunityDefinition}},
		FetchEventsOptions{SkipNetwork: true},
	)
	require.NoError(t, err)
	require.Len(t, events, 2)
	require.Equal(t, second.ID, events[0].ID)
	require.Equal(t, first.ID, events[1].ID)
}

func TestFetchEventsByFilter_UsesFallbackRelaysWhenNoOverrideProvided(t *testing.T) {
	sys := newFetchEventsTestSystem(t)
	sys.FallbackRelays = NewRelayStream("wss://fallback-one.example", "wss://fallback-two.example")

	events, err := sys.FetchEventsByFilter(
		context.Background(),
		nostr.Filter{Kinds: []nostr.Kind{nostr.KindCommunityDefinition}},
		FetchEventsOptions{},
	)
	require.NoError(t, err)
	require.Empty(t, events)
	// red-phase assertion for API existence and default path behavior without panicking.
}

func TestDefaultRelaysForFilter_UsesRelayListRelaysForCommunityDefinitions(t *testing.T) {
	sys := newFetchEventsTestSystem(t)
	sys.RelayListRelays = NewRelayStream("wss://indexer.example", "wss://purple.example")
	sys.FallbackRelays = NewRelayStream("wss://fallback.example")

	relays := sys.defaultRelaysForFilter(context.Background(), nostr.Filter{
		Kinds: []nostr.Kind{nostr.KindCommunityDefinition},
	})

	require.Equal(t, []string{"wss://indexer.example", "wss://purple.example", "wss://fallback.example"}, relays)
}

func TestDefaultRelaysForFilter_UsesJustIDRelaysForIDQueries(t *testing.T) {
	sys := newFetchEventsTestSystem(t)
	sys.JustIDRelays = NewRelayStream("wss://id-one.example", "wss://id-two.example")
	sys.FallbackRelays = NewRelayStream("wss://fallback.example")

	relays := sys.defaultRelaysForFilter(context.Background(), nostr.Filter{
		IDs: []nostr.ID{mustID(t, "9999999999999999999999999999999999999999999999999999999999999999")},
	})

	require.Equal(t, []string{"wss://id-one.example", "wss://id-two.example", "wss://fallback.example"}, relays)
}

func TestFetchEventsByFilter_CanQueryCommunityDefinitionsFromStore(t *testing.T) {
	sys := newFetchEventsTestSystem(t)
	event := nostr.Event{
		ID:        mustID(t, "3333333333333333333333333333333333333333333333333333333333333333"),
		PubKey:    mustPubKey(t, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
		Kind:      nostr.KindCommunityDefinition,
		CreatedAt: 10,
		Tags: nostr.Tags{
			{"d", "cats"},
			{"name", "Cats"},
		},
	}
	require.NoError(t, sys.Store.SaveEvent(event))

	events, err := sys.FetchEventsByFilter(
		context.Background(),
		nostr.Filter{
			Kinds: []nostr.Kind{nostr.KindCommunityDefinition},
			Tags:  nostr.TagMap{"d": []string{"cats"}},
		},
		FetchEventsOptions{SkipNetwork: true},
	)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, event.ID, events[0].ID)
}
