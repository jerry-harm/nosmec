# switch database backends to lmdb

## Goal

Switch nosmec's SDK-managed persistent databases from bbolt back to LMDB so multiple nosmec processes do not block each other on `hints` / `kvstore` access, and update docs to match the new storage layout.

## What I already know

* `config/config.go` currently opens `hints.db` via `nostr_sdk/hints/bbolth`.
* `config/config.go` currently opens `kvstore.db` via `nostr_sdk/kvstore/bbolt`.
* Both LMDB implementations already exist in-repo: `nostr_sdk/hints/lmdbh` and `nostr_sdk/kvstore/lmdb`.
* `event store` is currently separate from the SDK stores: `nosmec_events.db` uses `fiatjaf.com/nostr/eventstore/boltdb`, and search uses Bleve at `search_index/`.
* Current docs/specs still describe BoltDB-backed hints / kvstore in several places.
* User constraint: we do not want to fork the entire `fiatjaf.com/nostr` module; only the local `nostr_sdk` fork is acceptable.

## Assumptions (temporary)

* The main desired fix is to eliminate multi-process blocking for SDK-managed stores (`HintsDB` and `KVStore`).
* We should keep existing higher-level behavior unchanged and minimize migration surface.
* Any event-store LMDB solution should avoid forking the full upstream `nostr` module.

## Open Questions

* None.

## Requirements (evolving)

* Use LMDB-backed implementations for the SDK `HintsDB` and `KVStore`.
* User preference is to move all persistent databases to LMDB, including the event store, if feasible without forking the entire `nostr` module.
* Event-store strategy: adopt upstream `fiatjaf.com/nostr/eventstore/lmdb` directly.
* Existing on-disk BoltDB data will be ignored; after switching, LMDB stores start fresh instead of migrating or dual-reading old BoltDB files.
* Update user-facing and Trellis docs to describe the actual storage backends and paths.
* Preserve existing application behavior outside backend selection.
* Avoid forking the full `fiatjaf.com/nostr` module.

## Decision (ADR-lite)

**Context**: The user wants all persistent databases moved to LMDB, but does not want to fork the entire `fiatjaf.com/nostr` module.

**Decision**: Reuse the upstream `fiatjaf.com/nostr/eventstore/lmdb` backend for the event store, and switch the SDK-managed `HintsDB` and `KVStore` to the in-repo LMDB implementations.

**Consequences**: We can move all persistent stores to LMDB without a broad upstream fork. On-disk layout changes from single-file BoltDB paths to LMDB environment directories, so migration/cleanup behavior must be decided explicitly.

## Acceptance Criteria (evolving)

* [ ] `config/config.go` initializes `HintsDB` from `nostr_sdk/hints/lmdbh`.
* [ ] `config/config.go` initializes `KVStore` from `nostr_sdk/kvstore/lmdb`.
* [ ] Event-store persistence backend uses upstream `fiatjaf.com/nostr/eventstore/lmdb` instead of `eventstore/boltdb`.
* [ ] Relevant docs/specs no longer claim hints / kvstore are bbolt-backed.
* [ ] Docs note that old BoltDB data is not migrated and new LMDB stores rebuild from scratch.
* [ ] Build/tests covering affected code pass.

## Research References

* [`research/upstream-eventstore-lmdb.md`](research/upstream-eventstore-lmdb.md) — pinned upstream `fiatjaf.com/nostr` version already contains `eventstore/lmdb`, and Bleve can wrap it like the current Bolt backend.

## Definition of Done (team quality bar)

* Tests added/updated (unit/integration where appropriate)
* Lint / typecheck / CI green
* Docs/notes updated if behavior changes
* Rollout/rollback considered if risky

## Out of Scope (explicit)

* Forking the entire `fiatjaf.com/nostr` module
* Large SDK refactors beyond backend wiring

## Technical Notes

* Current `HintsDB`: `config/config.go` -> `GlobalHints()` -> `bbolth.NewBoltHints(filepath.Join(dataDir, "hints.db"))`
* Current `KVStore`: `config/config.go` -> `GlobalPool()` -> `kvstore_bbolt.NewStore(filepath.Join(dataDir, "kvstore.db"))`
* Current event store: `config/config.go` -> `boltdb.BoltBackend{Path: filepath.Join(dataDir, "nosmec_events.db")}` wrapped by `bleve.BleveBackend{Path: filepath.Join(dataDir, "search_index")}` when Bleve init succeeds.
* Research confirmed the pinned upstream module already contains `fiatjaf.com/nostr/eventstore/lmdb` and Bleve test coverage wraps LMDB as `RawEventStore`.
* LMDB paths are directory-based (environment directory), unlike the current single-file BoltDB paths.
