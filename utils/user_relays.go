package utils

import (
	"context"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip65"
	"github.com/jerry-harm/nosmec/config"
)

// DiscoverUserRelays queries both local relay (cache) and remote relays
// for a user's NIP-65 relay list. Kind 10002 is a replaceable event.
func DiscoverUserRelays(ctx context.Context, app *config.AppContext, pubKey nostr.PubKey) ([]string, error) {
	filter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindRelayListMetadata},
		Authors: []nostr.PubKey{pubKey},
		Limit:   1,
	}

	// Query local relay + known relays in one call
	relays := app.AllReadableRelays()
	if len(relays) == 0 {
		return nil, nil
	}

	// Kind 10002 is replaceable - use FetchManyReplaceable
	results := app.Pool().FetchManyReplaceable(ctx, relays, filter, nostr.SubscriptionOptions{})

	var event *nostr.Event
	results.Range(func(key nostr.ReplaceableKey, ev nostr.Event) bool {
		event = &ev
		return false
	})

	if event == nil {
		return nil, nil
	}

	readRelays, writeRelays := nip65.ParseRelayList(*event)

	// Backup to local relay
	CacheEvent(event, app)

	// Register discovered relays in pool
	EnsureRelays(app, readRelays)
	EnsureRelays(app, writeRelays)

	// Track in known relays for future connections
	app.TrackRelays(readRelays)
	app.TrackRelays(writeRelays)

	return readRelays, nil
}

// EnsureRelays adds relay URLs to the global pool (lazy connection).
func EnsureRelays(app *config.AppContext, urls []string) {
	for _, url := range urls {
		app.Pool().EnsureRelay(url)
	}
}
