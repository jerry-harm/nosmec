# Research: bubblon implementation

- **Query**: bubblon usage, window switching, compose close
- **Scope**: internal
- **Date**: 2026-05-13

## Findings

### Files Using bubblon

| File Path | Description |
|---|---|
| `tui/bubblon/controller.go` | Core bubblon.Controller implementation |
| `tui/timeline/model.go` | Timeline model holds `bubblon.Controller` field |
| `tui/timeline/main.go` | Initializes bubblon.Controller with timeline model |
| `tui/window/event/event.go` | EventView holds `*bubblon.Controller`, opens compose |
| `tui/compose/model.go` | Compose model calls `bubblon.Close()` on quit/sendSuccess |

### bubblon.Controller API (tui/bubblon/controller.go)

- **`Open(model tea.Model) tea.Cmd`** (line 30) — pushes model onto stack
- **`Close() tea.Msg`** (line 36) — closes top model, sends `Closed{}` notification to parent
- **`Replace(model tea.Model) tea.Cmd`** (line 41) — close + open in one command
- **`Controller.Models() int`** (line 66) — returns stack depth
- **`Controller.Update(msg tea.Msg)`** (line 81) — routes messages, handles openMsg/closeMsg internally, delegates others to top model

### Key Usage Patterns

#### timeline/model.go — opening EventView (line 600)
```go
_, cmd := m.ctrl.Update(bubblon.Open(ev))
```
The timeline model receives `showDetailMsg`, creates an `event.New()`, then opens it via bubblon.

#### event/event.go — opening Compose (lines 239-241, 251-253)
```go
func (m *EventView) reply() tea.Cmd {
    composeModel := compose.NewModel(m.app)
    composeModel.AddReply(m.event)
    return bubblon.Open(composeModel)
}

func (m *EventView) quote() tea.Cmd {
    composeModel := compose.NewModel(m.app)
    composeModel.AddQuote(m.event)
    return bubblon.Open(composeModel)
}
```
Both `reply()` and `quote()` use `bubblon.Open()` directly as a command return value.

#### compose/model.go — closing (lines 247, 264)
```go
case sendSuccessMsg:
    m.success = true
    m.errMsg = ""
    m.ClearDraft()
    return m, func() tea.Msg { return bubblon.Close() }
    ...

if key.Matches(msg, m.keys.quit) {
    // Send bubblon close instead of tea.Quit to preserve draft state
    return m, func() tea.Msg { return bubblon.Close() }
}
```
Compose sends `bubblon.Close()` on both quit and successful send.

#### timeline/model.go — rendering bubblon stack (lines 761-766)
```go
func (m *model) View() tea.View {
    if m.ctrl.Models() > 0 {
        v := m.ctrl.View()
        v.AltScreen = true
        return v
    }
    v := tea.NewView(m.styles.app.Render(m.list.View()))
    v.AltScreen = true
    return v
}
```
Timeline checks `m.ctrl.Models() > 0` to decide whether to render the bubblon stack.

#### timeline/model.go — forwarding ProfileLoadedMsg to EventView (line 609)
```go
case event.ProfileLoadedMsg:
    _, cmd := m.ctrl.Update(msg)
    return m, cmd
```
Messages from profile loading are forwarded directly to the bubblon controller.

### Potential Issue: Message Forwarding

The `event/event.go` EventView is a model pushed onto the timeline's bubblon stack. However, the EventView has its own `ctrl *bubblon.Controller` field (line 55) for managing compose windows. When `timeline/model.go` forwards messages to `m.ctrl.Update(msg)` (line 609), those messages go to the EventView — but if the EventView then needs to forward them to its own compose controller, there is no explicit mechanism shown.

### Potential Issue: nil ctrl Check in event/event.go

Both `reply()` (line 236) and `quote()` (line 248) check `if m.ctrl == nil { return nil }`. This nil check suggests the EventView's controller could be nil in some code paths. The `EventView` struct stores `ctrl *bubblon.Controller` but it's passed in from outside — the EventView itself doesn't create it.

### Controller Initialization (timeline/main.go line 23)
```go
ctrl, err := bubblon.New(tlModel)
```
The timeline's Controller is initialized with the timeline model as the base. Compose and EventView are pushed on top as additional models.

### Related Specs

- `.trellis/spec/backend/index.md` — backend spec index (packages use `backend` layer)