# brainstorm: 修复 bubblon 窗口切换和调用

## Goal

修复 bubblon 窗口管理实现，使窗口切换和调用正常工作。

## What I already know

### 当前架构

- `timeline/main.go`: `ctrl, _ := bubblon.New(tlModel)` 创建 controller，`tea.NewProgram(ctrl).Run()` 运行
- `timeline/model.go`: `ctrl bubblon.Controller` 作为字段，`showDetailMsg` 时调用 `m.ctrl.Update(bubblon.Open(ev))` 打开 EventView
- `event/event.go`: `EventView` 持有 `ctrl *bubblon.Controller`，`reply()` / `quote()` 调用 `bubblon.Open(composeModel)`
- `compose/model.go`: esc/ctrl+enter 时返回 `func() tea.Msg { return bubblon.Close() }`

### Bubblon Controller 实现 (`tui/bubblon/controller.go`)

- `Open(model)` 返回 `tea.Cmd` 执行 `openMsg{model}`
- `Close()` 返回 `tea.Msg` 是 `closeMsg{notify: true}`
- `Update(openMsg)` → `c.push(model)` + 调用 `model.Init()`
- `Update(closeMsg)` → `c.pop()`，如果栈非空返回 `Closed{}` 给 parent
- 其他消息默认路由到 `top.Update(msg)`

### 关键问题分析

**问题 1: compose/model.go 第 251 行使用 `tea.KeyMsg` 而非 `tea.KeyPressMsg`**

Bubble Tea v2 中 `tea.Model` 的 `Update` 方法接收 `tea.Msg`，实际运行时发送的是 `tea.KeyPressMsg`（在 v2 中 `KeyMsg` 被废弃）。

```go
// 第 250-265 行
switch msg := msg.(type) {
case tea.KeyMsg:  // ← 错误！v2 中应该是 tea.KeyPressMsg
    if m.sending {
        return m, nil
    }
    if key.Matches(msg, m.keys.quit) {
        // ...
        return m, func() tea.Msg { return bubblon.Close() }
    }
```

这意味着整个 `case tea.KeyMsg:` 块永远不会执行，所有 key 处理都落入下面的通用 `m.kindInput.Update(msg)` 等，导致用户体验混乱（输入框行为异常）。

**问题 2: `bubblon.Open()` 返回 `tea.Cmd`，但调用方式不一致**

在 `timeline/model.go` 第 600 行：
```go
_, cmd := m.ctrl.Update(bubblon.Open(ev))
return m, cmd
```

这里把 `bubblon.Open(ev)` 当作消息传给 `ctrl.Update()`，然后用其返回值 `cmd`。这实际上是正确的，因为 `bubblon.Open(ev)` 返回 `tea.Cmd`（一个函数），当 `ctrl.Update()` 执行这个 cmd 时，它会接收到 `openMsg`。

但是！`ctrl.Update(msg)` 返回 `(tea.Model, tea.Cmd)`。这里的 `cmd` 是从 `openMsg{model}` 的处理中返回的 `model.Init()` 的结果。所以这实际上是对的。

**问题 3: event/event.go 中 `reply()` / `quote()` 调用 `bubblon.Open()` 直接返回**

```go
func (m *EventView) reply() tea.Cmd {
    // ...
    composeModel := compose.NewModel(m.app)
    composeModel.AddReply(m.event)
    return bubblon.Open(composeModel)  // 返回 tea.Cmd
}
```

这返回的是一个 `tea.Cmd`，在 Bubble Tea 中正确。当 `Update` 返回 `tea.Cmd` 时，tea 程序会在下一帧执行这个 cmd。

**问题 4: event/event.go 第 224-226 行的 esc 处理**

```go
case "esc":
    logger.Debug("ESC pressed, sending CloseMsg")
    return func() tea.Msg { return CloseMsg{} }
```

这里返回一个**函数** `func() tea.Msg { return CloseMsg{} }`，这是一个 `tea.Cmd`。但问题是：

1. `EventView.Update()` 返回 `(m, cmd)`，cmd 是这个闭包函数
2. `CloseMsg{}` 是一个普通消息，不是 `bubblon.Close()`
3. `CloseMsg` 在 `timeline/model.go` 中**没有** handler！
4. timeline 的 `Update` 最后会把消息传给 `m.list.Update(msg)`，这不会处理 `CloseMsg`

所以 esc 在 EventView 中实际上什么都不做！

**问题 5: EventView 关闭后 timeline 无法恢复**

即使 esc 能发送 `CloseMsg` 并被正确处理，`EventView.CloseMsg` 返回 `tea.Quit`：
```go
case CloseMsg:
    logger.Debug("CloseMsg received, quitting")
    return m, tea.Quit
```

`tea.Quit` 会导致整个程序退出，而不是返回 timeline！

## 根因总结

1. **compose 使用 `tea.KeyMsg` 而非 `tea.KeyPressMsg`** — 导致 key 处理完全失效
2. **EventView 的 esc 没有正确的关闭机制** — 应该返回 `bubblon.Close()` 而不是 `CloseMsg{}`
3. **EventView 的 `CloseMsg` handler 返回 `tea.Quit`** — 会导致整个程序退出

## 修复方案

### 修复 1: compose/model.go — 把 `tea.KeyMsg` 改成 `tea.KeyPressMsg`

```go
// 第 251 行
case tea.KeyPressMsg:  // 原来是 tea.KeyMsg
```

### 修复 2: event/event.go — esc 返回 `bubblon.Close()`

```go
case "esc":
    logger.Debug("ESC pressed, closing")
    return func() tea.Msg { return bubblon.Close() }
```

### 修复 3: 删除 event/event.go 中不再需要的 `CloseMsg` 类型和相关处理

`CloseMsg` 本身可以保留（因为其他代码可能依赖），但 esc 处理应该用 bubblon 的 close 机制。

### 修复 4: timeline/model.go — 处理 `bubblon.Closed` 消息

当 compose 关闭后，bubblon 会发送 `Closed{}` 消息给 timeline。timeline 的 `Update` 需要处理这个消息。

但是注意：当前 `timeline/model.go` 的 `Update` 方法中，所有未被特殊处理的消息都会落入最后的 `m.list.Update(msg)`。`bubblon.Closed` 消息也会被传进去。

需要添加：
```go
case bubblon.Closed:
    logger.Debug("bubblon.Closed received, returning to timeline")
    return m, nil
```

## Open Questions

无 — 根因已分析清楚，可以开始实现。

## Requirements

1. compose/model.go 中 `tea.KeyMsg` → `tea.KeyPressMsg`
2. event/event.go 中 esc 处理返回 `bubblon.Close()` 而不是闭包返回 CloseMsg
3. timeline/model.go 处理 `bubblon.Closed` 消息

## Acceptance Criteria

- [ ] `go build ./...` 通过
- [ ] `go vet ./...` 无警告
- timeline 启动正常，显示事件列表
- 在 timeline 按 Enter 打开 event detail
- 在 event detail 按 r 打开 compose
- 在 compose 按 esc 或 ctrl+enter 关闭并返回 event detail
- 在 event detail 按 esc 返回 timeline

## Out of Scope

- 不改 EventView 的 reply/quote/follow/delete 等业务逻辑
- 不改 compose 的发送逻辑

## Technical Notes

- `bubblon.Closed` 在 `tui/bubblon/controller.go` 第 12 行定义
- `bubblon.Close()` 在第 35-38 行定义，返回 `closeMsg{notify: true}` 消息
- 当 closeMsg 被处理且栈非空时，第 94-95 行返回 `Cmd(Closed{})` 给 parent