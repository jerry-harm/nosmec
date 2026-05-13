# compose-ctrl-enter-fix

## Goal

修复 compose 页面中两个按键问题：
1. `ctrl+enter` 在 `contentInput` focused 时无法触发 send
2. `note compose`（standalone RunNoteCompose）按 esc 无法退出

## What I already know

### 当前按键处理（compose/model.go）

- `case tea.KeyMsg:`（错误 — v2 应为 `tea.KeyPressMsg`）
- key handling 在 `contentInput.Update(msg)` 之前
- `contentInput.Focused()` 时 `ctrl+enter` → `key.Matches(msg, m.keys.send)` 检查在 `textarea.Update()` 之前

### 两种打开方式

1. **wm 模式**：timeline → wm.OpenCompose() → compose 是 wm 栈顶
2. **standalone 模式**：`RunNoteCompose()` 直接运行 tea.Program

### wm 模式的 esc 流程（正常）

```
esc → compose.Update → case tea.KeyPressMsg (quit) → return CloseComposeMsg{}
CloseComposeMsg → timeline 收到 → wm.Close("compose")
```

### standalone 模式的 esc 流程（失败）

```
esc → compose.Update → return CloseComposeMsg{}
CloseComposeMsg → 发回给 compose.Update → 无 handler → 忽略
```

## Open Questions

1. `tea.KeyMsg` vs `tea.KeyPressMsg` — 是导致问题的根因吗？
2. standalone 模式下 CloseComposeMsg 的 handler 缺失 — 怎么修？
3. `contentInput.Update` 调用顺序是否影响 ctrl+enter 捕获？

## Requirements

### Bug 1: ctrl+enter 无效

- 当 contentInput focused 时，ctrl+enter 应触发 send
- 根因可能是 `tea.KeyMsg` 写错，或 textarea 自己消费了按键

### Bug 2: note compose esc 无效

- standalone RunNoteCompose 模式下按 esc 无响应
- 根因：CloseComposeMsg 没有被处理（standalone 模式没有 timeline 接收）
- 修复后：esc 应关闭程序（tea.Quit）

## Acceptance Criteria

- [ ] `ctrl+enter` 在 contentInput focused 时触发 send
- [ ] standalone note compose 按 esc 能退出
- [ ] wm 模式下的 esc 行为保持不变

## Out of Scope

- 不改 wm 架构
- 不改 compose 的业务逻辑（send 逻辑不变）