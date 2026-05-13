# compose-ctrl-enter-v2

## Goal

修复 compose 页面中 `ctrl+enter` 无法触发 send 的问题（bubbletea v2 升级后）。

## What I already know

### 问题根因

根据 [UPGRADE_GUIDE_V2.md](https://github.com/charmbracelet/bubbletea/raw/refs/heads/main/UPGRADE_GUIDE_V2.md)，bubbletea v2 中 textarea 的 `Update()` 方法内部会处理 key events。如果 `ctrl+enter` 在 textarea 之前没有被拦截，textarea 内部的 key handler 可能消费这个事件。

当前代码结构（compose/model.go）：
1. 第 258 行：`case tea.KeyPressMsg:` — 正确使用 v2 API
2. 第 322 行：`if m.contentInput.Focused()` 且 `key.Matches(msg, m.keys.send)` — **但是**，此时 `textarea.Update(msg)` 还未被调用
3. 第 409 行：`m.contentInput.Update(msg)` — textarea 在 key handling 之后才调用

### 关键问题

`ctrl+enter` 在 `contentInput.Focused()` 时应该在 `textarea.Update()` **之前**被检查并处理。当前代码中 key handling 确实在 `textarea.Update()` 之前，所以理论上不应该有问题。

但实际测试发现 `ctrl+enter` 无法触发 send。需要检查：
- `key.Matches()` 在 v2 中对 `ctrl+enter` 的处理是否正确
- 是否需要使用 `msg.String() == "ctrl+enter"` 代替 `key.Matches()`

### v2 Ctrl+Key Matching

从 UPGRADE_GUIDE_V2.md：
```go
// Option A — string matching:
case tea.KeyPressMsg:
    switch msg.String() {
    case "ctrl+c":
        // ctrl+c
    }

// Option B — field matching:
case tea.KeyPressMsg:
    if msg.Code == 'c' && msg.Mod == tea.ModCtrl {
        // ctrl+c
    }
```

## Open Questions

1. `key.Matches(msg, m.keys.send)` 是否在 v2 中对 `ctrl+enter` 正常工作？
2. 是否需要改用 `msg.String() == "ctrl+enter"` 来匹配？

## Requirements

### 修复 ctrl+enter 发送

当 `contentInput` focused 时，`ctrl+enter` 必须触发 send：
- 在 `textarea.Update()` 之前检查 `ctrl+enter`
- 如果 `key.Matches()` 不工作，尝试 `msg.String() == "ctrl+enter"`

### 验证方法

- `key.WithKeys("ctrl+enter")` 绑定的 keys 格式是否正确
- 测试 standalone 和 wm 模式两种场景

## Acceptance Criteria

- [ ] `ctrl+enter` 在 contentInput focused 时触发 send
- [ ] 修复后 lint/typecheck 通过

## Technical Notes

参考：`tui/compose/model.go` 第 258-328 行（key handling）和第 409 行（textarea.Update）