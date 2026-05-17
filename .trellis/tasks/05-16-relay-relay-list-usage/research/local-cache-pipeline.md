# Research: Local Relay Event Caching Pipeline

- **Query**: How do events get persisted to local relay cache? Are all entry points covered?
- **Scope**: internal
- **Date**: 2026-05-17

## Findings

### Core Mechanism: `CacheEvent()`

**File**: `utils/get.go:135-149`

```go
func CacheEvent(event *nostr.Event, app *config.AppContext) {
    if event == nil || app == nil {
        return
    }
    if !shouldCache(event, app) {
        return
    }
    localURL := config.GetLocalRelayURL()
    if localURL == "" {
        return
    }
    go func() {
        app.Pool().PublishMany(context.Background(), []string{localURL}, *event)
    }()
}
```

**How it works**:
1. `shouldCache()` (`get.go:125-133`) checks event against `app.Config().CacheFilters` — only whitelisted event kinds are cached
2. Gets local relay URL from `config.GetLocalRelayURL()`
3. Publishes the event to the local relay **asynchronously** via `Pool().PublishMany()`
4. The local relay's `StoreEvent` callback (`config/config.go:366-371`) persists to BoltDB + Bleve search index in a single transaction

### Pool EventMiddleware (HintsDB only)

**File**: `config/config.go:194-224`

The `PoolOptions.EventMiddleware` hook fires for EVERY event flowing through the pool. It updates the **HintsDB** (relay hints/presence tracking) but does **NOT** cache events to local relay. It:
- Saves `(pubkey, relayURL)` pairs to HintsDB for events received
- Extracts `p` tag relay hints from events
- Extracts relay list entries for Kind 10002

This is a separate persistence layer from event caching.

### CacheEvent Call Sites — Complete Inventory

#### Synchronous (blocking return) — All in `utils/get.go`

| Function | File:Line | Trigger | Cache? |
|---|---|---|---|
| `GetEvent` (replaceable, local hit) | `get.go:63` | QuerySingle → local relay | ✅ |
| `GetEvent` (replaceable, remote hit) | `get.go:74` | FetchManyReplaceable → remote | ✅ |
| `GetEvent` (non-replaceable, local hit) | `get.go:84` | QuerySingle → local relay | ❌ (already cached) |
| `GetEvent` (non-replaceable, remote hit) | `get.go:95` | QuerySingle → remote | ✅ |
| `GetProfile` | `get.go:194` | QuerySingle → all relays | ✅ |
| `GetParentEvent` | `get.go:353` | QuerySingle → hints/AllReadable | ✅ |
| `GetEventAsync` (replaceable) | `get.go:486` | FetchManyReplaceable → remote | ✅ |
| `GetEventAsync` (non-replaceable) | `get.go:493` | QuerySingle → remote | ✅ |
| `GetProfiles` (per event) | `get.go:432` | FetchManyReplaceable → remote | ✅ |

**Notable omission**: `GetProfileAsync` (`get.go:366-407`) — returns `&result.Event` at line 406 **without caching**. This is the async version that fetches from combined relays (AllReadable + KnownRelays) but skips local relay and skips caching. This means profiles fetched via `GetProfileAsync` / `GetProfileNameAsync` are never cached.

#### Streaming (channel-based) — All in `utils/get.go` and `utils/community.go`

| Function | File:Line | Pool Method | Cache? |
|---|---|---|---|
| `SubscribeWithCache` | `get.go:108` | SubscribeMany (wraps) | ✅ (every event) |
| `GetMyTimeline` | `get.go:525` | FetchMany | ✅ (per event) |
| `GetGlobalTimeline` | `get.go:554` | FetchMany | ✅ (per event) |
| `GetFollowedTimeline` | `get.go:635` | FetchMany (after dedup) | ✅ (per event) |
| `GetCommunityPosts` | `community.go:245` | FetchMany | ✅ (per event) |
| `GetMyCreatedCommunities` | `community.go:346` | FetchMany | ✅ (per event) |
| `GetPostedCommunities` | `community.go:375` | FetchMany | ✅ (per event) |

#### Publish-time caching — All in `utils/community.go`

| Function | File:Line | Trigger | Cache? |
|---|---|---|---|
| `CreateCommunity` | `community.go:77` | After PublishMany to writable relays | ✅ |
| `PublishPost` | `community.go:146` | After PublishMany to writable relays | ✅ |

### Entry Points NOT Caching Events

| Entry Point | File:Line | Pool Method | Reason |
|---|---|---|---|
| `GetProfileAsync` | `get.go:366-407` | QuerySingle | Missing CacheEvent call — the function fetches from combined relays but doesn't cache the result |
| `GetProfileNameAsync` | `get.go:357-364` | calls GetProfileAsync | Inherits the omission |
| `FindDMConversations` | `dm.go:131` | SubscribeMany | Uses raw SubscribeMany without SubscribeWithCache |
| `FetchDMMessages` | `dm.go:224` | SubscribeMany | Uses raw SubscribeMany without SubscribeWithCache |
| `QueryRepliesToRoot` | `get.go:282-310` | FetchMany | Thread replies query — no CacheEvent in the range loop |
| `ThreadTreeView.fetchAllReplies` | `thread_treeview.go:284` | FetchMany | TUI thread fetch — iterates relay events but doesn't call CacheEvent |
| `GetEvent` (non-replaceable, local hit) | `get.go:84-85` | QuerySingle (local) | Event already in local relay — correct to skip |
| `GetEvent` (replaceable, local hit) | `get.go:60-63` | QuerySingle (local) | ⚠️ Redundant: event is already in local relay but gets re-published (harmless but wasteful) |

### TUI Subscription Entry Points

| Source | File:Line | Caching Method |
|---|---|---|
| Timeline (main feed) | `tui/timeline/model.go:450` | `SubscribeWithCache` ✅ |
| DM TUI | `tui/dm/model.go:213` | `SubscribeWithCache` ✅ |
| Thread tree view | `tui/window/event/thread_treeview.go:284` | Raw `FetchMany` ❌ |

### Consistency Summary

| Category | Cached? |
|---|---|
| Synchronous queries (GetEvent, GetProfile, GetParent, GetEventAsync, GetProfiles) | ✅ (except GetProfileAsync) |
| Timeline streaming (My, Global, Followed, Community, DM TUI) | ✅ (via SubscribeWithCache or per-event CacheEvent) |
| DM utils (FindDMConversations, FetchDMMessages) | ❌ raw SubscribeMany |
| Thread query (QueryRepliesToRoot, ThreadTreeView.fetchAllReplies) | ❌ raw FetchMany |
| Publish operations (CreateCommunity, PublishPost) | ✅ (explicit CacheEvent after publish) |

### Key Insight: `GetEvent` Dual-Phase Caching

`GetEvent` (`get.go:41-101`) has a two-phase query strategy:
1. **Phase 1**: Query local relay first (2s timeout) — fast cache hit
2. **Phase 2**: If not found locally, query remote relays

For non-replaceable events (line 79-84): if the local relay has it, the event is returned without re-caching (correct — it's already stored).
For replaceable events (line 56-63): if the local relay has it, CacheEvent is called anyway — this is a harmless redundancy (re-publishes the replaceable event to the same local relay).

## Caveats / Not Found

- **`GetProfileAsync` missing cache**: This is a real gap. Any profile fetched asynchronously (for display in the TUI) never gets written to the local relay, so it won't be found on the next local-first query. The synchronous `GetProfile` DOES cache, but the async version skips it.
- **DM utils skip caching**: `FindDMConversations` and `FetchDMMessages` use raw `SubscribeMany` without caching. This means DM metadata (gift wraps) never enters the local relay — inconsistent with timeline events going through `SubscribeWithCache`.
- **Thread replies skip caching**: Both `QueryRepliesToRoot` and `ThreadTreeView.fetchAllReplies` skip CacheEvent. Thread replies are fetched from remote relays every time a thread is opened — they don't benefit from the local cache.
- **`shouldCache` filter**: CacheFilters are configured but the exact filter values are in the user's config, not in this codebase. Events not matching CacheFilters silently skip caching even if CacheEvent is called.
