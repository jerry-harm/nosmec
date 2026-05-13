# 05-13-tui

## Goal

让所有 TUI 窗口（timeline、event、compose、dm）全屏显示。

## What I already know

### 当前窗口实现

**timeline/model.go**：
- `View()` 第 766-768 行：已经是全屏 — `tea.NewView(...).AltScreen = true`

**event/event.go**：
- `View()` 第 366 行：`return tea.NewView(m.viewport.View() + "\n" + m.help.View(m.keys))`
- **问题**：没有设置 `AltScreen = true`，不是全屏

**compose/model.go**：
- `View()` 返回 `tea.NewView(...)` 带 `.AltScreen = true`（已全屏）

**dm/model.go**：未检查

### 全屏模式实现方式

BubbleTea v2 中全屏通过 `tea.View` 的 `AltScreen` 字段实现：
```go
v := tea.NewView(someContent)
v.AltScreen = true  // 启用 alt screen 模式（隐藏 chrome，换全屏）
return v
```

## Open Questions

1. dm 窗口是否也需要全屏？（standalone 模式）
2. event detail 窗口是否需要全屏？（当前 wm 栈中是 overlay）

## Requirements

### 统一全屏行为

所有 TUI 窗口的 `View()` 方法必须返回带 `AltScreen = true` 的 `tea.View`：
- event/event.go：设置 `AltScreen = true`
- dm/model.go：检查并同样处理

### 不需要修改

- timeline 已全屏，无需修改
- compose 已全屏，无需修改

## Acceptance Criteria

- [ ] event View() 返回的 tea.View 设置了 AltScreen = true
- [ ] dm View() 返回的 tea.View 设置了 AltScreen = true（如需全屏）

## Out of Scope

- 不改窗口逻辑，只改 View() 返回的包装
- 不改 AltScreen 以外的样式

## Technical Notes

需要修改的文件：
- `tui/window/event/event.go` — View() 方法
- `tui/dm/model.go` — 待检查