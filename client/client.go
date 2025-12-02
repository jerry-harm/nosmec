package client

import (
	"log"
	"path/filepath"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/eventstore/lmdb"
	"fiatjaf.com/nostr/sdk"
	"fiatjaf.com/nostr/sdk/hints/lmdbh"
	"github.com/jerry-harm/nosmec/config"
)

var (
	// System is the global SDK system instance
	System *sdk.System
	// Store is the LMDB event store
	Store *lmdb.LMDBBackend
)

// Init initializes the client system with configuration
func Init() {
	// Initialize LMDB event store (same as relay.go)
	Store = &lmdb.LMDBBackend{
		Path: filepath.Join(config.Global.BasePath, "client.db"),
	}

	if err := Store.Init(); err != nil {
		log.Fatalf("Failed to initialize LMDB store: %v", err)
	}

	// Initialize LMDB hints database
	hintsDB, err := lmdbh.NewLMDBHints(filepath.Join(config.Global.BasePath, "hints.db"))
	if err != nil {
		log.Fatalf("Failed to initialize LMDB hints: %v", err)
	}

	// Create a new system
	System = sdk.NewSystem()

	// Configure the system
	System.Store = Store
	System.Hints = hintsDB

	// Create a pool with relays from config
	pool := nostr.NewPool(nostr.PoolOptions{})

	// Add all relays from config using EnsureRelay
	for _, relayURL := range config.Global.Client.Relays {
		if _, err := pool.EnsureRelay(relayURL); err != nil {
			log.Printf("Warning: Failed to ensure relay %s: %v", relayURL, err)
		} else {
			log.Printf("Added relay: %s", relayURL)
		}
	}

	System.Pool = pool

	// Set up relay streams
	System.RelayListRelays = sdk.NewRelayStream(config.Global.Client.Relays...)
	System.FollowListRelays = sdk.NewRelayStream(config.Global.Client.Relays...)
	System.MetadataRelays = sdk.NewRelayStream(config.Global.Client.Relays...)
	System.FallbackRelays = sdk.NewRelayStream(config.Global.Client.Relays...)
	System.JustIDRelays = sdk.NewRelayStream(config.Global.Client.Relays...)
	System.UserSearchRelays = sdk.NewRelayStream(config.Global.Client.Relays...)
	System.NoteSearchRelays = sdk.NewRelayStream(config.Global.Client.Relays...)

	log.Println("Client system initialized successfully")
}

// Close releases resources held by the client system
func Close() {
	if System != nil {
		System.Close()
	}
	if Store != nil {
		Store.Close()
	}
}

// GetSystem returns the global system instance
func GetSystem() *sdk.System {
	return System
}

// GetStore returns the global store instance
func GetStore() *lmdb.LMDBBackend {
	return Store
}

// GetPool returns the global pool instance
func GetPool() *nostr.Pool {
	if System != nil {
		return System.Pool
	}
	return nil
}
