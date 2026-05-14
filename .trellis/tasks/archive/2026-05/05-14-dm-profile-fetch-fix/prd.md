# dm-profile-fetch-fix

## Goal

Fix DM chat to display recipient's profile name in the header instead of just showing the npub.

## What I Already Know

* **Problem**: DM header shows `DM: <npub[:32]>...` without fetching profile name
* **Root cause**: `tui/dm/model.go` View() only uses `recipientNpub`, no call to `GetProfileNameAsync()` to fetch username
* **Precedent**: `tui/window/event/event.go` has `fetchProfileNameAsync()` that fetches and displays profile name in event detail view
* **Fix**: Add `fetchRecipientProfileNameAsync()` to DM model, use it in Init(), show name in header

## Requirements

* DM model fetches recipient profile name asynchronously on init
* Header displays `DM: <name>` when name is available, falls back to `DM: <npub[:32]>...`
* Profile name fetched via `GetProfileNameAsync()` (which now uses the new parallel query implementation)

## Acceptance Criteria

* [ ] Recipient profile name displayed in DM header when available
* [ ] Falls back to npub display when profile not found
* [ ] `go build ./...` succeeds
* [ ] `go vet ./...` passes

## Definition of Done

* DM header shows profile name (or npub fallback)
* Build and vet pass

## Out of Scope

* Profile caching improvements
* Other DM UI changes

## Technical Notes

* File: `tui/dm/model.go`
* New field: `recipientName string` in model struct
* New Init() step: `fetchRecipientProfileNameAsync()`
* View() update: show `recipientName` if non-empty, else `recipientNpub[:32]+"...`