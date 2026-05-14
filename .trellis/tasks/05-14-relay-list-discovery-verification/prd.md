# relay-list-discovery-verification

## Goal

Add connectivity verification when discovering relays - verify each relay can connect before using it for critical operations.

## What I Already Know

* `Pool.EnsureRelay(url)` returns `*Relay` (lazy connection, doesn't actually connect)
* `Relay.IsConnected()` checks connection state
* `Relay.Connect(ctx)` actually establishes connection
* Problem: `EnsureRelays` just adds relays to pool without verifying they're reachable

## Requirements

* Add `VerifyRelayConnectivity(ctx context.Context, url string) (bool, error)` to check if relay is reachable
* Add `VerifyRelaysConnectivity(ctx context.Context, urls []string) ([]string, error)` that returns only reachable relays
* Use connectivity check before adding relays to pool for critical operations
* Non-blocking: don't fail if some relays are unreachable

## Acceptance Criteria

* [ ] `VerifyRelayConnectivity` returns true if relay connects, false otherwise
* [ ] `VerifyRelaysConnectivity` filters out unreachable relays
* [ ] `go build ./...` succeeds
* [ ] `go vet ./...` passes

## Definition of Done

* Connectivity verification functions exist
* Build passes

## Technical Approach

```go
func VerifyRelayConnectivity(ctx context.Context, url string) (bool, error) {
    relay := nostr.NewRelay(ctx, url, nostr.RelayOptions{})
    if err := relay.Connect(ctx); err != nil {
        return false, nil  // not an error, just unreachable
    }
    return relay.IsConnected(), nil
}
```