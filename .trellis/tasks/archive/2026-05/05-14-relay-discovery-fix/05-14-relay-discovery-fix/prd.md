# relay-discovery-fix

## Goal

Fix relay discovery for profile and event fetch by implementing missing event relay hint extraction and improving fallback behavior.

## What I Already Know

* `ExtractRelayHints` is defined in spec (`relay-guidelines.md:25-38`) but NOT implemented in code
* `DiscoverAndVerifyRelays` is defined in spec but NOT implemented
* `DiscoverUserRelays` (in `utils/user_relays.go`) works but returns `nil, nil` silently on failure — no fallback retry
* `GetProfile` (`utils/get.go:100-144`) calls `DiscoverUserRelays` but swallows errors; if user relays are not found, falls back to `AllReadableRelays()` only
* `GetNote`/`GetNoteAsync` fetches by event ID only, does not extract relay hints from `e` tags
* Profile fetch does not use `p` tag relay hints from events referencing the profile owner

## Requirements

* **Implement `ExtractRelayHints`** in `utils/get.go`:
  - Extract relay URLs from `e`, `p`, `a`, `q` tags (len >= 3, tag[2] non-empty)
  - Return deduplicated slice of relay URLs
* **`GetNote`/`GetNoteAsync` should use e-tag relay hints**:
  - When fetching an event by ID, extract relay hints from any `e` tags in the query context
  - Query hinted relays first, fallback to `AllReadableRelays()`
* **`GetProfile` should use p-tag relay hints**:
  - When fetching profile metadata, extract relay hints from `p` tags in events referencing that profile owner
  - Query hinted relays first, fallback to `AllReadableRelays()`
* **Improve fallback when `DiscoverUserRelays` returns nil**:
  - Current: silently falls back to `AllReadableRelays()`
  - Better: also try `KnownRelays` as second fallback before giving up
* **Error propagation**:
  - If `DiscoverUserRelays` fails with a real error (not just "no event found"), log it

## Acceptance Criteria

* [ ] `ExtractRelayHints(event)` returns relay URLs from `e/p/a/q` tags with relay field
* [ ] `GetNote(noteID, opts)` queries hinted relays from `e` tag first, then `AllReadableRelays()`
* [ ] `GetProfile(pubKey, opts)` queries hinted relays from `p` tags first, then `DiscoverUserRelays`, then `AllReadableRelays()`
* [ ] `DiscoverUserRelays` logs errors when they occur (not just swallowing them)
* [ ] `go build ./...` succeeds
* [ ] `go vet ./...` passes

## Definition of Done

* `ExtractRelayHints` implemented in `utils/get.go`
* Relay hints used in `GetNote` and `GetProfile`
* Better error handling in `DiscoverUserRelays`
* Build and vet pass

## Out of Scope

* `DiscoverAndVerifyRelays` implementation (different task)
* NIP-65 relay list publishing to network
* Changes to local relay caching behavior

## Technical Notes

* Files to modify: `utils/get.go`, `utils/user_relays.go`
* NIP-01 event tag formats:
  - `e`: `["e", <id>, <relay>, <marker>, <pubkey>]`
  - `p`: `["p", <pubkey>, <relay>]`
  - `a`: `["a", "<kind>:<pubkey>:<d>", <relay>]`
  - `q`: `["q", <event-id>, <relay>]`
* The relay field (tag[2] for e/p/a/q) is optional — use it when present
* Deduplication needed when multiple tags point to same relay