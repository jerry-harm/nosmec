package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/PowerDNS/lmdb-go/lmdb"
	"github.com/spf13/cobra"
)

var (
	hintsDBIName      = "hints"
	kvStoreDBIName    = "store"
	eventRelayPrefix  = byte('r')
	hintsKeyPubkeyLen = 32
)

func registerRelayCommands() {
	relayCmd := &cobra.Command{
		Use:   "relay",
		Short: "Relay operations",
	}

	relayListCmd := &cobra.Command{
		Use:   "list",
		Short: "List relays discovered from SDK databases",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp()
			if app == nil {
				return newError("app not initialized", nil)
			}

			if err := writeRelayList(cmd.OutOrStdout(), app.Config().DataDir); err != nil {
				return newError("failed to list relays", err)
			}
			return nil
		},
	}

	relayCmd.AddCommand(relayListCmd)
	RegisterCommandGroup("Relay", "Relay operations", relayCmd)
}

func writeRelayList(w io.Writer, dataDir string) error {
	hintsRelays, err := collectHintsDBRelays(filepath.Join(dataDir, "hints"))
	if err != nil {
		return err
	}

	kvRelays, err := collectKVStoreEventRelays(filepath.Join(dataDir, "kvstore"))
	if err != nil {
		return err
	}

	for _, relay := range mergeUniqueSortedRelayURLs(hintsRelays, kvRelays) {
		if _, err := fmt.Fprintln(w, relay); err != nil {
			return err
		}
	}

	return nil
}

func mergeUniqueSortedRelayURLs(groups ...[]string) []string {
	seen := make(map[string]struct{})
	merged := make([]string, 0)

	for _, group := range groups {
		for _, relay := range group {
			if relay == "" {
				continue
			}
			if _, ok := seen[relay]; ok {
				continue
			}
			seen[relay] = struct{}{}
			merged = append(merged, relay)
		}
	}

	sort.Strings(merged)
	return merged
}

func collectHintsDBRelays(dbPath string) ([]string, error) {
	var relays []string
	err := readLMDB(dbPath, hintsDBIName, func(k, _ []byte) {
		if len(k) <= hintsKeyPubkeyLen {
			return
		}
		relays = append(relays, string(k[hintsKeyPubkeyLen:]))
	})
	if err != nil {
		return nil, err
	}
	return mergeUniqueSortedRelayURLs(relays), nil
}

func collectKVStoreEventRelays(dbPath string) ([]string, error) {
	var relays []string
	err := readLMDB(dbPath, kvStoreDBIName, func(k, v []byte) {
		if len(k) != 9 || k[0] != eventRelayPrefix {
			return
		}
		relays = append(relays, decodeKVRelayList(v)...)
	})
	if err != nil {
		return nil, err
	}
	return mergeUniqueSortedRelayURLs(relays), nil
}

func readLMDB(dbPath string, dbiName string, visit func(k, v []byte)) error {
	// If the directory doesn't exist, there's no data to read.
	if info, err := os.Stat(dbPath); err != nil || !info.IsDir() {
		return nil
	}

	env, err := lmdb.NewEnv()
	if err != nil {
		return err
	}
	defer env.Close()

	if err := env.SetMaxDBs(1); err != nil {
		return err
	}
	if err := env.SetMapSize(1 << 30); err != nil {
		return err
	}

	if err := env.Open(dbPath, lmdb.Readonly|lmdb.NoTLS, 0); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	return env.View(func(txn *lmdb.Txn) error {
		txn.RawRead = true
		dbi, err := txn.OpenDBI(dbiName, 0)
		if err != nil {
			if lmdb.IsNotFound(err) {
				return nil
			}
			return err
		}

		cursor, err := txn.OpenCursor(dbi)
		if err != nil {
			return err
		}
		defer cursor.Close()

		for k, v, err := cursor.Get(nil, nil, lmdb.First); err == nil; k, v, err = cursor.Get(nil, nil, lmdb.Next) {
			key := append([]byte(nil), k...)
			val := append([]byte(nil), v...)
			visit(key, val)
		}
		if !lmdb.IsNotFound(err) {
			return err
		}
		return nil
	})
}

func decodeKVRelayList(data []byte) []string {
	relays := make([]string, 0, 6)
	for offset := 0; offset < len(data); {
		if offset+1 > len(data) {
			return nil
		}
		length := int(data[offset])
		offset++
		if offset+length > len(data) {
			return nil
		}
		relays = append(relays, string(data[offset:offset+length]))
		offset += length
	}
	return relays
}
