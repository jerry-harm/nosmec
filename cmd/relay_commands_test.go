package cmd

import (
	"bytes"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"fiatjaf.com/nostr"
	"go.etcd.io/bbolt"
)

func TestMergeUniqueSortedRelayURLs(t *testing.T) {
	t.Parallel()

	got := mergeUniqueSortedRelayURLs(
		[]string{"wss://relay-b.example", "wss://relay-a.example", ""},
		[]string{"wss://relay-a.example", "wss://relay-c.example"},
	)

	want := []string{
		"wss://relay-a.example",
		"wss://relay-b.example",
		"wss://relay-c.example",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mergeUniqueSortedRelayURLs() = %#v, want %#v", got, want)
	}
}

func TestCollectHintsDBRelays(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "hints.db")
	pk1 := mustPubKeyFromSecret(t, strings.Repeat("1", 64))
	pk2 := mustPubKeyFromSecret(t, strings.Repeat("2", 64))

	seedBoltDB(t, dbPath, []byte("hints"), map[string][]byte{
		string(append(pk1[:], []byte("wss://relay-b.example")...)): {1},
		string(append(pk1[:], []byte("wss://relay-a.example")...)): {1},
		string(append(pk2[:], []byte("wss://relay-a.example")...)): {1},
	})

	got, err := collectHintsDBRelays(dbPath)
	if err != nil {
		t.Fatalf("collectHintsDBRelays() error = %v", err)
	}

	want := []string{
		"wss://relay-a.example",
		"wss://relay-b.example",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("collectHintsDBRelays() = %#v, want %#v", got, want)
	}
}

func TestCollectKVStoreEventRelays(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "kvstore.db")
	id1 := mustID(t, strings.Repeat("a", 64))
	id2 := mustID(t, strings.Repeat("b", 64))

	seedBoltDB(t, dbPath, []byte("default"), map[string][]byte{
		string(makeEventRelayKVKey(id1)): encodeRelayListForTest([]string{"wss://relay-b.example", "wss://relay-a.example"}),
		string(makeEventRelayKVKey(id2)): encodeRelayListForTest([]string{"wss://relay-a.example", "wss://relay-c.example"}),
		"xignored":                       []byte("ignored"),
	})

	got, err := collectKVStoreEventRelays(dbPath)
	if err != nil {
		t.Fatalf("collectKVStoreEventRelays() error = %v", err)
	}

	want := []string{
		"wss://relay-a.example",
		"wss://relay-b.example",
		"wss://relay-c.example",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("collectKVStoreEventRelays() = %#v, want %#v", got, want)
	}
}

func TestWriteRelayList(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	pk := mustPubKeyFromSecret(t, strings.Repeat("3", 64))
	id := mustID(t, strings.Repeat("c", 64))

	seedBoltDB(t, filepath.Join(dataDir, "hints.db"), []byte("hints"), map[string][]byte{
		string(append(pk[:], []byte("wss://relay-b.example")...)): {1},
	})
	seedBoltDB(t, filepath.Join(dataDir, "kvstore.db"), []byte("default"), map[string][]byte{
		string(makeEventRelayKVKey(id)): encodeRelayListForTest([]string{"wss://relay-a.example", "wss://relay-b.example"}),
	})

	var out bytes.Buffer
	if err := writeRelayList(&out, dataDir); err != nil {
		t.Fatalf("writeRelayList() error = %v", err)
	}

	got := out.String()
	want := "wss://relay-a.example\nwss://relay-b.example\n"
	if got != want {
		t.Fatalf("writeRelayList() output = %q, want %q", got, want)
	}
}

func mustPubKey(t *testing.T, hex string) nostr.PubKey {
	t.Helper()
	pk, err := nostr.PubKeyFromHex(hex)
	if err != nil {
		t.Fatalf("PubKeyFromHex(%q): %v", hex, err)
	}
	return pk
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

func seedBoltDB(t *testing.T, dbPath string, bucket []byte, entries map[string][]byte) {
	t.Helper()

	db, err := bbolt.Open(dbPath, 0o600, nil)
	if err != nil {
		t.Fatalf("open bbolt %s: %v", dbPath, err)
	}
	defer db.Close()

	err = db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}
		for key, value := range entries {
			if err := b.Put([]byte(key), value); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed bbolt %s: %v", dbPath, err)
	}
}

func makeEventRelayKVKey(id nostr.ID) []byte {
	key := make([]byte, 9)
	key[0] = 'r'
	copy(key[1:], id[:8])
	return key
}

func encodeRelayListForTest(relays []string) []byte {
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
