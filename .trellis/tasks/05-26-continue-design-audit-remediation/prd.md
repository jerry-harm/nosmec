# continue design audit remediation

## Goal

Continue the 2026-05-26 design audit remediation by removing dead relay encoding code in `config`, and refactoring `relay list` to get relay data through `System` instead of opening LMDB directly in the command layer.

## What I already know

* Resource lifecycle and AppContext runtime ownership work has already been completed in earlier tasks.
* The remaining audit items are mostly around relay-layer duplication, command layering, and deferred SDK concerns.
* `config/config.go` still contains `encodeRelayListCompat` / `decodeRelayListCompat`, and they currently have no callers.
* `cmd/relay_commands.go` still opens LMDB directly and duplicates relay-list decoding in `decodeKVRelayList`.
* `nostr_sdk/event_relays.go` contains the only live relay-list codec currently used by the SDK event-relay tracking path.
* The user chose the next scope: delete dead code first, then refactor `relay list` to use `System` to fetch the list.

## Assumptions (temporary)

* We should avoid broad SDK refactors unless an outer-layer fix is clearly insufficient.
* `relay list` can be satisfied by a small SDK-facing read API instead of command-layer LMDB scanning.

## Open Questions

* What is the smallest `System` API shape that can expose the current relay list without leaking LMDB details upward?

## Requirements

* Delete dead relay codec helpers from `config/config.go`.
* Refactor `cmd/relay_commands.go` so `relay list` no longer opens LMDB or decodes KV relay bytes directly.
* Add or use a `nostr_sdk.System`-level read API for collecting the relay list used by `relay list`.
* Preserve the current user-facing `relay list` command behavior: print merged unique sorted relay URLs.
* Keep the change scoped to relay codec cleanup and relay list layering.

## Acceptance Criteria

* [ ] `encodeRelayListCompat` and `decodeRelayListCompat` are removed from `config/config.go`.
* [ ] `cmd/relay_commands.go` no longer imports LMDB or contains `readLMDB` / `decodeKVRelayList`-style direct storage parsing for `relay list`.
* [ ] `relay list` still prints the expected merged unique sorted relay URLs through a `System`-backed path.
* [ ] Relevant command / SDK tests pass.

## Decision (ADR-lite)

**Context**: The audit found one live relay-list codec in `nostr_sdk`, dead duplicate codec helpers in `config`, and command-layer LMDB parsing in `relay list`.

**Decision**: Treat the SDK path as the canonical owner for event-relay storage details, delete dead config helpers, and move relay-list data access behind a `System` API so the command layer no longer reads LMDB directly.

**Consequences**: This reduces duplication and layering violations without changing on-disk format or touching unrelated relay policy code.

## Out of Scope

* Changing the on-disk relay-list encoding format.
* Refactoring default relay policy / fallback strategy.
* Broader SDK cleanup outside the relay-list read path.

## Technical Notes

* Primary source: `.trellis/workspace/jerry/design-audit-2026-05-26.md`
* Likely code paths: `config/config.go`, `cmd/relay_commands.go`, `nostr_sdk/event_relays.go`, possibly `config/context.go` or `nostr_sdk/system.go` for API exposure.
