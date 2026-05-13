# Research: bubbles v2 list delegate UpdateFunc and Enter key

- **Query**: In bubbles v2 list, how does the delegate's UpdateFunc work with Enter key?
- **Scope**: internal + external mixed
- **Date**: 2026-05-13

## Findings

### Files Found

| File Path | Description |
|---|---|
| `tui/timeline/delegate.go` | Timeline list delegate with UpdateFunc for Enter key handling |
| `tui/timeline/model.go` | Timeline model that creates the list with the delegate |
| `tui/bubblon/controller.go` | bubblon Controller for model-stack navigation |

### Code Patterns

#### How `list.New` Works

From `tui/timeline/model.go:205-206`:
```go
delegate := newItemDelegate(m.delegateKeys, &m.styles)
groceryList := list.New(nil, delegate, 0, 0)
```

The `list.New` signature from `charm.land/bubbles/v2/list`:
- `list.New(items list.Item, delegate list.Delegate, width int, height int) list.Model`
- The third and fourth args are width and height (0 = default/caculated)

#### How `DefaultDelegate.UpdateFunc` Works

From `tui/timeline/delegate.go:14-30`:
```go
d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
    if i, ok := m.SelectedItem().(item); ok {
        switch msg := msg.(type) {
        case tea.KeyPressMsg:
            logger.Debug("delegate UpdateFunc received key", "key", msg.String())
            switch {
            case key.Matches(msg, keys.open):
                logger.Debug("delegate matches open key")
                return func() tea.Msg {
                    logger.Debug("delegate creating showDetailMsg")
                    return showDetailMsg{event: i.event, authorName: i.authorName}
                }
            }
        }
    }
    return nil
}
```

The `UpdateFunc` is called by `list.Model.Update()` for every message that reaches the list. It receives the current message and access to the list model (to call `m.SelectedItem()`).

#### Key Binding for Enter

From `tui/timeline/delegate.go:65-68`:
```go
open: key.NewBinding(
    key.WithKeys("enter"),
    key.WithHelp("enter", "view"),
),
```

Enter is bound to the `open` action, which is matched in `UpdateFunc` via `key.Matches(msg, keys.open)`.

#### How `list.Model.Update` Delegates Key Presses

From `tui/timeline/model.go:744-746`:
```go
newListModel, cmd := m.list.Update(msg)
m.list = newListModel
cmds = append(cmds, cmd)
```

The timeline model's `Update` calls `m.list.Update(msg)` which passes the message to `list.Model.Update()`. The list's `UpdateFunc` is called within that update.

#### Why Enter Might Not Trigger UpdateFunc

**Possible causes** (based on code analysis):

1. **Message routing in bubblon**: Looking at `tui/bubblon/controller.go:102-107`:
```go
default:
    if top := c.top(); top != nil {
        m, cmd := top.Update(msg)
        c.models[len(c.models)-1] = m
        return c, cmd
    }
```
When bubblon has models on stack, messages go to top model only. However, timeline model IS the top model (it's what was passed to `bubblon.New(m)` in `timeline/main.go`).

2. **Key handling precedence in timeline model**: Looking at `model.go:702-741`, the timeline model handles keys BEFORE calling `m.list.Update(msg)`. The list keymap handles `r`, `q`, `s`, `T`, `S`, `P`, `H`. The delegate handles `enter` via `UpdateFunc`.

3. **Filtering state**: From `model.go:703-705`:
```go
if m.list.FilterState() == list.Filtering {
    break
}
```
If the list is in filtering mode, the key press breaks out before reaching delegate UpdateFunc.

4. **The delegate UpdateFunc only fires when list Update is called**: The timeline model calls `m.list.Update(msg)` at line 744 AFTER handling its own key bindings. So the sequence is:
   - Timeline `Update(msg)` receives key press
   - Timeline checks its own keymap (r, q, s, T, S, P, H) — Enter is not among these
   - Timeline calls `m.list.Update(msg)` — THIS is when UpdateFunc fires
   - List's UpdateFunc checks `key.Matches(msg, keys.open)` where open = Enter

### External References

- `charm.land/bubbles/v2/list` — official bubbles v2 list package
- `charm.land/bubbletea/v2` — official Bubble Tea v2

### Related Specs

- `.trellis/spec/backend/index.md` — backend spec index (may contain relevant guidelines)

## Caveats / Not Found

- Web search failed due to API authentication issue, so external docs could not be fetched
- The bubbles v2 source code was not available locally to inspect; analysis based on usage patterns in codebase