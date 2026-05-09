# Refactor all query functions to return chan *nostr.Event

## Goal

Replace all `FetchMany` channel loops with goroutine-wrapped functions that return `chan *nostr.Event`. Events flow through as they arrive (no dedup/sort/limit inside query functions). TUI handles dedup (via seenEventIDs), sorting, and list management.

## What I already know

7 `FetchMany` call sites across `utils/get.go` and `utils/community.go`:

**utils/get.go:**
- Line 33: `FetchManyReplaceable` (in GetEvent replaceable path)
- Line 197: GetMyTimeline
- Line 249: GetGlobalTimeline
- Line 345: GetFollowedTimeline

**utils/community.go:**
- Line 221: GetCommunityPosts
- Line 327: GetMyCreatedCommunities
- Line 356: GetPostedCommunities

All currently: `for relayEvent := range ch { ... }` collecting into map → sort → return `([]Event, error)`

## Requirements

* Replace each FetchMany loop with goroutine that yields events via `chan *nostr.Event`
* `CacheEvent` fires per event inside the goroutine (same as current)
* Error goes to log, not propagated (discarded)
* No dedup/sort/limit inside query functions — TUI handles dedup via `seenEventIDs` map
* Query functions return `chan *nostr.Event` (not iter.Seq)
* TUI callers updated to `for e := range ch { dispatch(newEventMsg{e}) }` pattern

## API Shape

```go
// All query functions follow this pattern:
func GetMyTimeline(ctx context.Context, limit int, until nostr.Timestamp, opts *GetOptions) chan *nostr.Event {
    ch := opts.App.Pool().FetchMany(ctx, relays, filter, opts)
    out := make(chan *nostr.Event)
    go func() {
        for relayEvent := range ch {
            CacheEvent(&relayEvent.Event, opts.App)
            out <- &relayEvent.Event
        }
        close(out)
    }()
    return out
}
```

## TUI Update Pattern

```go
ch := utils.GetMyTimeline(ctx, limit, until, opts)
go func() {
    for e := range ch {
        model.dispatch(newEventMsg{event: e})
    }
}()
```

## Acceptance Criteria

* [ ] All FetchMany replaced with goroutine-wrapped channel returns in utils/
* [ ] TUI callers updated to goroutine + dispatch pattern
* [ ] `go build ./...` passes
* [ ] Behavior: events arrive asynchronously, TUI deduplicates and sorts

## Definition of Done

* Lint / typecheck / CI green

## Out of Scope

* iter.Seq approach (rejected — channel is sufficient)
* Changes to cache layer (already done)

## Technical Notes

* No dedup/sort/limit inside query functions — TUI owns that
* goroutine per call (non-blocking caller)
* channel closed after FetchMany EOSE
* CacheEvent fires inside goroutine before yield