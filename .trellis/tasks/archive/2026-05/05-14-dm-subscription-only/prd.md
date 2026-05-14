# dm-subscription-only

## Goal

Fix DM chat to show only subscription-received events, removing optimistic local display of sent messages.

## What I Already Know

* **Problem**: When user sends a DM, `sendDM` immediately appends the message to local display with `fromMe: true`. Then if the subscription echoes it back, it would show duplicate. Currently there's no deduplication.
* **Current behavior**: `sendDM` returns `newMessageMsg{fromMe: true}` which is immediately displayed locally.
* **Correct behavior**: Only show messages received via subscription. Sent messages should appear only when the relay echoes them back via `pollSubscription`.
* **Fix**: Remove the immediate local display from `sendDM` — just send and wait for subscription to deliver.

## Requirements

* Remove optimistic local display from `sendDM`
* Messages only appear when received via `pollSubscription` subscription
* `sendDM` still sends the message (no change to sending logic)
* `sendDM` no longer returns `newMessageMsg`

## Acceptance Criteria

* [ ] Sent messages don't appear immediately in local display
* [ ] Sent messages appear when subscription delivers them
* [ ] No duplicate messages when subscription delivers sent message
* [ ] `go build ./...` succeeds
* [ ] `go vet ./...` passes

## Definition of Done

* Optimistic display removed
* Build and vet pass

## Out of Scope

* Deduplication logic (subscription already unique per event ID if properly filtered)
* Other DM UI changes

## Technical Notes

* File: `tui/dm/model.go`
* Change `sendDM` to return only error, not `newMessageMsg`
* Remove `newMessageMsg` return from `sendDM`
* Subscription via `pollSubscription` delivers all messages including sent ones