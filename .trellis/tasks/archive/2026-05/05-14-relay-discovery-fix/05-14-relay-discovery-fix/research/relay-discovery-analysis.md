# Research: relay-discovery-fix — Current Implementation Analysis

- **Query**: How does relay discovery work in nosmec? Analyze DiscoverUserRelays, DiscoverAndVerifyRelays (spec only), event relay hints, and relay combination logic.
- **Scope**: internal
- **Date**: 2026-05-14

## Findings

### Files Found

| File Path | Description |
|---|---|
| `utils/user_relays.go` | `DiscoverUserRelays` (NIP-65 discovery), `EnsureRelays` |
| `utils/relay_list.go` | `SyncRelaysFromNetwork`, `PublishRelayList` for kind 10002/10050 |
| `utils/get.go` | `GetProfile`, `GetEvent`, relay selection strategy, local relay caching |
| `config/context.go` | `AllReadableRelays`, `WritableRelays`, `TrackRelays`, `KnownRelays` |
| `.trellis/spec/backend/relay-guidelines.md` | Design spec including `ExtractRelayHints` (not yet implemented) |

---

## 1. DiscoverUserRelays (`utils/user_relays.go:13-53`)

```go
func DiscoverUserRelays(ctx context.Context, app *config.AppContext, pubKey nostr.PubKey) ([]string, error)
```

**How it works:**
1. Builds a filter for `KindRelayListMetadata` (kind 10002) from the target user
2. Queries `app.AllReadableRelays()` — this includes local relay prepended to configured relays
3. Uses `FetchManyReplaceable` to get the latest kind 10002 event
4. Parses the event with `nip65.ParseRelayList` → returns `readRelays`, `writeRelays`
5. **Ensures** relays are registered in the pool (lazy connection)
6. **Tracks** relays via `app.TrackRelays` (merged into `KnownRelays` on `Close()`)
7. Returns only **read relays** (line 52)

**Returns `nil, nil` when:**
- `AllReadableRelays()` is empty (line 22-24)
- No kind 10002 event found on any relay (line 35-37)

**Problem**: If `AllReadableRelays()` returns only the local relay and the target user has never posted to it, discovery fails and returns nil.

---

## 2. DiscoverAndVerifyRelays (spec only — NOT implemented)

Defined in `relay-guidelines.md:88-95` but **not present in the codebase**:

```go
// Spec says:
func DiscoverAndVerifyRelays(ctx context.Context, app *config.AppContext, filter nostr.Filter) ([]string, error)
```

- Queries relays for events matching a filter (e.g., kind 10002 from any user)
- Parses `["r", <url>]` or `["r", <url>, "read"|"write"]` tags
- **Verifies connectivity** via `RelayConnect` + `IsConnected` — only reachable relays returned
- Returns empty list if all relays unreachable

**Status**: Not implemented. This is a planned function.

---

## 3. GetProfile Relay Combination (`utils/get.go:100-144`)

`GetProfile` follows this relay precedence:

```
1. User's own NIP-65 read relays (from DiscoverUserRelays)
2. app.AllReadableRelays() — local relay + configured read relays
```

**Code path:**
```go
// Line 104: Discover user relays
discovered, err := DiscoverUserRelays(ctx, opts.App, pubKey)

// Line 118-140: If user relays found, prepend them to AllReadableRelays
if relayOpts != nil && len(userRelays) > 0 {
    combined := make([]string, 0, len(userRelays)+len(baseRelays))
    seen := make(map[string]bool)
    // User relays first (most specific)
    for _, r := range userRelays { ... }
    // Then base relays (local + configured)
    for _, r := range baseRelays { ... }
}
```

**Fallback chain (in GetEvent, `get.go:22-24`):**
```go
relays := opts.Relays
if len(relays) == 0 {
    relays = opts.App.AllReadableRelays()  // local + configured
}
```

---

## 4. Event Relay Hints (e/p/a/q tags) — NOT USED in Fetch

**Spec defines** (`relay-guidelines.md:25-38`):
```go
func ExtractRelayHints(event *nostr.Event) (relays []string) {
    for _, tag := range event.Tags {
        if len(tag) < 2 { continue }
        switch tag[0] {
        case "e", "p", "a", "q":
            if len(tag) >= 3 && tag[2] != "" {
                relays = append(relays, tag[2])
            }
        }
    }
    return relays
}
```

**Current status**: `ExtractRelayHints` is **defined in the spec but NOT implemented in code**. There is no function that extracts relay hints from event tags and uses them for follow-up queries.

**Places that format these tags for input** (TUI compose, NIP.md):
- `tui/compose/model.go:171` — placeholder text shows `e:eventId p:pubkey a:addr r:relay:purpose q:eventId`
- `docs/NIP.md` — documents tag formats

**Where hints could be used but aren't**:
- `GetNote` / `GetNoteAsync` — fetches by event ID only, does not extract or use `e` tag relay hints
- Profile fetch does not use `p` tag relay hints from events referencing the profile owner
- Thread/comment resolution does not use `e`/`a` relay hints from reply tags

---

## 5. Local Relay Factor

**Local relay is always included in read path** (`config/context.go:91-97`):
```go
func (a *AppContext) AllReadableRelays() []string {
    relays := a.ReadableRelays()
    if localURL := a.localRelayURL(); localURL != "" {
        relays = append([]string{localURL}, relays...)  // PREPENDED first
    }
    return relays
}
```

**Local relay is EXCLUDED from write path**:
```go
func (a *AppContext) AllWritableRelays() []string {
    return a.WritableRelays()  // No local relay
}
```

**Local relay URL**: `ws://localhost:<port>` (default 8989), configurable via `LocalRelay.Enabled` and `LocalRelay.Port`.

**Used for**: Caching via `CacheEvent` (`get.go:84-98`), which publishes events to local relay asynchronously.

---

## 6. What Happens When DiscoverUserRelays Returns nil/empty

**In GetProfile (`get.go:101-108`):**
```go
discovered, err := DiscoverUserRelays(ctx, opts.App, pubKey)
if err == nil {
    userRelays = discovered
}
```

**The error is silently swallowed.** If `DiscoverUserRelays` returns `nil, nil` (no event found):
- `userRelays` is nil/empty
- The `if len(userRelays) > 0` check at line 118 **fails**
- `relayOpts` remains as `opts` (the original options)
- `GetEvent` falls back to `opts.App.AllReadableRelays()`

**So discovery failure is graceful** — it falls back to configured relays. But if `AllReadableRelays()` is also empty (no local relay, no configured relays), the query simply won't happen.

---

## Summary: Current Discovery Chain

```
GetProfile / GetEvent
  ↓
opts.App.AllReadableRelays()        ← local relay PREPENDED to configured
  ↓
(no event relay hints extraction — spec only)
  ↓
DiscoverUserRelays (for profiles only)
  ↓
userRelays returned → prepended to AllReadableRelays
  ↓
nip65.ParseRelayList(event)         ← parses kind 10002
  ↓
TrackRelays / EnsureRelays           ← registers relays in pool
```

**Missing pieces:**
1. `DiscoverAndVerifyRelays` — not implemented (spec only)
2. `ExtractRelayHints` — not implemented (spec only)
3. No use of `e/p/a/q` relay hints when fetching referenced events
4. `DiscoverUserRelays` returns nil silently on failure — no fallback retry with `KnownRelays`

---

## Related Specs

- `.trellis/spec/backend/relay-guidelines.md` — full relay design spec including `ExtractRelayHints` and `DiscoverAndVerifyRelays`
