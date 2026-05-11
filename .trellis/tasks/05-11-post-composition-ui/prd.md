# brainstorm: 帖子编写界面

## Goal

实现一个通用的 **Compose 组件/页面**，可在多个场景复用：
- 从 event 详情页触发 reply / quote
- 从零开始 compose note
- 从零开始 compose community post
- 支持不同 kind 的 event 构建

## What I already know

**现有实现：**
- `note post <content>` — 单行 CLI 发帖
- `note reply <note-id> <content>` — 单行 CLI 回复
- `tui/timeline/` — Timeline TUI（list）
- `tui/dm/` — DM TUI（viewport + textarea）
- `tui/window/event/` — Event 详情 TUI，已有 `reply` / `quote` key binding

**发帖逻辑在 `utils/post.go`：**
- `PostNote(ctx, app, content)` — Kind 1
- `ReplyToNote(ctx, app, parentID, content)` — Kind 1 with e/p tags
- `QuoteNote(ctx, app, quotedID, content)` — Kind 1 with q tag

**关键发现：**
- Event 详情页已有 `reply` / `quote` key binding，但无 TUI compose 实现
- DM TUI (`tui/dm/model.go`) 有完整的 textarea + send 模式可参考

## Open Questions

1. **compose 模式**：需要支持哪些 compose 类型？
   - note compose（Kind 1，纯文本）
   - community compose（Kind 1111，带 community tag）
   - generic event compose（其他 kind）

2. **调用方式**：如何从不同入口触发 compose？
   - Event 详情页按 `r` 打开 reply compose
   - Event 详情页按 `q` 打开 quote compose
   - CLI 命令 `note compose` / `community compose` 打开独立页面

3. **组件复用**：compose 是否需要支持 timeline 内联使用？

## Requirements (evolving)

- [ ] 支持 note compose（Kind 1）
- [ ] 支持 reply（带 e/p tags）
- [ ] 支持 quote（带 q tag）
- [ ] 支持 community compose（Kind 1111）
- [ ] Event 详情页可触发 compose
- [ ] CLI 命令可触发 compose

## Acceptance Criteria (evolving)

- [ ] `note compose` 打开 TUI compose 页面
- [ ] Event 详情页按 `r` 可 reply
- [ ] Event 详情页按 `q` 可 quote
- [ ] 发送后 timeline 刷新显示新帖

## Definition of Done

* Compose TUI 可正常发送帖子
* Event 详情页 reply/quote 可用
* go test ./... 通过

## Out of Scope

* 多媒体附件
* 回复嵌套
* 草稿保存

## Technical Notes

* 使用 Bubble Tea textarea + viewport
* 参考 `tui/dm/model.go` 的 textarea + send 模式
* Compose 作为独立 Window 可被 timeline/dm/event 等多入口调用