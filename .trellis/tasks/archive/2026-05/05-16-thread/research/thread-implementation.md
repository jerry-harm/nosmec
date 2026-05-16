# Research: Thread Implementation

- **Query**: How thread view is launched, fetches data, renders, and closes in the nosmec TUI
- **Scope**: internal
- **Date**: 2026-05-16

## Findings

### 1. Entry Point — How Thread View is Launched

**Full call chain:**

```
Timeline (tui/timeline/)
  → User presses Enter on an event item in the list
    → delegate.go:20-25: delegateKeyMap.open ("enter") matches → showDetailMsg created
      → model.go:600-607: showDetailMsg handled → event.New(...) called → bubblon.Open(ev) pushes EventView onto controller stack
        → User presses "t" in EventView
          → event.go:227-228: case "t" → m.thread()
            → event.go:261-270: thread() method → NewThreadTreeView(m.event, ...) → bubblon.Open(threadView) pushes onto controller stack
```

**Key files:**
| File | Role |
|---|---|
| `tui/timeline/delegate.go` | Enter key on timeline item triggers `showDetailMsg` |
| `tui/timeline/model.go:600-607` | Creates `event.New()` + pushes via `bubblon.Open()` |
| `tui/window/event/event.go:143-156` | Initializes keybindings; `"t"` maps to `"thread"` keybinding |
| `tui/window/event/event.go:261-270` | `thread()` method creates `NewThreadTreeView` and pushes via `bubblon.Open()` |

**Keybinding definition** (event.go:151):
```go
thread: key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "thread")),
```

**Open keybinding** (delegate.go:65-68):
```go
open: key.NewBinding(
    key.WithKeys("enter"),
    key.WithHelp("enter", "view"),
)
```

### 2. Thread TreeView — Main Implementation

**File:** `tui/window/event/thread_treeview.go` (445 lines)

**Imports:** Uses `github.com/Digital-Shane/treeview/v2` for the `TuiTreeModel` tree visualization and keyboard navigation.

#### 2.1 Core Data Structures

```go
type threadTreeView struct {
    event    *nostr.Event         // the event the user pressed 't' from
    root     *nostr.Event         // identified root per NIP-10
    app      *config.AppContext   // app context for relay pool, config, timeouts
    tuiModel *treeview.TuiTreeModel[nostr.Event]  // the treeview v2 model for keyboard nav
    provider *NostrEventProvider  // FlatDataProvider interface implementation
    styles   threadStyles         // lipgloss styles
    keys     threadKeyMap         // key bindings (just esc/quit)
    ctrl     *bubblon.Controller  // for bubblon.Close()
    width    int
    height   int
    loading   bool                // loading flag
    loadError error               // error state
    currentEventID string         // for focus/highlight tracking
    mu sync.Mutex
}
```

#### 2.2 NostrEventProvider — treeview.FlatDataProvider

**File:** `thread_treeview.go:90-108`

Implements `treeview.FlatDataProvider[nostr.Event]` with three methods:
- `ID(event)` → `event.ID.Hex()`
- `Name(event)` → truncated content (50 chars max) + first 8 chars of pubkey: `"content... (01234567)"`
- `ParentID(event)` → calls `extractParentID(&event)` which finds first e tag with `"reply"` marker

#### 2.3 NIP-10 Root/Parent Resolution

**`extractParentID(event)`** (thread_treeview.go:20-33):
- Iterates through e tags; returns the value of the first e tag with marker `"reply"`
- Returns `""` if no reply marker (treat as root → no parent)

**`extractRootEvent(event)`** (thread_treeview.go:37-87):
- Collects all e tags via `event.Tags.FindAll("e")`
- No e tags → event IS the root (return event.ID, true, nil)
- Has `"reply"` marker + `"root"` marker → NOT root → return root marker's hex as rootID
- Has `"reply"` marker but NO `"root"` marker → treat event as root (line 74: `return event.ID, true, nil`)
- Has `"root"` marker but NO `"reply"` marker → event IS root (line 81: `return event.ID, true, nil`)
- Has e tags but no markers → treat as root (line 86: `return event.ID, true, nil`)

**Note:** There is also `utils.FindRootEvent()` in `utils/get.go:225-272` which has slightly different logic for the `"reply"` but no `"root"` case — it returns `nostr.ID{}` (zero-value), NOT treating the event as root. This is a divergence between the two implementations.

#### 2.4 Data Fetching Flow

**`fetchThread()`** (thread_treeview.go:161-215) — runs as a tea.Cmd on Init:

1. **Step 1: Identify root** via `extractRootEvent(m.event)`:
   - If `isRoot`: set `m.root = m.event`, add m.event to events list
   - If NOT isRoot: call `fetchRootEvent(ctx, rootID)` to fetch from relays

2. **Step 2: Fetch replies** via `fetchRepliesToRoot(ctx, m.root.ID)` (only if root exists)

3. **Step 3: Build tree** via `buildTuiModel(events)` which:
   - Deduplicates events by ID
   - Creates placeholder nodes for missing parents (with `[loading...]` content)
   - Calls `treeview.NewTreeFromFlatData(...)` to build the tree structure
   - Focuses the current event via `SetFocusedID(...)`, falling back to root
   - Wraps in `TuiTreeModel` with custom keymap, width, height, and disabled nav bar

**`fetchRootEvent(ctx, rootID)`** (thread_treeview.go:217-247):
1. Gets relay hints from current event's e tags: `utils.ExtractRelayHints(m.event)`
2. Falls back to `m.app.AllReadableRelays()` if no hints
3. Builds a note filter for the root ID
4. Queries via `m.app.Pool().QuerySingle(ctx, relays, filter, ...)` on relay hints first
5. If that fails, falls back to `AllReadableRelays()` (line 239)
6. Returns `&result.Event` or nil

**`fetchRepliesToRoot(ctx, rootID)`** (thread_treeview.go:249-272):
1. Uses ALL readable relays (no relay hints for replies)
2. Creates filter: kinds 1 (TextNote) and 1111 (Comment), `#e = rootID`, limit 100
3. Queries via `m.app.Pool().FetchMany(ctxQuery, relays, filter, ...)`
4. Collects all results from channel, returns `[]*nostr.Event`

#### 2.5 Tree Building

**`buildTuiModel(events)`** (thread_treeview.go:277-345):
- Two-pass dedup: Pass 1 adds real events; Pass 2 creates placeholder nodes for any parent IDs not already in the set
- Calls `treeview.NewTreeFromFlatData(context.Background(), items, m.provider)` 
- Sets focus via `tree.SetFocusedID()` to the current event; falls back to root if current event not found
- Wraps in `TuiTreeModel` with:
  - Custom keymap: `threadKeyMapCustom()` — removes esc from Quit binding so `tea.Quit` is not sent
  - Width: `m.width`, Height: `m.height - 4` (leaves room for title + help bar)
  - Nav bar disabled (renders own help bar)

#### 2.6 Update (Event Handling)

**`Update(msg)`** (thread_treeview.go:351-392):

| Message | Behavior |
|---|---|
| `threadTreeLoadedMsg` | Stores error in `m.loadError` |
| `tea.WindowSizeMsg` | Updates width/height, propagates to TuiTreeModel |
| `tea.KeyPressMsg` + `esc` | Returns `bubblon.Close()` (NOT tea.Quit) |
| `tea.KeyPressMsg` + other keys | Delegates to `m.tuiModel.Update(msg)` for navigation |
| Any other message | Delegates to `m.tuiModel.Update(msg)` |

#### 2.7 View (Rendering)

**`View()`** (thread_treeview.go:394-433):
1. Renders `"Thread"` title bar (green background)
2. If loading: `"  [loading thread...]"`
3. If error: `"  [error: ...]"`
4. If `m.tuiModel != nil`: renders TuiTreeModel viewport output
5. Fallback (no tree model but has event): shows single current event
6. Footer: `"↑↓ navigate · →← expand/collapse · enter search · esc back"`
7. Returns `tea.NewView(b.String())` — NO alt screen (unlike old threadView)

**Thread styles** (shared from `thread.go:43-73`):
```go
title:         white text on green (#25A065) background
header:        green (#00FF00) bold
statusMessage: green (#04B575)
helpStyle:     gray (#AAAAAA)
currentEvent:  yellow (#FFFF00)
placeholder:   darker gray (#888888)
rootEvent:     cyan (#00FFFF) bold
```

### 3. How Thread View Closes

The thread view closes via the bubblon stack mechanism, NOT `tea.Quit`:

1. User presses `esc` in thread view
2. `Update()` matches against `m.keys.quit` (key: `"esc"`)
3. Returns command: `func() tea.Msg { return bubblon.Close() }`
4. Bubbletea runtime calls this command → sends `bubblon.closeMsg{notify: true}` 
5. `bubblon.Controller.Update()` receives it → calls `c.pop()` (removes top model from stack)
6. Controller also sends `bubblon.Closed{}` to notify parent that stack changed
7. Now the EventView is the top model again

**Critical design decision:** The TuiTreeModel's default keymap has `esc → tea.Quit` which would exit the entire app. `threadKeyMapCustom()` solves this by setting `km.Quit = nil`, so esc is handled in the thread treeview's own Update method as `bubblon.Close()`.

**Contrast with old threadView** (thread.go:188): old threadView also uses `bubblon.Close()`.

**Contrast with EventView** (event.go:326-330): EventView handles `CloseMsg` differently — if ctrl exists, uses `bubblon.Close()`; if no ctrl, falls back to `tea.Quit`.

### 4. Utility Functions Used

#### `utils.ExtractRelayHints(event)` (get.go:20-39)
- Extracts relay URLs from `e`, `p`, `a`, `q` tags where `tag[2]` is non-empty
- Deduplicates results
- Used by `fetchRootEvent()` to prioritize relay hints from the current event's tags

#### `utils.FindRootEvent(event)` (get.go:225-272)
- Duplicate of `extractRootEvent` in thread_treeview.go with slightly different "reply but no root" handling
- Returns `nostr.ID{}, false, nil` for reply events without root marker (does NOT treat event as root)
- NOT used by the thread treeview — the treeview has its own `extractRootEvent`

#### `utils.QueryRepliesToRoot(ctx, rootID, opts)` (get.go:276-304)
- Query all events referencing rootID via `#e` tag
- Filters: kinds 1 and 1111, limit 100
- NOT used by thread_treeview.go (it has its own `fetchRepliesToRoot`)

#### `utils.GetParentEvent(ctx, event, opts)` (get.go:308-349)
- Finds direct parent via `"reply"` e tag marker
- Uses relay hint from the reply tag (tag[2]) if present
- NOT used by thread_treeview.go (no parent fetching in the treeview)

#### `utils.BuildNoteFilter(id)` (called from get.go)
- Creates a nostr.Filter for a single note ID
- Used by `fetchRootEvent()`

### 5. Test Coverage

#### Tested (thread_treeview_test.go — 239 lines):

| Test Function | What It Tests |
|---|---|
| `TestExtractParentID_RootEvent` | Event with no e tags → empty parent ID |
| `TestExtractParentID_RootMarker` | e tag with "root" marker → empty parent ID |
| `TestExtractParentID_ReplyMarker` | e tag with "reply" marker → returns that tag's value |
| `TestExtractParentID_ReplyMarkerWithRelay` | Reply marker with relay hint → still returns parent ID |
| `TestExtractParentID_MultipleETags` | Multiple e tags → picks first "reply" marker |
| `TestExtractParentID_NoMarker` | e tag without marker → empty parent ID (root) |
| `TestExtractRootEvent_NilEvent` | nil event → error |
| `TestExtractRootEvent_NoETags` | No e tags → event IS root |
| `TestExtractRootEvent_RootMarker` | "root" marker → event IS root |
| `TestExtractRootEvent_ReplyMarker` | "reply" + "root" markers → extracts root from "root" marker |
| `TestExtractRootEvent_ReplyNoRoot` | "reply" but no "root" → event treated as root |
| `TestNostrEventProvider_ID` | ID() returns hex string |
| `TestNostrEventProvider_Name` | Name() returns truncated content + pubkey |
| `TestNostrEventProvider_ParentID_Root` | Root event → empty parent |
| `TestNostrEventProvider_ParentID_Reply` | Reply event → returns parent from "reply" marker |

#### Tested (thread_test.go — 130 lines, for OLD threadView):

| Test Function | What It Tests |
|---|---|
| `TestThreadView_EmptyReplies` | View renders with empty slice |
| `TestThreadView_NilReplies` | View renders with nil replies |
| `TestThreadView_WithParentAndEmptyReplies` | View renders with parent + empty replies |
| `TestThreadView_WithParentAndNilReplies` | View renders with parent + nil replies |
| `TestThreadView_NoParentNoReplies` | View renders with nil parent and no replies |
| `TestNewThreadView` | Constructor sets fields correctly |
| `TestUpdate_ThreadLoadedMsg` | threadLoadedMsg returns nil command |
| `TestUpdate_QuitKey` | esc key returns non-nil command (bubblon.Close) |

#### NOT Tested (gaps in test coverage):

- **`threadTreeView.fetchThread()`** — the entire async fetch flow is untested
- **`threadTreeView.fetchRootEvent()`** — relay query with hints, fallback logic
- **`threadTreeView.fetchRepliesToRoot()`** — FetchMany relay query
- **`threadTreeView.buildTuiModel()`** — tree construction, dedup, placeholder creation
- **`threadTreeView.Update()`** — WindowSizeMsg handling, key delegation to TuiTreeModel
- **`threadTreeView.View()`** — rendering with all state combinations (loading, error, loaded, nil model)
- **`threadTreeView.Init()`** — returns fetchThread command
- **`threadKeyMapCustom()`** — esc Quit binding removed
- **`NewThreadTreeView()`** — constructor sets all fields correctly
- **Integration**: full end-to-end flow (enter on timeline → 't' for thread → esc to close)

### 6. Old vs New Thread Implementation

| Aspect | Old (`thread.go`) | New (`thread_treeview.go`) |
|---|---|---|
| Display | Flat list: root, parent, replies | Tree view with expand/collapse |
| Library | Custom rendering | `treeview/v2` TuiTreeModel |
| Navigation | None (static view) | ↑↓→← keyboard navigation, search |
| Thread depth | 2 levels max (root → replies) | Nested tree via parent-child links |
| Parent fetching | Fetches direct parent via `GetParentEvent` | Does NOT fetch direct parent |
| Placeholders | None | Creates `[loading...]` placeholders for missing parents |
| Current event highlight | `>` marker in text | Focus indicator from TuiTreeModel |
| Alt screen | Yes (`v.AltScreen = true`) | No |
| Usage | Dead code (not bound in UI) | Active (bound to "t" in EventView) |

### 7. Known Issues and Edge Cases

1. **Dead code**: `tui/window/event/thread.go` (old threadView) is still in the package but never instantiated. `event.go:268` calls `NewThreadTreeView`, not `NewThreadView`. The old `threadView` and `thread_test.go` should be removed or the thread view should be renamed.

2. **Duplicate root extraction logic**: `thread_treeview.go:37-87` (`extractRootEvent`) and `utils/get.go:225-272` (`FindRootEvent`) both exist with subtly different logic for the "reply marker without root marker" case. The treeview version treats the event as root; the utils version returns zero-value rootID.

3. **Only 1 level of replies**: `fetchRepliesToRoot()` only queries for events with `#e = rootID`. Nested replies (replies to replies) are NOT fetched. The tree structure relies on existing events' e tags to build parent-child relationships, but if a reply has a parent that wasn't fetched, only a placeholder is created.

4. **Placeholder nodes never resolved**: When a parent event is missing from the fetched set, a `[loading...]` placeholder is inserted. These placeholders are never resolved — there is no mechanism to fetch them later.

5. **No caching**: Thread fetch always hits the network. There is no local cache lookup before relay queries.

6. **Fallback relay logic inconsistency**: `fetchRootEvent()` first queries relay hints, then falls back to `AllReadableRelays()` if the hint query returns nil. But `fetchRepliesToRoot()` always uses `AllReadableRelays()` directly, ignoring any relay hints that might be present on the root event.

7. **Timeout chaining issue**: `fetchThread()` creates one context with `m.app.QueryTimeout()`, then `fetchRootEvent()` and `fetchRepliesToRoot()` each create their own contexts (also with `QueryTimeout()`), effectively doubling the total possible wait time.

8. **No progress feedback during loading**: The loading state only shows `[loading thread...]` — no indication of which step (fetching root vs. fetching replies) is in progress.

9. **Error on no root and no reply events**: If `fetchRootEvent()` returns nil and `fetchRepliesToRoot()` returns empty, the tree model stays nil. The View fallback shows just the current event with no context.

10. **TuiTreeModel focus on missing event**: If `SetFocusedID` fails for the current event AND the root is nil, no fallback focus is set — the tree might render without any visible focus indicator.

### 8. Related Specs

- `.trellis/spec/tui/index.md` — TUI spec index (for coding conventions)
- `.trellis/spec/backend/index.md` — Backend spec index
- `.trellis/spec/guides/cross-layer-thinking-guide.md` — Cross-layer guidelines (thread touches both TUI and nostr relay layers)
- `.trellis/spec/guides/code-reuse-thinking-guide.md` — Code reuse thinking guide (relevant for duplicate root resolution logic)

### 9. Files Summary

| File Path | Description | Lines |
|---|---|---|
| `tui/window/event/thread_treeview.go` | New tree-based thread view (active) | 445 |
| `tui/window/event/thread_treeview_test.go` | Tests for NIP-10 resolution + NostrEventProvider | 239 |
| `tui/window/event/event.go` | Event view with thread launch keybinding | 389 |
| `tui/window/event/thread.go` | Old flat thread view (dead code) | 304 |
| `tui/window/event/thread_test.go` | Tests for old thread view | 130 |
| `utils/get.go` | Fetch helpers (ExtractRelayHints, FindRootEvent, etc.) | 636 |
| `tui/bubblon/controller.go` | Stack-based model navigation | 147 |
| `tui/timeline/delegate.go` | Enter key triggers event detail view | 77 |
| `tui/timeline/model.go` | Timeline model, handles showDetailMsg | 837 |

### Caveats / Not Found

- The `treeview/v2` library (`github.com/Digital-Shane/treeview/v2`) is a third-party dependency. The current version and API surface were not investigated for deprecation or update risks.
- The `nostr` library (`fiatjaf.com/nostr`) is a fork/domain-specific package. Specific version was not checked.
- Removed `thread_treeview_old_test.go` was found in git history (commit `508cedd` "migrate to TuiTreeModel") — the old tests were deleted as part of the migration. The new tests only cover pure functions (NIP-10 resolution, provider interface), not the model itself.
