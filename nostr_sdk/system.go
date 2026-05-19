package nostr_sdk

import (
	"context"
	"math/rand"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/eventstore"
	"fiatjaf.com/nostr/nip10"
	"fiatjaf.com/nostr/eventstore/nullstore"
	"fiatjaf.com/nostr/eventstore/wrappers"
	"github.com/jerry-harm/nosmec/nostr_sdk/cache"
	cache_memory "github.com/jerry-harm/nosmec/nostr_sdk/cache/memory"
	"github.com/jerry-harm/nosmec/nostr_sdk/dataloader"
	"github.com/jerry-harm/nosmec/nostr_sdk/hints"
	"github.com/jerry-harm/nosmec/nostr_sdk/hints/memoryh"
	"github.com/jerry-harm/nosmec/nostr_sdk/kvstore"
	kvstore_memory "github.com/jerry-harm/nosmec/nostr_sdk/kvstore/memory"
	"github.com/btcsuite/btcd/btcec/v2"
)

// System represents the core functionality of the SDK, providing access to
// various caches, relays, and dataloaders for efficient Nostr operations.
//
// Usually an application should have a single global instance of this and use
// its internal Pool for all its operations.
//
// Store, KVStore and Hints are databases that should generally be persisted
// for any application that is intended to be executed more than once. By
// default they're set to in-memory stores, but ideally persisteable
// implementations should be given (some alternatives are provided in subpackages).
type System struct {
	KVStore                   kvstore.KVStore
	metadataCacheOnce         sync.Once
	MetadataCache             cache.Cache32[ProfileMetadata]
	relayListCacheOnce        sync.Once
	RelayListCache            cache.Cache32[GenericList[string, Relay]]
	followListCacheOnce       sync.Once
	FollowListCache           cache.Cache32[GenericList[nostr.PubKey, ProfileRef]]
	muteListCacheOnce         sync.Once
	MuteListCache             cache.Cache32[GenericList[nostr.PubKey, ProfileRef]]
	bookmarkListCacheOnce     sync.Once
	BookmarkListCache         cache.Cache32[GenericList[string, EventRef]]
	pinListCacheOnce          sync.Once
	PinListCache              cache.Cache32[GenericList[string, EventRef]]
	blockedRelayListCacheOnce sync.Once
	BlockedRelayListCache     cache.Cache32[GenericList[string, RelayURL]]
	searchRelayListCacheOnce  sync.Once
	SearchRelayListCache      cache.Cache32[GenericList[string, RelayURL]]
	topicListCacheOnce        sync.Once
	TopicListCache            cache.Cache32[GenericList[string, Topic]]
	relaySetsCacheOnce        sync.Once
	RelaySetsCache            cache.Cache32[GenericSets[string, RelayURL]]
	followSetsCacheOnce       sync.Once
	FollowSetsCache           cache.Cache32[GenericSets[nostr.PubKey, ProfileRef]]
	topicSetsCacheOnce        sync.Once
	TopicSetsCache            cache.Cache32[GenericSets[string, Topic]]
	zapProviderCacheOnce      sync.Once
	ZapProviderCache          cache.Cache32[nostr.PubKey]
	mintKeysCacheOnce         sync.Once
	MintKeysCache             cache.Cache32[map[uint64]*btcec.PublicKey]
	nutZapInfoCacheOnce       sync.Once
	NutZapInfoCache           cache.Cache32[NutZapInfo]
	Hints                     hints.HintsDB
	Pool                      *nostr.Pool
	RelayListRelays           *RelayStream
	FollowListRelays          *RelayStream
	MetadataRelays            *RelayStream
	FallbackRelays            *RelayStream
	JustIDRelays              *RelayStream
	UserSearchRelays          *RelayStream
	NoteSearchRelays          *RelayStream
	Store                     eventstore.Store

	Publisher nostr.Publisher

	replaceableLoaders    map[nostr.Kind]*dataloader.Loader[nostr.PubKey, nostr.Event]
	addressableLoaders   map[nostr.Kind]*dataloader.Loader[nostr.PubKey, []nostr.Event]
}

// SystemModifier is a function that modifies a System instance.
// It's used with NewSystem to configure the system during creation.
type SystemModifier func(sys *System)

// RelayStream provides a rotating list of relay URLs.
// It's used to distribute requests across multiple relays.
type RelayStream struct {
	URLs   []string
	serial atomic.Int32
}

// NewRelayStream creates a new RelayStream with the provided URLs.
func NewRelayStream(urls ...string) *RelayStream {
	rs := &RelayStream{URLs: urls}
	rs.serial.Add(rand.Int31n(int32(len(urls))))
	return rs
}

// Next returns the next URL in the rotation.
func (rs *RelayStream) Next() string {
	v := rs.serial.Add(1)
	return rs.URLs[int(v)%len(rs.URLs)]
}

// NewSystem creates a new System with default configuration,
// which can be customized using the provided modifiers.
//
// The list of provided With* modifiers isn't exhaustive and
// most internal fields of System can be modified after the System
// creation -- and in many cases one or another of these will have
// to be modified, so don't be afraid of doing that.
func NewSystem() *System {
	sys := &System{
		KVStore:          kvstore_memory.NewStore(),
		RelayListRelays:  NewRelayStream("wss://indexer.coracle.social", "wss://purplepag.es", "wss://relay.primal.net", "wss://relay.nos.social"),
		FollowListRelays: NewRelayStream("wss://purplepag.es", "wss://antiprimal.net", "wss://relay.damus.io", "wss://relay.nos.social"),
		MetadataRelays:   NewRelayStream("wss://purplepag.es", "wss://antiprimal.net", "wss://relay.damus.io", "wss://relay.nos.social"),
		FallbackRelays: NewRelayStream(
			"wss://offchain.pub",
			"wss://relay.damus.io",
			"wss://nostr.mom",
			"wss://nos.lol",
			"wss://relay.mostr.pub",
			"wss://nostr.land",
			"wss://relay.ditto.pub",
		),
		JustIDRelays: NewRelayStream(
			"wss://cache2.primal.net/v1",
			"wss://relay.nostr.band",
		),
		UserSearchRelays: NewRelayStream(
			"wss://search.nos.today",
			"wss://nostr.wine",
			"wss://relay.nostr.band",
		),
		NoteSearchRelays: NewRelayStream(
			"wss://nostr.wine",
			"wss://relay.nostr.band",
			"wss://search.nos.today",
		),
		Hints: memoryh.NewHintDB(),
	}

	sys.Pool = nostr.NewPool(nostr.PoolOptions{
		AuthorKindQueryMiddleware: sys.TrackQueryAttempts,
		EventMiddleware:           sys.TrackEventHintsAndRelays,
		DuplicateMiddleware:       sys.TrackEventRelaysD,
		PenaltyBox:                true,
	})

	sys.metadataCacheOnce.Do(func() {
		if sys.MetadataCache == nil {
			sys.MetadataCache = cache_memory.New[ProfileMetadata](8000)
		}
	})
	sys.relayListCacheOnce.Do(func() {
		if sys.RelayListCache == nil {
			sys.RelayListCache = cache_memory.New[GenericList[string, Relay]](8000)
		}
	})
	sys.zapProviderCacheOnce.Do(func() {
		if sys.ZapProviderCache == nil {
			sys.ZapProviderCache = cache_memory.New[nostr.PubKey](8000)
		}
	})
	sys.mintKeysCacheOnce.Do(func() {
		if sys.MintKeysCache == nil {
			sys.MintKeysCache = cache_memory.New[map[uint64]*btcec.PublicKey](8000)
		}
	})
	sys.nutZapInfoCacheOnce.Do(func() {
		if sys.NutZapInfoCache == nil {
			sys.NutZapInfoCache = cache_memory.New[NutZapInfo](8000)
		}
	})

	if sys.Store == nil {
		sys.Store = &nullstore.NullStore{}
		sys.Store.Init()
	}
	sys.Publisher = wrappers.DynamicPublisher{GetStore: func() eventstore.Store { return sys.Store }, MaxLimit: 1000}

	sys.initializeReplaceableDataloaders()
	sys.initializeAddressableDataloaders()

	return sys
}

// Close releases resources held by the System.
func (sys *System) Close() {
	if sys.KVStore != nil {
		sys.KVStore.Close()
	}
	if sys.Pool != nil {
		sys.Pool.Close("sdk.System closed")
	}
}

// FetchGlobalTimelinePage fetches the global timeline starting from the given timestamp.
// It queries relays for text notes and publishes them to the local store.
func (sys *System) FetchGlobalTimelinePage(
	ctx context.Context,
	limit int,
	until nostr.Timestamp,
) ([]nostr.Event, error) {
	filter := nostr.Filter{
		Kinds: []nostr.Kind{nostr.KindTextNote},
		Limit: limit,
	}
	if until > 0 {
		filter.Until = until
	}

	relays := sys.FallbackRelays.URLs
	if len(relays) == 0 {
		relays = []string{"wss://relay.damus.io", "wss://nos.lol"}
	}

	events := make([]nostr.Event, 0, limit)
	for ie := range sys.Pool.FetchMany(ctx, relays, filter, nostr.SubscriptionOptions{Label: "global"}) {
		sys.Publisher.Publish(ctx, ie.Event)
		if ie.Event.CreatedAt < until || until == 0 {
			events = append(events, ie.Event)
			if len(events) >= limit {
				break
			}
		}
	}

	slices.SortFunc(events, nostr.CompareEventReverse)
	return events, nil
}

// FetchMyTimelinePage fetches timeline events for a given pubkey.
// It first queries the local store, then fetches from relays if needed.
func (sys *System) FetchMyTimelinePage(
	ctx context.Context,
	pubkey nostr.PubKey,
	limit int,
	until nostr.Timestamp,
) ([]nostr.Event, error) {
	oldestKey := makePubkeyStreamKey(pubkeyStreamOldestPrefix, pubkey)
	var oldestTimestamp nostr.Timestamp

	if data, _ := sys.KVStore.Get(oldestKey); data != nil {
		oldestTimestamp = decodeTimestamp(data)
		if oldestTimestamp == 0 {
			oldestTimestamp = nostr.Now()
		}
	}

	events := make([]nostr.Event, 0, limit)

	if until > oldestTimestamp {
		filter := nostr.Filter{
			Authors: []nostr.PubKey{pubkey},
			Kinds:   []nostr.Kind{nostr.KindTextNote},
			Limit:   limit,
			Until:   until,
		}
		count := 0
		for evt := range sys.Store.QueryEvents(filter, limit) {
			events = append(events, evt)
			count++
			if count >= limit {
				slices.SortFunc(events, nostr.CompareEventReverse)
				return events, nil
			}
		}
	}

	relays := sys.FetchOutboxRelays(ctx, pubkey, 2)
	if len(relays) == 0 {
		relays = []string{"wss://relay.damus.io", "wss://nos.lol"}
	}

	filter := nostr.Filter{
		Authors: []nostr.PubKey{pubkey},
		Kinds:   []nostr.Kind{nostr.KindTextNote},
		Limit:   limit,
	}
	if oldestTimestamp > 0 {
		filter.Until = oldestTimestamp + 1
		filter.Since = 0
	} else if until > 0 {
		filter.Until = until
	}

	for ie := range sys.Pool.FetchMany(ctx, relays, filter, nostr.SubscriptionOptions{Label: "mytimeline"}) {
		sys.Publisher.Publish(ctx, ie.Event)
		if ie.Event.CreatedAt < until || until == 0 {
			events = append(events, ie.Event)
			if len(events) >= limit {
				break
			}
		}
		if ie.Event.CreatedAt < oldestTimestamp || oldestTimestamp == 0 {
			oldestTimestamp = ie.Event.CreatedAt
		}
	}

	sys.KVStore.Set(oldestKey, encodeTimestamp(oldestTimestamp))
	slices.SortFunc(events, nostr.CompareEventReverse)
	return events, nil
}

// FetchFollowedTimelinePage fetches timeline events for followed pubkeys and community addresses.
// It queries both local store and relays for events.
func (sys *System) FetchFollowedTimelinePage(
	ctx context.Context,
	pubkeys []nostr.PubKey,
	communityAddrs []string,
	limit int,
	until nostr.Timestamp,
) ([]nostr.Event, error) {
	if len(pubkeys) == 0 && len(communityAddrs) == 0 {
		return nil, nil
	}

	kinds := []nostr.Kind{nostr.KindTextNote, nostr.KindComment}
	limitPerKey := limit
	if len(pubkeys) > 0 {
		limitPerKey = (limit + len(pubkeys) - 1) / len(pubkeys)
	}
	if limitPerKey < 1 {
		limitPerKey = 1
	}

	events := make([]nostr.Event, 0, limit)
	wg := sync.WaitGroup{}
	wg.Add(len(pubkeys) + len(communityAddrs))

	for _, pubkey := range pubkeys {
		oldestKey := makePubkeyStreamKey(pubkeyStreamOldestPrefix, pubkey)
		var oldestTimestamp nostr.Timestamp
		if data, _ := sys.KVStore.Get(oldestKey); data != nil {
			oldestTimestamp = decodeTimestamp(data)
			if oldestTimestamp == 0 {
				oldestTimestamp = nostr.Now()
			}
		}

		go func(pk nostr.PubKey, oldTs nostr.Timestamp) {
			defer wg.Done()

			localEvents := make([]nostr.Event, 0, limitPerKey)
			newOldest := oldTs

			if until > oldTs {
				filter := nostr.Filter{
					Authors: []nostr.PubKey{pk},
					Kinds:   kinds,
					Until:   until,
					Limit:   limitPerKey,
				}
				count := 0
				for evt := range sys.Store.QueryEvents(filter, limitPerKey) {
					localEvents = append(localEvents, evt)
					count++
					if count >= limitPerKey {
						break
					}
				}
			}

			relays := sys.FetchOutboxRelays(ctx, pk, 2)
			if len(relays) == 0 {
				relays = []string{"wss://relay.damus.io", "wss://nos.lol"}
			}

			filter := nostr.Filter{
				Authors: []nostr.PubKey{pk},
				Kinds:   kinds,
				Limit:   limitPerKey,
			}
			if oldTs > 0 {
				filter.Until = oldTs + 1
				filter.Since = 0
			}

			for ie := range sys.Pool.FetchMany(ctx, relays, filter, nostr.SubscriptionOptions{Label: "followed"}) {
				sys.Publisher.Publish(ctx, ie.Event)
				if ie.Event.CreatedAt < until || until == 0 {
					localEvents = append(localEvents, ie.Event)
				}
				if ie.Event.CreatedAt < newOldest || newOldest == 0 {
					newOldest = ie.Event.CreatedAt
				}
			}

			sys.KVStore.Set(makePubkeyStreamKey(pubkeyStreamOldestPrefix, pk), encodeTimestamp(newOldest))

			for _, ev := range localEvents {
				if ev.CreatedAt < until || until == 0 {
					events = append(events, ev)
				}
			}
		}(pubkey, oldestTimestamp)
	}

	for _, addr := range communityAddrs {
		go func(communityAddr string) {
			defer wg.Done()

			filter := nostr.Filter{
				Tags:  nostr.TagMap{"a": []string{communityAddr}},
				Kinds: kinds,
				Limit: limitPerKey,
			}
			if until > 0 {
				filter.Until = until
			}

			relays := sys.FetchOutboxRelays(ctx, nostr.PubKey{}, 2)
			if len(relays) == 0 {
				relays = []string{"wss://relay.damus.io", "wss://nos.lol", "wss://relay.nostr.band"}
			}

			count := 0
			for ie := range sys.Pool.FetchMany(ctx, relays, filter, nostr.SubscriptionOptions{Label: "community"}) {
				sys.Publisher.Publish(ctx, ie.Event)
				if ie.Event.CreatedAt < until || until == 0 {
					events = append(events, ie.Event)
					count++
					if count >= limitPerKey {
						break
					}
				}
			}
		}(addr)
	}

	wg.Wait()
	slices.SortFunc(events, nostr.CompareEventReverse)
	if len(events) > limit {
		events = events[:limit]
	}
	return events, nil
}

// FetchProfilesBatch fetches profile metadata for multiple pubkeys in batch.
func (sys *System) FetchProfilesBatch(
	ctx context.Context,
	pubkeys []nostr.PubKey,
) map[nostr.PubKey]*nostr.Event {
	results := make(map[nostr.PubKey]*nostr.Event)
	for _, pk := range pubkeys {
		pm := sys.FetchProfileMetadata(ctx, pk)
		if pm.Event != nil {
			results[pk] = pm.Event
		}
	}
	return results
}

// FetchEventByFilter fetches a specific event using a filter.
// It uses the filter to construct a pointer and fetches the event.
func (sys *System) FetchEventByFilter(
	ctx context.Context,
	filter nostr.Filter,
	timeoutMs int,
) *nostr.Event {
	pointer := filterToPointer(filter)
	if pointer == nil {
		return nil
	}

	params := FetchSpecificEventParameters{}
	if timeoutMs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
		defer cancel()
	}

	evt, _, _ := sys.FetchSpecificEvent(ctx, pointer, params)
	return evt
}

func filterToPointer(filter nostr.Filter) nostr.Pointer {
	if len(filter.IDs) == 1 && len(filter.Authors) == 0 && len(filter.Kinds) == 0 && len(filter.Tags) == 0 {
		return nostr.EventPointer{ID: filter.IDs[0]}
	}
	if len(filter.Authors) == 1 && len(filter.Kinds) == 1 && filter.Kinds[0].IsReplaceable() {
		return nostr.EntityPointer{
			PublicKey:  filter.Authors[0],
			Kind:       filter.Kinds[0],
			Identifier: getFirstTagValue(filter.Tags, "d"),
		}
	}
	return nil
}

func getFirstTagValue(tags nostr.TagMap, key string) string {
	if vals, ok := tags[key]; ok && len(vals) > 0 {
		return vals[0]
	}
	return ""
}

// FetchNote fetches a note (text event) by its ID string.
func (sys *System) FetchNote(ctx context.Context, noteID string, timeoutMs int) *nostr.Event {
	id, err := nostr.IDFromHex(noteID)
	if err != nil {
		return nil
	}
	pointer := nostr.EventPointer{ID: id}
	evt, _, _ := sys.FetchSpecificEvent(ctx, pointer, FetchSpecificEventParameters{})
	return evt
}

// FetchRepliesToRoot fetches all replies to a root event (by root ID).
func (sys *System) FetchRepliesToRoot(ctx context.Context, rootID nostr.ID, limit int) []*nostr.Event {
	filter := nostr.Filter{
		Kinds: []nostr.Kind{nostr.KindTextNote, nostr.KindComment},
		Tags:  nostr.TagMap{"e": []string{rootID.Hex()}},
		Limit: limit,
	}

	relays := sys.FallbackRelays.URLs
	if len(relays) == 0 {
		relays = []string{"wss://relay.damus.io", "wss://nos.lol"}
	}

	var events []*nostr.Event
	for ie := range sys.Pool.FetchMany(ctx, relays, filter, nostr.SubscriptionOptions{Label: "replies"}) {
		sys.Publisher.Publish(ctx, ie.Event)
		events = append(events, &ie.Event)
	}
	return events
}

// FetchParent fetches the immediate parent event of a reply.
// It uses NIP-10 to get the parent pointer and resolves relay hints.
func (sys *System) FetchParent(ctx context.Context, event *nostr.Event, timeoutMs int) *nostr.Event {
	if event == nil {
		return nil
	}

	replyTag := event.Tags.FindWithValue("e", "reply")
	if len(replyTag) < 2 {
		return nil
	}

	relayHint := ""
	if len(replyTag) >= 3 && replyTag[2] != "" {
		relayHint = replyTag[2]
	}

	ptr := nip10.GetImmediateParent(event.Tags)
	if ptr == nil {
		return nil
	}

	if ep, ok := ptr.(nostr.EventPointer); ok && relayHint != "" {
		ep.Relays = []string{relayHint}
		ptr = ep
	}

	if timeoutMs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
		defer cancel()
	}

	evt, _, _ := sys.FetchSpecificEvent(ctx, ptr, FetchSpecificEventParameters{})
	return evt
}
