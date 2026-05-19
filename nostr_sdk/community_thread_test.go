package nostr_sdk

import (
	"context"
	"testing"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/eventstore/slicestore"
	"github.com/stretchr/testify/require"
)

func newThreadTestSystem(t *testing.T) *System {
	t.Helper()
	sys := NewSystem()
	store := &slicestore.SliceStore{}
	require.NoError(t, store.Init())
	sys.Store = store
	return sys
}

func mustID(t *testing.T, hex string) nostr.ID {
	t.Helper()
	id, err := nostr.IDFromHex(hex)
	require.NoError(t, err)
	return id
}

func TestFetchEventByIDInScope_UsesLocalStore(t *testing.T) {
	sys := newThreadTestSystem(t)
	scope := "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:cats"
	event := nostr.Event{
		ID:   mustID(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		Kind: nostr.KindComment,
		Tags: nostr.Tags{{"A", scope}},
	}
	require.NoError(t, sys.Store.SaveEvent(event))

	got, _, err := sys.FetchEventByIDInScope(context.Background(), event.ID, nil, scope)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, event.ID, got.ID)
}

func TestFetchParentChainInScope_StopsOnScopeMismatch(t *testing.T) {
	sys := newThreadTestSystem(t)
	scope := "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:cats"
	otherScope := "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:dogs"

	root := nostr.Event{
		ID:   mustID(t, "1111111111111111111111111111111111111111111111111111111111111111"),
		Kind: nostr.KindComment,
		Tags: nostr.Tags{{"A", scope}},
	}
	parent := nostr.Event{
		ID:   mustID(t, "2222222222222222222222222222222222222222222222222222222222222222"),
		Kind: nostr.KindComment,
		Tags: nostr.Tags{{"A", otherScope}, {"e", root.ID.Hex(), "", "root"}},
	}
	child := nostr.Event{
		ID:   mustID(t, "3333333333333333333333333333333333333333333333333333333333333333"),
		Kind: nostr.KindComment,
		Tags: nostr.Tags{{"A", scope}, {"e", root.ID.Hex(), "", "root"}, {"e", parent.ID.Hex(), "", "reply"}},
	}
	require.NoError(t, sys.Store.SaveEvent(root))
	require.NoError(t, sys.Store.SaveEvent(parent))
	require.NoError(t, sys.Store.SaveEvent(child))

	chain := sys.FetchParentChainInScope(context.Background(), &child, scope, 0, 10)
	require.Empty(t, chain)
}

func TestFetchRepliesBreadthFirstInScope_UsesLocalStoreAndFiltersScope(t *testing.T) {
	sys := newThreadTestSystem(t)
	scope := "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:cats"
	otherScope := "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:dogs"
	root := nostr.Event{
		ID:   mustID(t, "4444444444444444444444444444444444444444444444444444444444444444"),
		Kind: nostr.KindComment,
		Tags: nostr.Tags{{"A", scope}},
	}
	reply := nostr.Event{
		ID:   mustID(t, "5555555555555555555555555555555555555555555555555555555555555555"),
		Kind: nostr.KindComment,
		Tags: nostr.Tags{{"A", scope}, {"e", root.ID.Hex(), "", "root"}},
	}
	otherReply := nostr.Event{
		ID:   mustID(t, "6666666666666666666666666666666666666666666666666666666666666666"),
		Kind: nostr.KindComment,
		Tags: nostr.Tags{{"A", otherScope}, {"e", root.ID.Hex(), "", "root"}},
	}
	nested := nostr.Event{
		ID:   mustID(t, "7777777777777777777777777777777777777777777777777777777777777777"),
		Kind: nostr.KindComment,
		Tags: nostr.Tags{{"A", scope}, {"e", root.ID.Hex(), "", "root"}, {"e", reply.ID.Hex(), "", "reply"}},
	}
	require.NoError(t, sys.Store.SaveEvent(root))
	require.NoError(t, sys.Store.SaveEvent(reply))
	require.NoError(t, sys.Store.SaveEvent(otherReply))
	require.NoError(t, sys.Store.SaveEvent(nested))

	events := sys.FetchRepliesBreadthFirstInScope(context.Background(), root.ID, nil, scope, 10, 50)
	require.Len(t, events, 2)
	require.ElementsMatch(t, []nostr.ID{reply.ID, nested.ID}, []nostr.ID{events[0].ID, events[1].ID})
}
