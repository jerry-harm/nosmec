package sdkplus

import (
	"context"
	"slices"
	"sync"
	"time"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip10"
	"fiatjaf.com/nostr/sdk"
)

const (
	pubkeyStreamLatestPrefix = byte('L')
	pubkeyStreamOldestPrefix = byte('O')
)

type System struct {
	*sdk.System
}

func Wrap(sys *sdk.System) *System {
	return &System{System: sys}
}

func makePubkeyStreamKey(prefix byte, pubkey nostr.PubKey) []byte {
	key := make([]byte, 1+8)
	key[0] = prefix
	copy(key[1:], pubkey[0:8])
	return key
}

func decodeTimestamp(data []byte) nostr.Timestamp {
	if len(data) < 8 {
		return 0
	}
	var v nostr.Timestamp
	for i := 0; i < 8; i++ {
		v = v<<8 | nostr.Timestamp(data[i])
	}
	return v
}

func encodeTimestamp(v nostr.Timestamp) []byte {
	data := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		data[i] = byte(v)
		v >>= 8
	}
	return data
}

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

	relays := sys.System.FallbackRelays.URLs
	if len(relays) == 0 {
		relays = []string{"wss://relay.damus.io", "wss://nos.lol"}
	}

	events := make([]nostr.Event, 0, limit)
	for ie := range sys.System.Pool.FetchMany(ctx, relays, filter, nostr.SubscriptionOptions{Label: "global"}) {
		sys.System.Publisher.Publish(ctx, ie.Event)
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

func (sys *System) FetchMyTimelinePage(
	ctx context.Context,
	pubkey nostr.PubKey,
	limit int,
	until nostr.Timestamp,
) ([]nostr.Event, error) {
	oldestKey := makePubkeyStreamKey(pubkeyStreamOldestPrefix, pubkey)
	var oldestTimestamp nostr.Timestamp

	if data, _ := sys.System.KVStore.Get(oldestKey); data != nil {
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
		for evt := range sys.System.Store.QueryEvents(filter, limit) {
			events = append(events, evt)
			count++
			if count >= limit {
				slices.SortFunc(events, nostr.CompareEventReverse)
				return events, nil
			}
		}
	}

	relays := sys.System.FetchOutboxRelays(ctx, pubkey, 2)
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

	for ie := range sys.System.Pool.FetchMany(ctx, relays, filter, nostr.SubscriptionOptions{Label: "mytimeline"}) {
		sys.System.Publisher.Publish(ctx, ie.Event)
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

	sys.System.KVStore.Set(oldestKey, encodeTimestamp(oldestTimestamp))
	slices.SortFunc(events, nostr.CompareEventReverse)
	return events, nil
}

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
		if data, _ := sys.System.KVStore.Get(oldestKey); data != nil {
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
				for evt := range sys.System.Store.QueryEvents(filter, limitPerKey) {
					localEvents = append(localEvents, evt)
					count++
					if count >= limitPerKey {
						break
					}
				}
			}

			relays := sys.System.FetchOutboxRelays(ctx, pk, 2)
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

			for ie := range sys.System.Pool.FetchMany(ctx, relays, filter, nostr.SubscriptionOptions{Label: "followed"}) {
				sys.System.Publisher.Publish(ctx, ie.Event)
				if ie.Event.CreatedAt < until || until == 0 {
					localEvents = append(localEvents, ie.Event)
				}
				if ie.Event.CreatedAt < newOldest || newOldest == 0 {
					newOldest = ie.Event.CreatedAt
				}
			}

			sys.System.KVStore.Set(makePubkeyStreamKey(pubkeyStreamOldestPrefix, pk), encodeTimestamp(newOldest))

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

			relays := sys.System.FetchOutboxRelays(ctx, nostr.PubKey{}, 2)
			if len(relays) == 0 {
				relays = []string{"wss://relay.damus.io", "wss://nos.lol", "wss://relay.nostr.band"}
			}

			count := 0
			for ie := range sys.System.Pool.FetchMany(ctx, relays, filter, nostr.SubscriptionOptions{Label: "community"}) {
				sys.System.Publisher.Publish(ctx, ie.Event)
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

func (sys *System) FetchProfilesBatch(
	ctx context.Context,
	pubkeys []nostr.PubKey,
) map[nostr.PubKey]*nostr.Event {
	results := make(map[nostr.PubKey]*nostr.Event)
	for _, pk := range pubkeys {
		pm := sys.System.FetchProfileMetadata(ctx, pk)
		if pm.Event != nil {
			results[pk] = pm.Event
		}
	}
	return results
}

func (sys *System) FetchEventByFilter(
	ctx context.Context,
	filter nostr.Filter,
	timeoutMs int,
) *nostr.Event {
	pointer := filterToPointer(filter)
	if pointer == nil {
		return nil
	}

	params := sdk.FetchSpecificEventParameters{}
	if timeoutMs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
		defer cancel()
	}

	evt, _, _ := sys.System.FetchSpecificEvent(ctx, pointer, params)
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

func (sys *System) FetchNote(ctx context.Context, noteID string, timeoutMs int) *nostr.Event {
	id, err := nostr.IDFromHex(noteID)
	if err != nil {
		return nil
	}
	pointer := nostr.EventPointer{ID: id}
	evt, _, _ := sys.FetchSpecificEvent(ctx, pointer, sdk.FetchSpecificEventParameters{})
	return evt
}

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

	evt, _, _ := sys.FetchSpecificEvent(ctx, ptr, sdk.FetchSpecificEventParameters{})
	return evt
}