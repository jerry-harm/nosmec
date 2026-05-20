# add relay list command

## Goal

Add a top-level `nosmec relay list` command that prints a simple deduplicated relay URL list collected from the new SDK-backed local databases, so the user can inspect what relay data is currently stored.

## What I already know

* The command should be `nosmec relay list`.
* Output should be a simple flat list, one relay URL per line.
* No counts, scores, or grouped sections are needed.
* Relay URLs should be collected from the new SDK storage only.
* The relay list should be built by deduplicating data from `HintsDB` and `sdk.System` KVStore.
* Old config-backed relay tracking should be removed from scope rather than supported.
* `KnownRelays` in config is no longer part of the intended design and stale docs need to be corrected.

## Requirements

* Add a top-level `relay` command group if one does not already exist.
* Add `nosmec relay list` under that command group.
* Collect relay URLs from all pubkey entries stored in `HintsDB`.
* Collect relay URLs from SDK event->relay tracking stored in `sdk.System` KVStore.
* Merge both sources and deduplicate relay URLs.
* Print a simple stable list suitable for manual database inspection.
* Do not read from or preserve old config-based event->relay helpers.
* Remove old config-based event->relay tracking code that is no longer wanted.
* Update stale task documentation/context that still refers to `KnownRelays` config persistence or old relay tracking behavior.

## Acceptance Criteria

* [ ] Running `nosmec relay list` prints unique relay URLs discovered from SDK `HintsDB` and SDK KVStore.
* [ ] Relay URLs are not duplicated even if present in both stores.
* [ ] Output is a simple ungrouped list without counts.
* [ ] Old config event->relay tracking helpers are removed from the implementation path.
* [ ] Documentation touched by this task no longer claims `KnownRelays` config storage is active for this flow.

## Definition of Done

* Tests added or updated for new relay enumeration helpers and command behavior where practical.
* Relevant verification commands pass.
* Task context files reference the correct SDK/database specs.

## Technical Approach

Implement SDK-native enumeration helpers instead of scraping debug output.

Likely shape:

* Add an SDK-facing way to enumerate all relays present in `HintsDB` across all pubkeys.
* Add an SDK-facing way to enumerate all event relay mappings from KVStore entries written by `nostr_sdk/event_relays.go`.
* Add a command handler that unions, sorts, and prints the combined relay list.
* Remove or stop using the older config-level event->relay helpers that store plain `eventID -> relayURL` mappings.

## Decision (ADR-lite)

**Context**: Relay URLs are currently learned in multiple storage layers. The user wants a database inspection command and explicitly does not want to preserve the old config-based relay tracking path.

**Decision**: Build `nosmec relay list` only on top of the new SDK persistence layers (`HintsDB` and SDK KVStore), print a flat deduplicated list, and remove the old config event->relay path from scope.

**Consequences**: The command reflects actual SDK-managed data and avoids prolonging migration glue. Some SDK store APIs may need small enumeration helpers because current public methods are optimized for lookup rather than full traversal.

## Out of Scope

* Per-source grouping in CLI output
* Relay score/count display
* Legacy config relay inspection flows
* Keeping backward compatibility for old config event->relay tracking

## Technical Notes

* `nostr_sdk/hints/interface.go` exposes per-pubkey APIs plus `PrintScores()`, but no current all-relay enumeration API.
* `nostr_sdk/hints/memoryh/db.go`, `nostr_sdk/hints/lmdbh/db.go`, and `nostr_sdk/hints/bbolth/db.go` each already iterate their full stores internally for debug printing; that traversal pattern can inform a proper enumeration API.
* `nostr_sdk/event_relays.go` stores SDK event relay mappings under a compact prefixed key format in KVStore.
* `nostr_sdk/kvstore/interface.go` currently supports point lookup/update only, so full KVStore relay enumeration may require a narrow extension or SDK-local helper over the chosen backend.
* Existing `config/config.go` docs and helpers still describe an older `eventID -> relayURL` approach and should not be treated as the target architecture for this task.
