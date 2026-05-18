# Query Patterns

> Standard patterns for querying events from relays.

---

## Synchronous Query

```go
event := GetEvent(ctx, filter, opts)
```

- Uses `Pool().QuerySingle()` with timeout context
- Returns first result or nil
- Timeout: `opts.App.QueryTimeout()` (default 5s, configurable)

```go
func GetEvent(ctx context.Context, filter nostr.Filter, opts *GetOptions) *nostr.Event {
    ctx, cancel := context.WithTimeout(ctx, opts.App.QueryTimeout())
    defer cancel()
    return opts.App.Pool().QuerySingle(ctx, relays, filter, nostr.SubscriptionOptions{})
}
```

---

## Async Query

```go
event := GetEventAsync(ctx, filter, opts)
```

- Uses `context.WithTimeout`
- Returns `*nostr.Event` (blocks until result or context cancel)

---

## Streaming Query

```go
ch := GetTimeline(ctx, limit, until, opts)
for event := range ch {
    // process event
}
```

- Returns `chan *nostr.Event`
- Events yielded as they arrive (no buffering)
- Channel closed when subscription ends

---

## Replaceable Events

```go
results := opts.App.Pool().FetchManyReplaceable(ctx, relays, filter, nostr.SubscriptionOptions{})
results.Range(func(key nostr.ReplaceableKey, ev nostr.Event) bool {
    // process
})
```

- Used for Kind 0 (metadata), Kind 10002 (relay list), Kind 10050 (DM relay list)
- Context must have timeout

---

## Timeout Rule

**Never use hardcoded timeouts.** Always use `opts.App.QueryTimeout()`:

```go
// WRONG
ctx, cancel := context.WithTimeout(ctx, 10*time.Second)

// CORRECT
ctx, cancel := context.WithTimeout(ctx, opts.App.QueryTimeout())
```

`QueryTimeout()` reads from config, defaults to 5 seconds if not set.

---

## GetQueryRelays Priority

When `GetOptions` has no explicit relay list, `GetQueryRelays` determines the relay set:

1. **tag[2] relay hints** — `ExtractRelayHints(event)` from e/p/a/q tags
2. **HintsDB outbox** — `app.Hints().TopN(pubkey, 3)` from e tag author pubkeys (checks tag[4] first for NIP-10 5-field, falls back to tag[3])
3. **AllReadableRelays()** — configured relays only
4. **KnownRelays** — NIP-65 discovered + gossip fallback

```go
func GetQueryRelays(event *nostr.Event, app *config.AppContext) []string
```

Implemented in `utils/user_relays.go`.

---

## Event Middleware (Auto-learning)

`Pool.EventMiddleware` runs on every incoming event:

- **HintsDB**: learns relay→pubkey associations (scoring: HintEventFetched 700pts, HintInRelayList 350pts, HintFromTag 20pts)
- **TrackEventRelay**: records event→relay mapping for NIP-10 relay hints (first-write-wins via KVStore)

---

## Profile Fetching

Profile metadata (Kind 0) is fetched via `sdk.System.FetchProfileMetadata(ctx, pubkey)` which handles:

1. In-memory cache check (MetadataCache LRU, 8000 entries, 6h TTL)
2. BoltDB/Bleve store query (persisted events)
3. Network fetch with 7-day debounce (replaces if newer)
4. Cache + store update