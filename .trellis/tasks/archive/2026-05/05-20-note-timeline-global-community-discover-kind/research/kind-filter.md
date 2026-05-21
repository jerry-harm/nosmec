# Research: kind filtering in note timeline and community discover

- **Query**: How does "note timeline --global" filter by kind? How does "community discover" filter by kind? How does "kind:34550" filtering work? What recent changes could have broken it?
- **Scope**: internal (codebase analysis)
- **Date**: 2026-05-20

## Findings

### Architecture Overview

The codebase has two distinct filter paths:

1. **CLI TUI path**: `tui/timeline/model.go` → `nostr_sdk/system.go`
2. **CLI command path**: `cmd/note_commands.go` → `nostr_sdk/system.go`

### Timeline Kinds Usage

| Timeline Type | File:Line | Kinds Filtered |
|---|---|---|
| `global` (FetchGlobalTimelinePage) | `system.go:217-218` | `KindTextNote` (hardcoded) |
| `mine` (FetchMyTimelinePage) | `system.go:267-268` | `KindTextNote` (hardcoded) |
| `followed` (FetchFollowedTimelinePage) | `system.go:331` | `KindTextNote, KindComment` (hardcoded) |
| `community` (FetchFollowedTimelinePage) | `system.go:331` | `KindTextNote, KindComment` (hardcoded) |

### FetchGlobalTimelinePage (system.go:212-243)

```go
func (sys *System) FetchGlobalTimelinePage(ctx context.Context, limit int, until nostr.Timestamp) ([]nostr.Event, error) {
    filter := nostr.Filter{
        Kinds: []nostr.Kind{nostr.KindTextNote},  // HARDCODED - cannot filter by other kinds
        Limit: limit,
    }
    // ... uses Pool.FetchMany directly, NOT FetchEventsByFilter
```

**Key finding**: `FetchGlobalTimelinePage` does NOT use `FetchEventsByFilter`. It queries `Pool.FetchMany` directly with a hardcoded `KindTextNote` filter.

### FetchFollowedTimelinePage (system.go:320-450)

```go
kinds := []nostr.Kind{nostr.KindTextNote, nostr.KindComment}
filter := nostr.Filter{
    Tags:  nostr.TagMap{"a": []string{communityAddr}},
    Kinds: kinds,  // Also hardcoded to 1 and 1111
    Limit: limitPerKey,
}
```

### FetchEventsByFilter (system.go:490-551)

This function is used by `utils.FetchCommunityEvents` for kind:34550 queries:

```go
func (sys *System) FetchEventsByFilter(ctx context.Context, filter nostr.Filter, opts FetchEventsOptions) ([]nostr.Event, error) {
    // ...
    relays := opts.Relays
    if len(relays) == 0 {
        relays = sys.defaultRelaysForFilter(ctx, filter)  // Relay selection happens HERE
    }
    // ...
}
```

### defaultRelaysForFilter (system.go:553-586)

This function selects relays based on filter content:

```go
func (sys *System) defaultRelaysForFilter(ctx context.Context, filter nostr.Filter) []string {
    // IDs query → JustIDRelays + FallbackRelays
    if len(filter.IDs) > 0 { ... }

    // Single author query → FetchOutboxRelays + FallbackRelays
    if len(filter.Authors) == 1 { ... }

    // KindCommunityDefinition (34550) → RelayListRelays + FallbackRelays
    if slices.Contains(filter.Kinds, nostr.KindCommunityDefinition) && len(sys.RelayListRelays.URLs) > 0 {
        relays := append([]string{}, sys.RelayListRelays.URLs...)
        if len(sys.FallbackRelays.URLs) > 0 {
            relays = nostr.AppendUnique(relays, sys.FallbackRelays.URLs...)
        }
        return relays
    }

    // Default → FallbackRelays
    return sys.FallbackRelays.URLs
}
```

### Real-Time Subscription Filter (timeline/model.go:528-597)

The `startSubscription` function also hardcodes kinds:

```go
case "global":
    filter = nostr.Filter{
        Kinds: []nostr.Kind{nostr.KindTextNote},  // HARDCODED
        Since: since,
        Limit: 100,
    }
case "community":
    kinds := []nostr.Kind{nostr.KindTextNote, nostr.KindComment}  // HARDCODED
    filter = nostr.Filter{
        Kinds: kinds,
        Since: since,
        Limit: 100,
    }
```

### Community Discover Flow

1. `cmd/community_commands.go:324` → `discover.RunCommunityDiscover`
2. `tui/community/discover/model.go:180` → `utils.FetchCommunityEvents`
3. `utils/community.go:27` → `nostr.Filter{Kinds: []nostr.Kind{nostr.KindCommunityDefinition}}`
4. `nostr_sdk/system.go:490` → `FetchEventsByFilter`
5. `nostr_sdk/system.go:532` → `defaultRelaysForFilter` → **correctly routes to RelayListRelays**

### Relevant Spec Files

- `.trellis/spec/backend/nip-conventions.md` — NIP-72 community post convention
- `.trellis/spec/backend/forked-sdk-architecture.md` — Kind 34550 addressable loader design
- `.trellis/spec/backend/query-patterns.md` — Query patterns for kind:34550

### Recent Changes (Last 2 Weeks)

| Commit | Change | Potential Impact |
|---|---|---|
| `34c19d1` | Switch persistent backends from bbolt to LMDB | Could affect relay/KVStore behavior |
| `1bba05c` | Persist kvstore and add relay inspection | May have changed relay selection |
| `fa45ecd` | Remove sdkplus thin wrappers, inline simple parsers | Could affect filter handling |
| `0bf9151` | Extract filter builders for pure-unit-testability | Filter building refactored |
| `795f1f3` | Auto-publish subscriptions and DM relays after mutations | Could affect subscription handling |

### Root Cause Analysis

**The problem**: `FetchGlobalTimelinePage` hardcodes `KindTextNote` and does NOT use `FetchEventsByFilter`, so it bypasses `defaultRelaysForFilter` which has the logic to properly route kind:34550 queries.

**Why "community discover" likely still works**: It uses `FetchEventsByFilter` which properly routes to `RelayListRelays` for kind:34550 queries.

**Why "note timeline --global" cannot filter by kind:34550**: The CLI command `note timeline --global` has no `--kind` flag, and even if it did, `FetchGlobalTimelinePage` would ignore it.

**The "discover" command issue**: The `startSubscription` in timeline/model.go builds its own filter for real-time updates, and for "community" type it hardcodes `KindTextNote, KindComment`. However, the initial `fetchTimeline()` call uses `FetchFollowedTimelinePage` which queries by `a` tag (community address), so the initial load might work but real-time updates would miss events.

### Code Patterns

1. **Hardcoded kinds in system.go timeline functions**: `system.go:217-218`, `system.go:267-268`, `system.go:290`, `system.go:331`
2. **Hardcoded kinds in timeline/model.go subscription**: `model.go:533`, `model.go:543`, `model.go:551`, `model.go:578`
3. **FetchEventsByFilter properly routes kinds**: `system.go:574` — contains `KindCommunityDefinition` check

### External References

- [NIP-72](https://github.com/nostr-protocol/nips/blob/master/72.md) — Community Boards using kind:34550 and kind:1111
- `nostr.KindCommunityDefinition` = 34550
- `nostr.KindComment` = 1111 (NIP-72 community posts)
- `nostr.KindTextNote` = 1
