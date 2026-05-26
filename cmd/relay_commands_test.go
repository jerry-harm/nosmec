package cmd

import (
	"bytes"
	"strings"
	"testing"

	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/nostr_sdk/hints/memoryh"
	"github.com/jerry-harm/nosmec/nostr_sdk/kvstore/memory"
	"github.com/stretchr/testify/require"
)

func TestWriteRelayList(t *testing.T) {
	t.Parallel()

	app := config.NewAppContext(nil, config.Config{}, nil)
	app.System().Hints = memoryh.NewHintDB()
	app.System().KVStore = memory.NewStore()

	pk := mustPubKeyFromSecret(t, strings.Repeat("3", 64))
	app.System().Hints.Save(pk, "wss://relay-b.example", 0, 1)
	relayID := mustID(t, strings.Repeat("c", 64))
	require.NoError(t, app.System().KVStore.Set(makeEventRelayKey(relayID), encodeRelayList([]string{"wss://relay-a.example", "wss://relay-b.example"})))

	var out bytes.Buffer
	if err := writeRelayList(&out, app); err != nil {
		t.Fatalf("writeRelayList() error = %v", err)
	}

	got := out.String()
	want := "wss://relay-a.example\nwss://relay-b.example\n"
	if got != want {
		t.Fatalf("writeRelayList() output = %q, want %q", got, want)
	}
}

func mustPubKeyFromSecret(t *testing.T, hex string) nostr.PubKey {
	t.Helper()
	sk, err := nostr.SecretKeyFromHex(hex)
	if err != nil {
		t.Fatalf("SecretKeyFromHex(%q): %v", hex, err)
	}
	return sk.Public()
}

func mustID(t *testing.T, hex string) nostr.ID {
	t.Helper()
	id, err := nostr.IDFromHex(hex)
	if err != nil {
		t.Fatalf("IDFromHex(%q): %v", hex, err)
	}
	return id
}

func encodeRelayList(relays []string) []byte {
	total := 0
	for _, relay := range relays {
		total += 1 + len(relay)
	}
	buf := make([]byte, total)
	offset := 0
	for _, relay := range relays {
		buf[offset] = byte(len(relay))
		offset++
		copy(buf[offset:], relay)
		offset += len(relay)
	}
	return buf
}

func makeEventRelayKey(id nostr.ID) []byte {
	key := make([]byte, 9)
	key[0] = 'r'
	copy(key[1:], id[:8])
	return key
}
