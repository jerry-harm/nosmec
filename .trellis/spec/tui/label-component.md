# Label Component

> Pubkey/username chip for TUI — clickable, async profile fetch.

## Overview

`label` is a self-contained `tea.Model` that displays a pubkey as a styled chip. It resolves the profile name asynchronously and renders three visual states: loading (gray), resolved (green chip), error (dark red chip).

**Location:** `tui/component/label/`

---

## Public API

### `Config`

```go
type Config struct {
    Pubkey string            // hex pubkey
    App    *config.AppContext // for relay queries
}
```

### `Model`

```go
func New(cfg Config) *Model

func (m *Model) Init() tea.Cmd         // starts async profile fetch
func (m *Model) Update(tea.Msg) (tea.Model, tea.Cmd)
func (m *Model) View() tea.View         // returns tea.NewView with lipgloss styling + OnMouse handler

func (m *Model) Focus()
func (m *Model) Blur()
func (m *Model) IsFocused() bool
```

### `State`

```go
const (
    StateLoading  State = iota
    StateResolved
    StateError
)
```

### `LabelClickedMsg`

```go
type LabelClickedMsg struct {
    Pubkey string
}
```

Emitted when user clicks the label (left mouse button).

---

## Static Render Helper

For embedding label-style text in a parent view without managing a sub-model:

```go
func RenderLabel(pubkey, name string, state State) string
```

Usage in timeline/item:
```go
labelStr := label.RenderLabel(pubkeyHex, authorName, label.StateResolved)
```

---

## Visual States

| State | Style | Rendered text |
|-------|-------|---------------|
| `StateLoading` | Gray foreground, no background | `@abc123...` |
| `StateResolved` | Green foreground, dark green background, padding | `@username` |
| `StateError` | Gray foreground, dark red background, padding | `@abc123...` |
| `focused` | White foreground, green background, border | `@username` |

---

## Interaction

- **Mouse click (left)** — emits `LabelClickedMsg{Pubkey}` via `tea.View.OnMouse`
- **Focus** — via `Focus()`/`Blur()` (for future Tab navigation by parent)

---

## Integration Points

| View | File | Method | Notes |
|------|------|--------|-------|
| Timeline | `tui/timeline/model.go` | `formatItemTitle()` | Replaced `npub[:16]` placeholder |
| Event detail | `tui/event/view.go` | `renderHeader()` | Replaced author name |
| Thread | `tui/thread/thread.go` | `eventProvider.Name()` | Replaced pubkey prefix |

---

## Mouse Handling

`tea.View.OnMouse` is set in `View()`:

```go
v.OnMouse = func(msg tea.MouseMsg) tea.Cmd {
    click, ok := msg.(tea.MouseClickMsg)
    if !ok { return nil }
    mouse := click.Mouse()
    if mouse.Button == tea.MouseLeft {
        return func() tea.Msg { return LabelClickedMsg{Pubkey: m.pubkey} }
    }
    return nil
}
```

Note: The `OnMouse` handler does NOT perform bounds checking against the rendered label dimensions. Click events within the terminal window will fire the message regardless of whether the cursor was over the label text. This is a known limitation — parent views that need precise hit testing should add it in their own `Update()` before propagating to child label models.

---

## Future Extensions

- **Bounds-checked click detection** — use rendered dimensions to filter clicks
- **Hover state** — via `MouseMotionMsg` tracking cursor position
- **Keyboard focus+enter** — parent calls `Focus()`/`Blur()` and routes Enter to focused label
- **Context menu** — right-click emits a different message type