# Fix "too many concurrent REQs" errors

## Goal

Suppress spurious NOTICE-level error logs and add TUI refresh rate limiting.

## What I already know

* "ERROR: too many concurrent REQs" is a NOTICE from external relays — not a client-side detection
* nostr.Pool has no built-in concurrency limit
* Default NoticeHandler logs to stderr via `log.Printf`
* "Too many concurrent REQs" is a relay-side limit — client cannot change server behavior
* The only thing client can do: manage subscription lifecycle + add rate limits

## Requirements

### 1. Suppress NOTICE logs

Set custom `NoticeHandler` in `NewPool()` — logs at DEBUG level instead of printing to stderr.

### 2. TUI refresh rate limiting

Add cooldown period (2 seconds) to prevent rapid-fire `fetchTimeline` calls when user presses 'r' repeatedly.

## Implementation

* `config/config.go`: `NoticeHandler` in `PoolOptions.RelayOptions` → logs at `logger.Debug`
* `tui/timeline/model.go`: `lastRefresh` field + 2-second cooldown check at start of `fetchTimeline()`

## Acceptance Criteria

* [x] NOTICE log "too many concurrent REQs" no longer printed to stderr (now DEBUG)
* [x] TUI `fetchTimeline` rate-limited to once per 2 seconds
* [x] `go build ./...` passes

## Out of Scope

* ThrottledPool — not needed; throttle handled by TUI rate limit
* Server-side relay changes — client cannot control