# event-thread-view-treeview

## Goal

Add thread view in event detail showing parent post and thread replies in a tree structure.

## What I Already Know

* Event detail is in `tui/window/event/event.go`
* `GetParentEvent` just added for fetching parent from reply tag
* No thread view exists yet

## Requirements

* Add thread view showing parent event and direct replies
* Tree structure with indentation
* ESC to go back to event detail
* Fetch replies using e tag hints from parent

## Acceptance Criteria

* [ ] Thread view shows parent event above
* [ ] Thread view shows direct replies below
* [ ] Navigation back to event detail
* [ ] `go build ./...` succeeds
* [ ] `go vet ./...` passes

## Definition of Done

* Thread view exists and builds