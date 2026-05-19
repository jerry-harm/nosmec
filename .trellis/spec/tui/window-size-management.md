# TUI Window Size Management

> How window dimensions flow through bubblon stack views.

---

## The Problem

Bubble Tea sends `tea.WindowSizeMsg` only once at startup. When using the bubblon stack controller to push child views, the child model's `Init()` runs but `WindowSizeMsg` is **not re-sent** to it — only forwarded through the controller's `Update`. This means child views initialized via `bubblon.Open` receive zero-valued `width`/`height` unless explicitly injected.

---

## Pattern: InjectSize for Child Views

When pushing a model onto the bubblon stack as a child view, the parent must inject its own dimensions **before** calling `bubblon.Open`:

```go
tlModel := timeline.NewModel(m.app, "community", nil, 10, communityAddr)
tlModel.SetBubblonController(m.ctrl)
tlModel.InjectSize(m.width, m.height)  // <-- must call BEFORE bubblon.Open
return m, bubblon.Open(tlModel)
```

`InjectSize` sets the model's `width`/`height` fields and calls `updateListProperties()` to resize the internal list.

---

## Pattern: List Frame Size Subtraction

All list-based views in this project wrap the list in an `app` style with `Padding(1, 2)`:

```go
app: lipgloss.NewStyle().
    Padding(1, 2),  // top/bottom=1, left/right=2  →  frame = (4px horizontal, 2px vertical)
```

When calling `list.SetSize()`, the frame must be subtracted:

```go
// CORRECT — subtract frame
h, v := m.styles.app.GetFrameSize()
m.list.SetSize(m.width-h, m.height-v)

// WRONG — raw dimensions cause overflow, list renders outside terminal bounds
m.list.SetSize(m.width, m.height)
```

This is done in `updateListProperties()` for the timeline model. New views must follow the same pattern.

---

## WindowSizeMsg Handler Contract

Every `tea.Model` that uses a list or viewport must handle `tea.WindowSizeMsg`:

```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    h, v := m.styles.app.GetFrameSize()
    m.list.SetSize(msg.Width-h, msg.Height-v)
```

The pattern is identical across all views:
1. Store raw `width`/`height`
2. Query frame size from `app` style
3. Subtract frame from dimensions
4. Call `list.SetSize` (or `viewport.SetWidth/Height`)

---

## List Initialization

Always initialize list dimensions to `(0, 0)`, not hardcoded values:

```go
// CORRECT — waits for WindowSizeMsg or InjectSize
m.list = list.New(nil, delegate, 0, 0)

// WRONG — hardcoded placeholder (80, 20) fights against real terminal size
m.list = list.New(nil, delegate, 80, 20)
```

---

## Bubblon Controller Message Flow

The bubblon `Controller.Update` forwards **all** messages (including `WindowSizeMsg`) to the top model via the `default` case. Only `openMsg` bypasses this — it calls `model.Init()` on the newly pushed model but does **not** send `WindowSizeMsg`.

```
tea.WindowSizeMsg → Controller.Update → default case → top.Update(msg)
openMsg          → Controller.Update → push model   → model.Init()
```

This is why `InjectSize` exists: parent dimensions must be explicitly propagated to child views that are created and pushed in the same message handler.

---

## Common Mistakes

### Mistake: Child view shows empty / list not visible

**Symptom**: Child view pushed via `bubblon.Open` renders nothing.

**Cause**: Child model's `width`/`height` are zero because `Init()` runs before `WindowSizeMsg` is received, and bubblon doesn't re-send it on push.

**Fix**: Call `InjectSize(m.width, m.height)` on the child model before `bubblon.Open`.

### Mistake: Help bar not visible / rendered off-screen

**Symptom**: Bottom help bar disappears or is cut off at terminal edge.

**Cause**: `list.SetSize` called with raw terminal dimensions without subtracting `app` style frame padding. The list is larger than the rendered area, causing content to overflow.

**Fix**: Use `h, v := m.styles.app.GetFrameSize()` and `m.list.SetSize(msg.Width-h, msg.Height-v)`.

### Mistake: List shows hardcoded 80x20 on startup

**Symptom**: List appears at fixed size before WindowSizeMsg arrives.

**Cause**: `list.New(nil, delegate, 80, 20)` uses hardcoded initial dimensions.

**Fix**: Use `list.New(nil, delegate, 0, 0)` and handle `WindowSizeMsg` properly.

---

## Related

- [tui-testing skill](../skills/tui-testing/SKILL.md)
- `tui/component/bubblon/controller.go` — bubblon stack controller
- `tui/timeline/model.go` — reference implementation (InjectSize + updateListProperties + WindowSizeMsg)
- `tui/community/discover/model.go` — reference for the frame-size subtraction fix