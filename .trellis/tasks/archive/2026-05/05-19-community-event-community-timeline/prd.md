# community 发现功能 - 进入 event 详情和 community timeline

## Goal

在 community discover 列表界面中，`Ctrl+E` 打开选中 community 的底层 Nostr event（kind:34550 的 raw event 详情），`Enter` 打开该 community 的 timeline（查看社区内帖子）。

## What I already know

* Community discover 是 **standalone program**（`tea.NewProgram(m).Run()`），没有 bubblon controller，无法切换 view
* Timeline 使用 bubblon stack 模式：`bubblon.New(tlModel)` 包装，用 `bubblon.Open(ev)` 推入 EventView
* EventView (`tui/event/event.go`) 需要 `*nostr.Event`、`*config.AppContext`、width/height、authorName、`*bubblon.Controller`
* Community Timeline 已存在：`timeline.RunTimeline(app, "community", nil, limit, communityAddr)`，communityAddr 格式为 `34550:<moderator_pubkey>:<d_tag>`
* Enter 当前在 community discover 中已绑定但 `openCommunity` message 无 handler（被忽略）
* CommunityItem 只存 `def utils.CommunityDefinition`（解析后的字段），不存 raw event

## Requirements

* `Ctrl+E` → 打开选中 community 的底层 Nostr event 详情 view（`tui/event/EventView`）
* `Enter` → 打开选中 community 的 community timeline view
* Esc/q → 从 event 详情 / timeline 返回到 community discover 列表
* Ctrl+C → 退出整个程序

## Technical Approach

1. **将 community discover 转换为 bubblon 架构**（`RunCommunityDiscover` 用 `bubblon.New` 包装，类似 `RunTimeline`）
2. **communityItem 增加 raw event 字段**：`communityItem` 加 `event *nostr.Event`，供 `Ctrl+E` 创建 EventView 使用
3. **Ctrl+E**：`event.New(item.event, ...)` → `bubblon.Open(ev)`
4. **Enter**：构建 communityAddr（已有逻辑）→ `NewModel(app, "community", nil, limit, addr)` → `bubblon.Open(tlModel)`

## Open Questions

* 无

## Acceptance Criteria

* [ ] `Ctrl+E` 在社区列表按中打开选中 community 的 EventView
* [ ] `Enter` 在社区列表按中打开选中 community 的 community timeline
* [ ] Esc 从 EventView / timeline 返回到 community discover 列表
* [ ] Ctrl+C 从任意层退出
* [ ] `go build ./...` 通过
* [ ] `go vet ./...` 通过

## Out of Scope

* 修改 EventView 的 keybinding
* 修改 Community Timeline 的 filter 逻辑
* community discover 以外的任何改动

## Decision (ADR-lite)

**Context**: Community discover 需要支持内联导航到 EventView 和 community timeline
**Decision**: 将 discover model 从 standalone 改为 bubblon stack 架构；communityItem 增加 `event *nostr.Event` 字段以支持 Ctrl+E 创建 EventView；CommunityDefinition 增加 RawEvent 字段
**Consequences**: discover model 需要持有 `*bubblon.Controller`；`utils/community.go` 的 FetchCommunityEvents 需要同时返回 raw events

## Implementation Plan

1. `utils/community.go`: `CommunityDefinition` 加 `Event *nostr.Event` 字段，`FetchCommunityEvents` 填充它
2. `tui/community/discover/model.go`:
   - model struct 加 `ctrl *bubblon.Controller`
   - keyMap 加 `eventDetail`（Ctrl+E）、`open`（Enter）
   - Update() 处理 `Ctrl+E` → `bubblon.Open(EventView)`、`Enter` → `bubblon.Open(timeline model)`
   - `RunCommunityDiscover` 用 `bubblon.New(m)` 包装
   - 删除废 `openCommunity` message
3. 编译验证：`go build ./... && go vet ./...`

## Technical Notes

* `tui/community/discover/model.go` — 需改 model struct、Update()、RunCommunityDiscover()
* `utils/community.go:FetchCommunityEvents` — `CommunityDefinition` 加 `Event` 字段
* `tui/timeline/main.go:RunTimeline` — 参考 bubblon 包装方式
* `tui/event/event.go:New` — EventView 构造函数签名
