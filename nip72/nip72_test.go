package nip72

import (
	"testing"

	"fiatjaf.com/nostr"
	"github.com/stretchr/testify/require"
)

func mustID(t *testing.T, hex string) nostr.ID {
	t.Helper()
	id, err := nostr.IDFromHex(hex)
	require.NoError(t, err)
	return id
}

func mustPubKey(t *testing.T, hex string) nostr.PubKey {
	t.Helper()
	pk, err := nostr.PubKeyFromHexCheap(hex)
	require.NoError(t, err)
	return pk
}

func TestGetCommunityPointer(t *testing.T) {
	event := &nostr.Event{
		Kind: nostr.KindComment,
		Tags: nostr.Tags{
			{"A", "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:cats"},
			{"a", "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:cats"},
			{"P", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			{"p", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			{"K", "34550"},
			{"k", "34550"},
		},
	}

	ptr := GetCommunityPointer(event)
	require.NotNil(t, ptr)
	ref, ok := ptr.(nostr.EntityPointer)
	require.True(t, ok)
	require.Equal(t, nostr.KindCommunityDefinition, ref.Kind)
	require.Equal(t, "cats", ref.Identifier)
	require.Equal(t, mustPubKey(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), ref.PublicKey)
}

func TestGetParentPointer(t *testing.T) {
	event := &nostr.Event{
		Kind: nostr.KindComment,
		Tags: nostr.Tags{
			{"A", "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:cats"},
			{"P", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			{"K", "34550"},
			{"e", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "wss://relay.example.com", "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"},
			{"p", "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"},
			{"k", "1111"},
		},
	}

	ptr := GetParentPointer(event)
	require.NotNil(t, ptr)
	ref, ok := ptr.(nostr.EventPointer)
	require.True(t, ok)
	require.Equal(t, mustID(t, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"), ref.ID)
	require.Equal(t, []string{"wss://relay.example.com"}, ref.Relays)
	require.Equal(t, nostr.KindComment, ref.Kind)
	require.Equal(t, mustPubKey(t, "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"), ref.Author)
}

func TestGetRootPointer(t *testing.T) {
	event := &nostr.Event{
		Kind: nostr.KindComment,
		Tags: nostr.Tags{
			{"A", "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:cats"},
			{"a", "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:cats"},
			{"P", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			{"p", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			{"K", "34550"},
			{"k", "34550"},
		},
	}

	ptr := GetRootPointer(event)
	require.NotNil(t, ptr)
	ref, ok := ptr.(nostr.EntityPointer)
	require.True(t, ok)
	require.Equal(t, nostr.KindCommunityDefinition, ref.Kind)
	require.Equal(t, mustPubKey(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), ref.PublicKey)
}

func TestClassifyRole(t *testing.T) {
	communityEvent := &nostr.Event{
		Kind: nostr.KindComment,
		Tags: nostr.Tags{
			{"A", "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:cats"},
			{"a", "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:cats"},
			{"P", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			{"p", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			{"K", "34550"},
			{"k", "34550"},
		},
	}
	role, ok := ClassifyRole(communityEvent)
	require.True(t, ok)
	require.Equal(t, TopLevelPost, role)

	replyEvent := &nostr.Event{
		Kind: nostr.KindComment,
		Tags: nostr.Tags{
			{"A", "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:cats"},
			{"P", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			{"K", "34550"},
			{"e", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "wss://relay.example.com", "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"},
			{"p", "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"},
			{"k", "1111"},
		},
	}
	role, ok = ClassifyRole(replyEvent)
	require.True(t, ok)
	require.Equal(t, Reply, role)
}

func TestRejectsLegacyLowercaseOnly(t *testing.T) {
	event := &nostr.Event{
		Kind: nostr.KindComment,
		Tags: nostr.Tags{{"a", "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:cats"}},
	}

	require.Nil(t, GetCommunityPointer(event))
	require.Nil(t, GetParentPointer(event))
	require.Nil(t, GetRootPointer(event))
	_, ok := ClassifyRole(event)
	require.False(t, ok)
}

func TestGetParentPointer_TopLevelPostReturnsNil(t *testing.T) {
	event := &nostr.Event{
		Kind: nostr.KindComment,
		Tags: nostr.Tags{
			{"A", "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:cats"},
			{"a", "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:cats"},
			{"P", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			{"p", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			{"K", "34550"},
			{"k", "34550"},
			{"e", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
		},
	}

	require.Nil(t, GetParentPointer(event))
}

func TestCommunityDefinitionGetters(t *testing.T) {
	event := &nostr.Event{
		Kind:   nostr.KindCommunityDefinition,
		PubKey: mustPubKey(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		Tags: nostr.Tags{
			{"d", "cats"},
			{"name", "Cats"},
			{"description", "For cat people"},
			{"image", "https://example.com/cats.png", "256x256"},
			{"p", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "wss://relay.example.com", "moderator"},
			{"p", "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"},
			{"relay", "wss://authors.example.com", "author"},
			{"relay", "wss://requests.example.com", "requests"},
			{"relay", "wss://default.example.com"},
		},
	}

	require.True(t, IsCommunityDefinition(event))
	require.Equal(t, "cats", GetDefinitionIdentifier(event))
	require.Equal(t, "Cats", GetDefinitionName(event))
	require.Equal(t, "For cat people", GetDefinitionDescription(event))
	require.Equal(t, "https://example.com/cats.png", GetDefinitionImage(event))
	require.Equal(
		t,
		[]nostr.PubKey{
			mustPubKey(t, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
		},
		GetDefinitionModerators(event),
	)
	require.Equal(
		t,
		[]CommunityRelay{
			{URL: "wss://authors.example.com", Purpose: "author"},
			{URL: "wss://requests.example.com", Purpose: "requests"},
			{URL: "wss://default.example.com", Purpose: ""},
		},
		GetDefinitionRelays(event),
	)
}

func TestCommunityDefinitionGetters_AreLightweightOnPartialTags(t *testing.T) {
	event := &nostr.Event{
		Kind: nostr.KindCommunityDefinition,
		Tags: nostr.Tags{
			{"d", "cats"},
			{"relay", "wss://default.example.com"},
			{"relay"},
			{"p", "not-a-pubkey", "", "moderator"},
		},
	}

	require.True(t, IsCommunityDefinition(event))
	require.Equal(t, "cats", GetDefinitionIdentifier(event))
	require.Equal(t, "", GetDefinitionName(event))
	require.Equal(t, "", GetDefinitionDescription(event))
	require.Equal(t, "", GetDefinitionImage(event))
	require.Empty(t, GetDefinitionModerators(event))
	require.Equal(
		t,
		[]CommunityRelay{{URL: "wss://default.example.com", Purpose: ""}},
		GetDefinitionRelays(event),
	)
}
