# event-detail-navigation

## Goal

Add navigation from event detail to user timeline and thread view.

## What I Already Know

* Event detail is in `tui/window/event/event.go`
* Thread view exists in `tui/window/event/thread.go`
* Navigation uses `bubblon.Open`

## Requirements

* Add key binding (e.g., 't' for thread view) to open thread view from event detail
* Add key binding to open user timeline for event author
* ESC or 'q' to go back

## Acceptance Criteria

* [ ] Thread view opens from event detail
* [ ] User timeline navigation works
* [ ] `go build ./...` succeeds
* [ ] `go vet ./...` passes

## Definition of Done

* Navigation works from event detail