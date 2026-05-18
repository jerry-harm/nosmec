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
3. **AllReadableRelays()** — configured relays + local relay
4. **KnownRelays** — NIP-65 discovered + gossip fallback

```go
func GetQueryRelays(event *nostr.Event, app *config.AppContext) []string
```

Implemented in `utils/user_relays.go`.

---

## Event Middleware (Auto-learning)

`Pool.EventMiddleware` runs on every incoming event:

- **HintsDB**: learns relay→pubkey associations (scoring: HintEventFetched 700pts, HintInRelayList 350pts, HintFromTag 20pts)
- **TrackEventRelay**: records event→relay mapping for NIP-10 relay hints (first-write-wins, local relay skipped)

```go
// config/config.go EventMiddleware
opts.EventMiddleware = func(ie nostr.RelayEvent) {
    ev := ie.Event
    if ev.PubKey != [32]byte{} {
        h.Save(ev.PubKey, ie.Relay.URL, sdk_hints.MostRecentEventFetched, nostr.Now())
    }
    if ev.ID != [32]byte{} && ie.Relay.URL != localRelayURL {
        TrackEventRelay(ev.ID.Hex(), ie.Relay.URL)
    }
    // ...
}
```

---

## CacheEvent

`CacheEvent()` publishes events to local relay asynchronously (fire-and-forget):

```go
func CacheEvent(event *nostr.Event, app *config.AppContext) {
    if !shouldCache(event, app) { return }
    go func() {
        app.Pool().PublishMany(context.Background(), []string{localRelayURL}, *event)
    }()
}
```

Triggered by `shouldCache()` which checks `Config.CacheFilters`. Only publishes to local relay (`ws://localhost:PORT`).