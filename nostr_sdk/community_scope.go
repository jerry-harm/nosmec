package nostr_sdk

import (
	"context"
	"strings"

	"fiatjaf.com/nostr"
)

const communityScopeKindPrefix = "34550:"

// ExtractCommunityScope returns the community root scope for an event.
// It prefers the NIP-22/NIP-72 uppercase A tag and falls back to legacy
// lowercase a tags that directly reference a kind:34550 community address.
func ExtractCommunityScope(event *nostr.Event) string {
	if event == nil {
		return ""
	}

	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "A" && strings.HasPrefix(tag[1], communityScopeKindPrefix) {
			return tag[1]
		}
	}

	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "a" && strings.HasPrefix(tag[1], communityScopeKindPrefix) {
			return tag[1]
		}
	}

	return ""
}

// MatchesCommunityScope reports whether an event belongs to the given
// community root scope. Empty scope means unscoped and matches everything.
func MatchesCommunityScope(event *nostr.Event, scope string) bool {
	if scope == "" {
		return true
	}

	return ExtractCommunityScope(event) == scope
}

// FetchSpecificEventInScope fetches a specific event and drops it if it does
// not belong to the requested community scope.
func (sys *System) FetchSpecificEventInScope(
	ctx context.Context,
	pointer nostr.Pointer,
	scope string,
	params FetchSpecificEventParameters,
) (event *nostr.Event, successRelays []string, err error) {
	event, successRelays, err = sys.FetchSpecificEvent(ctx, pointer, params)
	if event == nil || err != nil {
		return event, successRelays, err
	}
	if !MatchesCommunityScope(event, scope) {
		return nil, successRelays, nil
	}
	return event, successRelays, nil
}

// FetchEventsReferencingIDsInScope fetches note/comment events that reference
// any of the given event IDs via e tags, then filters them by community scope.
func (sys *System) FetchEventsReferencingIDsInScope(
	ctx context.Context,
	ids []nostr.ID,
	relays []string,
	scope string,
) []*nostr.Event {
	if len(ids) == 0 {
		return nil
	}

	if len(relays) == 0 {
		relays = sys.FallbackRelays.URLs
	}
	if len(relays) == 0 {
		return nil
	}

	idHexes := make([]string, 0, len(ids))
	for _, id := range ids {
		idHexes = append(idHexes, id.Hex())
	}

	filter := nostr.Filter{
		Kinds: []nostr.Kind{nostr.KindTextNote, nostr.KindComment},
		Tags:  nostr.TagMap{"e": idHexes},
	}

	var events []*nostr.Event
	seen := make(map[string]struct{})
	for evt := range sys.Store.QueryEvents(filter, len(ids)*32) {
		if !MatchesCommunityScope(&evt, scope) {
			continue
		}
		if _, ok := seen[evt.ID.Hex()]; ok {
			continue
		}
		eventCopy := evt
		events = append(events, &eventCopy)
		seen[evt.ID.Hex()] = struct{}{}
	}

	if len(relays) == 0 {
		return events
	}

	for ie := range sys.Pool.FetchMany(ctx, relays, filter, nostr.SubscriptionOptions{Label: "threadscope"}) {
		sys.Publisher.Publish(ctx, ie.Event)
		if !MatchesCommunityScope(&ie.Event, scope) {
			continue
		}
		if _, ok := seen[ie.Event.ID.Hex()]; ok {
			continue
		}
		eventCopy := ie.Event
		events = append(events, &eventCopy)
		seen[ie.Event.ID.Hex()] = struct{}{}
	}

	return events
}
