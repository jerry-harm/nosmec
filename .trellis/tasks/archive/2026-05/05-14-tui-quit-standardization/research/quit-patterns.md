# Research: TUI Quit/Exit Patterns

- **Query**: Quit key bindings and functions across all TUI screens in nosmec
- **Scope**: internal (tui/ directory)
- **Date**: 2026-05-14

## Findings

### TUI Screens Found

| File | Type | Quit Keys |
|------|------|-----------|
| `tui/dm/model.go` | DM chat window | `q`, `ctrl+c`, `esc` |
| `tui/timeline/model.go` | Timeline list | `q`, `ctrl+c` (no `esc`) |
| `tui/compose/model.go` | Note/Reply compose | `q`, `ctrl+c`, `esc` |
| `tui/window/event/event.go` | Event detail view | `esc` only |

### Quit Key Bindings by Screen

#### 1. DM Model (`tui/dm/model.go`)
- **Keys**: `q`, `ctrl+c`, `esc` (line 64)
- **Help text**: `q` → "quit" (line 65)
- **Behavior**:
  - Cancels subscription via `m.subCancel()` before quitting (lines 346-348)
  - Returns `tea.Quit` (line 349)
- **No standalone mode distinction**

#### 2. Timeline Model (`tui/timeline/model.go`)
- **Keys**: `q`, `ctrl+c` (line 138) — **MISSING `esc`**
- **Help text**: `q` → "quit" (line 139)
- **Behavior**:
  - Cancels subscription via `m.subCancel()` before quitting (lines 716-718)
  - Returns `tea.Quit` (line 720)

#### 3. Compose Model (`tui/compose/model.go`)
- **Keys**: `q`, `ctrl+c`, `esc` (line 82)
- **Help text**: `esc` → "quit" (line 83)
- **Special behavior**:
  - Has `SetStandalone()` mode (line 138) where `esc` sends `tea.Quit` instead of `bubblon.Close()`
  - Under bubblon: `esc` sends `bubblon.Close()` (line 283) to preserve draft state
  - Successful send: sends `tea.Quit` (line 263)
  - Standalone `ctrl+c`: sends `tea.Quit` (line 280)
- **Two distinct quit paths** based on standalone flag

#### 4. Event View (`tui/window/event/event.go`)
- **Keys**: `esc` only (line 150) — **`q` is used for "quote"**
- **Help text**: `esc` → "close" (line 150)
- **Behavior**: Returns `tea.Quit` (line 314)
- **Note**: `q` is bound to `quote` action (line 145), not quit

### Window Stack Management (bubblon)

The `tui/bubblon/controller.go` provides stack-based navigation:

| Function | Purpose |
|----------|---------|
| `bubblon.Open(model)` | Push new model onto stack |
| `bubblon.Close()` | Close current model, notify parent via `Closed{}` message |
| `bubblon.Replace(model)` | Close + open atomically |

**Notification behavior**: When a model closes with `bubblon.Close()`, a `Closed{}` message is sent to the parent model. When closing with `tea.Quit`, no notification is sent (stack is just exited).

### Inconsistencies Found

1. **Timeline missing ESC** — `timeline/model.go` only binds `q`/`ctrl+c`, not `esc`. All other screens accept `esc`.

2. **Quit help text inconsistency**:
   - DM: "q" → "quit"
   - Compose: "esc" → "quit"  
   - Timeline: "q" → "quit"

3. **Event view uses `q` for quote** — In event view, `q` triggers "quote", not "quit". Quit is `esc` only. This is intentional but worth noting.

4. **Compose has dual quit path** — Compose distinguishes standalone vs bubblon modes, sending `tea.Quit` vs `bubblon.Close()` respectively. This preserves draft state under the window manager.

5. **No shared quit helper** — Each model implements its own key binding and quit handler. No common `Quit()` function or interface.

### Related Specs

- `.trellis/spec/backend/relay-guidelines.md` — unrelated to TUI quit patterns

### Code Patterns

**Key binding pattern** (all models follow this):
```go
type keyMap struct {
    quit key.Binding
}

func newKeyMap() *keyMap {
    return &keyMap{
        quit: key.NewBinding(
            key.WithKeys("q", "ctrl+c", "esc"),
            key.WithHelp("q", "quit"),
        ),
    }
}
```

**Quit handling pattern**:
```go
case key.Matches(msg, m.keys.quit):
    if m.subCancel != nil {
        m.subCancel()  // cleanup subscription
    }
    return m, tea.Quit
```

**Compose special case** (standalone vs wm):
```go
case key.Matches(msg, m.keys.quit):
    if m.isStandalone {
        return m, tea.Quit
    }
    return m, func() tea.Msg { return bubblon.Close() }
```

### Caveats

- `tui/bubblon/controller.go` handles window management, not quit directly
- Event view uses `tea.Quit` directly without subscription cleanup (no `subCancel` field observed)
- DM and timeline both have `subCancel` cleanup before quit