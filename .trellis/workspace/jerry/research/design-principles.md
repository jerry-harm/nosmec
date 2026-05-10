# Research: Design Principles — nosmec

- **Query**: Analyze nosmec codebase and produce design principles document
- **Scope**: internal
- **Date**: 2026-05-10

## Findings

### 1. Query Patterns

#### Files Examined
| File Path | Description |
|---|---|
| `utils/get.go` | Main query functions for events, profiles, timelines |
| `utils/dm.go` | DM sending, listening, conversation listing |
| `utils/community.go` | Community creation, posts, following |
| `utils/profile.go` | Profile metadata, full profile, relay list fetching |

#### Sync vs Async Patterns

**Sync patterns** (`utils/get.go:17-71`):
- `GetEvent` — local relay first (2s timeout), then remote relays (10s timeout)
- Uses `context.WithTimeout` for deadline propagation
- Returns single `*nostr.Event`
- Replaceable events use `FetchManyReplaceable` (context-aware)

**Async patterns** (`utils/get.go:219-266`):
- `GetEventAsync` — fires goroutine, uses channel-based `found` signal
- Uses select on `found` channel vs `ctx.Done()`
- Returns `*nostr.Event` (blocking until result or context cancel)

**Channel-based streaming** (`utils/get.go:268-342`):
- `GetMyTimeline`, `GetGlobalTimeline`, `GetFollowedTimeline` return `chan *nostr.Event`
- Use `FetchMany` from the nostr pool, range over channel
- Goroutine wraps the subscription to return a typed channel

**SubscribeMany pattern** (`utils/dm.go:127-171`, `utils/community.go:220-229`):
- Direct iteration over `SubscribeMany` channel
- Used for DM conversations and community posts
- No wrapping — callers iterate directly

#### Inconsistencies Observed

| Pattern | Location | Issue |
|---|---|---|
| Local relay priority | `get.go:38-53` | Only for non-replaceable events; replaceable uses `FetchManyReplaceable` directly |
| Timeout values | `get.go:46-47` | Local: 2s, Remote: 10s, but no env var to configure |
| Context handling | `get.go:241-256` | `GetEventAsync` uses custom channel instead of `context.WithTimeout` |
| Goroutine cleanup | `get.go:65-68` | Fire-and-forget `go func()` for caching, no context |
| DM SubscribeMany | `dm.go:127` | No timeout/context on subscription |
| Community posts | `community.go:222` | No timeout/context on subscription |

### 2. Caching Strategy

#### Implementation

**CacheEvent** (`utils/get.go:105-119`):
```go
func CacheEvent(event *nostr.Event, app *config.AppContext) {
    if !shouldCache(event, app) { return }
    go func() {
        app.Pool().PublishMany(context.Background(), privateRelays, *event)
    }()
}
```
- Publishes to private relays asynchronously (goroutine, no context)
- `shouldCache` matches against `CacheFilters` from config

**SubscribeWithCache** (`utils/get.go:73-84`):
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
- Wraps subscription channel, caches each incoming event

**Default CacheFilters** (`config/config.go:149-168`):
- Kinds: 0 (profile), 3 (contacts), 10002 (relay list), 10050 (channel)
- Plus all events authored by self

#### Observations

| Aspect | Finding |
|---|---|
| Cache backend | No client-side cache (LMDB/boltdb only used for local relay storage, not client-side) |
| Cache invalidation | None — events cached by publishing to private relays |
| Cache consistency | Fire-and-forget, no confirmation of write |
| Filter matching | `CacheFilter.ToNostr()` creates nostr.Filter, matched via `Matches()` |

### 3. Relay Pool Architecture

#### Relay Categorization (`config/context.go`)

| Method | Returns | Used By |
|---|---|---|
| `WritableRelays()` | relays with Write=true | Publishing events |
| `ReadableRelays()` | relays with Read=true | Querying events |
| `AllWritableRelays()` | Writable + prepend local relay | Publishing |
| `AllReadableRelays()` | Readable + prepend local relay | Querying |
| `PrivateRelays()` | private relays + prepend local relay | DM, cache propagation |
| `ListDMRelays()` | separate DM relay list | DM operations |
| `ListSearchRelays()` | separate search relay list | Search operations |
| `ListSubscriptions(subType)` | followed users/communities/hashtags | Timeline filtering |

#### Local Relay Priority (`config/context.go:85-98`)

```go
func (a *AppContext) AllReadableRelays() []string {
    relays := a.ReadableRelays()
    if localURL := a.localRelayURL(); localURL != "" {
        relays = append([]string{localURL}, relays...)
    }
    return relays
}
```

- Local relay (`ws://localhost:8989`) prepended to all relay lists when enabled
- Single point of control: `LocalRelayEnabled()` check in `localRelayURL()`
- Port configurable via `local_relay.port` (default 8989)
- Data dir: `~/.cache/nosmec` (via `os.UserCacheDir()`)

#### Observations

| Finding | Notes |
|---|---|
| Relay list management | AppContext has AddRelay/RemoveRelay/SetRelayRead/SetRelayWrite |
| Known relays tracking | `TrackRelays()` accumulates discovered relays, persisted on `Close()` |
| No relay health monitoring | All relays treated equally, no latency/availability tracking |
| No relay selection strategy | All readable relays queried for every request |

### 4. TUI Architecture

#### Components

| Component | File | Pattern |
|---|---|---|
| `timeline.model` | `tui/timeline/model.go` | `tea.Model`, `Init() Update() View()` |
| `EventView` | `tui/window/event/event.go` | `tea.Model`, implements `Window` interface |
| `WindowManager` | `tui/windowmanager/windowmanager.go` | Manages multiple `Window` instances, stack-based z-order |
| `Window` interface | `tui/window/window.go` | `Init() Update() View() ID()` |
| `ToolKit` | `tui/toolkit/toolkit.go` | Keymap registration, message routing |

#### Message Flow

**Timeline model** (`tui/timeline/model.go`):
- Initial load: `fetchTimeline()` via `tea.Cmd` returning `fetchMsg`
- Profile names: `fetchProfileNames()` via WaitGroup, returns `namesMsg`
- Subscription: `startSubscription()` sets up `SubscribeWithCache`, returns `pollSubMsg`
- Polling: `pollSubscription()` uses 100ms `tea.Tick` to check subscription channel

**WindowManager** (`tui/windowmanager/windowmanager.go:28-61`):
- `Open(win Window)` — adds to map, pushes to stack top, calls `Init()`, returns `tea.Cmd`
- `Close(id string)` — removes from map and stack, refocuses next in stack
- `UpdateFocused(msg)` — routes to topmost window only

**EventView** (`tui/window/event/event.go:294-340`):
- Handles `CloseMsg` → returns `tea.Quit`
- `EventLoadedMsg` triggers profile name fetch
- `ProfileLoadedMsg` updates author name
- Key handling via `ToolKit` (`m.tk.HandleMsg(WindowID, msg)`)

#### Observations

| Finding | Notes |
|---|---|
| Infinite scroll | `fetchMoreOld()` triggered when paginator approaches last page |
| Subscription deduplication | `seenEventIDs map[nostr.ID]bool` prevents duplicates |
| Window resize propagation | `tea.WindowSizeMsg` forwarded to WindowManager → all windows |
| CloseMsg handling | Timeline handles `event.CloseMsg` at lines 585-588, EventView handles at line 298 |
| Message routing | KeyPressMsg goes to focused window, CloseMsg goes to windowmanager.Close |

### 5. CLI Command Structure

#### Registration Pattern (`cmd/registry.go`)

```go
type CommandRegistrar func(*cobra.Command) error

var commandRegistrar CommandRegistrar

func RegisterCommands(fn CommandRegistrar) { ... }
func initCommands() {
    if commandRegistrar == nil {
        registerDefaultCommands()
    } else {
        commandRegistrar(rootCmd)
    }
}
```

**Default commands** (`cmd/registry.go:23-30`):
- `registerNoteCommands()` — note create, delete, reply, quote
- `registerConfigCommands()` — relay, dm-relay, search-relay, alias
- `registerProfileCommands()` — profile get, set, sync
- `registerDMCommands()` — dm send, conversations, history
- `registerCommunityCommands()` — community create, post, reply, list, info
- `registerEventCommands()` — event detail, raw

#### Observations

| Finding | Notes |
|---|---|
| Note vs Event distinction | `note_commands.go` (CRUD), `event_commands.go` (detail/raw) — unclear why separated |
| Grouping | `RegisterCommandGroup(name, description, ...cmds)` allows logical grouping |
| No subcommand hierarchy | All flat under root, no `note create`, just `create` with note-specific behavior |
| No command-level context | Each command creates its own AppContext? Need to verify |

### 6. Configuration Layering

#### Viper Configuration (`config/config.go:69-172`)

**Config file search paths** (line 82-85):
```go
globalViper.AddConfigPath(configDir)        // ~/.config/nosmec
globalViper.AddConfigPath("$HOME/.config")
globalViper.AddConfigPath("./")
globalViper.AddConfigPath(".")
```

**Environment variable mapping** (line 87-89):
```go
globalViper.SetEnvPrefix("NOSMEC")
globalViper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
globalViper.AutomaticEnv()
```
- `local_relay.enabled` → `NOSMEC_LOCAL_RELAY_ENABLED`

**Config defaults** (lines 91-103):
- `local_relay.enabled = true`
- `local_relay.port = 8989`
- `known_relays = []`
- `private_key = ""`
- `proxy.socks = ""`

#### AppContext Configuration Access

- All config access via `AppContext.Config()` which returns a copy (RLock)
- `AppContext.GetProfile()`, `AppContext.GetPrivateKey()` — specific getters
- `AppContext.SetProfile()`, `AppContext.SetPrivateKey()` — setters that write via viper

#### Observations

| Finding | Notes |
|---|---|
| Single instance lock | `NewLMDB()` creates lock file with PID, checks for stale lock on startup |
| Auto key generation | If no private key, generates one on first load (lines 138-147) |
| Config file creation | Writes default config if file doesn't exist (lines 116-120) |
| No env override for relays | Relays must be in config file, not env vars |

---

## Current Design Patterns (Summary)

| Category | Pattern | Location |
|---|---|---|
| Query: sync | Local-first with timeout, then remote fallback | `utils/get.go:17-71` |
| Query: async | Goroutine + channel signal + select | `utils/get.go:219-266` |
| Query: streaming | FetchMany channel iteration | `utils/get.go:268-342` |
| Query: replaceable | FetchManyReplaceable with Range | `utils/get.go:30-36` |
| Caching | Async publish to private relays | `utils/get.go:105-119` |
| Caching: subscription | SubscribeWithCache wrapper | `utils/get.go:73-84` |
| Relay priority | Local relay prepended to AllReadable/Writable | `config/context.go:85-98` |
| TUI: model | tea.Model with Init/Update/View | `tui/timeline/model.go:89-117` |
| TUI: window | Window interface with ID() | `tui/window/window.go` |
| TUI: multi-window | WindowManager stack-based z-order | `tui/windowmanager/windowmanager.go` |
| TUI: subscription polling | 100ms tea.Tick for channel non-blocking read | `tui/timeline/model.go:427-462` |
| CLI: registration | CommandRegistrar callback + RegisterCommandGroup | `cmd/registry.go:7-49` |
| Config: layering | Viper with env prefix + multiple file paths | `config/config.go:69-172` |

---

## Inconsistencies and Violations

| # | Issue | Location |
|---|---|---|
| 1 | Query timeout not configurable — 2s/10s hardcoded | `utils/get.go:46,55` |
| 2 | Local relay priority only for non-replaceable events | `utils/get.go:30-36 vs 38-62` |
| 3 | `GetEventAsync` uses manual channel instead of `context.WithTimeout` like `GetEvent` | `utils/get.go:241-256` |
| 4 | Fire-and-forget goroutines for caching with no context or cancellation | `utils/get.go:65-68,116-118` |
| 5 | No client-side persistent cache — all caching is relay-to-relay propagation | `utils/get.go:105-119` |
| 6 | DM SubscribeMany has no timeout context | `utils/dm.go:127` |
| 7 | Community posts SubscribeMany has no timeout context | `utils/community.go:222` |
| 8 | No relay health/latency tracking — all relays queried equally | `config/context.go` |
| 9 | Note vs Event command split unclear — no clear principle | `cmd/note_commands.go` vs `cmd/event_commands.go` |
| 10 | WindowManager.Update returns only topmost window model — rest ignored | `tui/windowmanager/windowmanager.go:93-113` |
| 11 | EventView.Close() always returns true — no actual close check | `tui/window/event/event.go:363-365` |
| 12 | Config reload not implemented — changes require restart | `config/context.go` |
| 13 | knownRelays persisted only on AppContext.Close(), not on discovery | `config/context.go:406-412` |

---

## Recommended Design Principles

### Query Principles

1. **Timeout propagation**: All query functions should accept context and use `context.WithTimeout` consistently. Hardcoded timeout values should be configurable via config.

2. **Local relay priority for all queries**: The local-first pattern used in `GetEvent` for non-replaceable events should be applied uniformly — both for read and write path consistency.

3. **Replaceable event handling**: Use `FetchManyReplaceable` for all replaceable/event-kind queries, not just in the sync `GetEvent` path.

4. **Subscription context**: All `SubscribeMany` calls should pass a context with timeout for clean termination.

### Caching Principles

5. **Client-side cache**: Consider using the existing LMDB store (via `GlobalLMDB()`) for client-side caching of fetched events to reduce relay queries.

6. **Cache confirmation**: Consider tracking cache write confirmations instead of fire-and-forget publish.

7. **Cache invalidation**: Implement cache invalidation or TTL for cached events.

### Relay Architecture Principles

8. **Relay health tracking**: Add latency or availability tracking to allow intelligent relay selection (e.g., prefer relays with lower latency).

9. **Relay list persistence**: Persist discovered relays more frequently than just on shutdown.

### TUI Principles

10. **Window update return values**: WindowManager.Update should return all window models and commands for proper composition, not just the topmost.

11. **Close semantics**: EventView.Close() should return actual close state, not always true.

12. **Subscription cleanup**: Ensure subscriptions are properly cancelled on window close.

### CLI Principles

13. **Command grouping**: Clarify the distinction between "note" and "event" commands or consolidate them into a single domain.

14. **Context reuse**: Verify that commands share AppContext instance rather than creating new ones.

### Configuration Principles

15. **Hot config reload**: Implement config file watching with `viper.OnConfigChange` for runtime config updates.

16. **Timeout configuration**: Move hardcoded timeout values (2s, 10s) to config with env var overrides.