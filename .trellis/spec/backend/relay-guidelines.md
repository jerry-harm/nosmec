# Relay Guidelines

> NIP compliance and relay configuration patterns.

---

## Event-Provided Relay Hints (NIP-01, NIP-10)

Events can embed recommended relay URLs in their tags. These hints should be used for targeted querying.

### Tag Format

| Tag | Format | Description |
|-----|--------|-------------|
| `e` | `["e", <id>, <relay>, <marker>, <pubkey>]` | Reference to another event (NIP-10). `<marker>` is `"reply"` or `"root"` |
| `p` | `["p", <pubkey>, <relay>]` | Reference to another user |
| `a` | `["a", "<kind>:<pubkey>:<d>", <relay>]` | Reference to an addressable event |
| `q` | `["q", <event-id>, <relay>, <pubkey>]` | Quote/reference to an event (NIP-10) |

`<relay>` is the **recommended relay URL** for querying the referenced event/user. It is optional but should be used when present.

### Relay Hint Extraction

```go
func ExtractRelayHints(event *nostr.Event) (relays []string) {
    for _, tag := range event.Tags {
        if len(tag) < 2 {
            continue
        }
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

### Query Strategy with Relay Hints

When fetching an event that was referenced by another event:

1. Extract relay hints from the referencing event's tags
2. Query those relays first
3. If not found, fallback to `AllReadableRelays()`

```go
func GetEventWithHint(ctx context.Context, eventID string, hintRelay string, opts *GetOptions) *nostr.Event {
    relays := []string{}
    if hintRelay != "" {
        relays = append(relays, hintRelay)
    }
    if len(relays) == 0 {
        relays = opts.App.AllReadableRelays()
    }
    return GetEvent(ctx, nostr.Filter{IDs: []nostr.ID{eventID}}, &GetOptions{
        App:    opts.App,
        Relays: relays,
    })
}
```

---

## Relay Discovery from NIP-65 (kind:10002)

When querying a user's profile or other data, discover their NIP-65 relay list first.

### DiscoverUserRelays

```go
func DiscoverUserRelays(ctx context.Context, app *config.AppContext, pubKey nostr.PubKey) ([]string, error)
```

- Queries `AllReadableRelays()` for Kind 10002 from the user
- Parses read/write relay list from event tags
- Calls `EnsureRelays` to register discovered relays in the pool
- Calls `TrackRelays` to add to `knownRelays` (persisted on `Close()`)
- Returns the user's read relays

### Relay List Discovery with Verification

For fetching other's relay lists (not the user's own):

```go
func DiscoverAndVerifyRelays(ctx context.Context, app *config.AppContext, filter nostr.Filter) ([]string, error)
```

- Queries relays for events matching the filter (e.g., Kind 10002 from any user)
- Parses `["r", <url>]` or `["r", <url>, "read"|"write"]` tags
- Verifies each relay connectivity via `RelayConnect` + `IsConnected`
- Returns only reachable relays (read + write combined, not distinguished)
- Returns empty list if all relays unreachable

### When KnownRelays Are Updated

| Trigger | Method | Timing |
|---------|--------|--------|
| NIP-65 discovery (per-user) | `DiscoverUserRelays` | On-demand during profile queries |
| Config persistence | `TrackRelays` → `Close()` | Only on app shutdown |
| Network sync (self only) | `SyncRelaysFromNetwork` | Manual `config sync` command |

**Note**: Relay connectivity is verified only at config persistence time. During runtime, Pool uses lazy connection — unreachable relays are ignored by the pool.

---

## NIP Relay List Events

### NIP-65 — Relay List Metadata (kind:10002)

Published to advertise user's preferred read/write relays.

```go
// kind:10002 structure
{
  "kind": 10002,
  "tags": [
    ["r", "wss://relay1.com"],              // both read+write
    ["r", "wss://relay2.com", "write"],    // write only
    ["r", "wss://relay3.com", "read"]      // read only
  ],
  "content": ""
}
```

Published via `utils.PublishRelayList(ctx, app)`.

### NIP-17 — DM Relay List (kind:10050)

Published to advertise user's DM inbox relays.

```go
// kind:10050 structure
{
  "kind": 10050,
  "tags": [
    ["relay", "wss://inbox.nostr.wine"],
    ["relay", "wss://myrelay.nostr1.com"]
  ],
  "content": ""
}
```

Published via `utils.PublishRelayList(ctx, app)` (same function handles both).

---

## Relay Selection Strategy

When a `GetOptions` has no explicit relay list, functions must follow this fallback order:

```go
relays := opts.Relays
if len(relays) == 0 {
    relays = opts.App.AllReadableRelays()   // local + configured relays
}
if len(relays) == 0 {
    relays = opts.App.Config().KnownRelays  // discovered fallback
}
```

**Why**: `AllReadableRelays()` includes the local relay (cache) first, which provides resilience when configured relays fail. `KnownRelays` is a last-resort pool of relays discovered from NIP-65.

**Functions following this pattern**: `GetEvent`, `GetEventAsync`, `GetProfile`, `GetMyTimeline`, `GetGlobalTimeline`, `GetFollowedTimeline`

**Functions with special handling**:
- `GetNote`/`GetNoteAsync`: Cannot discover author relays without fetching event first — the event contains the author pubkey
- `GetProfile`: Calls `DiscoverUserRelays` first to find author's NIP-65 relays, then prepends them to the relay list

---

## Local Relay Role

| Direction | Local Relay Included? | Rationale |
|-----------|---------------------|-----------|
| Read path | ✅ Yes (prepended first) | Local relay is cache — serves hits without network round-trip |
| Write path | ❌ No | Local relay is backup/cache only — never the primary write target |

```go
func (a *AppContext) AllReadableRelays() []string {
    relays := a.ReadableRelays()
    if localURL := a.localRelayURL(); localURL != "" {
        relays = append([]string{localURL}, relays...)
    }
    return relays
}

func (a *AppContext) AllWritableRelays() []string {
    return a.WritableRelays()  // local relay EXCLUDED
}
```

---

## Convention: Auto-publish on Config Mutation

**When relay configuration is mutated via CLI, always publish the updated relay list.**

```bash
# After these commands, PublishRelayList MUST be called:
nosmec config relay add <url>        # then PublishRelayList
nosmec config relay remove <url>     # then PublishRelayList
nosmec config relay sync             # then PublishRelayList
nosmec config dm-relay add <url>     # then PublishRelayList
nosmec config dm-relay remove <url>  # then PublishRelayList
nosmec config dm-relay sync         # then PublishRelayList
```

**Why**: `PublishRelayList` existed in `utils/relay_list.go` but was not wired to CLI mutations — relay lists were configured locally but never broadcast to the network.

---

## Relay Configuration Semantics

| Field | Purpose |
|-------|---------|
| `RelayList` | User's relays with read/write flags (NIP-65) |
| `PrivateRelays` | Private data relay (DMs, follows — sensitive data) |
| `DMRelays` | DM inbox relays (NIP-17 kind:10050) |
| `SearchRelays` | Search-only relay (BLEVE index queries) |
| `KnownRelays` | Fallback relay discovery list |

---

## Files

- `utils/relay_list.go` — `PublishRelayList`, `SyncRelaysFromNetwork`
- `utils/user_relays.go` — `DiscoverUserRelays`, `EnsureRelays`
- `config/types.go` — relay config structs
- `config/context.go` — relay helper methods, `TrackRelays`, `Close`
- `utils/get.go` — `GetEvent`, `GetProfile`, relay selection strategy
