package utils

import (
	"context"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip65"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/logger"
)

// VerifyRelayConnectivity checks if a relay can be connected to.
func VerifyRelayConnectivity(ctx context.Context, url string) (bool, error) {
	relay := nostr.NewRelay(ctx, url, nostr.RelayOptions{})
	if err := relay.Connect(ctx); err != nil {
		return false, nil
	}
	return relay.IsConnected(), nil
}

// VerifyRelaysConnectivity filters out unreachable relays from a list.
func VerifyRelaysConnectivity(ctx context.Context, app *config.AppContext, urls []string) ([]string, error) {
	if len(urls) == 0 {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, app.QueryTimeout())
	defer cancel()

	var reachable []string
	for _, url := range urls {
		if ok, _ := VerifyRelayConnectivity(ctx, url); ok {
			reachable = append(reachable, url)
		}
	}
	return reachable, nil
}

// DiscoverUserRelays queries both local relay (cache) and remote relays
// for a user's NIP-65 relay list. Kind 10002 is a replaceable event.
func DiscoverUserRelays(ctx context.Context, app *config.AppContext, pubKey nostr.PubKey) ([]string, error) {
	filter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindRelayListMetadata},
		Authors: []nostr.PubKey{pubKey},
		Limit:   1,
	}

	relays := getAllCandidateRelays(app)
	if len(relays) == 0 {
		logger.Debug("DiscoverUserRelays: no candidate relays", "pubkey", pubKey.Hex())
		return nil, nil
	}

	logger.Debug("DiscoverUserRelays: querying", "pubkey", pubKey.Hex(), "relays", relays)

	results := app.Pool().FetchManyReplaceable(ctx, relays, filter, nostr.SubscriptionOptions{})

	var event *nostr.Event
	results.Range(func(key nostr.ReplaceableKey, ev nostr.Event) bool {
		logger.Debug("DiscoverUserRelays: got event", "pubkey", pubKey.Hex(), "eventID", ev.ID.Hex())
		event = &ev
		return false
	})

	if event == nil {
		logger.Debug("DiscoverUserRelays: no event found", "pubkey", pubKey.Hex())
		return nil, nil
	}

	readRelays, writeRelays := nip65.ParseRelayList(*event)
	allRelays := append(readRelays, writeRelays...)
	logger.Debug("DiscoverUserRelays: parsed relays", "pubkey", pubKey.Hex(), "read", readRelays, "write", writeRelays)

	EnsureRelays(app, allRelays)
	app.TrackRelays(allRelays)

	return allRelays, nil
}

func getAllCandidateRelays(app *config.AppContext) []string {
	seen := make(map[string]struct{})
	var result []string

	readable := app.AllReadableRelays()
	logger.Debug("getAllCandidateRelays: AllReadableRelays", "count", len(readable), "relays", readable)

	for _, r := range readable {
		if _, ok := seen[r]; !ok {
			seen[r] = struct{}{}
			result = append(result, r)
		}
	}

	known := app.Config().KnownRelays
	logger.Debug("getAllCandidateRelays: KnownRelays", "count", len(known), "relays", known)

	for _, r := range known {
		if _, ok := seen[r]; !ok {
			seen[r] = struct{}{}
			result = append(result, r)
		}
	}

	logger.Debug("getAllCandidateRelays: total candidates", "count", len(result))
	return result
}

func DiscoverUserRelaysWithFallback(ctx context.Context, app *config.AppContext, pubKey nostr.PubKey) ([]string, error) {
	relays, err := DiscoverUserRelays(ctx, app, pubKey)
	if err != nil {
		logger.Debug("DiscoverUserRelays failed, trying KnownRelays fallback", "error", err.Error(), "pubkey", pubKey.Hex())
		relays, err = nil, nil
	}
	if len(relays) == 0 {
		relays = app.Config().KnownRelays
	}
	return relays, err
}

// EnsureRelays adds relay URLs to the global pool (lazy connection).
func EnsureRelays(app *config.AppContext, urls []string) {
	for _, url := range urls {
		app.Pool().EnsureRelay(url)
	}
}

// GetQueryRelays returns the ordered relay list for querying events related to the given event.
// Priority: 1. tag[2] relay hints  2. HintsDB outbox (from e tag[3] author pubkeys)
// 3. AllReadableRelays  4. KnownRelays fallback.
func GetQueryRelays(event *nostr.Event, app *config.AppContext) []string {
	seen := make(map[string]struct{})
	var result []string

	// 1. Relay hints from tags (tag[2])
	hints := ExtractRelayHints(event)
	for _, r := range hints {
		if _, ok := seen[r]; !ok {
			seen[r] = struct{}{}
			result = append(result, r)
		}
	}

	// 2. Outbox relays from e tag author pubkeys via HintsDB
	// Per NIP-10 (marked): ["e", <id>, <relay>, <marker>, <pubkey>] — pubkey at tag[4]
	// Per NIP-01 (legacy): ["e", <id>, <relay>, <pubkey>] — pubkey at tag[3]
	for tag := range event.Tags.FindAll("e") {
		pubkey := ""
		if len(tag) >= 5 && nostr.IsValid32ByteHex(tag[4]) {
			pubkey = tag[4] // NIP-10 marked
		} else if len(tag) >= 4 && nostr.IsValid32ByteHex(tag[3]) {
			pubkey = tag[3] // NIP-01 legacy
		}
		if pubkey != "" {
			outbox := app.Hints().TopN(pubkey, 3)
			for _, r := range outbox {
				if _, ok := seen[r]; !ok {
					seen[r] = struct{}{}
					result = append(result, r)
				}
			}
		}
	}

	// 3. All readable relays (configured + local)
	for _, r := range app.AllReadableRelays() {
		if _, ok := seen[r]; !ok {
			seen[r] = struct{}{}
			result = append(result, r)
		}
	}

	// 4. KnownRelays fallback
	for _, r := range app.Config().KnownRelays {
		if _, ok := seen[r]; !ok {
			seen[r] = struct{}{}
			result = append(result, r)
		}
	}

	return result
}
