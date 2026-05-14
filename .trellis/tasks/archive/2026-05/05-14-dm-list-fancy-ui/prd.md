# dm-list-fancy-ui

## Goal

Create a DM conversation list TUI screen using FancyList showing username/npub + latest message, then navigate to individual DM chat.

## What I Already Know

* `ListDMConversations` returns `[]Conversation` with PubKey, LatestDM, LatestAt
* `Conversation` has `PubKey string`, `LatestDM DMMessage`, `LatestAt nostr.Timestamp`
* DMMessage has `Content string`, `FromMe bool`, `Timestamp nostr.Timestamp`
* Existing DM chat is at `tui/dm/main.go`
* Need to fetch profile names for each conversation

## Requirements

* Create `tui/dm/list/` directory with list screen
* Use `FancyList` or similar component showing:
  - Username or npub (fetched via `GetProfileName`)
  - Latest message content (truncated)
  - Timestamp
* Selection navigates to individual DM chat (`dm.RunDM`)
* ESC goes back to previous screen
* Load conversations on init

## Acceptance Criteria

* [ ] DM conversation list shows all conversations
* [ ] Each row shows username/npub + latest message + timestamp
* [ ] Selection navigates to DM chat with that person
* [ ] `go build ./...` succeeds
* [ ] `go vet ./...` passes

## Definition of Done

* List screen works, build passes