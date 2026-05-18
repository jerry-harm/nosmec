# Research: access.System → sdk.System Migration Surface

- **Query**: Map every place that uses `access.System` (directly or via `access.GlobalSystem`) and assess migration fit to `nostr/sdk.System`
- **Scope**: internal + external (sdk docs)
- **Date**: 2026-05-18

## Files Found

| File Path | Description |
|---|---|
| `access/system.go` | Defines `access.System`, `access.GlobalSystem`, `access.NewSystem`, and all methods |
| `access/relay_stream.go` | Defines `access.RelayStream` (round-robin URL iterator) |
| `access/kvstore.go` | Defines `access.NewKVStore` (BoltDB-backed KVStore factory) |
| `access/system_test.go` | Tests for `access.System` |
| `access/relay_stream_test.go` | Tests for `access.RelayStream` |
| `config/config.go` | Creates `access.GlobalSystem`, wraps `TrackEventRelay`/`GetEventRelay` |
| `config/context.go` | Holds `*access.System` in `AppContext`, exposes via `System()` |
| `utils/get.go` | Uses `access.GlobalSystem` for profile cache Get/Set |
| `utils/post.go` | Uses `config.GetEventRelay` (which delegates to `access.GlobalSystem`) |

---

## 1. access.System Definition

**File**: `access/system.go:22-38`

```go
type System struct {
    Pool  *nostr.Pool
    Hints sdk_hints.HintsDB
    Store kvstore.KVStore          // event→relay persistence

    ReadRelays     *RelayStream
    WriteRelays    *RelayStream
    DMRelays       *RelayStream
    SearchRelays   *RelayStream
    FallbackRelays *RelayStream

    localRelayURL string

    MetadataCache   sdk_cache.Cache32[sdk.ProfileMetadata]
    RelayListCache  sdk_cache.Cache32[sdk.GenericList[string, sdk.Relay]]
    FollowListCache sdk_cache.Cache32[sdk.GenericList[nostr.PubKey, sdk.ProfileRef]]
}
```

## 2. Every File:Line Reference (Grouped by Category)

### 2a. Type references (`access.System` as a type)
| File:Line | Code | Usage |
|---|---|---|
| `config/context.go:24` | `sys *access.System` | Field in `AppContext` struct |
| `config/context.go:77` | `func (a *AppContext) System() *access.System` | Accessor method |

### 2b. GlobalSystem reads (`access.GlobalSystem`)
| File:Line | Code | Usage |
|---|---|---|
| `utils/get.go:157` | `access.GlobalSystem.GetProfileCached(pubKey)` | Cache read in `GetProfile` |
| `utils/get.go:202` | `access.GlobalSystem.SetProfileCached(pubKey, pm)` | Cache write in `GetProfile` |
| `utils/get.go:380` | `access.GlobalSystem.GetProfileCached(pubKey)` | Cache read in `GetProfileAsync` |
| `utils/get.go:421` | `access.GlobalSystem.SetProfileCached(pubKey, pm)` | Cache write in `GetProfileAsync` |
| `utils/get.go:452` | `access.GlobalSystem.SetProfileCached(ev.PubKey, pm)` | Cache write in `GetProfiles` batch |
| `config/context.go:28` | `sys := access.GlobalSystem` | `NewAppContext` — read to reuse or create |
| `config/config.go:281` | `if access.GlobalSystem == nil` | `GlobalPool()` — check if exists |
| `config/config.go:306` | `if access.GlobalSystem != nil` | `GlobalPool()` — back-link pool |
| `config/config.go:504` | `sys := access.GlobalSystem` | `TrackEventRelay()` wrapper |
| `config/config.go:517` | `sys := access.GlobalSystem` | `GetEventRelay()` wrapper |
| `config/config.go:550` | `sys := access.GlobalSystem` | `MigrateEventRelaysToSystem()` |

### 2c. GlobalSystem writes
| File:Line | Code | Usage |
|---|---|---|
| `config/config.go:292` | `access.GlobalSystem = access.NewSystem(...)` | `GlobalPool()` — create singleton |
| `config/config.go:307` | `access.GlobalSystem.Pool = globalPool` | `GlobalPool()` — back-link |
| `config/context.go:46` | `access.GlobalSystem = sys` | `NewAppContext` — set if nil |

### 2d. Constructor calls (`access.NewSystem`)
| File:Line | Code | Usage |
|---|---|---|
| `config/config.go:292` | `access.NewSystem(nil, hints, kvStore, readRelays, writeRelays, dmRelays, searchRelays, fallback, localRelayURL)` | `GlobalPool()` — create early |
| `config/context.go:35` | `access.NewSystem(pool, hints, nil, readRelays, writeRelays, dmRelays, searchRelays, knownRelays, localRelayURL)` | `NewAppContext` — fallback if GlobalSystem is nil |

**Constructor signature** (`access/system.go:44-54`):
```go
func NewSystem(
    pool *nostr.Pool,
    hints sdk_hints.HintsDB,
    kvStore kvstore.KVStore,
    readRelays []string,
    writeRelays []string,
    dmRelays []string,
    searchRelays []string,
    fallbackRelays []string,
    localRelayURL string,
) *System
```

### 2e. access.RelayStream (type + constructor)
| File:Line | Code | Usage |
|---|---|---|
| `access/relay_stream.go:7-11` | Type definition | `RelayStream` with `mu`, `urls`, `index` |
| `access/relay_stream.go:14-16` | `func NewRelayStream(urls ...string) *RelayStream` | Constructor |
| `access/relay_stream.go:22-31` | `func (rs *RelayStream) Next() string` | Thread-safe round-robin |
| `access/system.go:59-63` | `NewRelayStream(readRelays...)` etc. | Creates 5 RelayStreams in NewSystem |

### 2f. access.NewKVStore
| File:Line | Code | Usage |
|---|---|---|
| `access/kvstore.go:14-18` | `func NewKVStore(dataDir string) (kvstore.KVStore, error)` | Creates BoltDB store at `event_relays.db` |
| `config/config.go:284` | `kvStore, err := access.NewKVStore(dataDir)` | `GlobalPool()` — called to create KVStore |

### 2g. Method calls on `*access.System`
| File:Line | Method | Usage |
|---|---|---|
| `access/system_test.go:42` | `.TrackEventRelay(eventID, relay1)` | Test: first write |
| `access/system_test.go:47` | `.TrackEventRelay(eventID, relay2)` | Test: second write (should be ignored) |
| `access/system_test.go:51` | `.GetEventRelay(eventID)` | Test: verify first-write-wins |
| `access/system_test.go:63` | `.GetEventRelay("nonexistent")` | Test: not found |
| `access/system_test.go:72` | `.TrackEventRelay("id", "wss://...")` | Test: nil store |
| `access/system_test.go:79` | `.GetEventRelay("id")` | Test: nil store |
| `access/system_test.go:90` | `.TrackEventRelay("id", "wss://...")` | Test: Close test |
| `access/system_test.go:94` | `.GetEventRelay("id")` | Test: Close test |
| `access/system_test.go:100` | `.Close()` | Test: Close |
| `config/config.go:506` | `sys.TrackEventRelay(eventID, relayURL)` | `TrackEventRelay()` wrapper |
| `config/config.go:519` | `sys.GetEventRelay(eventID)` | `GetEventRelay()` wrapper |
| `config/config.go:557` | `sys.TrackEventRelay(id, relay)` | Migration from fallback |
| `utils/get.go:157` | `access.GlobalSystem.GetProfileCached(pubKey)` | Cache read |
| `utils/get.go:202` | `access.GlobalSystem.SetProfileCached(pubKey, pm)` | Cache write |
| `utils/get.go:380` | `access.GlobalSystem.GetProfileCached(pubKey)` | Cache read (async) |
| `utils/get.go:421` | `access.GlobalSystem.SetProfileCached(pubKey, pm)` | Cache write (async) |
| `utils/get.go:452` | `access.GlobalSystem.SetProfileCached(ev.PubKey, pm)` | Cache write (batch) |

### 2h. Indirect callers (via config wrapper functions)
| File:Line | Code | Usage |
|---|---|---|
| `config/config.go:216` | `TrackEventRelay(ev.ID.Hex(), ie.Relay.URL)` | EventMiddleware in `NewPool` |
| `utils/post.go:109` | `config.GetEventRelay(rootID.Hex())` | NIP-10 e tag relay hint |
| `utils/post.go:110` | `config.GetEventRelay(parentEvent.ID.Hex())` | NIP-10 e tag relay hint |

---

## 3. sdk.System Public API (compared to access.System)

### 3a. Fields that exist in sdk.System

| sdk.System field | Type | access.System equivalent |
|---|---|---|
| `Pool` | `*nostr.Pool` | `Pool *nostr.Pool` — **identical** |
| `Hints` | `hints.HintsDB` | `Hints sdk_hints.HintsDB` — **identical** |
| `KVStore` | `kvstore.KVStore` | `Store kvstore.KVStore` — **same type, different name** |
| `Store` | `eventstore.Store` | *(none)* — sdk has an additional event store |
| `Publisher` | `nostr.Publisher` | *(none)* |
| `MetadataCache` | `cache.Cache32[ProfileMetadata]` | `MetadataCache sdk_cache.Cache32[sdk.ProfileMetadata]` — **identical** |
| `RelayListCache` | `cache.Cache32[GenericList[string, Relay]]` | `RelayListCache ...` — **identical** |
| `FollowListCache` | `cache.Cache32[GenericList[nostr.PubKey, ProfileRef]]` | `FollowListCache ...` — **identical** |

### 3b. RelayStream equivalents in sdk.System

sdk.System has many specialized RelayStream fields (all type `*RelayStream`):
- `RelayListRelays`, `FollowListRelays`, `MetadataRelays`, `FallbackRelays`
- `JustIDRelays`, `UserSearchRelays`, `NoteSearchRelays`

access.System has:
- `ReadRelays`, `WriteRelays`, `DMRelays`, `SearchRelays`, `FallbackRelays`

**Both use the same `NewRelayStream(urls ...string) *RelayStream` function** — same constructor. However, they are **different types**: `access.RelayStream` vs `sdk.RelayStream`. Both expose a `Next() string` method.

### 3c. Methods unique to access.System (NOT on sdk.System)

| access.System method | Signature | Purpose |
|---|---|---|
| `GetProfileCached` | `(pubKey nostr.PubKey) (sdk.ProfileMetadata, bool)` | Read from MetadataCache with nil-safety guard |
| `SetProfileCached` | `(pubKey nostr.PubKey, pm sdk.ProfileMetadata)` | Write to MetadataCache with 5-min TTL |
| `GetRelayListCached` | `(pubKey nostr.PubKey) (GenericList, bool)` | Read from RelayListCache with nil-safety |
| `SetRelayListCached` | `(pubKey nostr.PubKey, rl GenericList)` | Write with 5-min TTL |
| `GetFollowListCached` | `(pubKey nostr.PubKey) (GenericList, bool)` | Read from FollowListCache with nil-safety |
| `SetFollowListCached` | `(pubKey nostr.PubKey, fl GenericList)` | Write with 5-min TTL |
| `TrackEventRelay` | `(eventID, relayURL string) error` | First-write-wins event→relay to KVStore |
| `GetEventRelay` | `(eventID string) string` | Read single event→relay from KVStore |

### 3d. Methods on sdk.System (NOT on access.System)

| sdk.System method | Signature | Purpose |
|---|---|---|
| `TrackEventRelaysD` | `(relay string, id nostr.ID)` | Companion to duplicate middleware — tracks relay for event ID |
| `TrackEventHintsAndRelays` | `(ie nostr.RelayEvent)` | Combined EventMiddleware callback |
| `TrackEventHints` | `(ie nostr.RelayEvent)` | EventMiddleware: hints only |
| `GetEventRelays` | `(id nostr.ID) []string` | Returns **all** known relays for an event |
| `TrackEventAccessTime` | `(id nostr.ID)` | Records access time |
| `GetEventAccessTime` | `(id nostr.ID) nostr.Timestamp` | Reads access time |
| `EraseEventRelays` | `(id nostr.ID) error` | Deletes relay records for an event |
| `EraseAccessTime` | `(id nostr.ID) error` | Deletes access time records |
| `FetchProfileMetadata` | `(ctx, pubkey) ProfileMetadata` | High-level profile fetch with cache+network |
| `FetchFollowList` | `(ctx, pubkey) GenericList` | High-level follow list fetch |
| `FetchRelayList` | `(ctx, pubkey) GenericList` | High-level relay list fetch |
| `FetchOutboxRelays` | `(ctx, pubkey, n) []string` | Discover outbox relays |
| `FetchInboxRelays` | `(ctx, pubkey, n) []string` | Discover inbox relays |
| `FetchFeedPage` | `(ctx, pubkeys, kinds, ...) ([]Event, error)` | Paginated feed fetch |
| `SearchUsers` | `(ctx, query) []ProfileMetadata` | User search |
| `PrepareNoteEvent` | `(ctx, evt) []string` | Prepare note with target relays |

---

## 4. Migration Analysis

### 4a. Patterns that map CLEANLY

| access.System usage | sdk.System replacement | Migration effort |
|---|---|---|
| `access.GlobalSystem` (singleton var) | `sdk.GlobalSystem` (same pattern — use a global `*sdk.System`) | **Trivial** — rename |
| `access.NewSystem(...)` | Use `sdk.NewSystem()` then manually set fields | **Medium** — sdk.NewSystem takes no args; fields must be assigned post-creation |
| `.Pool` field | `.Pool` — same type | **Trivial** |
| `.Hints` field | `.Hints` — same type | **Trivial** |
| `.Store` (KVStore) | `.KVStore` — same type, different name | **Trivial** — rename field access |
| `.MetadataCache` | `.MetadataCache` — same type | **Trivial** |
| `.RelayListCache` | `.RelayListCache` — same type | **Trivial** |
| `.FollowListCache` | `.FollowListCache` — same type | **Trivial** |
| `.Close()` | `.Close()` — same signature | **Trivial** |
| RelayStream fields (`ReadRelays`, etc.) | sdk.System has similar fields (`MetadataRelays`, `FallbackRelays`, etc.) but different naming | **Low** — field renaming |
| `access.NewRelayStream(urls...)` | `sdk.NewRelayStream(urls...)` — identical constructor | **Trivial** — same signature |
| `relayStream.Next()` | `sdk.RelayStream.Next()` — identical method | **Trivial** |
| `access.NewKVStore(dataDir)` | sdk doesn't have `NewKVStore`, but can directly call `bboltKv.NewStore(path)` | **Trivial** — inline the 3-line function |

### 4b. Patterns that DO NOT map cleanly (need custom code)

| access.System pattern | Why it doesn't map to sdk.System | Migration approach |
|---|---|---|
| `GetProfileCached(pubKey)` + nil-safety | sdk.System has `MetadataCache` field but no wrapper method with nil-safety | **Add a helper function** that checks `sys != nil && sys.MetadataCache != nil` before calling `Get()` |
| `SetProfileCached(pubKey, pm)` + TTL | sdk.System has `MetadataCache` field but no wrapper with TTL | Similar: wrapper accessing `sys.MetadataCache.SetWithTTL(pubKey, pm, 5*time.Minute)` |
| `GetRelayListCached(..)` / `SetRelayListCached(..)` | Same as above | Same pattern |
| `GetFollowListCached(..)` / `SetFollowListCached(..)` | Same as above | Same pattern |
| `TrackEventRelay(eventID string, relayURL string) error` — first-write-wins | sdk uses `TrackEventRelaysD(relay string, id nostr.ID)` with different args order; also it appends (does NOT use first-write-wins) | Either: (a) keep custom wrapper over KVStore, or (b) switch to `TrackEventRelaysD` and accept multi-relay tracking |
| `GetEventRelay(eventID string) string` — returns single string | sdk returns `GetEventRelays(id nostr.ID) []string` — returns all known relays | Change callers to use `[]string` return, get first element when single relay needed |
| `TrackEventRelay(ev.ID.Hex(), ie.Relay.URL)` in EventMiddleware | sdk has `TrackEventHintsAndRelays(ie)` that does hints + relay tracking together | **Could replace** the entire EventMiddleware function with `sys.TrackEventHintsAndRelays(ie)` — but need to verify it handles the `h.Save` hints logic identically |
| Fallback in-memory buffer (`fallbackEventRelays`) in config.go | sdk.System has no fallback; tracks directly to KVStore | Remove fallback — create System early with KVStore, drop fallback buffer logic |
| `config.TrackEventRelay()` wrapper | Wraps access.GlobalSystem with nil check + fallback | Simplify: access sdk.GlobalSystem directly, remove wrapper |
| `config.GetEventRelay()` wrapper | Same as above | Same |

### 4c. Field mismatch: sdk.System naming vs access.System naming

| access.System field | sdk.System field | Notes |
|---|---|---|
| `ReadRelays` | No direct equivalent. Closest: `MetadataRelays`, `FollowListRelays`, etc. | sdk.System uses task-specific relay streams. Could map `ReadRelays` → `FallbackRelays` for general reads, or use multiple specialized fields |
| `WriteRelays` | No direct equivalent. Use `PrepareNoteEvent()` instead | sdk handles write relay selection automatically via hints |
| `DMRelays` | No direct equivalent | Need custom field or use `FallbackRelays` |
| `SearchRelays` | `UserSearchRelays`, `NoteSearchRelays` | Separate streams for user vs note search |
| `FallbackRelays` | `FallbackRelays` | **Direct match** |
| `Store` | `KVStore` | Same type, different name |
| *(none)* | `Store` (eventstore.Store) | sdk has an extra event store field |

---

## 5. Code Change Summary

### Files that need modification

| File | Changes needed |
|---|---|
| `access/system.go` | **Delete or reduce to thin helper functions** (GetProfileCached, etc. become standalone helpers on sdk.System) |
| `access/relay_stream.go` | **Delete** — replaced by `sdk.RelayStream` |
| `access/kvstore.go` | **Delete or inline** — `bboltKv.NewStore(path)` call replaces it |
| `access/system_test.go` | **Rewrite** for new helper functions or remove if helpers are deleted |
| `access/relay_stream_test.go` | **Delete** — sdk.RelayStream has its own tests |
| `config/config.go` | Replace `access.GlobalSystem` with `sdk.GlobalSystem`; replace `access.NewSystem(...)` with `sdk.NewSystem()` + field assignment; simplify `TrackEventRelay`/`GetEventRelay` wrappers |
| `config/context.go` | Change `sys *access.System` to `sys *sdk.System`; update `NewAppContext`; update `System()` accessor return type |
| `utils/get.go` | Change `access.GlobalSystem.GetProfileCached` → `sdk.GlobalSystem.GetProfileCached` (or equivalent helper) |
| `utils/post.go` | Change `config.GetEventRelay` → return `[]string` instead of `string` |

### Import changes

| File | Remove | Add |
|---|---|---|
| `config/config.go` | `"github.com/jerry-harm/nosmec/access"` | *(none extra — already imports sdk types)* |
| `config/context.go` | `"github.com/jerry-harm/nosmec/access"` | *(none extra — already imports sdk types)* |
| `utils/get.go` | `"github.com/jerry-harm/nosmec/access"` | *(none extra — already imports sdk)* |

---

## 6. Caveats / Not Found

1. **sdk.NewSystem() takes no arguments** — unlike `access.NewSystem(...)` which takes 9 args. After `sdk.NewSystem()`, each field must be assigned manually:
   ```go
   sys := sdk.NewSystem()
   sys.Pool = pool
   sys.Hints = hints
   sys.KVStore = kvStore
   sys.FallbackRelays = sdk.NewRelayStream(fallbackRelays...)
   // ... etc
   ```

2. **Event→relay tracking semantics differ**: `access.TrackEventRelay` is first-write-wins (single relay per event), while `sdk.TrackEventRelaysD` appends (tracks multiple relays). The callers in `utils/post.go` expect a single relay for NIP-10 hints — if switching to sdk's multi-relay tracking, the callers need to pick the first relay.

3. **Relay stream field mapping is not 1:1** — access.System has 5 general-purpose relay streams (Read, Write, DM, Search, Fallback), while sdk.System has ~10 task-specific streams. The mapping depends on how each relay list is used in queries. May need to keep custom fields alongside sdk's built-in ones.

4. **Local relay URL (`localRelayURL`)** is an unexported field in access.System used to skip tracking events from the local relay. sdk.System does not have this — need to add a custom field or handle the skip logic differently.

5. **Cache TTL**: access.System uses 5-minute TTL for MetadataCache, RelayListCache, FollowListCache. sdk.System's cache behavior depends on the cache implementation used — if switching to sdk, verify TTL behavior.

6. **The `access/` directory itself** would likely be reduced to a thin helper package with nil-safe cache wrappers, or eliminated entirely.
