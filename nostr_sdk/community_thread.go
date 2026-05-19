package nostr_sdk

import (
	"context"

	"fiatjaf.com/nostr"
)

// FetchEventByIDInScope fetches a specific event ID constrained to a community
// scope, using the provided relays as priority hints.
func (sys *System) FetchEventByIDInScope(ctx context.Context, id nostr.ID, relays []string, scope string) (*nostr.Event, []string, error) {
	pointer := nostr.EventPointer{ID: id, Relays: relays}
	return sys.FetchSpecificEventInScope(ctx, pointer, scope, FetchSpecificEventParameters{})
}

// FetchRootEventInScope fetches a root event ID constrained to a community
// scope. It is a semantic wrapper over FetchEventByIDInScope for thread use.
func (sys *System) FetchRootEventInScope(ctx context.Context, rootID nostr.ID, relays []string, scope string) (*nostr.Event, []string, error) {
	return sys.FetchEventByIDInScope(ctx, rootID, relays, scope)
}

// FetchParentInScope fetches the immediate parent event and filters it by the
// provided community scope.
func (sys *System) FetchParentInScope(ctx context.Context, event *nostr.Event, scope string, timeoutMs int) *nostr.Event {
	parent := sys.FetchParent(ctx, event, timeoutMs)
	if !MatchesCommunityScope(parent, scope) {
		return nil
	}
	return parent
}

// FetchParentChainInScope walks up the parent chain using NIP-10/NIP-22 parent
// semantics and stops when the scope no longer matches, the root is reached, or
// maxDepth is exceeded.
func (sys *System) FetchParentChainInScope(ctx context.Context, event *nostr.Event, scope string, timeoutMs int, maxDepth int) []*nostr.Event {
	if event == nil || maxDepth <= 0 {
		return nil
	}

	var chain []*nostr.Event
	seen := map[string]bool{event.ID.Hex(): true}
	current := event

	for depth := 0; depth < maxDepth; depth++ {
		ptr := GetThreadParentPointer(current)
		if ptr == nil {
			break
		}
		parentRef := ptr.AsTagReference()
		if parentRef == "" || seen[parentRef] {
			break
		}
		seen[parentRef] = true

		parent := sys.FetchParentInScope(ctx, current, scope, timeoutMs)
		if parent == nil {
			break
		}

		chain = append(chain, parent)
		current = parent

		rootID, isRoot, err := GetThreadRootID(parent)
		if err != nil || isRoot || rootID == parent.ID {
			break
		}
	}

	return chain
}

// FetchRepliesBreadthFirstInScope recursively fetches replies by e-tag links,
// breadth-first, constrained to the given community scope.
func (sys *System) FetchRepliesBreadthFirstInScope(ctx context.Context, rootID nostr.ID, relays []string, scope string, maxDepth int, batchSize int) []*nostr.Event {
	if rootID == (nostr.ID{}) || maxDepth <= 0 || batchSize <= 0 {
		return nil
	}

	var allEvents []*nostr.Event
	seen := map[string]bool{rootID.Hex(): true}
	queryIDs := []nostr.ID{rootID}

	for depth := 0; depth < maxDepth && len(queryIDs) > 0; depth++ {
		var nextIDs []nostr.ID

		for start := 0; start < len(queryIDs); start += batchSize {
			end := start + batchSize
			if end > len(queryIDs) {
				end = len(queryIDs)
			}

			events := sys.FetchEventsReferencingIDsInScope(ctx, queryIDs[start:end], relays, scope)
			for _, ev := range events {
				if seen[ev.ID.Hex()] {
					continue
				}
				seen[ev.ID.Hex()] = true
				allEvents = append(allEvents, ev)
				nextIDs = append(nextIDs, ev.ID)
			}
		}

		queryIDs = nextIDs
	}

	return allEvents
}
