# refactor-thread-view

## Goal

Refactor the thread view to correctly display thread context per NIP-10: query all events referencing the root event (via `["e", <root-id>, <relay>, "root"]` tag) and organize the display by reply relationships.

## What I already know

**Current buggy implementation** (`tui/window/event/thread.go`):
- `getRepliesToEvent` uses `QuerySingle` which returns only 1 event (bug)
- No proper root identification logic
- Shows "No thread data" incorrectly

**NIP-10** (https://github.com/nostr-protocol/nips/raw/refs/heads/master/10.md):
- `["e", <event-id>, <relay-url>, "root"]` — top-level reply to root
- `["e", <event-id>, <relay-url>, "reply"]` — reply to direct parent
- For kind 1, no "e" tag = original root note

## Thread Logic (NIP-10 Compliant)

### Step 1: Identify Root Event

Given current event E:
- If E has e tag with `"root"` marker → E IS the root
- If E has e tag with `"reply"` marker → find root by:
  - First e tag with "root" marker (if exists) = root
  - Otherwise, traverse up via "reply" chain until find "root"
- If E has no e tag → E IS the root (original note)

### Step 2: Query Replies

Query for all events where:
- `["e", <root-id>, <relay>, "root"]` — direct top-level replies
- OR `["e", <root-id>]` without marker (backward compat for kind 1)

### Step 3: Organize Display

Organize by "reply" tag:
- Events with `"reply"` marker to same parent → grouped together
- Show as tree/list structure based on parent-child relationships

### Step 4: Async Loading with Placeholders

**UX Flow:**
1. Immediately show current event with placeholder positions for parent/root
2. Highlight current event position (visual indicator)
3. Async fetch parent event (if current is a reply)
4. Async fetch all replies to root
5. Update display as data arrives (animated or instant)

**Placeholder Strategy:**
- Parent position: Show "[loading parent...]" or skeleton UI
- Root position: Show "[loading root...]" if different from parent
- Replies: Show "[loading X replies...]" with count

## Requirements

### Core
1. **Identify root event** from current event's e tags per NIP-10
2. **Query all replies** using `Pool.Query()` (not `QuerySingle`) with filter:
   - `{"#e": [<root-id>], "#标记": ["root"]}`
3. **Organize by reply chain** - group by direct parent
4. **Show current event's position** in the thread

### Display
- Show root event prominently
- Show direct replies grouped by parent
- Indicate which event is the current one being viewed
- **Placeholder for events not yet loaded (lazy load)**

### Async Loading UX
1. **Immediate display**: Show current event first with placeholder positions
2. **Current event focus**: Highlight/scroll to current event position
3. **Placeholders**: Show loading indicators for parent/root/replies
4. **Progressive update**: As data arrives, replace placeholders with real events
5. **Loading states**: Different placeholder text based on what's loading

### Edge Cases
- Root event not found on relays → show placeholder "[root not found]"
- No replies found → show "[no replies yet]" with root event
- Partial data → show placeholders for missing events
- Current event is root with no replies → show "[no replies yet]"

## Acceptance Criteria

* [ ] Correctly identify root event per NIP-10
* [ ] Query returns ALL events referencing root (not just 1)
* [ ] Display organized by reply chain
* [ ] Current event highlighted in thread view
* [ ] `go build ./...` passes
* [ ] `go test ./...` passes

## Definition of Done

* Thread view shows complete thread context
* NIP-10 compliance verified
* No "No thread data" for valid events with replies

## Technical Approach

1. Add `FindRootEvent(event *nostr.Event) *nostr.ID` helper
2. Add `QueryRepliesToRoot(rootID nostr.ID, opts *GetOptions) []*nostr.Event` using `Pool.Query()`
3. Add `OrganizeByReplyChain(replies []*nostr.Event) map[string][]*nostr.Event`
4. Update `threadView` struct and `fetchThread()` logic
5. Update `View()` to display organized thread

## Out of Scope

* Nested thread tree (> 2 levels) - show flat parent/children only
* Live subscription (realtime updates)
* Thread write operations (reply, quote)