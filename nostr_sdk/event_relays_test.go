package nostr_sdk

import (
	"strings"
	"testing"

	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/nostr_sdk/kvstore/memory"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecodeRelayList(t *testing.T) {
	tests := []struct {
		name   string
		relays []string
	}{
		{
			name:   "empty list",
			relays: []string{},
		},
		{
			name:   "single relay",
			relays: []string{"wss://relay.example.com"},
		},
		{
			name: "multiple relays",
			relays: []string{
				"wss://relay1.example.com",
				"wss://relay23.example.com",
				"wss://relay456.example.com",
			},
		},
		{
			name: "relays with varying lengths",
			relays: []string{
				"wss://a.com",
				"wss://very-long-relay-url.example.com",
				"wss://b.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// test encoding
			encoded := encodeRelayList(tt.relays)
			require.NotNil(t, encoded)

			// test decoding
			decoded := decodeRelayList(encoded)
			require.Equal(t, tt.relays, decoded)
		})
	}

	t.Run("malformed data", func(t *testing.T) {
		// test with truncated data
		decoded := decodeRelayList([]byte{5, 'h', 'e'}) // length prefix of 5 but only 2 bytes of data
		require.Nil(t, decoded)

		// test with invalid length prefix
		decoded = decodeRelayList([]byte{255}) // length prefix but no data
		require.Nil(t, decoded)
	})

	t.Run("skip too long relay URLs", func(t *testing.T) {
		// create a long URL by repeating 'a' 257 times
		longURL := "wss://" + strings.Repeat("a", 257) + ".com"
		longRelays := []string{
			"wss://normal.example.com",
			longURL,
			"wss://also-normal.example.com",
		}

		encoded := encodeRelayList(longRelays)
		decoded := decodeRelayList(encoded)

		// should only contain the normal URLs
		require.Equal(t, []string{
			"wss://normal.example.com",
			"wss://also-normal.example.com",
		}, decoded)
	})

	t.Run("list known event relays merges unique sorted relays", func(t *testing.T) {
		sys := NewSystem()
		sys.KVStore = memory.NewStore()

		id1 := mustEventID(t, strings.Repeat("a", 64))
		id2 := mustEventID(t, strings.Repeat("b", 64))
		require.NoError(t, sys.KVStore.Set(makeEventRelayKey(id1), encodeRelayList([]string{"wss://relay-b.example", "wss://relay-a.example"})))
		require.NoError(t, sys.KVStore.Set(makeEventRelayKey(id2), encodeRelayList([]string{"wss://relay-a.example", "wss://relay-c.example"})))
		require.NoError(t, sys.KVStore.Set([]byte("xignored"), []byte("ignored")))

		relays, err := sys.ListKnownEventRelays()
		require.NoError(t, err)
		require.Equal(t, []string{
			"wss://relay-a.example",
			"wss://relay-b.example",
			"wss://relay-c.example",
		}, relays)
	})
}

func mustEventID(t *testing.T, hex string) nostr.ID {
	t.Helper()
	id, err := nostr.IDFromHex(hex)
	require.NoError(t, err)
	return id
}
