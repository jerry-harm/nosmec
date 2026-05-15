# Refactor utils: separate filter building from network requests

## Goal

Refactor utils request functions to separate pure filter-building logic from network request orchestration, enabling simple unit testing without mocks.

## What I already know

Current `get.go` pattern:
```
GetEvent(ctx, filter, opts)
  ├─ Determine relays (opts.Relays / AllReadableRelays)
  ├─ Local relay first (2s timeout)
  ├─ Replaceable kind check (FetchManyReplaceable)
  ├─ Network query (QuerySingle / FetchMany)
  └─ Cache result

GetNote(noteID) → BuildNoteFilter(noteID) → GetEvent()
GetPost(postID) → BuildPostFilter(postID) → GetEvent()
GetProfile(pubKey) → BuildProfileFilter(pubKey) → GetEvent()
```

nostr/sdk.System pattern:
```
FetchProfileMetadata(pubKey)
  ├─ MetadataCache.TryGet(pubKey)  ← cache first
  ├─ Store.TryGet(pubKey)           ← local store
  ├─ MetadataRelays.Stream()         ← parallel fetch from relays
  └─ Cache result
```

## Problem

Filter building is mixed with request orchestration. Can't test filter logic without mocking Pool.

## Solution

### Layer 1: Pure Filter Builders (testable without mocks)

```go
// Filter builders - pure functions, no side effects
func BuildNoteFilter(noteID string) (nostr.Filter, error)
func BuildPostFilter(postID string) (nostr.Filter, error)
func BuildProfileFilter(pubKey nostr.PubKey) nostr.Filter
func BuildFollowListFilter(pubKey nostr.PubKey) nostr.Filter
func BuildRelayListFilter(pubKey nostr.PubKey) nostr.Filter
```

### Layer 2: Refactored Get Functions

```go
// Get functions compose builders + nostr SDK directly
func GetNote(ctx context.Context, noteID string, opts *GetOptions) *nostr.Event {
    filter, err := BuildNoteFilter(noteID)  // pure, testable
    if err != nil {
        return nil
    }
    // All relays including local, no special handling
    relays := opts.Relays
    if len(relays) == 0 {
        relays = opts.App.AllReadableRelays()
    }
    result := opts.App.Pool().QuerySingle(ctx, relays, filter, nostr.SubscriptionOptions{})
    // ...
}
```

### Key Decisions

1. **No custom request functions** — use nostr SDK directly
2. **No special local relay treatment** — all relays together, SDK handles it
3. **Filter builders are pure functions** — directly testable
4. **Only extract what can be tested in isolation**

## Implementation Plan

1. Extract filter builders from existing functions (GetNote, GetPost, etc.)
2. Rewrite existing functions to call builders + nostr SDK directly
3. Add table-driven tests for filter builders (pure functions)

## Out of Scope

- Custom request wrapper functions
- Special local relay handling
- Cache strategy changes

## Definition of Done

- [ ] All filter builders extracted and tested (table-driven)
- [ ] Existing functions refactored to use builders
- [ ] No custom request wrapper functions added
- [ ] `go build ./... && go test ./...` passes