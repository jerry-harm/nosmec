# Research: fiatjaf.com/nostr SDK Pool Implementation

- **Query**: How does fiatjaf.com/nostr SDK's Pool manage connections, does it support relay discovery, and how do QuerySingle/FetchMany work internally?
- **Scope**: External (SDK source code analysis via web)
- **Date**: 2026-05-11

## Findings

### SDK Source

The canonical SDK is now at `fiatjaf.com/nostr` (formerly `github.com/nbd-wtf/go-nostr`). The README notes the core codebase is the same but API breaks slightly.

### Pool Architecture (`SimplePool`)

File: `pool.go` from `nbd-wtf/go-nostr` (identical API to `fiatjaf.com/nostr`)

**Structure**:
```go
type SimplePool struct {
    Relays  *xsync.MapOf[string, *Relay]  // shared global map of active connections
    Context context.Context
    // ... middleware, penalty box, options
}
```

**Key behaviors**:
1. **Single shared connection map**: All relays managed in one `xsync.MapOf[string, *Relay]`. There is NO per-user or per-query isolation — a relay is connected once and reused across all subscriptions.
2. **`EnsureRelay(url)`**: Lazy-connect on first use. If already connected, returns cached `*Relay`. Uses a named lock + double-checked locking pattern.
3. **`PublishMany`**: Spawns goroutines per URL, each calling `EnsureRelay` then `Publish`.
4. **`SubscribeMany` / `FetchMany`**: Opens subscriptions across multiple relays concurrently. All share the same global relay connections.
5. **`QuerySingle`**: Wraps `SubManyEose` and cancels the context after the first event arrives. All other goroutines continue until they hit EOSE or timeout.
6. **`FetchManyReplaceable`**: Deduplicates replaceable/addressable events by `(PubKey, d-tag)` — useful for getting latest versions.

**No concept of "temporary pools" or "per-user relay views"** — the pool is global and flat. Every query shares the same connection pool.

### Relay Lifecycle

1. `EnsureRelay(url)` — called before any operation on a relay
2. If not connected: creates `NewRelay()`, calls `Connect(ctx)` with 15s timeout
3. Relay stored in `pool.Relays` map — persists until `pool.Close()`
4. Penalty box: failed relays are temporarily ignored (exponential backoff)

### NIP-65 Support

**Package**: `fiatjaf.com/nostr/nip65` (also exists in `nbd-wtf/go-nostr/nip65`)

```go
func ParseRelayList(event nostr.Event) (readRelays []string, writeRelays []string)
```

This parses a Kind 10002 event and returns separate read/write relay lists based on `r` tags with `marker` values.

**No built-in relay discovery**: The SDK does NOT fetch NIP-65 events automatically. Client code must:
1. Query for `KindRelayListMetadata` events for a given pubkey
2. Call `nip65.ParseRelayList()` on the result
3. Use the returned relays for user-specific queries

### Connection Per Query?

**No** — `SimplePool` maintains a global pool. `QuerySingle` and `FetchMany` do NOT create new connections per query. They:
- Call `EnsureRelay` which reuses existing connections if already connected
- Creates new connections only if the relay isn't in the map yet
- All concurrent queries share the same underlying `*Relay` connection

### Is There a "View" or "Perspective" Based on a User's Relay List?

**No**. The SDK has no concept of:
- Per-user temporary relay pools
- "Perspective" objects that encapsulate a user's relay list
- Automatic switching to a user's relays for their events

This is entirely the client's responsibility. The user would need to:
1. Maintain a mapping of `pubkey -> []relays`
2. Create a `SimplePool` per user (or per session), passing only their relays
3. Use that pool for queries about that user

## Code Patterns Observed in Local Codebase

From `utils/get.go`, `utils/profile.go`, `utils/relay_list.go`:

```go
// Fetch user's NIP-65 relay list
filter := nostr.Filter{
    Authors: []string{pubKey},
    Kinds:   []nostr.Kind{nostr.KindRelayListMetadata},
}
result := app.Pool().QuerySingle(ctx, knownRelays, filter, nostr.SubscriptionOptions{})
readRelays, writeRelays := nip65.ParseRelayList(result.Event)

// Then use those relays for further queries
profile := FetchProfile(ctx, app, readRelays, pubKey)
```

This pattern is already in use locally — the local codebase handles relay discovery manually.

## Related Specs

- `.trellis/spec/backend/relay-guidelines.md` — NIP-65 relay list metadata guidelines
- `.trellis/tasks/05-11-nip65-relay-discovery/prd.md` — current task requirements

## Caveats

- The SDK's `SimplePool` is global/flat — no built-in per-user relay views
- Temporary per-user pools would need to be created externally (new `SimplePool` instances)
- The `QuerySingle` cancellation only cancels the context for the first-event handler; other goroutines continue until EOSE
- `fiatjaf.com/nostr` and `nbd-wtf/go-nostr` are nearly identical but with API differences — the local codebase uses `fiatjaf.com/nostr`