# Research: relay selection for kind:34550 queries

- **Query**: How does `defaultRelaysForFilter` handle kind:34550 queries and whether the hardcoded `RelayListRelays` in `NewSystem()` are working correctly for community discovery?
- **Scope**: internal
- **Date**: 2026-05-20

## Findings

### Files Found

| File Path | Description |
|---|---|
| `nostr_sdk/system.go:123-155` | `NewSystem()` — initializes hardcoded `RelayListRelays` |
| `nostr_sdk/system.go:553-586` | `defaultRelaysForFilter()` — relay routing logic |
| `nostr_sdk/system.go:490-549` | `FetchEventsByFilter()` — calls `defaultRelaysForFilter` |
| `config/context.go:26-42` | `NewAppContext()` — wires `GlobalSystem` singleton |
| `utils/community.go:26-33` | `FetchCommunityEvents()` — calls `app.System().FetchEventsByFilter()` |
| `cmd/community_commands.go:193-207` | `community list` — uses `app.Pool().FetchMany` directly, not `FetchEventsByFilter` |
| `tui/community/discover/model.go:168-180` | `model.Init()` — loads communities via `FetchCommunityEvents` |

### Code Patterns

**`NewSystem()` hardcodes `RelayListRelays`** (system.go:126):
```go
RelayListRelays: NewRelayStream(
    "wss://indexer.coracle.social",
    "wss://purplepag.es",
    "wss://relay.primal.net",
    "wss://relay.nos.social",
),
```

**`defaultRelaysForFilter` special-cases kind:34550** (system.go:574-580):
```go
if slices.Contains(filter.Kinds, nostr.KindCommunityDefinition) && len(sys.RelayListRelays.URLs) > 0 {
    relays := append([]string{}, sys.RelayListRelays.URLs...)
    if len(sys.FallbackRelays.URLs) > 0 {
        relays = nostr.AppendUnique(relays, sys.FallbackRelays.URLs...)
    }
    return relays
}
```

**`FetchCommunityEvents` flow** (utils/community.go:26-33):
```go
filter := nostr.Filter{Kinds: []nostr.Kind{nostr.KindCommunityDefinition}}
events, err := app.System().FetchEventsByFilter(ctx, filter, nostr_sdk.FetchEventsOptions{
    SaveToLocalStore: true,
})
```

**`FetchEventsByFilter` relay selection** (system.go:530-533):
```go
relays := opts.Relays
if len(relays) == 0 {
    relays = sys.defaultRelaysForFilter(ctx, filter)
}
```

### Relay Routing Chain

1. `tui/community/discover/model.go:loadCommunities()` → `utils.FetchCommunityEvents()`
2. `utils.FetchCommunityEvents()` → `app.System().FetchEventsByFilter()` with kind:34550 filter
3. `FetchEventsByFilter()` → `sys.defaultRelaysForFilter(ctx, filter)`
4. `defaultRelaysForFilter()` → detects `KindCommunityDefinition` → returns `RelayListRelays + FallbackRelays`

### `config/context.go` Wiring

`NewAppContext()` (lines 26-42) uses a singleton `GlobalSystem`:
```go
sys := GlobalSystem
if sys == nil {
    sys = nostr_sdk.NewSystem()
    GlobalSystem = sys
}
```

The `RelayListRelays` is initialized **hardcoded in `NewSystem()`**, NOT from `config.RelayList`. The user's configured relays from `config/context.go` (`ReadableRelays()`) are NOT used for this purpose.

### `cmd/community_commands.go` Direct Pool Query

The `community list` CLI command (cmd/community_commands.go:193-207) queries using `app.Pool().FetchMany()` directly with `FallbackRelays`, NOT `FetchEventsByFilter` — it does NOT benefit from the `defaultRelaysForFilter` kind:34550 routing.

### Key Observation

**`RelayListRelays` is hardcoded in `NewSystem()`** (system.go:126), NOT derived from user's relay config. The four relays are:
- `wss://indexer.coracle.social`
- `wss://purplepag.es`
- `wss://relay.primal.net`
- `wss://relay.nos.social`

The `defaultRelaysForFilter` check at line 574 **requires `len(sys.RelayListRelays.URLs) > 0`** — since it's hardcoded with 4 relays, this condition is always true in `NewSystem()`.

### Condition Check

The condition at system.go:574 is:
```go
if slices.Contains(filter.Kinds, nostr.KindCommunityDefinition) && len(sys.RelayListRelays.URLs) > 0
```

Since `RelayListRelays` is initialized with 4 URLs in `NewSystem()`, `len(sys.RelayListRelays.URLs) > 0` is always true. The function should return `RelayListRelays + FallbackRelays` for kind:34550 queries.

### Community Discover TUI Flow

`RunCommunityDiscover` → `loadCommunities()` → `FetchCommunityEvents()` → `app.System().FetchEventsByFilter()` → `defaultRelaysForFilter()` → `RelayListRelays + FallbackRelays`

### Caveats / Not Found

- `community discover` PTY command failed to spawn — could not verify runtime behavior
- `config/context.go`'s `ReadableRelays()` is NOT connected to `RelayListRelays` — `RelayListRelays` is hardcoded in `NewSystem()`
- `cmd/community_commands.go:193-207` does NOT use `FetchEventsByFilter` — it queries `FallbackRelays` directly for kind:34550

## Related Specs

- `.trellis/spec/backend/forked-sdk-architecture.md` — mentions kind:34550 addressable loader registration
- `.trellis/tasks/05-20-note-timeline-global-community-discover-kind/prd.md` — task PRD