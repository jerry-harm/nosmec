package access

import (
	"os"
	"path/filepath"

	"fiatjaf.com/nostr/sdk/kvstore"
	bboltKv "fiatjaf.com/nostr/sdk/kvstore/bbolt"
)

// NewKVStore opens or creates a BoltDB-backed key-value store for event→relay
// persistence under the given data directory. The store file is named
// "event_relays.db".
func NewKVStore(dataDir string) (kvstore.KVStore, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	return bboltKv.NewStore(filepath.Join(dataDir, "event_relays.db"))
}
