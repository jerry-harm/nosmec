package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
)

type GetOptions struct {
	App    *config.AppContext
	Relays []string
}

func GetEvent(ctx context.Context, filter nostr.Filter, opts *GetOptions) *nostr.Event {
	if opts == nil || opts.App == nil {
		return nil
	}

	relays := opts.Relays
	if len(relays) == 0 {
		relays = opts.App.AllReadableRelays()
	}
	privateRelays := opts.App.PrivateRelays()

	var event *nostr.Event

	if isReplaceableKind(filter.Kinds) {
		results := opts.App.Pool().FetchManyReplaceable(ctx, relays, filter, nostr.SubscriptionOptions{})
		results.Range(func(key nostr.ReplaceableKey, ev nostr.Event) bool {
			event = &ev
			return false
		})
	} else {
		localURL := config.GetLocalRelayURL()
		remoteRelays := make([]string, 0, len(relays))
		for _, r := range relays {
			if r != localURL {
				remoteRelays = append(remoteRelays, r)
			}
		}

		ctxLocal, cancelLocal := context.WithTimeout(ctx, 2*time.Second)
		defer cancelLocal()
		if localURL != "" {
			result := opts.App.Pool().QuerySingle(ctxLocal, []string{localURL}, filter, nostr.SubscriptionOptions{})
			if result != nil {
				event = &result.Event
			}
		}
		if event == nil && len(remoteRelays) > 0 {
			result := opts.App.Pool().QuerySingle(ctx, remoteRelays, filter, nostr.SubscriptionOptions{})
			if result != nil {
				event = &result.Event
			}
		}
	}

	if event != nil && shouldCache(event, opts.App) {
		go func() {
			opts.App.Pool().PublishMany(context.Background(), privateRelays, *event)
		}()
	}

	return event
}

func QueryEventsCached(ctx context.Context, pool *nostr.Pool, relays []string,
	filter nostr.Filter, limit int, opts *GetOptions) ([]nostr.Event, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	eventMap := make(map[nostr.ID]nostr.Event)
	ch := pool.FetchMany(ctx, relays, filter, nostr.SubscriptionOptions{})

	for relayEvent := range ch {
		if _, ok := eventMap[relayEvent.Event.ID]; !ok {
			eventMap[relayEvent.Event.ID] = relayEvent.Event
		}
		if len(eventMap) >= limit*3 {
			break
		}
	}

	if len(eventMap) == 0 {
		return nil, nil
	}

	events := make([]nostr.Event, 0, len(eventMap))
	for _, e := range eventMap {
		events = append(events, e)
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].CreatedAt > events[j].CreatedAt
	})

	if len(events) > limit {
		events = events[:limit]
	}

	if opts != nil && opts.App != nil {
		privateRelays := opts.App.PrivateRelays()
		if len(privateRelays) > 0 {
			go func() {
				for _, e := range events {
					if shouldCache(&e, opts.App) {
						opts.App.Pool().PublishMany(context.Background(), privateRelays, e)
					}
				}
			}()
		}
	}

	return events, nil
}

func isReplaceableKind(kinds []nostr.Kind) bool {
	for _, k := range kinds {
		if k.IsReplaceable() || k.IsAddressable() {
			return true
		}
	}
	return false
}

func shouldCache(event *nostr.Event, app *config.AppContext) bool {
	filters := app.Config().CacheFilters
	for _, cf := range filters {
		if cf.ToNostr().Matches(*event) {
			return true
		}
	}
	return false
}

func GetProfile(ctx context.Context, pubKey nostr.PubKey, opts *GetOptions) *nostr.Event {
	filter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindProfileMetadata},
		Authors: []nostr.PubKey{pubKey},
		Limit:   1,
	}
	return GetEvent(ctx, filter, opts)
}

func GetProfileName(ctx context.Context, pubKey nostr.PubKey, opts *GetOptions) string {
	if opts == nil || opts.App == nil {
		return ""
	}

	profile := GetProfile(ctx, pubKey, opts)
	if profile == nil {
		return ""
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(profile.Content), &data); err == nil {
		if name, ok := data["name"].(string); ok && name != "" {
			return name
		}
	}

	for _, tag := range profile.Tags {
		if len(tag) >= 2 && tag[0] == "name" {
			return tag[1]
		}
	}

	return ""
}

func GetNote(ctx context.Context, noteID string, opts *GetOptions) *nostr.Event {
	var id nostr.ID
	if len(noteID) != 64 {
		return nil
	}
	copy(id[:], noteID)

	filter := nostr.Filter{
		IDs:   []nostr.ID{id},
		Limit: 1,
	}
	return GetEvent(ctx, filter, opts)
}

func GetMyTimeline(ctx context.Context, limit int, until nostr.Timestamp, opts *GetOptions) ([]nostr.Event, error) {
	if opts == nil || opts.App == nil {
		return nil, fmt.Errorf("nil options")
	}

	pubKey, err := opts.App.GetMyPubKey()
	if err != nil {
		return nil, err
	}

	relays := opts.Relays
	if len(relays) == 0 {
		relays = opts.App.AllReadableRelays()
	}
	privateRelays := opts.App.PrivateRelays()
	relays = append(relays, privateRelays...)

	filter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindTextNote},
		Authors: []nostr.PubKey{pubKey},
		Limit:   limit,
	}
	if until > 0 {
		filter.Until = until
	}

	return QueryEventsCached(ctx, opts.App.Pool(), relays, filter, limit, opts)
}

func GetGlobalTimeline(ctx context.Context, limit int, until nostr.Timestamp, opts *GetOptions) ([]nostr.Event, error) {
	if opts == nil || opts.App == nil {
		return nil, fmt.Errorf("nil options")
	}

	relays := opts.Relays
	if len(relays) == 0 {
		relays = opts.App.Config().KnownRelays
	}
	privateRelays := opts.App.PrivateRelays()
	relays = append(relays, privateRelays...)

	filter := nostr.Filter{
		Kinds: []nostr.Kind{nostr.KindTextNote},
		Limit: limit,
	}
	if until > 0 {
		filter.Until = until
	}

	return QueryEventsCached(ctx, opts.App.Pool(), relays, filter, limit, opts)
}

type TimelineEvent struct {
	Event       nostr.Event
	CommunityID string
	IsCommunity bool
}

func GetFollowedTimeline(ctx context.Context, limit int, until nostr.Timestamp, hashtags []string, opts *GetOptions) ([]TimelineEvent, error) {
	if opts == nil || opts.App == nil {
		return nil, fmt.Errorf("nil options")
	}

	subs := opts.App.ListSubscriptions("")

	relays := opts.Relays
	if len(relays) == 0 {
		relays = opts.App.Config().KnownRelays
	}
	privateRelays := opts.App.PrivateRelays()
	relays = append(relays, privateRelays...)

	var authors []nostr.PubKey
	var communityAddrs []string
	hashtagSet := make(map[string]bool)

	for _, sub := range subs {
		switch sub.Type {
		case "user":
			pk, err := ResolveAliasToPubKey(opts.App, sub.ID)
			if err == nil {
				authors = append(authors, pk)
			}
		case "community":
			communityAddrs = append(communityAddrs, sub.ID)
			hashtagSet[sub.ID] = true
		case "hashtag":
			hashtagSet[sub.ID] = true
		}
	}

	for _, tag := range hashtags {
		hashtagSet[tag] = true
	}

	var timelineEvents []TimelineEvent

	if len(authors) > 0 || len(communityAddrs) > 0 || len(hashtags) > 0 {
		kinds := []nostr.Kind{nostr.KindTextNote, nostr.KindComment}

		filter := nostr.Filter{
			Kinds: kinds,
			Limit: limit * 3,
		}
		if until > 0 {
			filter.Until = until
		}

		if len(authors) > 0 {
			filter.Authors = authors
		}

		if len(communityAddrs) > 0 {
			filter.Tags = nostr.TagMap{"a": communityAddrs}
		}

		events, err := QueryEventsCached(ctx, opts.App.Pool(), relays, filter, limit*3, opts)
		if err != nil {
			return nil, err
		}

		for _, event := range events {
			if len(hashtags) > 0 {
				hasMatch := false
				for _, tag := range event.Tags {
					if len(tag) >= 2 && tag[0] == "t" {
						for _, h := range hashtags {
							if tag[1] == h {
								hasMatch = true
								break
							}
						}
					}
				}
				if !hasMatch {
					continue
				}
			}

			te := TimelineEvent{Event: event}
			if event.Kind == nostr.KindComment {
				te.IsCommunity = true
				for _, tag := range event.Tags {
					if len(tag) >= 2 && tag[0] == "a" && len(tag[1]) > 6 && tag[1][:6] == "34550:" {
						te.CommunityID = tag[1]
						break
					}
				}
			}

			timelineEvents = append(timelineEvents, te)
		}

		sort.Slice(timelineEvents, func(i, j int) bool {
			return timelineEvents[i].Event.CreatedAt > timelineEvents[j].Event.CreatedAt
		})

		if len(timelineEvents) > limit {
			timelineEvents = timelineEvents[:limit]
		}
	}

	return timelineEvents, nil
}