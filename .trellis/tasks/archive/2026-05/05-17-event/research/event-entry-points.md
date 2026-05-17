# Research: Event Detail Entry Points

- **Query**: All entry points to the event detail view in nosmec and what features each one has
- **Scope**: internal
- **Date**: 2026-05-17

## Findings

### Overview

There are **3 entry points** to the EventView, all residing in different call sites but sharing the same `EventView` struct and methods in `tui/window/event/`. The core struct has two constructors:

| Constructor | File:Line | Purpose |
|---|---|---|
| `New(event, app, width, height, authorName, ctrl)` | `tui/window/event/event.go:81` | Event already loaded in memory |
| `NewFromID(eventID, app, width, height, ctrl)` | `tui/window/event/event.go:103` | Event ID provided; fetches asynchronously |

---

### Entry Point 1: CLI — `nosmec event <event-id>`

**File**: `cmd/event_commands.go`

**How it creates EventView**:
```go
// cmd/event_commands.go:56
m := event.NewFromID(eventID, app, 80, 24, nil)
```
- Uses `NewFromID` (fetches event async via `fetchEventAsync` → `EventLoadedMsg`)
- `ctrl` is **`nil`** — no bubblon controller
- Fixed dimensions: width=80, height=24
- Wraps in a standalone `tea.NewProgram(ctrl).Run()` (line 61)

**Feature availability**:

| Feature | Key | Works? | Why |
|---------|-----|--------|-----|
| Reply | `r` | ❌ | `m.ctrl == nil` → returns nil (`event.go:241-243`) |
| Quote | `q` | ❌ | `m.ctrl == nil` → returns nil (`event.go:253-255`) |
| Thread | `t` | ❌ | `m.ctrl == nil` → returns nil (`event.go:265-267`) |
| Delete | `d` | ✅ | No ctrl check; calls `utils.DeleteNote` directly (`event.go:272-282`) |
| Follow/Unfollow | `f` | ✅ | No ctrl check; uses `m.app.ListSubscriptions` (`event.go:284-307`) |
| Open in browser | `o` | ✅ | No ctrl check; runs `xdg-open nostr:...` (`event.go:309-319`) |
| Toggle raw JSON | `j` | ✅ | Simple boolean toggle, no ctrl needed (`event.go:223-226`) |
| Close | `esc` | ✅ | `ctrl == nil` path → `tea.Quit` (quits whole program) (`event.go:327-330`) |

**Information shown**:
- Header: full pubkey (npub), author name (fetched from profile async), timestamp, kind
- Content: note text
- Tags section (numbered list)
- Signature section: nevent, sig hex
- Raw JSON toggle: all event fields as formatted JSON
- Loading state shown while async fetch is in progress

**Note on input**: Accepts raw 64-char hex ID or nevent/note format (NIP-19 decoding).

---

### Entry Point 2: Timeline List — Enter on item

**Files**:
- `tui/timeline/delegate.go` (Enter key binding → `showDetailMsg`)
- `tui/timeline/model.go:600-607` (`showDetailMsg` handler)

**How it creates EventView**:
```go
// tui/timeline/model.go:603
ev := event.New(&msg.event.Event, m.app, m.width, m.height, msg.authorName, &m.ctrl)
return m, bubblon.Open(ev)
```
- Uses `New` (event already loaded in timeline memory — no async fetch needed)
- `ctrl` is **`&m.ctrl`** (non-nil) — all bubblon features enabled
- Dimensions: inherits timeline width/height (full terminal)
- `authorName`: passed from the list item (resolved from `fetchProfileNames`)

**Feature availability**:

| Feature | Key | Works? |
|---------|-----|--------|
| Reply | `r` | ✅ Opens compose via `bubblon.Open(composeModel)` |
| Quote | `q` | ✅ Opens compose via `bubblon.Open(composeModel)` |
| Thread | `t` | ✅ Opens `NewThreadTreeView` via `bubblon.Open(threadView)` |
| Delete | `d` | ✅ Sends Kind 5 deletion event |
| Follow/Unfollow | `f` | ✅ Toggles user subscription |
| Open in browser | `o` | ✅ Opens `nostr:<id>` via `xdg-open` |
| Toggle raw JSON | `j` | ✅ Toggles formatted JSON view |
| Close | `esc` | ✅ `ctrl != nil` path → `bubblon.Close()` (returns to timeline) |

**Information shown**: Same as CLI, but author name is already resolved (no async profile fetch needed in most cases).

**Key flow**:
```
Timeline list
  → Enter on item (delegate.go:18-26, key.Matches(msg, keys.open))
  → showDetailMsg{event, authorName} (delegate.go:24)
  → model.go:600-607: event.New(...) → bubblon.Open(ev)
  → EventView rendered on top via bubblon controller stack
```

---

### Entry Point 3: Thread Tree View — Enter on focused node

**File**: `tui/window/event/thread_treeview.go`

**How it creates EventView**:
```go
// tui/window/event/thread_treeview.go:433
eventView := New(&ev, m.app, m.width, m.height, "", m.ctrl)
return m, bubblon.Open(eventView)
```
- Uses `New` (event already loaded in thread tree)
- `ctrl` is **`m.ctrl`** (non-nil, inherited from parent EventView's controller) — all bubblon features enabled
- `authorName` is **`""`** — empty; triggers `fetchProfileNameAsync()` in `Init()` (`event.go:179-181`)
- Dimensions: inherited from thread view dimensions
- **Guard**: skips placeholder nodes (`[loading...]`) and zero-ID events (`event.go:432`)

**Feature availability**: Same as Timeline entry point — all 8 features work.

| Feature | Key | Works? |
|---------|-----|--------|
| Reply | `r` | ✅ |
| Quote | `q` | ✅ |
| Thread | `t` | ✅ (opens a *new nested* ThreadTreeView) |
| Delete | `d` | ✅ |
| Follow/Unfollow | `f` | ✅ |
| Open in browser | `o` | ✅ |
| Toggle raw JSON | `j` | ✅ |
| Close | `esc` | ✅ Returns to thread tree view |

**Information shown**: Same as other entry points; author name fetched asynchronously.

**Key flow**:
```
EventView → press 't' → ThreadTreeView opens via bubblon.Open
  → ↵↑↓ navigate tree, Enter on a node
  → New(&ev, ...) → bubblon.Open(eventView)
  → Nested EventView (esc returns to thread tree)
  → Esc from thread tree → back to original EventView
```

---

### Comparison Table

| Feature | CLI (`nosmec event`) | Timeline Enter | Thread Tree Enter |
|---------|:---:|:---:|:---:|
| **Constructor** | `NewFromID` | `New` | `New` |
| **ctrl (bubblon)** | `nil` | `&m.ctrl` | `m.ctrl` |
| **Reply (r)** | ❌ | ✅ | ✅ |
| **Quote (q)** | ❌ | ✅ | ✅ |
| **Thread (t)** | ❌ | ✅ | ✅ |
| **Delete (d)** | ✅ | ✅ | ✅ |
| **Follow (f)** | ✅ | ✅ | ✅ |
| **Open (o)** | ✅ | ✅ | ✅ |
| **JSON toggle (j)** | ✅ | ✅ | ✅ |
| **Close (esc)** | Quits app | Returns to timeline | Returns to thread tree |
| **Author name** | Fetched async | Pre-resolved | Fetched async |
| **Dimensions** | Fixed 80×24 | Terminal size | Inherited from parent |

---

### Code Sharing Analysis

All three entry points use the **same `EventView` struct** (`tui/window/event/event.go:37-56`) and all its methods. There is **no code duplication** of the event detail view.

The two code paths diverge only at two points:

1. **Constructor** — `New` vs `NewFromID` (identical except for loading state)
2. **bubblon controller presence** — `ctrl == nil` gates `reply()`, `quote()`, `thread()` and the `CloseMsg` handler branch (`tea.Quit` vs `bubblon.Close`)

The `delete()`, `follow()`, `openInBrowser()`, and `showRawJSON` toggle have **no ctrl dependency** and work identically across all entry points.

---

### Delete/Restore Functionality

- **Delete**: `EventView.delete()` (`event.go:272-282`) calls `utils.DeleteNote()` (`utils/post.go:120-150`)
  - Sends a Kind 5 (deletion) event with `e` tag pointing to the target event
  - Publishes to all writable relays
  - Returns nil (no UI feedback on success/failure beyond logging)
- **Restore/Undelete**: **Not implemented**. No `RestoreNote` function exists anywhere in the codebase.

---

### Files Found

| File Path | Description |
|---|---|
| `tui/window/event/event.go` | Core EventView model (389 lines) — struct, constructors, Update, View, all action methods |
| `tui/window/event/view.go` | Rendering: `renderHeader()`, `renderContent()`, `renderRawJSON()` (125 lines) |
| `tui/window/event/styles.go` | Lipgloss styles for EventView (49 lines) |
| `tui/window/event/thread.go` | Old linear thread view (304 lines) — still exists but superseded by tree view |
| `tui/window/event/thread_treeview.go` | New tree-based thread view (507 lines) — NIP-10 parsing, recursive fetch, Enter→EventView |
| `tui/window/event/thread_treeview_test.go` | Thread tree view tests (802 lines) |
| `cmd/event_commands.go` | CLI `nosmec event` command (63 lines) — decodes nevent/note, calls `RunEventDetail` |
| `tui/timeline/model.go` | Timeline model (837 lines) — `showDetailMsg` handler at lines 600-607 |
| `tui/timeline/delegate.go` | Timeline item delegate (77 lines) — Enter key → `showDetailMsg` |
| `utils/post.go` | Post utilities — `DeleteNote()` at lines 120-150 |

### Related Specs

- `.trellis/spec/tui/index.md` — TUI layer guidelines
- `.trellis/spec/backend/index.md` — Backend layer guidelines

### Caveats / Not Found

1. **CLI entry point has reduced functionality**: reply, quote, thread are all silent no-ops because `ctrl == nil`. No error message or feedback is given to the user when pressing r/q/t in CLI mode — the key press is just silently consumed.
2. **No restore/undelete**: Once deleted (Kind 5 published), there is no way to undo it from the UI.
3. **Old thread view still exists**: `thread.go` (`threadView`) is the old linear thread view. It's still in the package but the `EventView.thread()` method now calls `NewThreadTreeView` instead (`event.go:268`). The old `threadView` has no references in the codebase outside its own file.
4. **No community or search views** open EventView. Only the timeline list and thread tree view serve as TUI entry points.
5. **Delete has no UI feedback**: `delete()` returns `nil` cmd — the user sees no success/failure indication beyond log output.
