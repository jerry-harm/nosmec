# Research: nostr Relay Hint Usage for Event Fetching

- **Query**: How does the nostr SDK handle relay hints? What's the NIP-65 outbox model?
- **Scope**: mixed (internal codebase + external SDK + NIP specs)
- **Date**: 2026-05-17

## Part A: nostr SDK Built-In Functions

### 1. Core nostr Package (`fiatjaf.com/nostr`)

The core `nostr` package does **not** have built-in relay hint extraction. It provides:
- `Pool` — manages connections to multiple relays (EnsureRelay, QuerySingle, FetchMany, PublishMany, SubscribeMany, etc.)
- `Event` / `Tag` / `Tags` types — the raw data structures
- No functions named `ExtractRelayHints`, `FetchRelayList`, or `TrackEventHints` in the core package

### 2. SDK Package (`fiatjaf.com/nostr/sdk`)

The `sdk.System` type provides all high-level relay management. Key functions:

#### FetchRelayList
```go
func (sys *System) FetchRelayList(ctx context.Context, pubkey nostr.PubKey) GenericList[string, Relay]
```
- Fetches and parses the user's kind 10002 (NIP-65) relay list
- Returns `GenericList[string, Relay]` where each `Relay` has:
  - `URL string` — the relay URL
  - `Inbox bool` — true if marked "read" or no marker
  - `Outbox bool` — true if marked "write" or no marker
- Uses a cache (`RelayListCache`, LRU-based, defaults to 1000 entries)
- Source: `lists_relay.go:22-25`, parsing function `parseRelayFromKind10002` at lines 60-88

#### FetchOutboxRelays (NIP-65 outbox model)
```go
func (sys *System) FetchOutboxRelays(ctx context.Context, pubkey nostr.PubKey, n int) []string
```
- **This is the main function for finding where to read a user's notes from**
- Uses multiple data sources in priority order:
  1. 2-minute short-term cache (indexed by pubkey[7])
  2. Fetches kind 10002 to feed hints (triggers `fetchGenericList` → updates HintsDB with LastInRelayList)
  3. `sys.Hints.TopN(pubkey, 6)` — top 6 scored relays from HintsDB
  4. Falls back to `["wss://relay.damus.io", "wss://nos.lol"]` if no hints at all
- Returns at most `n` relays
- Source: `outbox.go:23-51`

#### FetchInboxRelays (NIP-65 read relays)
```go
func (sys *System) FetchInboxRelays(ctx context.Context, pubkey nostr.PubKey, n int) []string
```
- Reads from kind 10002 relay list only
- Returns relays marked as "read" (inbox). If list is empty or >10, falls back to defaults
- **Use case**: fetching events where this user is mentioned (notified)
- Comment in code: "just reads relays from a kind:10002, that's the only canonical place where a user reveals the relays they intend to receive notifications from"
- Source: `outbox.go:53-69`

#### FetchWriteRelays
```go
func (sys *System) FetchWriteRelays(ctx context.Context, pubkey nostr.PubKey) []string
```
- Reads from kind 10002, returns only outbox (write) relays
- **Use case**: deciding where to publish on behalf of a user
- Source: `outbox.go:71-87`

#### TrackEventHints
```go
func (sys *System) TrackEventHints(ie nostr.RelayEvent)
```
- Meant to be used as event middleware (WithEventMiddleware)
- Skips virtual relays and ephemeral events (kind >= 20000 && < 30000)
- For kind 10002 events: saves LastInRelayList hint for each relay with "write" marker
- For kind 3 events: saves LastInHint for each relay in p-tag[2]
- For other events: saves MostRecentEventFetched for the author on this relay, saves LastInHint from p-tag[2] relay URLs, and parses nip27 content references for additional hints
- Source: `tracker.go:44-53`, internal `trackEventHints` at lines 55-148

#### TrackEventHintsAndRelays
```go
func (sys *System) TrackEventHintsAndRelays(ie nostr.RelayEvent)
```
- Same as TrackEventHints but also records event→relay associations in KVStore via `trackEventRelay`
- Source: `tracker.go:27-40`

#### TrackEventRelaysD
```go
func (sys *System) TrackEventRelaysD(relay string, id nostr.ID)
```
- Partner function for WithDuplicateMiddleware
- Only records relay for an event if the event already has existing relay entries
- Source: `tracker.go:150-156`

#### GetEventRelays
```go
func (sys *System) GetEventRelays(id nostr.ID) []string
```
- Returns all known relay URLs where an event has been seen
- Stored in KVStore with key format: `'r' + first 8 bytes of event ID`
- Source: `event_relays.go:107-118`

#### FetchSpecificEvent — The Key Function
```go
func (sys *System) FetchSpecificEvent(ctx context.Context, pointer nostr.Pointer, params FetchSpecificEventParameters) (*nostr.Event, []string, error)
```
- **The main event-fetching entry point in the SDK.** Implements the relay priority strategy:
  1. Check local event store first
  2. Use relay hints from the Pointer (nevent/naddr relay URLs) as highest priority
  3. Call `FetchOutboxRelays(ctx, author, 3)` for the author (if pubkey is known from tag[3])
  4. Append author's outbox relays after hint relays
  5. Query in two passes: first with hint+outbox relays, then with fallback relays
  6. Sorts successful relay URLs with priority relays first
- **This is why the `pubkey` field at tag[4] in e tags matters**: without it, `author` is ZeroPK and FetchOutboxRelays is skipped entirely
- Source: `specific_event.go:62-190`

#### PrepareNoteEvent
```go
func (sys *System) PrepareNoteEvent(ctx context.Context, evt *nostr.Event) (targetRelays []string)
```
- Prepares a note for publishing (fixes URLs, nostr references, adds tags)
- Returns target relays: `FetchInboxRelays(ctx, pk, 4)` for every tagged pubkey (so the mentioned user gets notified)
- Source: `note.go:23-114`

### 3. NIP-65 Utilities (`fiatjaf.com/nostr/nip65`)

```go
func ParseRelayList(event nostr.Event) (readRelays []string, writeRelays []string)
```
- Parses kind 10002 event, extracts URLs from "r" tags
- Tag without marker → appears in both readRelays and writeRelays
- Tag with "write" marker → writeRelays only
- Tag with "read" marker → readRelays only
- **Used in the project**: `utils/user_relays.go:70` in `DiscoverUserRelays()`

### 4. HintsDB Scoring System (`fiatjaf.com/nostr/sdk/hints`)

HintsDB is the core of the outbox model — it learns which relays work for each pubkey.

#### Interface
```go
type HintsDB interface {
    TopN(pubkey nostr.PubKey, n int) []string
    Save(pubkey nostr.PubKey, relay string, key HintKey, score nostr.Timestamp)
    PrintScores()
    GetDetailedScores(pubkey nostr.PubKey, n int) []RelayScores
}
```

#### Hint Key Types and Base Points
| Hint Key | Base Points | Meaning |
|---|---|---|
| `LastFetchAttempt` (0) | **-500** | Tried to query this relay for this author (may fail) |
| `MostRecentEventFetched` (1) | **700** | Successfully got an event from this author on this relay |
| `LastInRelayList` (2) | **350** | Author listed this relay in their kind 10002 (strong signal) |
| `LastInHint` (3) | **20** | Hint from tags, nprofile, nevent, or NIP-05 |

#### Scoring Algorithm (TopN)
From `hints/bbolth/db.go:185-196`:
```
For each hint key (0-3):
  if timestamp != 0:
    value = basePoints * 10^10 / max(now+24h - timestamp, 1)^1.3
    sum += value
```
- Score decays over time with power 1.3 — strongly favors recent hints
- `now` is offset by +24h so current hints aren't infinite
- Relay list presence (350) is weighted above hints (20) but below successful fetches (700)

#### IsVirtualRelay
```go
func IsVirtualRelay(url string) bool
```
- Excludes `wss://feeds.nostr.band`, `wss://filter.nostr.wine`, `wss://cache*` from hint tracking
- Also excludes localhost in non-test environments
- Source: `sdk/utils.go:13-32`

---

## Part B: NIP-65 Outbox Model

### NIP-65: Relay List Metadata (kind 10002)

**Full text source**: https://raw.githubusercontent.com/nostr-protocol/nips/master/65.md

#### Key Definitions

1. **Event kind 10002** is a replaceable event advertising the user's preferred relays
2. Tags use `["r", "<relay URL>", "<marker>"]` format:
   - No marker → both read and write
   - `"write"` marker → author publishes to this relay
   - `"read"` marker → author reads mentions from this relay

#### Outbox Model

| Operation | Which Relays to Use |
|---|---|
| **Download events FROM a user** | The user's **write** relays |
| **Download events ABOUT a user** (where the user was tagged) | The user's **read** relays |
| **Publish an event** | Author's **write** relays + ALL tagged users' **read** relays |
| **Publish kind 10002 itself** | Broadcast to ALL relays the event was published to |

#### Recommendations
- Keep lists small: 2-4 relays per category
- Spread kind 10002 events to as many relays as possible

### NIP-01: Relay Hints in Tags

**Full text source**: https://raw.githubusercontent.com/nostr-protocol/nips/master/01.md

#### Tag Formats with Relay Hints

| Tag | Format | Relay Hint Position |
|---|---|---|
| `e` | `["e", <event-id>, <relay URL, optional>, <author pubkey, optional>]` | tag[2] |
| `p` | `["p", <pubkey>, <relay URL, optional>]` | tag[2] |
| `a` (addressable) | `["a", "<kind>:<pubkey>:<d-tag>", <relay URL, optional>]` | tag[2] |
| `a` (replaceable) | `["a", "<kind>:<pubkey>:", <relay URL, optional>]` | tag[2] |

#### Critical Detail: e tag[3] (author pubkey)
The e tag's fourth element (`tag[3]`) is the **author's pubkey**. This is essential for the outbox model:
- Without it, `FetchSpecificEvent` cannot call `FetchOutboxRelays` for the author
- With it, the SDK can discover additional relays where the author publishes
- This is how `FetchSpecificEvent` at `specific_event.go:105-109` switches from hint-only to hint+outbox relay strategy

### Recommended Relay Priority (as implemented by SDK)

The SDK's `FetchSpecificEvent` implements this priority:

1. **Relay hints from tags** (tag[2] in e/p/a tags) — highest priority
2. **Author's outbox relays** (from FetchOutboxRelays, which combines HintsDB scoring + kind 10002)
3. **Fallback relays** (JustIDRelays, general-purpose fallback)

The HintsDB itself scores relays by combining:
- Successful past fetches from this relay for this author (700 points, decayed by recency)
- Presence in author's kind 10002 relay list (350 points)
- Hints from tags/nprofile/NIP-05 (20 points)
- Minus penalty for fetch attempts (-500 points)

---

## Current Project Usage

### ExtractRelayHints (custom implementation)
- **File**: `utils/get.go:18-39`
- Extracts relay URLs from tag[2] of e/p/a/q tags
- Simple deduplication using a map
- Does NOT use any SDK hint tracking or HintsDB

### Where ExtractRelayHints Is Called
1. **`tui/window/event/thread_treeview.go:241`** — `fetchRootEvent()` uses relay hints from current event's e tags as the first query target, falls back to AllReadableRelays() if hints fail
2. **`utils/community.go:199`** — `GetCommunity()` extracts relay hints from community definition events to add them to the pool

### DiscoverUserRelays (custom implementation)
- **File**: `utils/user_relays.go:41-78`
- Queries for kind 10002 event from both local cache and remote relays
- Uses `nip65.ParseRelayList(*event)` to parse read/write relays
- Adds discovered relays via EnsureRelays and TrackRelays
- Does NOT use SDK FetchRelayList or HintsDB

### Key Gap: No SDK Relay Strategy
The project currently does NOT use:
- `sdk.System` (no System instance exists)
- `sdk.System.FetchOutboxRelays` / `FetchSpecificEvent` (the SDK's outbox model)
- `sdk.System.TrackEventHints` (automatic hint learning)
- `sdk.System.GetEventRelays` (event→relay tracking)
- `sdk.HintsDB` (relay scoring)

The project uses raw `nostr.Pool` for all queries and has its own parallel relay infrastructure (`config.RelayList`, `KnownRelays`, etc.).

---

## Caveats / Not Found

- The `sdk.System.FetchOutboxRelays` uses a 2-minute short-term cache (global, not per-System), which means its results may be stale for recently-discovered pubkeys but fresh for frequent lookups
- The HintsDB scoring formula (`basePoints * 10^10 / max(now+24h - ts, 1)^1.3`) means scores decay to near-zero after a few days, so inactive pubkeys lose their relay associations
- The SDK's `FetchOutboxRelays` falls back to hardcoded defaults (`wss://relay.damus.io`, `wss://nos.lol`) if HintsDB has no data — this is a hardcoded fallback that may not match user-configured relays
- NIP-01 does not define a standard relay hint extraction function — the `ExtractRelayHints` pattern is a convention, not part of the spec
- The `tag[3]` field (author pubkey) in e tags is optional per NIP-01 and many clients don't include it, which breaks the outbox lookup in FetchSpecificEvent
