# migrate access.System to nostr/sdk.System

## Goal

Replace the custom `access.System` + local relay architecture with `fiatjaf.com/nostr/sdk.System`, which already provides in-memory caches, local event store, relay discovery, and store-aware queries — eliminating redundant network requests at the architecture level.

## Why

Our current architecture has three layers that don't talk to each other:
1. `access.System` — our custom cache layer (just added)
2. Local khatru relay — event storage (used by CacheEvent but never queried)
3. Network queries — all go directly to relays

`sdk.System` unifies all three:
- `MetadataCache` — in-memory profile cache (with TTL)
- `Store` (eventstore) — local database, queried before network
- `FetchProfileMetadata` — Cache → Store → Network automatically
- `FetchFeedPage` — Store-aware pagination (skips network if local has data)

## Pre-MVP Requirements

[ ] Write PRD (this file)
[ ] Research sdk.System API thoroughly
[ ] Create migration plan with phases
[ ] User approval

## Current Architecture

```
local khatru relay (ws://localhost:8989)
    ├── BoltDB + Bleve eventstore
    └── CacheEvent writes events here (async)
        └── NEVER QUERIED BY GetProfile/GetEvent!

access.System (our custom layer)
    ├── Pool *nostr.Pool
    ├── MetadataCache (Ristretto, 5min TTL)
    ├── RelayListCache / FollowListCache
    └── TrackEventRelay / GetEventRelay (KVStore)

Network queries (all directly to relays)
    ├── GetProfile → QuerySingle (all relays)
    ├── GetEvent → QuerySingle
    ├── GetTimeline → FetchMany
    └── No local store fallback
```

## Target Architecture

```
sdk.System (unified, single instance)
    ├── Pool *nostr.Pool
    ├── Store eventstore.Store (BoltDB + Bleve)
    │   ├── SaveEvent (on every incoming event)
    │   ├── QueryEvents (local store fallback)
    │   └── ReplaceEvent (replaceable kinds)
    ├── Hints sdk_hints.HintsDB
    ├── KVStore kvstore.KVStore (event→relay tracking)
    ├── MetadataCache (in-memory, TTL)
    ├── RelayListCache / FollowListCache / ...
    ├── RelayStreams (MetadataRelays, FollowListRelays, FallbackRelays)
    └── Methods
        ├── FetchProfileMetadata → Cache → Store → Network
        ├── FetchSpecificEvent → Store → Known → Fallback
        ├── FetchFeedPage → Store-aware pagination
        ├── FetchRelayList / FetchFollowList
        ├── FetchOutboxRelays / FetchInboxRelays
        ├── PrepareNoteEvent (write relay selection)
        ├── TrackEventHintsAndRelays (EventMiddleware)
        └── GetEventRelays ([]string, not single string)
```

## Migration Phases

### Phase 1: sdk.System initialization (config/config.go)

Replace `access.GlobalSystem` with `sdk.GlobalSystem`:
- Create `sdk.NewSystem()` instead of `access.NewSystem(...)`
- Assign fields post-creation (sdk.NewSystem takes no args)
- Set up `sdk.System.Store` with BoltDB + Bleve eventstore
- Configure relay streams on sdk.System
- Set up `sdk.System.KVStore` for event→relay tracking
- Wire `sdk.System.Pool`

Files: `config/config.go`, `config/context.go`

### Phase 2: Event middleware (config/config.go)

Replace our EventMiddleware with sdk's `TrackEventHintsAndRelays`:
- Every incoming event saves to Store (automatic)
- Every incoming event tracks hints + relays
- No more `CacheEvent` or `shouldCache` needed

Files: `config/config.go`

### Phase 3: Replace query functions (utils/get.go)

| Old | New | Notes |
|-----|-----|-------|
| `GetProfile` | `sdk.FetchProfileMetadata` | Returns `sdk.ProfileMetadata` not `*nostr.Event` |
| `GetProfileAsync` | Same as above | SDK handles async internally |
| `GetProfiles` | Batch `FetchProfileMetadata` | Need to batch since SDK does single-pubkey |
| `GetEvent` | `sdk.FetchSpecificEvent` | Store → Known → Fallback |
| `GetNote` | `sdk.FetchSpecificEvent` with note pointer | |
| `GetNoteAsync` | Same | |
| `GetTimeline` (global/mine/followed) | `sdk.FetchFeedPage` | Store-aware pagination |
| `GetCommunityPosts` | `sdk.FetchFeedPage` with community filter | |
| `SubscribeWithCache` | Remove | SDK handles via Store |

Files: `utils/get.go`

### Phase 4: Update TUI layer

Replace all `utils.GetProfile` / `utils.GetProfileName` calls:
- `tui/timeline/model.go` — `fetchProfileNames` → `sdk.FetchProfileMetadata`
- `tui/thread/thread.go` — `fetchProfileNames` → `sdk.FetchProfileMetadata`
- `tui/event/event.go` — `fetchEventAsync` → `sdk.FetchSpecificEvent`
- `tui/dm/model.go` — `fetchRecipientProfileNameAsync` → `sdk.FetchProfileMetadata`
- `tui/component/label/model.go` — `RenderLabel` profile fetch

Files: `tui/timeline/model.go`, `tui/thread/thread.go`, `tui/event/event.go`, `tui/dm/model.go`, `tui/component/label/model.go`

### Phase 5: Remove local relay

Delete:
- `config.StartLocalRelay()`
- `config.cacheEvent()` / `config.shouldCacheEvent()`
- `utils.CacheEvent()` / `utils.shouldCache()`
- `utils.SubscribeWithCache()`
- `config.CacheFilters` type and config
- Local relay server (khatru)

The SDK's `Store` replaces all of this.

Files: `config/config.go`, `utils/get.go`

### Phase 6: Clean up access/ package

Reduce or delete:
- `access/system.go` — replaced by sdk.System
- `access/relay_stream.go` — replaced by sdk.RelayStream
- `access/kvstore.go` — inline to config
- `access/system_test.go` / `access/relay_stream_test.go` — remove or rewrite

Keep (as utility wrappers):
- `config.TrackEventRelay()` / `config.GetEventRelay()` — thin wrappers over sdk.System

### Phase 7: Update config types

- `config.AppContext.sys` → change type from `*access.System` to `*sdk.System`
- `config.AppContext.System()` → change return type
- Remove `config.CacheFilter` type
- Remove `cache_filters` from config

## Key Design Decisions

### 1. Event→relay tracking semantics
sdk uses `TrackEventRelaysD(relay, id)` which appends (multi-relay). Our code uses `TrackEventRelay(id, relay)` which is first-write-wins (single relay).

Decision: Switch to sdk's multi-relay tracking. `GetEventRelay` becomes `GetEventRelays` returning `[]string`. Callers in `utils/post.go` pick first relay for NIP-10 hints.

### 2. Profile return type
sdk returns `sdk.ProfileMetadata` (struct with Name, DisplayName, Picture, etc.), not raw `*nostr.Event`.

Decision: Adapt callers. `sdk.ProfileMetadata.ShortName()` gives the best display name (Name → DisplayName → short npub).

### 3. Batch profile fetching
`FetchProfileMetadata` fetches one pubkey. We need batch for timeline/thread N+1 problem.

Decision: Wrap in a helper that calls `FetchProfileMetadata` in parallel goroutines. The SDK's MetadataCache + Store handles deduplication.

### 4. Local relay removal
The local relay serves two purposes: (a) offline cache, (b) search (Bleve).

Decision: `sdk.System.Store` handles (a). For (b), keep a searchable Store (Bleve-backed). `sdk.System` doesn't have a search method, so we may need `utils.SearchEvents` to query the local Store directly for NIP-50.

### 5. Timeline pagination
Our `GetTimeline` uses `FetchMany` (streaming). sdk's `FetchFeedPage` is paginated with `until` + `totalLimit`.

Decision: Refactor TUI to use paginated API instead of streaming. The SDK approach is more efficient (store-aware).

## Risk Areas

1. **ProfileMetadata vs nostr.Event** — many callers expect `*nostr.Event`. Need to update all callers.
2. **Search** — sdk.System doesn't have NIP-50 search. Need direct Store query.
3. **DM** — NIP-17 gift wrap logic is custom, won't be replaced by SDK.
4. **Community** — Kind 34550 queries are custom.
5. **Compose publish** — `PostNote`/`ReplyNote` use PublishMany, which sdk supports via `Pool.PublishMany`.

## Research References

- `research/access-system-migration.md` — Full migration surface audit
- `research/network-calls.md` — Network call inventory

## Out of Scope

- DM protocol redesign (NIP-17 stays as is)
- Community discover (kind 34550 stays as is)
- NIP-50 search redesign
- TUI layout changes
