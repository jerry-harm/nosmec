# 补全 event 详情界面功能，统一所有入口

## Goal

修复 CLI 入口 (`nosmec event <id>`) 的 bubblon controller 问题，使 reply/quote/thread 在 CLI 模式下可用。添加 delete 反馈。所有入口共用同一套 EventView 代码。

## What I already know

- 3 个入口点都使用同一个 `EventView` struct，代码无重复
- CLI 入口调用 `NewFromID(eventID, app, 80, 24, nil)` → ctrl=nil → reply/quote/thread 静默失效
- Timeline 入口调用 `New(&ev, ..., &m.ctrl)` → ctrl 存在 → 全部功能正常
- Thread 入口调用 `New(&ev, ..., m.ctrl)` → ctrl 存在 → 全部功能正常
- Delete 发送 Kind 5 后返回 nil，UI 无任何反馈
- Nostr 协议无 "restore" 标准（Kind 5 是建议性删除，无反向操作）

## Requirements

* [ ] CLI 入口：EventView 获得 bubblon controller → reply/quote/thread 可用
* [ ] Delete 反馈：发送成功后显示确认信息或自动关闭当前窗口
* [ ] 确保所有入口功能一致（同一套 keybinding，同一套 action）

## Acceptance Criteria

* [ ] `nosmec event <id>` 中按 r/q/t 能正常打开 compose/thread 界面
* [ ] Delete 后用户能看到反馈（成功提示或自动返回上一层）
* [ ] `go build ./...` + `go test ./...` 通过
* [ ] PTY 测试：CLI event 命令中按键功能验证

## Decision (ADR-lite)

**Context**: CLI 入口和 delete 行为需要统一。
**Decision**: 
1. 加 `SetController` 方法注入 bubblon controller，CLI 入口在 `bubblon.New()` 后调用
2. 按 `d` 时弹确认提示，按 `y` 执行删除 + 发布 Kind 5 事件 + 关闭详情界面
**Consequences**: 所有入口功能完全一致，删除有确认保护

## Out of Scope

* Restore/undelete（Nostr 协议不支持）
* 新增功能（现有 8 个 keybinding 已覆盖）

## Technical Notes

- 关键文件: `cmd/event_commands.go:55`, `tui/window/event/event.go:81/103`
- 问题点: `RunEventDetail` 创建 EventView 时 ctrl=nil，但立刻又创建 bubblon.New(m)
- 修复方案: EventView 加 SetController 方法，bubblon.New 后注入
