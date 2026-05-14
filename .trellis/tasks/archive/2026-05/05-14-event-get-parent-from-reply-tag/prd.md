# event-get-parent-from-reply-tag

## Goal

Add utility to extract parent event from reply tag and fetch it using relay hints from the tag.

## What I Already Know

* Reply events have `e` tag with format: `["e", <event-id>, <relay-url>, "reply"]`
* NIP-18 defines reply tag format
* Current `GetNote` fetches without using relay hints from tags
* `ExtractRelayHints` already exists and extracts relays from e/p/a/q tags

## Requirements

* Add `GetParentEvent(ctx, event, opts) *nostr.Event` to get parent from reply tag
* Extract relay from tag[2] if present
* Fetch parent using that relay first
* Fallback to general relays if no relay in tag or fetch fails

## Acceptance Criteria

* [ ] `GetParentEvent` returns parent event if present
* [ ] Uses relay hint from tag for faster fetch
* [ ] Falls back correctly when no relay in tag
* [ ] `go build ./...` succeeds
* [ ] `go vet ./...` passes

## Definition of Done

* Parent fetch works with relay hints
* Build passes