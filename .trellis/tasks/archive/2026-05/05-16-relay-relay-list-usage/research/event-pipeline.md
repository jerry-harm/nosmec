# Research: Event Pipelines in nosmec

- **Query**: Trace all event pipelines — where events enter the system, where we can hook auto-learning
- **Scope**: internal
- **Date**: 2026-05-17

## Findings

### 1. Pool Initialization & Configuration

#### Pool Creation
- **File**: `config/config.go:177-185`
- `NewPool()` creates a `nostr.Pool` with `PoolOptions` containing only a `NoticeHandler` logger callback
- **No `EventMiddleware` is set** — this is the global hook point where every incoming event could be intercepted

```go
// config/config.go:177-185
func NewPool() *nostr.Pool {
    return nostr.NewPool(nostr.PoolOptions{
        RelayOptions: nostr.RelayOptions{
            NoticeHandler: func(relay *nostr.Relay, notice string) {
                logger.Debug("NOTICE from %s: '%s'", relay.URL, notice)
            },
        },
    })
}
```

#### Pool singleton
- **File**: `config/config.go:187-197`
- `GlobalPool()` returns a singleton pool (created once)
- `SetPool()` allows tests to inject a mock pool

#### AppContext wiring
- **File**: `config/context.go:14-35`
- `AppContext` holds `pool *nostr.Pool` (private)
- `AppContext.Pool()` returns the pool
- Pool is created in `cmd/root.go:60-68` during `initApp()`:
  ```go
  pool := config.GlobalPool()
  store := config.GlobalLMDB()
  app = config.NewAppContext(pool, store, cfg, config.GetViper())
  ```

#### nostr PoolOptions — EventMiddleware support
- **File**: `fiatjaf.com/nostr@v0.0.0-20260310013726-4e490879b558/pool.go:76-99`
- `PoolOptions` supports four middleware hooks:
  1. **`EventMiddleware func(RelayEvent)`** — called with ALL events received from any relay
  2. **`DuplicateMiddleware func(relay string, id ID)`** — called with duplicate IDs
  3. **`AuthorKindQueryMiddleware func(relay string, pubkey PubKey, kind Kind)`** — called per relay+pubkey+kind combo
  4. **`AuthRequiredHandler func(context.Context, *Event) error`** — auth challenge handler
- **`RelayEvent`** (from `types.go:12-15`): `{Event, Relay *Relay}` — contains both the event AND the relay it came from
- **`SubscriptionOptions`** (from `subscription.go:62-76`): `{Label string, CheckDuplicate, CheckDuplicateReplaceable, MaxWaitForEOSE}`

**Key insight**: The `EventMiddleware` is the exact hook we need — it fires for every incoming event in `subMany()` (line 555-557), `subManyEose()` (line 716-718), and `fetchManyReplaceable()` (line 414-416). The `RelayEvent` provides the relay URL via `ie.Relay.URL`.

---

### 2. ALL Event Entry Points (where events flow INTO the app)

#### 2A. `SubscribeWithCache` — THE central subscribe wrapper
- **File**: `utils/get.go:103-114`
- **Called by**: timeline model, DM TUI model
- **Flow**: wraps `pool.SubscribeMany()` → caches each event → forwards to `chan nostr.RelayEvent`
- **Has relay info**: YES — passes through `nostr.RelayEvent` (contains `Relay *Relay`)
- **Hook opportunity**: Easy — intercept inside the goroutine (line 106-112) before caching

```go
func SubscribeWithCache(ctx context.Context, pool *nostr.Pool, relays []string, filter nostr.Filter, opts nostr.SubscriptionOptions, app *config.AppContext) chan nostr.RelayEvent {
    ch := pool.SubscribeMany(ctx, relays, filter, opts)
    out := make(chan nostr.RelayEvent)
    go func() {
        for ie := range ch {
            CacheEvent(&ie.Event, app)
            out <- ie
        }
        close(out)
    }()
    return out
}
```

#### 2B. Timeline model subscription (real-time)
- **File**: `tui/timeline/model.go:370-455` (`startSubscription`)
- **Calls**: `utils.SubscribeWithCache()` at line 450
- **Filter**: varies by timeline type (global/mine/community/followed), `kind=1` or `kind=1,1111`
- **Relays**: `app.Config().KnownRelays` → fallback `app.AllReadableRelays()`
- **Relay info**: DISCARDED — `SubscribeWithCache` returns `chan nostr.RelayEvent` but `pollSubscription()` (line 459-493) strips relay info and only forwards `relayEvent.Event`
- **Dispatch to UI**: `newEventMsg{event: utils.TimelineEvent{Event: event}}` — relay info is LOST

#### 2C. DM TUI model subscription (real-time)
- **File**: `tui/dm/model.go:193-225` (`startSubscription`)
- **Calls**: `utils.SubscribeWithCache()` at line 213 → then strips relay info, feeds only `nostr.Event` into `m.subCh`
- **Filter**: `kind=1059` (GiftWrap), `#p = [ourPubKey, recipientPubKey]`
- **Relays**: `app.ListDMRelays()` → `app.ReadableRelays()` → `app.AllReadableRelays()`
- **Relay info**: DISCARDED — similar to timeline, only `relayEvent.Event` forwarded

#### 2D. DM utility: raw `pool.SubscribeMany` calls
- **File**: `utils/dm.go`
  - **`ListDMConversations`** (line 131): `app.Pool().SubscribeMany(ctx, relays, filter, {...})` — kind=1059, #p
  - **`QueryDMHistory`** (line 224): `app.Pool().SubscribeMany(ctx, relays, filter, {...})` — kind=1059, #p
- **Relay info**: DISCARDED — both iterate `for ie := range` but only use `ie.Event`, never `ie.Relay`
- These are the CLI-level DM commands (not TUI), run in separate goroutines

#### 2E. `GetEvent` / `GetEventAsync` — one-shot fetches
- **File**: `utils/get.go:41-101`, `466-498`
- **Calls**: `pool.QuerySingle()` or `pool.FetchManyReplaceable()`
- **Relay info**: DISCARDED — `QuerySingle` returns `*RelayEvent` (has relay) but callers only extract `.Event`
- **Used by**: profile fetching, note fetching, parent event fetching, relay list syncing, etc.

#### 2F. `QueryRepliesToRoot` — thread reply fetch
- **File**: `utils/get.go:282-310`
- **Calls**: `pool.FetchMany()` — iterates `for relayEvent := range results`, uses only `relayEvent.Event`
- **Relay info**: DISCARDED

#### 2G. Timeline fetch functions (initial load)
- **File**: `utils/get.go:500-642`
  - **`GetMyTimeline`** (line 523): `pool.FetchMany()` — discards relay info (only `relayEvent.Event`)
  - **`GetGlobalTimeline`** (line 552): `pool.FetchMany()` — discards relay info
  - **`GetFollowedTimeline`** (line 608): `pool.FetchMany()` — discards relay info

#### 2H. Community posts
- **File**: `utils/community.go:221-251`
- **`GetCommunityPosts`** (line 243): `pool.FetchMany()` — discards relay info
- **`GetMyCreatedCommunities`** (line 344): `pool.FetchMany()` — discards relay info
- **`GetPostedCommunities`** (line 369): `pool.FetchMany()` — discards relay info

#### 2I. Thread tree view
- **File**: `tui/window/event/thread_treeview.go:255-261`, `296`
- **Calls**: `pool.QuerySingle()` and `pool.FetchMany()` — discards relay info

#### 2J. User relay discovery
- **File**: `utils/user_relays.go:56`
- **`DiscoverUserRelays`**: `pool.FetchManyReplaceable()` — uses result internally, no relay info needed

#### 2K. Search
- **File**: `utils/search.go:101`
- **`pool.FetchMany()`** — discards relay info

#### 2L. Profile fetching (one-shot)
- **File**: `utils/profile.go:169` — `PublishMany` (outbound only)
- **File**: `utils/profile.go:224,253,322,359,387,405` — `QuerySingle()` — discards relay info

#### 2M. Subscription sync
- **File**: `utils/subscription.go:108,153,210` — `QuerySingle()` — discards relay info

#### 2N. Relay list sync
- **File**: `utils/relay_list.go:47,113` — `QuerySingle()` — discards relay info

---

### 3. Event OUTBOUND Paths (when WE publish/create events)

All publish paths follow the same pattern:
1. Get secret key
2. Build event with tags
3. Sign event
4. `app.Pool().PublishMany(ctx, writableRelays, *event)` → iterate results for errors
5. **No hook** on publish success to detect the event we just published

#### 3A. Compose model (TUI)
- **File**: `tui/compose/model.go:459-511` (`sendContent`)
- **Publishes to**: `m.app.AllWritableRelays()` via `pool.PublishMany()`
- **Event kinds**: KindTextNote, KindComment (default KindTextNote=1)
- **Has relay info**: YES — `PublishMany` returns `chan PublishResult{Error, RelayURL, Relay}`
- **Note**: After publish, sends `sendSuccessMsg` → `tea.Quit` — no recording of which relays the event was sent to

#### 3B. CLI post commands
- **File**: `utils/post.go`
  - **`PostNote`** (line 32): `pool.PublishMany()` to writable relays
  - **`ReplyToNote`** (line 74): `pool.PublishMany()` to writable relays
  - **`QuoteNote`** (line 109): `pool.PublishMany()` to writable relays
  - **`DeleteNote`** (line 141): `pool.PublishMany()` to writable relays

#### 3C. DM send
- **File**: `utils/dm.go:16-58` (`SendDM`)
- Uses `nip17.PublishMessage()` which internally uses `pool.PublishMany()`
- Relays: our DM relays + recipient's DM relays

#### 3D. Profile publish
- **File**: `utils/profile.go:169` — `pool.PublishMany()` to writable relays

#### 3E. Community publish
- **File**: `utils/community.go:69,138` — `pool.PublishMany()` to writable relays

#### 3F. Subscription publish
- **File**: `utils/subscription.go:292,329,362` — `pool.PublishMany()` to writable relays

#### 3G. Relay list publish
- **File**: `utils/relay_list.go:177,210` — `pool.PublishMany()` to writable relays

#### 3H. Cache-to-local (also outbound)
- **File**: `utils/get.go:146-148` (`CacheEvent`)
- `pool.PublishMany(context.Background(), []string{localURL}, *event)` — writes to local relay only

---

### 4. Relay Info Flow Analysis

| Entry Point | Has Relay in Channel | Passes Relay to Consumer | Consumer Uses Relay? |
|---|---|---|---|
| `SubscribeWithCache` | YES (`nostr.RelayEvent`) | YES | NO (stripped in timeline/DM polling) |
| `QuerySingle` | YES (`*nostr.RelayEvent`) | N/A (synchronous) | NO (only `.Event` extracted) |
| `FetchMany` | YES (per `RelayEvent` in chan) | YES (in for-range) | NO (only `.Event` extracted) |
| `FetchManyReplaceable` | NO (returns `*xsync.MapOf`) | N/A | N/A |
| `PublishMany` | YES (`PublishResult`) | YES (in for-range) | PARTIAL (only for error reporting) |

**Conclusion**: Relay info is systematically discarded everywhere. The `nostr.RelayEvent` struct contains the relay, but every single consumer pattern is:
```go
for relayEvent := range ch {
    // use relayEvent.Event
    // NEVER use relayEvent.Relay
}
```

---

### 5. Where to Insert a Hook/Middleware

#### Option A: Pool-level `EventMiddleware` (RECOMMENDED)
- **Set at**: `config/config.go:177-185` in `NewPool()`
- **Coverage**: ALL incoming events from ALL subscription types (SubscribeMany, FetchMany, FetchManyReplaceable)
- **Provides**: `RelayEvent{Event, Relay}` — relay info is included
- **Pros**: Single insertion point, catches everything, cannot be forgotten
- **Cons**: Fires inside pool goroutines (thread safety), must not block
- **Verification**: `pool.go` lines 414-416 (fetchManyReplaceable), 555-557 (subMany), 716-718 (subManyEose) all call `pool.eventMiddleware(ie)` before forwarding

#### Option B: Inside `SubscribeWithCache` wrapper
- **Insert at**: `utils/get.go:106-112` goroutine, inside the `for ie := range ch` loop
- **Coverage**: ONLY continuous subscriptions (timeline, DM TUI) — misses `FetchMany`/`QuerySingle` calls
- **Provides**: `nostr.RelayEvent` — relay info available
- **Pros**: Simpler, doesn't touch pool creation
- **Cons**: Misses many entry points (all fetches, CLI commands)
- **Verification**: Only two callers: `tui/timeline/model.go:450` and `tui/dm/model.go:213`

#### Option C: Inside `CacheEvent` wrapper
- **Insert at**: `utils/get.go:135-149` in `CacheEvent()`
- **Coverage**: Only events that are explicitly passed to `CacheEvent` — misses events that aren't cached
- **Provides**: `*nostr.Event` — NO relay info
- **Pros**: Already a centralized function
- **Cons**: No relay info, only fires for cached events

---

### 6. Existing Hints Database / Auto-Learning Check

- **Search for "hints"**: No hints database, no `HintsDB` type found in the codebase
- **Search for "learn"**: No auto-learning mechanism exists
- **Search for "auto"**: No auto-recording of relay→event mappings
- **`TrackRelays`**: `config/context.go:402-408` — stores relay URLs (no event association)
- **`KnownRelays`**: `config/config.go` — plain `[]string`, no associated metadata

---

### 7. Summary of All Files and Lines

| Category | File | Lines | Pool Method | Relay Info Kept? |
|---|---|---|---|---|
| Pool init | `config/config.go` | 177-185 | - | N/A |
| Central sub wrapper | `utils/get.go` | 103-114 | `SubscribeMany` | YES (passed through) |
| Timeline sub | `tui/timeline/model.go` | 370-455, 459-493 | via `SubscribeWithCache` | NO (stripped) |
| DM TUI sub | `tui/dm/model.go` | 193-225 | via `SubscribeWithCache` | NO (stripped) |
| DM util sub | `utils/dm.go` | 131, 224 | `SubscribeMany` | NO (stripped) |
| Fetch many | `utils/get.go` | 523, 552, 608 | `FetchMany` | NO (stripped) |
| Fetch single | `utils/get.go` | 60, 82, 92, 347, 401, etc. | `QuerySingle` | NO (stripped) |
| Thread fetch | `tui/window/event/thread_treeview.go` | 256, 296 | `QuerySingle`, `FetchMany` | NO (stripped) |
| Community | `utils/community.go` | 243, 344, 369 | `FetchMany` | NO (stripped) |
| User relays | `utils/user_relays.go` | 56 | `FetchManyReplaceable` | N/A |
| Search | `utils/search.go` | 101 | `FetchMany` | NO (stripped) |
| Compose publish | `tui/compose/model.go` | 494 | `PublishMany` | PARTIAL |
| Post publish | `utils/post.go` | 32, 74, 109, 141 | `PublishMany` | PARTIAL |
| Profile publish | `utils/profile.go` | 169 | `PublishMany` | PARTIAL |
| Community publish | `utils/community.go` | 69, 138 | `PublishMany` | PARTIAL |
| Sub publish | `utils/subscription.go` | 292, 329, 362 | `PublishMany` | PARTIAL |
| Relay list publish | `utils/relay_list.go` | 177, 210 | `PublishMany` | PARTIAL |
| Cache to local | `utils/get.go` | 147 | `PublishMany` | N/A |

---

## Caveats / Not Found

1. **No existing HintsDB or auto-learning**: The project has no mechanism to record which relays served which events. This would need to be built from scratch.
2. **Relay info is universally discarded**: Despite `nostr.RelayEvent` containing relay info, every single consumer path strips it. This is the root cause of the relay-list-usage bug.
3. **Pool middleware is the simplest fix**: Setting `EventMiddleware` in `NewPool()` would capture every incoming event with relay info at a single point, without needing to touch any consumer code.
4. **Publish path has partial relay info**: `PublishMany` returns `chan PublishResult` with relay URLs, but consumers only check for errors — they don't record which relays accepted the event.
5. **`FetchManyReplaceable` does NOT pass relay info**: Uses `xsync.MapOf[ReplaceableKey, Event]` — the relay is not part of the key or value. The `eventMiddleware` hook fires BEFORE this dedup, so it still has relay info.
