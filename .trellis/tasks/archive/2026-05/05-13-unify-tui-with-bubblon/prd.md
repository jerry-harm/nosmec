# brainstorm: 统一 TUI 显示命令代码，使用 bubblon 实现

## Goal

将所有 TUI 显示命令（event、timeline、compose 等）统一使用 bubblon 栈管理器，消除直接调用 `tea.NewProgram()` 的不一致模式，并清理 orphaned 的 `tui/cmd/cmd.go`。

## What I already know

* bubblon 是栈式窗口管理器（`Controller` 持有 `[]tea.Model` 栈，`Open` 压入，`Close` 弹出）
* `timeline/main.go` 使用 `bubblon.New()` 包装
* `event_commands.go` 直接调用 `tea.NewProgram(m).Run()`，未使用 bubblon
* `tui/cmd/cmd.go` 包含旧的 windowmanager 消息（WinOpen/WinClose 等），是 orphaned 代码
* 近期已删除 windowmanager 包，bubblon 作为替代

## Assumptions (temporary)

* event 命令需要支持多窗口栈（从 timeline/compose 等打开）
* 统一后所有 TUI 使用 `bubblon.Controller` 管理生命周期
* `tui/cmd/cmd.go` 废弃，可以删除或迁移

## Open Questions

1. ~~event 命令需要支持从其他 TUI 窗口打开吗？~~ — 已确认：全部支持栈式导航
2. ~~tui/cmd/cmd.go 是否直接删除？~~ — 已确认：直接删除（windowmanager 已删除，消息已废弃）
3. 是否有其他 TUI 命令（dm 等）也需要统一？ — 调研时发现，暂不涉及

## Requirements (evolving)

* **统一所有 TUI 使用 bubblon** — event、timeline、compose 等命令全部通过 `bubblon.Controller` 管理
* **栈式导航** — timeline → event detail → compose 形成窗口栈，`Close` 命令返回上一层
* **清理 orphaned 代码** — 删除 `tui/cmd/cmd.go`（旧 windowmanager 消息已废弃）
* **一致性** — 所有 TUI 命令使用相同的窗口管理模式

## Acceptance Criteria (evolving)

* [ ] event 命令使用 bubblon.Controller 管理
* [ ] timeline 打开 event detail 使用 `bubblon.Open()`
* [ ] event detail 打开 compose 使用 `bubblon.Open()`
* [ ] `bubblon.Close()` 正确返回上一层窗口
* [ ] tui/cmd/cmd.go 被删除
* [ ] 所有 TUI 命令使用一致的窗口管理模式

## Definition of Done (team quality bar)

* Tests added/updated
* Lint / typecheck / CI green
* 如果有行为变化，更新文档

## Out of Scope (explicit)

* 修改 TUI 内部的业务逻辑（event view、timeline view 的具体渲染）
* 新功能，只做统一和清理

## Technical Notes

* bubblon API: `New(model)`, `Open(model) tea.Cmd`, `Close() tea.Msg`, `Replace(model) tea.Cmd`
* 窗口栈流程: `timeline.RunTimeline()` → 用户选中 event → `bubblon.Open(eventDetailModel)` → 用户回复 → `bubblon.Open(composeModel)` → `bubblon.Close()` 返回 event detail → 再 `bubblon.Close()` 返回 timeline
* 关键文件:
  - `tui/bubblon/controller.go` — bubblon 实现
  - `cmd/event_commands.go` — event 命令（需修改）
  - `cmd/note_commands.go` — note/timeline 命令
  - `tui/cmd/cmd.go` — orphaned 旧代码（待删除）
  - `tui/timeline/main.go` — bubblon 使用示例
  - `tui/window/event/event.go` — event detail 窗口（已有 bubblon 支持）
  - `tui/compose/model.go` — compose 窗口（已有 bubblon 支持）