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

## NIP-10 Reply Tag Generation (Full 5-Field)

Per [NIP-10](https://github.com/nostr-protocol/nips/blob/master/10.md), marked e tags use the full format:

```
["e", <event-id>, <relay-url>, <marker>, <pubkey>]
```

- `<relay-url>` — **SHOULD** be the relay where the referenced event was found
- `<pubkey>` — **OPTIONAL**; SHOULD be the hexagonal pubkey of the referenced event's author
- Backward compatible: parsers read tag[1] (ID) and tag[3] (marker); tag[2] and tag[4] are additive

> **Warning — tag length safety**: Always check `len(tag) >= N` before accessing `tag[N]`. The 5-field format means tags can be longer than 4 fields. Existing parsers in `extractParentID` (tui/thread/thread.go) and `FindRootEvent` (utils/get.go) only read tag[1] and tag[3], so tag[4] addition does not break parsing.

### BuildReplyTags — Contract

```go
// Located in utils/post.go

// BuildReplyTags creates NIP-10 marked e tags for a reply to a parent event.
// Relay URLs are looked up from the event→relay tracking map (see below).
// Pubkey is taken from parentEvent.PubKey.Hex().
func BuildReplyTags(parentEvent *nostr.Event) nostr.Tags
```

| Scenario | Returns |
|----------|---------|
| Direct reply (parent IS root) | 1 tag: `["e", rootID, relay, "root", pubkey]` |
| Nested reply (parent HAS root marker) | 2 tags: `["e", rootID, relay, "root"]` + `["e", parentID, relay, "reply", pubkey]` |
| Empty parent event | Empty tags |

**Root event pubkey**: For nested replies, the root event object is not available (only its ID from the parent's tags), so the root tag's `<pubkey>` field is left empty. If the root event is needed, fetch it with `FetchSpecificEvent`.

**Callers**: `ReplyToNote` (utils/post.go), `compose.AddReply` (tui/compose/model.go).

---

## Event→Relay Tracking for NIP-10 e Tags

When building reply tags (NIP-10), we need the relay URL where a referenced event was fetched from. This is tracked via `sdk.System` backed by LMDB `KVStore` (persists across restarts).

### API (config package — delegates to sdk.System)

```go
// TrackEventRelay records which remote relay(s) an event was fetched from.
// The compatibility wrapper appends into the SDK relay-list encoding.
func TrackEventRelay(eventID, relayURL string)

// GetEventRelay returns the first known relay URL for an event.
// Returns "" if never tracked.
func GetEventRelay(eventID string) string
```

### Population

Called inside `Pool.EventMiddleware` (config/config.go) for every incoming event:

```go
if ev.ID != [32]byte{} {
    TrackEventRelay(ev.ID.Hex(), ie.Relay.URL)
}
```

### Storage: KVStore (LMDB)

Event→relay mappings are stored in `sdk.System.KVStore` backed by LMDB (`nostr_sdk/kvstore/lmdb`). The KVStore directory lives at `{dataDir}/kvstore/`.

Keys: `'r' + first 8 bytes of event ID` → compact binary relay-list bytes

`TrackEventRelay` uses `kvstore.Update()` to preserve the SDK event-relay encoding while appending unseen relay URLs.

### Thread Safety

KVStore via LMDB is thread-safe within a single process. No external mutex needed — LMDB handles concurrent reads and serializes writes internally.

### Known Relay List Read Path

`relay list` must not open LMDB directly from the command layer.

```go
// cmd/relay_commands.go
func writeRelayList(w io.Writer, app *config.AppContext) error {
    sys := app.System()
    if sys == nil {
        return nil
    }
    relays, err := sys.ListKnownEventRelays()
    if err != nil {
        return err
    }
    ...
}
```

`System.ListKnownEventRelays()` is the owning read API for this behavior. It is responsible for:

1. reading relay URLs learned through `HintsDB`
2. reading event→relay mappings from `KVStore`
3. merging, deduplicating, and sorting the final list

Wrong:

```go
// cmd layer opens LMDB and decodes relay-list bytes itself
env, _ := lmdb.NewEnv()
... decodeKVRelayList(v)
```

Correct:

```go
relays, err := app.System().ListKnownEventRelays()
```

Why:
- command code should not know KVStore layout, DBI names, or relay byte encoding
- relay-list storage details belong to `nostr_sdk`
- changing on-disk relay encoding should not require CLI-layer updates

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
- Calls `TrackRelays` to update the legacy config fallback cache
- Returns the user's read relays

### Relay List Discovery with Verification

> **Deprecated**: This approach is no longer used. Relays are now added unconditionally and connectivity is verified only at config save time.

Previously (now abandoned):
- `DiscoverAndVerifyRelays` queried relays for events matching a filter, parsed relay tags, verified connectivity via `RelayConnect` + `IsConnected`, and returned only reachable relays.

### Relay Discovery Runtime Behavior

- `DiscoverUserRelays` queries currently configured readable relays
- discovered relays are `EnsureRelay`'d into the pool for the current session
- relay auto-learning continues through `HintsDB` via `Pool.EventMiddleware`

**Note**: There is no config-backed `KnownRelays` persistence path anymore. During runtime, Pool uses lazy connection — unreachable relays are ignored by the pool. HintsDB scoring decays over time (^1.3), so stale associations naturally fade.

---

## NIP Relay List Events

### NIP-65 — Relay List Metadata (kind:10002)

Published to advertise user's preferred read/write relays.

**Tag Rules** (from [NIP-65](https://github.com/nostr-protocol/nips/raw/refs/heads/master/65.md)):
- `["r", <url>]` — relay is **both read AND write** (no marker = both)
- `["r", <url>, "read"]` — relay is **read only**
- `["r", <url>, "write"]` — relay is **write only**
- **Never** create separate tags for the same relay with read AND write markers

```go
// kind:10002 structure
{
  "kind": 10002,
  "tags": [
    ["r", "wss://relay1.com"],              // both read+write (no marker)
    ["r", "wss://relay2.com", "write"],    // write only
    ["r", "wss://relay3.com", "read"]      // read only
  ],
  "content": ""
}
```

Published via `utils.PublishRelayList(ctx, app)`.

**Parsing**: When reading NIP-65 tags, if len(tag)==2 the relay is both read+write. Otherwise check for "read"/"write" markers.

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
    relays = utils.GetQueryRelays(opts.Event, opts.App)
}
```

**`GetQueryRelays` priority** (implemented in `utils/user_relays.go`):

1. **tag[2] relay hints** — `ExtractRelayHints(event)` from e/p/a/q tags
2. **HintsDB outbox** — `app.Hints().TopN(pubkey, 3)` from e tag[3] author pubkeys
3. **AllReadableRelays()** — configured relays only (no local relay)
**HintsDB** (`config/hints.go`): learns relay→pubkey from every incoming event via Pool.EventMiddleware.
- HintEventFetched (700pts): successfully received event from relay
- HintInRelayList (350pts): author listed relay in kind:10002
- HintFromTag (20pts): p-tag relay hint
- Scoring: `basePoints * 1e10 / (age + 86400)^1.3` (same formula as nostr SDK)

**Why**: `AllReadableRelays()` provides configured relays; `sdk.System.Store` (LMDB/Bleve) handles local caching of fetched events.

**Functions following this pattern**: `GetEvent`, `GetEventAsync`, `GetMyTimeline`, `GetGlobalTimeline`, `GetFollowedTimeline`

**Functions with special handling**:
- `GetNote`/`GetNoteAsync`: Cannot discover author relays without fetching event first — the event contains the author pubkey
- `GetProfile`: Uses `sdk.System.FetchProfileMetadata` which handles cache → store → network automatically

---

## Profile Fetch Strategy

`GetProfile` delegates to `sdk.System.FetchProfileMetadata(ctx, pubkey)`:

1. Check `MetadataCache` (in-memory LRU, 6h TTL)
2. Query `Store` (LMDB/Bleve) for persisted kind 0 event
3. If stale (>7 days) or miss, fetch from network via replaceable event loader
4. Save to `MetadataCache` and `Store`

The `DiscoverUserRelays` goroutine runs async to discover and ensure user relays for the current session.
```

**Implementation**:
1. `GetProfile` queries `AllReadableRelays()` in parallel for kind 0
2. Returns immediately when first relay responds (profile metadata is replaceable)
3. Simultaneously launches `DiscoverUserRelays` as async goroutine to ensure discovered relays in the pool
4. User's NIP-65 relay list is NOT used for the current profile fetch — only for future event fetches

**Why this works**: Profile metadata (kind 0) is a replaceable event — all relays hold the same latest version. Querying all relays in parallel and taking the first response is both faster AND correct.

---

## Event Fetch After Profile (Targeted Query)

**When fetching a user's events (timeline, etc.)**: use their NIP-65 relay list for targeted querying.

```
GetUserTimeline(pubKey)
  │
  ├─ DiscoverUserRelays (sync) → get user's read relays from NIP-65
  │    └─ If nil/empty: fallback to AllReadableRelays()
  │
  └─ GetEvents(filter, userRelays) → query ONLY user's relays
       └─ If not found: fallback to AllReadableRelays()
```

**Why separate from profile fetch**: Profile fetches are frequent (every name display), while timeline fetches are less frequent and benefit from targeted relay list to reduce network traffic.

---

## Event Persistence and Query

Events are persisted via `sdk.System.Store` (LMDB/Bleve) and queried through the normal relay network. There is no local relay — the store is only for caching events we've already fetched from relays.

```go
func (a *AppContext) AllReadableRelays() []string {
    return a.ReadableRelays()
}

func (a *AppContext) AllWritableRelays() []string {
    return a.WritableRelays()
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

---

## Files

- `utils/filters.go` — pure filter builder functions (no side effects, unit-testable without mocks)
- `utils/get.go` — `GetEvent`, `GetProfile`, relay selection strategy
