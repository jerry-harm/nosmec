package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

var (
	hintsBucketName   = []byte("hints")
	kvDefaultBucket   = []byte("default")
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
	hintsRelays, err := collectHintsDBRelays(filepath.Join(dataDir, "hints.db"))
	if err != nil {
		return err
	}

	kvRelays, err := collectKVStoreEventRelays(filepath.Join(dataDir, "kvstore.db"))
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
	err := readBoltBucket(dbPath, hintsBucketName, func(k, _ []byte) {
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
	err := readBoltBucket(dbPath, kvDefaultBucket, func(k, v []byte) {
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

func readBoltBucket(dbPath string, bucketName []byte, visit func(k, v []byte)) error {
	db, err := bbolt.Open(dbPath, 0o600, &bbolt.Options{ReadOnly: true})
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer db.Close()

	return db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			return nil
		}

		return bucket.ForEach(func(k, v []byte) error {
			key := append([]byte(nil), k...)
			val := append([]byte(nil), v...)
			visit(key, val)
			return nil
		})
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
