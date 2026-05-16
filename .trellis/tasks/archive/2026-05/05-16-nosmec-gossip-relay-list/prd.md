# nosmec gossip relay list

## Goal

Implement `nosmec gossip` command to batch fetch users' relay lists (NIP-65) and add discovered relays to KnownRelays. Also improve relay hint usage in e-tag scenarios for fetching parent events and thread data.

## What I already know

**Existing code**:
- `utils/user_relays.go` has `DiscoverUserRelays()` for single user
- `utils/get.go` has `ExtractRelayHints()` to extract relays from e/p/a/q tags
- `cmd/` uses cobra for commands, registered via `registerDefaultCommands()`
- `config.AppContext` has `AllReadableRelays()`, `KnownRelays`, `TrackRelays()`
- Subscriptions stored in `app.ListSubscriptions("user")` - contains followed pubkeys

**NIP-65 relay list**:
- Kind 10002 is replaceable event containing read/write relay lists
- `nip65.ParseRelayList(event)` returns `(readRelays, writeRelays)`

## Requirements

* `nosmec gossip` - batch fetch NIP-65 relay lists for all followed users
* Silent operation with spinner + progress count (users processed, relays discovered)
* CLI mode (no full screen TUI)
* Discovered relays added to KnownRelays on exit (persist to config)
* Relay verification only at save time (not during fetch)
* Thread view uses e-tag relay hints for fetching parent/root events

## Acceptance Criteria

* [ ] `nosmec gossip` command exists and works
* [ ] Spinner shows: "Discovering relays... (X users, Y relays found)"
* [ ] Discovered relays persisted to KnownRelays
* [ ] Thread view uses e-tag relay hints for fetching
* [ ] `go build ./...` passes

## Out of Scope

* Full screen TUI for gossip (CLI spinner only)
* Real-time relay verification during fetch

## Technical Approach

### Command: `nosmec gossip`
1. Get all pubkeys from subscriptions (type="user")
2. For each pubkey, call `DiscoverUserRelays()` (background, no verification)
3. Show spinner with count: users processed, relays found
4. On completion, save relays to config via `EnsureRelays()`

### Relay hint usage
1. Modify `thread_treeview.go` `fetchRootEvent()` to use e-tag relay hints first
2. Modify `fetchRepliesToRoot()` to use e-tag relay hints
3. Add relay hints to `GetParentEvent()` in `utils/get.go`