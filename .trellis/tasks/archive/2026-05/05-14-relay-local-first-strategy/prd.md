# relay-local-first-strategy

## Goal

Fix relay query strategy to prioritize local relay (cache) for faster responses and better UX. For replaceable events: return local result immediately, then async update from remote. For DM send: handle case when recipient has no published relay list.

## What I Already Know

* `AllReadableRelays()` prepends local relay to configured relays — local is already first in list
* Current `GetEvent` queries ALL relays in parallel via `FetchManyReplaceable` (for replaceable) or `QuerySingle` (for non-replaceable)
* `GetProfile` was recently changed to use `QuerySingle` in parallel
* `FetchRecipientDMRelays` / `FetchRecipientReadRelays` return empty if recipient never published NIP-65/NIP-17
* DM send fails with "recipient has no public relay list" when both return empty

## Problem Statement

1. **Replaceable events (kind 0, 10002, 10050)**: Currently queries all relays in parallel, waits for ALL to respond. Slow when remote relays are slow.
2. **Non-replaceable events**: Queries all relays in parallel, waits for ALL. Same problem.
3. **DM send without recipient relay list**: Fails entirely instead of falling back to sending to user's own relays or common public relays.

## Strategy: Local-First Query

### Replaceable Events

```
GetEvent(filter, replaceable=true)
  │
  ├─ Query LOCAL RELAY ONLY first
  │    └─ If found: return immediately
  │
  └─ Async: query ALL REMOTE RELAYS in parallel
       └─ If newer event found: update local cache
       └─ For UI updates: relying on next poll / refresh cycle
```

**Why**: Local relay has cached events. For replaceable events (same content everywhere), local = remote. No need to wait for network.

### Non-Replaceable Events

```
GetEvent(filter, replaceable=false)
  │
  ├─ Query LOCAL RELAY FIRST
  │    └─ If found: return immediately
  │
  └─ If not found: query REMOTE RELAYS (all, parallel)
       └─ Return first result
```

**Why**: Local might not have the event (e.g., a specific note ID). But if local has it, no network needed.

### DM Send Without Recipient Relay List

Current code:
```go
if len(theirRelays) == 0 {
    return fmt.Errorf("recipient has no public relay list...")
}
```

**New behavior**: Fall back to sending on user's own relays. The recipient can potentially receive from any relay their clients connect to.

```
DM Send
  │
  ├─ Try FetchRecipientDMRelays → fail
  ├─ Try FetchRecipientReadRelays → fail
  │
  └─ Fallback: send on our own relays (ListDMRelays or ReadableRelays)
       └─ nip17.PublishMessage handles this
```

## Implementation

### Changes to `utils/get.go`

1. **New `GetEventLocalFirst`** or modify existing `GetEvent`:
   - Split query into local-first and remote-fallback phases
   - For replaceable: return local immediately, async update from remote
   - For non-replaceable: return local if found, else query remote

2. **Helper to get local relay URL only**:
   - Extract `localRelayURL` from `AllReadableRelays()`

### Changes to `utils/dm.go`

1. **`SendDM`**: Remove the fatal error when `theirRelays == 0`. Instead, fall back to user's own relays for sending.

## Acceptance Criteria

* [ ] Replaceable event queries local relay first, returns immediately if cached
* [ ] Async remote query updates local cache for replaceable events
* [ ] Non-replaceable events query local first, fallback to remote
* [ ] DM send works even when recipient has no published relay list (falls back to our relays)
* [ ] `go build ./...` succeeds
* [ ] `go vet ./...` passes

## Definition of Done

* Relay query strategy implemented
* DM send fallback implemented
* Build and vet pass

## Out of Scope

* Changes to subscription/continuous query patterns
* Timeline-specific optimizations
* Changes to relay connectivity verification

## Technical Notes

* `config.GetLocalRelayURL()` returns local relay URL or empty string
* `AllReadableRelays()` returns `[]string{localURL} + configuredRelays`
* Need to split into local-only vs all-relays queries
* `isReplaceableKind()` checks if kind is replaceable or addressable