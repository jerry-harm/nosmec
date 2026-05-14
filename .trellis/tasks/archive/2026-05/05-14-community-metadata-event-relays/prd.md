# community-metadata-event-relays

## Goal

Fetch community metadata using event-provided relay hints when available.

## What I Already Know

* `GetCommunity` fetches community definition event
* `GetCommunityPosts` fetches posts using community address
* Community events may have relay hints in their tags

## Requirements

* Modify `GetCommunity` to use relay hints from community definition event if available
* Modify `GetCommunityPosts` to use relay hints from community definition event
* Fallback to `KnownRelays`/`AllReadableRelays` when no hints available

## Acceptance Criteria

* [ ] Community metadata fetch uses relay hints from event
* [ ] Community posts fetch uses relay hints from event
* [ ] `go build ./...` succeeds
* [ ] `go vet ./...` passes

## Definition of Done

* Community fetches use event relay hints
* Build passes