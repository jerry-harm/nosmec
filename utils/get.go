package utils

import (
	"context"
	"encoding/json"
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
			CacheEvent(&ev, opts.App)
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

func SubscribeWithCache(ctx context.Context, pool *nostr.Pool, relays []string, filter nostr.Filter, opts nostr.SubscriptionOptions, app *config.AppContext) chan nostr.RelayEvent {
	ch := pool.SubscribeMany(ctx, relays, filter, opts)
	out := make(chan nostr.RelayEvent)
	go func() {
		for ie := range ch {
			CacheEvent(&ie.Event, app)
			out <- ie
		}
		close(out)
	}()
	return out
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

func CacheEvent(event *nostr.Event, app *config.AppContext) {
	if event == nil || app == nil {
		return
	}
	if !shouldCache(event, app) {
		return
	}
	privateRelays := app.PrivateRelays()
	if len(privateRelays) == 0 {
		return
	}
	go func() {
		app.Pool().PublishMany(context.Background(), privateRelays, *event)
	}()
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

func GetMyTimeline(ctx context.Context, limit int, until nostr.Timestamp, opts *GetOptions) chan *nostr.Event {
	if opts == nil || opts.App == nil {
		ch := make(chan *nostr.Event)
		close(ch)
		return ch
	}

	pubKey, err := opts.App.GetMyPubKey()
	if err != nil {
		ch := make(chan *nostr.Event)
		close(ch)
		return ch
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

	out := make(chan *nostr.Event)
	go func() {
		ch := opts.App.Pool().FetchMany(ctx, relays, filter, nostr.SubscriptionOptions{})
		for relayEvent := range ch {
			CacheEvent(&relayEvent.Event, opts.App)
			out <- &relayEvent.Event
		}
		close(out)
	}()
	return out
}

func GetGlobalTimeline(ctx context.Context, limit int, until nostr.Timestamp, opts *GetOptions) chan *nostr.Event {
	if opts == nil || opts.App == nil {
		ch := make(chan *nostr.Event)
		close(ch)
		return ch
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

	out := make(chan *nostr.Event)
	go func() {
		ch := opts.App.Pool().FetchMany(ctx, relays, filter, nostr.SubscriptionOptions{})
		for relayEvent := range ch {
			CacheEvent(&relayEvent.Event, opts.App)
			out <- &relayEvent.Event
		}
		close(out)
	}()
	return out
}

type TimelineEvent struct {
	Event       nostr.Event
	CommunityID string
	IsCommunity bool
}

func GetFollowedTimeline(ctx context.Context, limit int, until nostr.Timestamp, hashtags []string, opts *GetOptions) chan *nostr.Event {
	if opts == nil || opts.App == nil {
		ch := make(chan *nostr.Event)
		close(ch)
		return ch
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

	for _, sub := range subs {
		switch sub.Type {
		case "user":
			pk, err := ResolveAliasToPubKey(opts.App, sub.ID)
			if err == nil {
				authors = append(authors, pk)
			}
		case "community":
			communityAddrs = append(communityAddrs, sub.ID)
		case "hashtag":
			hashtags = append(hashtags, sub.ID)
		}
	}

	out := make(chan *nostr.Event)
	go func() {
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

			seen := make(map[nostr.ID]bool)
			ch := opts.App.Pool().FetchMany(ctx, relays, filter, nostr.SubscriptionOptions{})
			for relayEvent := range ch {
				if seen[relayEvent.Event.ID] {
					continue
				}
				seen[relayEvent.Event.ID] = true

				if len(hashtags) > 0 {
					hasMatch := false
					for _, tag := range relayEvent.Event.Tags {
						if len(tag) >= 2 && tag[0] == "t" {
							for _, h := range hashtags {
								if tag[1] == h {
									hasMatch = true
									break
								}
							}
						}
						if hasMatch {
							break
						}
					}
					if !hasMatch {
						continue
					}
				}

				CacheEvent(&relayEvent.Event, opts.App)
				out <- &relayEvent.Event
			}
		}
		close(out)
	}()
	return out
}