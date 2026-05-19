package nostr_sdk

import (
	"testing"

	"fiatjaf.com/nostr"
	"github.com/stretchr/testify/require"
)

func TestGetThreadParentPointer_NIP72TopLevelPostHasNoParent(t *testing.T) {
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

	require.Nil(t, GetThreadParentPointer(event))
}

func TestGetThreadParentPointer_NIP72ReplyUsesLowercaseParentReference(t *testing.T) {
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

	ptr := GetThreadParentPointer(event)
	require.NotNil(t, ptr)
	ep, ok := ptr.(nostr.EventPointer)
	require.True(t, ok)
	require.Equal(t, mustID(t, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"), ep.ID)
	require.Equal(t, []string{"wss://relay.example.com"}, ep.Relays)
}

func TestGetThreadRootID_StrictCommunityPostUsesSelfAsRoot(t *testing.T) {
	id := mustID(t, "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd")
	event := &nostr.Event{
		ID:   id,
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

	rootID, isRoot, err := GetThreadRootID(event)
	require.NoError(t, err)
	require.True(t, isRoot)
	require.Equal(t, id, rootID)
}

func TestExtractCommunityScope_StrictNIP72RejectsLowercaseFallback(t *testing.T) {
	event := &nostr.Event{
		Kind: nostr.KindComment,
		Tags: nostr.Tags{{"a", "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:cats"}},
	}

	require.Empty(t, ExtractCommunityScope(event))
	if MatchesCommunityScope(event, "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:cats") {
		t.Fatal("MatchesCommunityScope() = true, want false")
	}
}
