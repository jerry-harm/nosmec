# Research: bubbles v2 list Enter key + UpdateFunc flow analysis

- **Query**: In `charm.land/bubbles/v2/list`, when does `DefaultDelegate.UpdateFunc` fire for the Enter key? Specifically: how `list.Model.Update()` calls the delegate's `UpdateFunc`, and whether there's any special Enter key handling that bypasses it.
- **Scope**: external (bubbles v2 source) + internal (timeline code)
- **Date**: 2026-05-13

## Findings

### bubbles v2 list source (`v2.1.0`)

| File | Relevance |
|---|---|
| `list/list.go:819-850` | `Model.Update()` — main update loop |
| `list/list.go:853-908` | `handleBrowsing()` — called when NOT filtering |
| `list/list.go:904` | `m.delegate.Update(msg, m)` — where UpdateFunc fires |
| `list/defaultitem.go:136-141` | `DefaultDelegate.Update()` — calls `UpdateFunc` if set |
| `list/keys.go` | KeyMap — no Enter binding in list keymap |

#### Key Finding: Enter is NOT in the list's KeyMap

The list's `DefaultKeyMap` (`keys.go:34-97`) has NO Enter binding:

```
CursorUp, CursorDown, PrevPage, NextPage, GoToStart, GoToEnd,
Filter, ClearFilter, CancelWhileFiltering, AcceptWhileFiltering,
ShowFullHelp, CloseFullHelp, Quit, ForceQuit
```

**Enter is only in `AcceptWhileFiltering`** (line 76) — and that only fires when `filterState == Filtering`.

#### Where UpdateFunc is Called

In `handleBrowsing()` (`list/list.go:853`), after all key handling, at line 904:

```go
// line 904
cmd := m.delegate.Update(msg, m)
```

This happens for **every message** after the key switch statement — including any key that wasn't matched by the list's own keymap.

So for Enter (which is not in the list keymap), the flow is:
1. `handleBrowsing` switch — no match for Enter
2. Falls through to `m.delegate.Update(msg, m)` at line 904
3. `DefaultDelegate.Update()` calls `UpdateFunc(msg, m)`

#### DefaultDelegate.UpdateFunc Called for All Messages

From `defaultitem.go:136-141`:
```go
func (d DefaultDelegate) Update(msg tea.Msg, m *Model) tea.Cmd {
    if d.UpdateFunc == nil {
        return nil
    }
    return d.UpdateFunc(msg, m)
}
```

`UpdateFunc` is called for **all messages** that reach `delegate.Update()`, not just key presses.

---

### Timeline Code Path

| File | Relevance |
|---|---|
| `tui/timeline/main.go:23` | `bubblon.New(tlModel)` — timeline IS the initial bubblon model |
| `tui/timeline/model.go:744` | `m.list.Update(msg)` — where list.Update is called |
| `tui/timeline/delegate.go:14-30` | `UpdateFunc` with Enter/open key handling |
| `tui/timeline/delegate.go:20` | `key.Matches(msg, keys.open)` — Enter check |
| `tui/timeline/delegate.go:63-69` | `keys.open` bound to "enter" |
| `tui/timeline/model.go:594-602` | `showDetailMsg` handler → calls `m.ctrl.Update(bubblon.Open(ev))` |

#### Message Flow

```
tea.KeyPressMsg(Enter)
  → timeline.Update(msg)
    → handleBrowsing switch (lines 707-741) — Enter not in timeline keys
    → m.list.Update(msg) [line 744]
      → list.handleBrowsing()
        → switch — no Enter match
        → m.delegate.Update(msg, m) [line 904]
          → DefaultDelegate.Update()
            → UpdateFunc(msg, m) [delegate.go:14]
              → key.Matches(msg, keys.open) [line 20] — Enter matches!
                → returns showDetailMsg{...} via closure
```

#### showDetailMsg → bubblon.Open

```go
// model.go:594-602
case showDetailMsg:
    logger.Debug("showDetailMsg received")
    ev := event.New(&msg.event.Event, m.app, m.width, m.height, msg.authorName, &m.ctrl)
    _, cmd := m.ctrl.Update(bubblon.Open(ev))
    return m, cmd
```

`m.ctrl` is the `bubblon.Controller` that wraps the timeline model in `main.go:23`.

---

### Filtering State Blocks Everything

From `model.go:703-705`:
```go
if m.list.FilterState() == list.Filtering {
    break
}
```

When `filterState == Filtering`, the timeline model's Update breaks out **before** calling `m.list.Update(msg)`. This means `UpdateFunc` never fires during filtering.

Additionally, in `list.handleFiltering()` (`list/list.go:911`), Enter is only matched for `AcceptWhileFiltering` (to apply the filter), and `delegate.Update` is NOT called in the filter path.

---

### Conclusion

**Enter key DOES reach `UpdateFunc`** in the normal (non-filtering) browse flow. There is no special-case in `list.Model.Update()` that intercepts Enter before calling `delegate.Update()`.

The issue must be elsewhere. Possible causes:
1. **Filtering state is active** (`filterState == Filtering`) — blocks the path before `list.Update`
2. **Delegate's `UpdateFunc` is nil** — but we can see it's set in `delegate.go:14`
3. **`keys.open` binding is disabled** — `key.Matches` returns false if binding is disabled
4. **Something intercepts the message before reaching `timeline.Update`** — e.g., bubblon routing issue
5. **The timeline model `Update` returns early before reaching `m.list.Update`** — but Enter is not in the timeline keymap so no early return

**Most likely**: The list is in `Filtering` state when Enter is pressed, or the `keys.open` binding somehow becomes disabled.

---

### External References

- [charm.land/bubbles/v2](https://charm.land/bubbles/) — official bubbles v2 package
- `go list -m -f '{{.Dir}}' charm.land/bubbles/v2` — local module path

### Related Specs

- `.trellis/spec/backend/index.md` — backend spec index

## Caveats / Not Found

- The filtering state at `model.go:703` is the most suspicious blocking mechanism — worth checking if a filter is accidentally active during the user's test
- Could not find any other code path that would intercept Enter before `m.list.Update`