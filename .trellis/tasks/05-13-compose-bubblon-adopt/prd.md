# brainstorm: 用 Bubblon 替换窗口管理

## Goal

将 `WindowManager` 中的自定义栈管理逻辑替换为 Bubblon 库，实现更干净的 model-stack 架构。

## What I already know

### 当前架构

- **`timeline.model`** 是 `tea.Model`（`var _ tea.Model = (*model)(nil)`），是整个程序的主 model
- **`EventView`** 也是独立的 `tea.Model`（`Init/Update/View/ID` 都有）
- **`WindowManager`** — 不是 `tea.Model`，是纯管理层：
  - `stack []string`（z-order 栈）
  - `windows map[string]Window`
  - `focused string`
  - 用 `sync.RWMutex` 保护
- 当 `WindowCount() > 0` 时，timeline 的 `View()` 只渲染 windowManager 的 view（compose 全屏覆盖 timeline）
- 当用户按 `r` / `q` 在 event detail 里回复时，`EventView.reply()` → `wm.PrepareReply()` + `wm.OpenCompose()` 把 compose 窗口 push 进栈

### Bubblon 关键 API

- `bubblon.Controller` 本身是 `tea.Model`，持有 `[]tea.Model` 栈
- 只栈顶 model 接收 `Update` / `View`
- 命令：`Open(model)`, `Close()`, `Replace(model)`, `Fail(err)`
- 关闭时给父 model 发 `Closed{}` 消息
- 只需 model 实现 `tea.Model` 接口

### 迁移方案

将 `WindowManager` 整体替换为 `bubblon.Controller`：

| 现有 | 替换为 |
|------|--------|
| `wm := windowmanager.New()` | `ctrl, _ := bubblon.New(rootModel)` |
| `wm.windows[id] = win` | bubblon 内部管理 |
| `wm.stack` | bubblon 内部管理 |
| `wm.Open(win)` | `bubblon.Open(win)` |
| `wm.Close(id)` | `bubblon.Close()` |
| `wm.focused` | bubblon 内部管理 |
| `wm.WindowCount() > 0` | `len(ctrl.Models()) > 0` |

`timeline.model` 仍然作为主 model（根 model），但 `WindowManager` 字段替换为 `bubblon.Controller`。

`EventView` 和 `ComposeModel` 都已经是 `tea.Model`，直接可以作为 bubblon 栈的元素。

## Open Questions

无阻塞问题 — 架构已理解清楚，可以开始实现。

## Requirements

- `timeline.model` 持有 `bubblon.Controller` 替代 `WindowManager`
- `EventView` 持有 `*bubblon.Controller`（而非 `*WindowManager`）
- `OpenCompose()` 返回 `bubblon.Open(composeModel)` 的命令
- `CloseComposeMsg` 替换为 `bubblon.Closed`（或自定义消息）
- `PrepareReply` / `PrepareQuote` 行为不变
- 当 bubblon 栈为空时，timeline 渲染 list；栈非空时渲染 bubblon 栈顶 view

## Acceptance Criteria

- [ ] `go build` 通过，无循环依赖
- [ ] `timeline.RunTimeline` 正常启动，timeline list 可渲染
- [ ] 在 event detail 按 `r` 打开 compose，写完按 `ctrl+enter` 或 `esc` 正确关闭并返回 event detail
- [ ] 原有 deadlock / overlay rendering 问题已修复

## Definition of Done

- `golangci-lint run` / `go vet` 无警告
- 行为与原有逻辑一致

## Out of Scope

- 不改各个子 model（`ComposeModel`、`EventView`）内部逻辑
- 不改 `RunTimeline` 等入口

## Technical Notes

### 关键实现点

1. **`timeline.model`** 的 `Update` 需要把消息委托给 bubblon.Controller（类似现在委托给 `wm.Update`）
2. **compose close 流程**：现在 compose 发送 `CloseComposeMsg` 触发 `wm.Close(ComposeWindowID)`；迁移后 compose 发送 `bubblon.Close()`，timeline 收到 `bubblon.Closed` 后自行处理（如果需要）
3. **EventView 创建**：EventView 目前由 timeline 在 `openEventDetail` 里 `newEventView()`，需要传入 `*bubblon.Controller`

### 文件改动清单（预估）

```
tui/timeline/model.go       — WindowManager 字段替换为 bubblon.Controller
tui/window/event/event.go   — windowManager *WindowManager 替换为 *bubblon.Controller
tui/windowmanager/          — 删除或降级为最小兼容层（如果其他模块依赖）
tui/compose/model.go        — CloseComposeMsg 改为 bubblon.Closed（或自定义）
go.mod                      — 添加 github.com/donderom/bubblon 依赖
```