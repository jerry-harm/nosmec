package utils

import (
	"context"
	"time"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/sdk"
	"github.com/jerry-harm/nosmec/config"
)

type GetOptions struct {
	App    *config.AppContext
	Relays []string
}

// ExtractRelayHints extracts relay URLs from e/p/a/q tags in an event.
// The relay field (tag[2]) is used when present. Results are deduplicated.
func ExtractRelayHints(event *nostr.Event) []string {
	if event == nil {
		return nil
	}
	seen := make(map[string]bool)
	var relays []string
	for _, tag := range event.Tags {
		if len(tag) < 3 {
			continue
		}
		switch tag[0] {
		case "e", "p", "a", "q":
			if relay := tag[2]; relay != "" && !seen[relay] {
				relays = append(relays, relay)
				seen[relay] = true
			}
		}
	}
	return relays
}

func GetEvent(ctx context.Context, filter nostr.Filter, opts *GetOptions) *nostr.Event {
	if opts == nil || opts.App == nil {
		return nil
	}

	relays := opts.Relays
	if len(relays) == 0 {
		relays = opts.App.AllReadableRelays()
	}

	localURL := config.GetLocalRelayURL()
	hasLocal := localURL != ""

	var event *nostr.Event

	if isReplaceableKind(filter.Kinds) {
		if hasLocal {
			ctxLocal, cancelLocal := context.WithTimeout(ctx, 2*time.Second)
			defer cancelLocal()
			result := opts.App.Pool().QuerySingle(ctxLocal, []string{localURL}, filter, nostr.SubscriptionOptions{})
			if result != nil && result.Event.ID != [32]byte{} {
				event = &result.Event
				CacheEvent(event, opts.App)
			}
		}

		if event == nil {
			timeout := opts.App.QueryTimeout()
			ctxQuery, cancelQuery := context.WithTimeout(ctx, timeout)
			defer cancelQuery()
			results := opts.App.Pool().FetchManyReplaceable(ctxQuery, relays, filter, nostr.SubscriptionOptions{})
			results.Range(func(key nostr.ReplaceableKey, ev nostr.Event) bool {
				event = &ev
				CacheEvent(&ev, opts.App)
				return false
			})
		}
	} else {
		if hasLocal {
			ctxLocal, cancelLocal := context.WithTimeout(ctx, 2*time.Second)
			defer cancelLocal()
			result := opts.App.Pool().QuerySingle(ctxLocal, []string{localURL}, filter, nostr.SubscriptionOptions{})
			if result != nil && result.Event.ID != [32]byte{} {
				event = &result.Event
			}
		}

		if event == nil {
			timeout := opts.App.QueryTimeout()
			ctxQuery, cancelQuery := context.WithTimeout(ctx, timeout)
			defer cancelQuery()
			result := opts.App.Pool().QuerySingle(ctxQuery, relays, filter, nostr.SubscriptionOptions{})
			if result != nil {
				event = &result.Event
				CacheEvent(&result.Event, opts.App)
			}
		}
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
	localURL := config.GetLocalRelayURL()
	if localURL == "" {
		return
	}
	go func() {
		app.Pool().PublishMany(context.Background(), []string{localURL}, *event)
	}()
}

func GetProfile(ctx context.Context, pubKey nostr.PubKey, opts *GetOptions) *nostr.Event {
	if opts == nil || opts.App == nil {
		return nil
	}

	filter := BuildProfileFilter(pubKey)

	// Build combined relay list: AllReadableRelays + KnownRelays
	allRelays := opts.App.AllReadableRelays()
	knownRelays := opts.App.Config().KnownRelays
	seen := make(map[string]bool)
	combined := make([]string, 0, len(allRelays)+len(knownRelays))
	for _, r := range allRelays {
		if !seen[r] {
			combined = append(combined, r)
			seen[r] = true
		}
	}
	for _, r := range knownRelays {
		if !seen[r] {
			combined = append(combined, r)
			seen[r] = true
		}
	}

	if len(combined) == 0 {
		return nil
	}

	// Launch DiscoverUserRelays async to update KnownRelays for future use
	go func() {
		DiscoverUserRelays(context.Background(), opts.App, pubKey)
	}()

	// Query all relays in parallel, return first result
	ctxQuery, cancelQuery := context.WithTimeout(ctx, opts.App.QueryTimeout())
	defer cancelQuery()

	result := opts.App.Pool().QuerySingle(ctxQuery, combined, filter, nostr.SubscriptionOptions{})
	if result == nil || result.Event.ID == [32]byte{} {
		return nil
	}

	CacheEvent(&result.Event, opts.App)
	return &result.Event
}

func GetProfileName(ctx context.Context, pubKey nostr.PubKey, opts *GetOptions) string {
	if opts == nil || opts.App == nil {
		return ""
	}

	profile := GetProfile(ctx, pubKey, opts)
	return extractProfileName(profile)
}

func GetNote(ctx context.Context, noteID string, opts *GetOptions) *nostr.Event {
	filter, err := BuildNoteFilter(noteID)
	if err != nil {
		return nil
	}
	return GetEvent(ctx, filter, opts)
}

func GetNoteAsync(ctx context.Context, noteID string, opts *GetOptions) *nostr.Event {
	filter, err := BuildNoteFilter(noteID)
	if err != nil {
		return nil
	}
	return GetEventAsync(ctx, filter, opts)
}

func GetParentEvent(ctx context.Context, event *nostr.Event, opts *GetOptions) *nostr.Event {
	if event == nil || opts == nil || opts.App == nil {
		return nil
	}

	replyTag := event.Tags.FindWithValue("e", "reply")
	if len(replyTag) < 2 {
		return nil
	}

	parentID := replyTag[1]

	filter, err := BuildParentEventFilter(parentID)
	if err != nil {
		return nil
	}

	var relays []string
	if len(replyTag) >= 3 && replyTag[2] != "" {
		relays = []string{replyTag[2]}
	}

	if len(relays) == 0 {
		relays = opts.App.AllReadableRelays()
	}

	if len(relays) == 0 {
		return nil
	}

	ctxQuery, cancelQuery := context.WithTimeout(ctx, opts.App.QueryTimeout())
	defer cancelQuery()

	result := opts.App.Pool().QuerySingle(ctxQuery, relays, filter, nostr.SubscriptionOptions{})
	if result == nil {
		return nil
	}

	parent := result.Event
	CacheEvent(&parent, opts.App)
	return &parent
}

func GetProfileNameAsync(ctx context.Context, pubKey nostr.PubKey, opts *GetOptions) string {
	if opts == nil || opts.App == nil {
		return ""
	}

	profile := GetProfileAsync(ctx, pubKey, opts)
	return extractProfileName(profile)
}

func GetProfileAsync(ctx context.Context, pubKey nostr.PubKey, opts *GetOptions) *nostr.Event {
	if opts == nil || opts.App == nil {
		return nil
	}

	filter := BuildProfileFilter(pubKey)

	allRelays := opts.App.AllReadableRelays()
	knownRelays := opts.App.Config().KnownRelays
	seen := make(map[string]bool)
	combined := make([]string, 0, len(allRelays)+len(knownRelays))
	for _, r := range allRelays {
		if !seen[r] {
			combined = append(combined, r)
			seen[r] = true
		}
	}
	for _, r := range knownRelays {
		if !seen[r] {
			combined = append(combined, r)
			seen[r] = true
		}
	}

	if len(combined) == 0 {
		return nil
	}

	go func() {
		DiscoverUserRelays(context.Background(), opts.App, pubKey)
	}()

	ctxQuery, cancelQuery := context.WithTimeout(ctx, opts.App.QueryTimeout())
	defer cancelQuery()

	result := opts.App.Pool().QuerySingle(ctxQuery, combined, filter, nostr.SubscriptionOptions{})
	if result == nil || result.Event.ID == [32]byte{} {
		return nil
	}

	return &result.Event
}

func GetProfiles(ctx context.Context, pubKeys []nostr.PubKey, opts *GetOptions) map[nostr.PubKey]*nostr.Event {
	if opts == nil || opts.App == nil || len(pubKeys) == 0 {
		return nil
	}

	filter := BuildProfilesFilter(pubKeys)

	relays := opts.Relays
	if len(relays) == 0 {
		relays = opts.App.AllReadableRelays()
	}
	if len(relays) == 0 {
		return nil
	}

	ctxQuery, cancelQuery := context.WithTimeout(ctx, opts.App.QueryTimeout())
	defer cancelQuery()

	results := opts.App.Pool().FetchManyReplaceable(ctxQuery, relays, filter, nostr.SubscriptionOptions{})

	profiles := make(map[nostr.PubKey]*nostr.Event)
	results.Range(func(key nostr.ReplaceableKey, ev nostr.Event) bool {
		profiles[ev.PubKey] = &ev
		CacheEvent(&ev, opts.App)
		return true
	})

	return profiles
}

func GetProfileNames(ctx context.Context, pubKeys []nostr.PubKey, opts *GetOptions) map[nostr.PubKey]string {
	profiles := GetProfiles(ctx, pubKeys, opts)
	if len(profiles) == 0 {
		return nil
	}

	names := make(map[nostr.PubKey]string)
	for pk, profile := range profiles {
		names[pk] = extractProfileName(profile)
	}

	return names
}

func extractProfileName(profile *nostr.Event) string {
	if profile == nil {
		return ""
	}

	pm, err := sdk.ParseMetadata(*profile)
	if err == nil && pm.Name != "" {
		return pm.Name
	}

	return ""
}

func GetEventAsync(ctx context.Context, filter nostr.Filter, opts *GetOptions) *nostr.Event {
	if opts == nil || opts.App == nil {
		return nil
	}

	relays := opts.Relays
	if len(relays) == 0 {
		relays = opts.App.AllReadableRelays()
	}

	timeout := opts.App.QueryTimeout()
	ctxTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var event *nostr.Event

	if isReplaceableKind(filter.Kinds) {
		results := opts.App.Pool().FetchManyReplaceable(ctxTimeout, relays, filter, nostr.SubscriptionOptions{})
		results.Range(func(key nostr.ReplaceableKey, ev nostr.Event) bool {
			event = &ev
			CacheEvent(&ev, opts.App)
			return false
		})
	} else {
		result := opts.App.Pool().QuerySingle(ctxTimeout, relays, filter, nostr.SubscriptionOptions{})
		if result != nil {
			event = &result.Event
			CacheEvent(&result.Event, opts.App)
		}
	}

	return event
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

	filter := BuildTimelineFilter(pubKey, limit, until)

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
		relays = opts.App.AllReadableRelays()
	}
	if len(relays) == 0 {
		relays = opts.App.Config().KnownRelays
	}

	filter := BuildGlobalTimelineFilter(limit, until)

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
		relays = opts.App.AllReadableRelays()
	}
	if len(relays) == 0 {
		relays = opts.App.Config().KnownRelays
	}

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
			filter := BuildFollowedTimelineFilter(authors, communityAddrs, hashtags, limit, until)

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