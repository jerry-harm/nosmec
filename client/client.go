package client

import (
	"log"
	"path/filepath"

	"fiatjaf.com/nostr/eventstore/lmdb"
	"fiatjaf.com/nostr/sdk"
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

	// Create a new system
	System = sdk.NewSystem()

	// Configure the system
	System.Store = Store

	// Add all relays from config using EnsureRelay
	for _, relayURL := range config.Global.Client.Relays {
		if _, err := System.Pool.EnsureRelay(relayURL); err != nil {
			log.Printf("Warning: Failed to ensure relay %s: %v", relayURL, err)
		} else {
			log.Printf("Added relay: %s", relayURL)
		}
	}

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

// TODO subscribe needed events and save it
