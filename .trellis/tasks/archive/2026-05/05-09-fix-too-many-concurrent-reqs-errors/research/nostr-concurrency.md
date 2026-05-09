# Research: nostr SDK Concurrency Control

- **Query**: How to limit concurrent FetchMany/SubscribeMany requests to prevent "too many concurrent REQs" errors
- **Scope**: internal (go-nostr SDK) + external (khatru relay)
- **Date**: 2026-05-09

## Findings

### 1. nostr.Pool Has No Built-in Concurrency Limit

**File**: `pool.go` lines 53-74, 239-292, 431-619

`nostr.Pool` spawns one goroutine per relay per call with no semaphore or queue:

```go
// subMany() at line 488 - one goroutine per URL, no limit
go func(nm string) {
    relay, err := pool.EnsureRelay(nm)
    sub, err := relay.Subscribe(ctx, filter, opts)
    // ...
}(NormalizeURL(url))
```

`PoolOptions` (lines 76-99) has NO concurrency field:
- `AuthRequiredHandler`
- `PenaltyBox`
- `EventMiddleware`
- `DuplicateMiddleware`
- `AuthorKindQueryMiddleware`
- `RelayOptions`

**Conclusion**: No semaphore, no channel, no queue config in the SDK itself.

### 2. Write Queue on Relay Has No Concurrency Limit

**File**: `relay.go` lines 57-58, 295-310

The `writeQueue` on each Relay is a `chan writeRequest` (unbuffered), but the writer loop processes it with no per-subscription backpressure — writes just block when the connection is slow. The queue is not bounded and does not apply any concurrency limit on REQs.

```go
// relay.go line 58
writeQueue chan writeRequest  // unbuffered, not size-limited

// line 295-310: writer loop
case wr := <-r.writeQueue:
    debugLogf("{%s} sending '%v'\n", r.URL, string(wr.msg))
    writeCtx, cancel := context.WithTimeoutCause(connCtx, time.Second*10, ...)
    err := c.Write(writeCtx, ws.MessageText, wr.msg)
    // blocks until write completes, but no per-REQ limit
```

### 3. khatru Relay — No Per-Client Subscription Limit

**File**: `khatru/relay.go`, `khatru/listener.go`

khatru's `Relay` struct:
```go
// relay.go lines 42-43
clients   map[*WebSocket][]listenerSpec  // no hard limit
listeners []listener                     // Grows dynamically
```

`handleRequest` (responding.go line 12) does NOT enforce a subscription count cap. It calls `OnRequest` hook if set, but the default `FilterIPRateLimiter` (policies/ratelimits.go line 48) only rate-limits by IP, not by subscription count.

**`FilterIPRateLimiter`** limits how many REQs per time window, not the total active subscriptions.

### 4. How "Too Many Concurrent REQs" Errors Appear

The error is a **NOTICE message** from the relay server (NIP-01). The `Pool` handles NOTICEs at `relay.go` lines 377-383:

```go
case *NoticeEnvelope:
    if r.noticeHandler != nil {
        r.noticeHandler(r, string(*env))
    } else {
        log.Printf("NOTICE from %s: '%s'\n", r.URL, string(*env))  // default: print to stderr
    }
```

The default `NoticeHandler` logs to stderr — there is no built-in suppression.

### 5. How to Suppress NOTICE Logs

**File**: `relay.go` lines 148-162, `config/config.go` line 175-177

Set a custom `NoticeHandler` via `RelayOptions`:

```go
// config/config.go line 175-177
return nostr.NewPool(nostr.PoolOptions{
    RelayOptions: nostr.RelayOptions{
        NoticeHandler: func(relay *nostr.Relay, notice string) {
            // suppress or log at Debug level
        },
    },
})
```

### 6. Available Rate-Limiting Policies in khatru

**File**: `khatru/policies/ratelimits.go` lines 12-58

```go
// Limits REQs per IP across all subscriptions (token bucket)
FilterIPRateLimiter(tokensPerInterval, interval, maxTokens)

// Same by PubKey
EventPubKeyRateLimiter(...)

// Per connection count
ConnectionRateLimiter(...)

// Per event publish
EventIPRateLimiter(...)
```

These are applied as `OnRequest` / `RejectConnection` hooks on the Relay, not on the Pool.

### 7. Client-Side Solutions (nosmec can implement)

#### 7a. Semaphore-Wrapped Pool (application-level)

Wrap calls in a bounded semaphore before calling `Pool.FetchMany` / `Pool.SubscribeMany`:

```go
type ThrottledPool struct {
    pool   *nostr.Pool
    sem    chan struct{}
}

func (tp *ThrottledPool) FetchMany(ctx context.Context, urls []string, filter nostr.Filter, opts nostr.SubscriptionOptions) chan nostr.RelayEvent {
    // Acquire before firing
    select {
    case tp.sem <- struct{}{}:
    case <-ctx.Done():
        // context canceled, don't proceed
    }
    events := tp.pool.FetchMany(ctx, urls, filter, opts)
    // Release when events channel closes (handled in a wrapper goroutine)
    go func() {
        <-events
        <-tp.sem
    }()
    return events
}
```

#### 7b. BatchedQueryMany for Reducing REQ Count

**File**: `pool.go` lines 782-868

`Pool.BatchedQueryMany` sends multiple filters to different relays with smart deduplication — reduces the total number of goroutines vs calling `FetchMany` separately:

```go
// Instead of multiple FetchMany calls:
ch := pool.FetchMany(ctx, []string{relay1}, filter1, opts)
ch2 := pool.FetchMany(ctx, []string{relay1}, filter2, opts)

// Use BatchedQueryMany to consolidate:
ch := pool.BatchedQueryMany(ctx, []nostr.DirectedFilter{
    {Relay: relay1, Filter: filter1},
    {Relay: relay1, Filter: filter2},
}, opts)
```

This doesn't add a hard limit — it's a structural improvement that reduces goroutine count.

#### 7c. Subscription Cancellation (already in codebase)

`SubscriptionOptions` has `MaxWaitForEOSE` (subscription.go line 75) to auto-close after 7s. Ensuring all subscriptions are properly cancelled via context is the primary mitigation — each goroutine watches `ctx.Done()`.

### 8. Local Relay: Add Request Limit Hook

Since nosmec starts its own local relay (config.go lines 290-304), add a filter limit:

```go
relay := khatru.NewRelay()
relay.UseEventstore(store, 500)

// Add per-IP REQ rate limit
relay.OnRequest = policies.FilterIPRateLimiter(100, time.Minute, 200)
```

This limits how many REQs the local relay accepts per IP.

## Caveats / Not Found

- `PoolOptions` in this SDK version (`fiatjaf.com/nostr v0.0.0-20260310`) has **no** `MaxConcurrent` or `RequestSemaphore` field — no built-in mechanism
- khatru relay does NOT track per-client subscription count, only rate-per-time-window
- The "too many concurrent REQs" NOTICE likely comes from **external relays** (not local khatru), which impose their own limits — nosmec cannot change server-side behavior
- `debugLogf` uses a global `debugLog` variable (utils.go) — debug logging cannot be suppressed per-relay