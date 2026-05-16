# thread-view-with-treeview

## Goal

Refactor thread view to use `Digital-Shane/treeview` library for proper tree structure display - showing thread hierarchy with root, parent, and replies in a navigable tree.

## What I already know

**Library**: [Digital-Shane/treeview](https://github.com/Digital-Shane/treeview)
- Full Bubble Tea support (与本项目 TUI 架构兼容)
- Lipgloss styling support
- 支持 keyboard navigation, search/filter, viewport
- GPL-3.0 license
- 83 stars, 3 forks

**TreeView 特性**:
- Build trees from flat data (parent ID relationships)
- Build trees from hierarchical data
- Keyboard navigation (arrow keys)
- Viewport support for large trees
- Custom styling via Lipgloss
- Provider system for custom styling

**Thread 需求**:
- 显示 root event (no e tag or root marker)
- 显示 parent event (reply tag with "reply" marker)
- 显示 direct replies (events referencing root with "root" marker)
- 显示 hierarchical structure (tree, not flat list)

## Thread Logic (UX Requirements)

### Display Rules
1. **Current event** — always visible, highlighted
2. **Downward** — show 1 level of direct replies (children), collapsed by default
3. **Upward** — show complete chain to root when navigating up
4. **Expand** — deeper thread chains require manual expansion (lazy loading)

### NIP-10 Compliance
- Root: event with no e tag OR e tag with "root" marker
- Direct reply: event with e tag pointing to root with "root" marker
- Reply chain: e tags with "reply" marker (parent chain)

### Thread Fetch Strategy

**Fetch entire thread in ONE call** (not per-node lazy loading):
- Build filter for all events in thread: root + direct replies
- Use relay hints from NIP-10 e tag: `["e", <id>, <relay-url>, <marker>]`
- Fallback to `App.AllReadableRelays()` if no relay hint
- Use `Pool.FetchMany()` or `Pool.Query()` to get all at once

**Relay priority**:
1. Relay from event's e tag (position 2 in NIP-10 format)
2. Relay from relay list (NIP-65)
3. All readable relays as fallback

### TreeView UX
- Root node = root event (collapsible)
- First level children = direct replies to root
- Expand interaction = keyboard navigation to load more
- Current event = focus/highlighted node
- Missing events = placeholder nodes with "[loading...]"

### Keyboard Navigation
- **Up/Down arrows** — navigate between nodes at same level (siblings)
- **Right arrow** — expand node (show children) / enter child thread
- **Left arrow** — collapse node (hide children) / go to parent level
- **Enter** — select/focus node (scroll to current event if different)
- **/** — search (find by event ID, pubkey, content)
- **Esc** — quit thread view

## Requirements (evolving)

* Use `Digital-Shane/treeview` library for tree display
* NIP-10 compliant root identification
* Async loading of events with placeholders
* Keyboard navigation support
* Display thread hierarchy

## Acceptance Criteria

* [ ] Thread displays as tree structure (not flat list)
* [ ] Root event shown at top
* [ ] Replies shown as children
* [ ] Keyboard navigation (Up/Down/Left/Right arrows) works
* [ ] Current event highlighted
* [ ] Placeholder nodes for loading
* [ ] Relay hints from NIP-10 e tags used for fetch
* [ ] `go build ./...` passes

## Definition of Done

* Thread view displays as tree using treeview library
* Treeview keyboard navigation functional
* NIP-10 relay hints properly used for thread fetching
* Current event position tracked and highlighted

## Out of Scope

* Thread write operations (reply, quote) - display only
* Live subscription (realtime updates)
* Search functionality (MVP)

## Technical Approach

### Step 1: Add treeview dependency
```
go get github.com/Digital-Shane/treeview/v2
```

### Step 2: Create NostrEvent Provider
```go
type NostrEventProvider struct{}
func (p *NostrEventProvider) ID(d nostr.Event) string { return d.ID.Hex() }
func (p *NostrEventProvider) Name(d nostr.Event) string {
    // truncated content + author npub
    return truncate(d.Content, 50)
}
func (p *NostrEventProvider) ParentID(d nostr.Event) string {
    // NIP-10: root marker or no marker = root (no parent)
    // reply marker = parent is e tag[1]
    return extractParentID(d.Tags)
}
```

### Step 3: Thread Fetch
1. Identify root event from current event's e tags
2. Build filter: `{"#e": [root.ID.Hex()], "kinds": [1, 1111]}`
3. Use relay hints from current event's e tags
4. Fetch all with `Pool.FetchMany()`

### Step 4: Build Tree
```go
tree, err := treeview.NewTreeFromFlatData(
    ctx, events, &NostrEventProvider{},
    treeview.WithExpandFunc(func(n *treeview.Node[nostr.Event]) bool {
        return n.Depth() <= 1  // expand 1 level down
    }),
)
```

### Step 5: Create TUI Model
```go
model := treeview.NewTuiTreeModel(tree,
    treeview.WithTuiWidth(80),
    treeview.WithTuiHeight(25),
    treeview.WithTuiAltScreen(true),
)
```

### Step 6: Keyboard Navigation
- Treeview default keys: Up/Down/Left/Right/Enter
- Custom handler for `/` search (optional MVP)
- Esc to quit → `bubblon.Close()`