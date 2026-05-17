# community 发现功能 - filter 抓取 kind:34550 元信息并列表显示

## Goal

替代当前 community list 的复杂流程（follow → list → timeline），改为直接从 relay 抓取 kind 34550 元信息事件，用和 timeline 一样的列表组件展示。

## What I already know

- Kind 34550 是 addressable event（30000-40000 范围），用 `d` tag 做 identifier
- NIP-72 没有标准 community 发现机制 — 只需要 `{"kinds": [34550]}` 查询所有 relay
- 当前 community 流程：CreateCommunity → FollowCommunity → GetCommunityPosts → timeline 显示
- DM list 模型 (`tui/dm/list/model.go`) 是理想复用模板：独立 tea 程序，异步加载，Enter 选中，esc 退出
- CommunityDefinition 含 `name`, `description`, `image`, `relay` tags + `p` moderator tags

## Requirements

* [ ] `nosmec community discover` 命令：抓取所有 kind 34550 事件并列表显示
* [ ] 复用 timeline 的列表组件（delegate/keymap/navigation）
* [ ] 显示 community 名称、描述、moderator 数量
* [ ] Enter 进入某个 community 的帖子列表
* [ ] esc 或 q 退出
* [ ] 异步加载，显示 spinner

## Acceptance Criteria

* [ ] `nosmec community discover` 能在终端列出 relay 上的 community
* [ ] 列表用 ↑↓ 导航，Enter 选中
* [ ] `go build ./...` + `go test ./...` 通过

## Out of Scope

* Community 创建（已有）
* Community 帖子发布（已有）
* Community 关注/取消关注（已有）
* 分页/搜索（MVP 用一次 FetchMany 获取全部）

## Technical Approach

复用 `tui/timeline/delegate.go` 的列表 delegate 模式 + `tui/dm/list/` 的独立 tea 程序模式：

```
nosmec community discover
  → utils.FetchCommunityEvents(ctx, app) {kinds: [34550]}
  → communityListModel (tea.Model)
    ├─ Init: fetch async
    ├─ Update: ↑↓ nav, Enter select, esc quit
    └─ View: list with name, description
  → Enter → nosmec community view <addr>
```

## Technical Notes

- 研究: `research/community-current.md`
- 模板: `tui/dm/list/model.go` (独立 tea 程序), `tui/timeline/delegate.go` (列表 delegate)
- 关键: kind 34550 是 addressable，用 `FetchMany` 不是 `FetchManyReplaceable`
