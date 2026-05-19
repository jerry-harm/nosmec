# community timeline 修复: 刷新功能+显示条目+显示完整地址

## Goal

修复 community timeline 的三个问题：
1. community timeline 能像 note timeline 一样用 `r` 刷新
2. community timeline 正确展示帖子条目（而非只显示数量）
3. community EventView 详情页顶部突出显示完整的 community 地址格式 `34550:hexpubkey:d`

## What I already know

* `tui/timeline/model.go` 是 note timeline 和 community timeline 共用的 model
* "r" refresh handler 在 model.go:850，已有逻辑：`m.fetchTimeline()` 重拉事件
* community timeline 的 `fetchTimeline()` (line 285-290) 调用 `ext.FetchFollowedTimelinePage(ctx, nil, []string{m.communityAddr}, m.limit, 0)`
* community posts 是 kind:1 和 kind:1111（community approve）；kind:34550 是 community 定义本身，不在 posts 里
* `startSubscription()` (line 539-550) 订阅 `{"a": [communityAddr]}` 过滤 kind:1 和 kind:1111
* `View()` 方法 (line 905-911) 渲染 `m.list.View()`，应该显示 items
* EventView (`tui/event/view.go`) 在 tags 区显示原始 tag，但不突出显示 community 地址格式
* EventView 头部 (view.go:19) 显示 `Kind: 34550`，tags 区 (view.go:59-63) 显示 `a` tag（但不是 `34550:hexpubkey:d` 格式）

## Requirements

* `r` 键在 community timeline 视图刷新列表
* community timeline 正确渲染帖子条目（author + content preview），不是只显示数量
* EventView 显示 kind:34550 时，顶部区域突出显示 community 地址：`34550:hexpubkey:d` 格式

## Technical Approach

1. **刷新功能**：确认 community timeline 有 `refresh` keyBinding（查看 listKeyMap 是否包含 refresh）。如果没有，在 keyMap 中加入 `refresh` 绑定。
2. **帖子渲染**：检查 `fetchMsg` handler 是否正确把 community posts 加到 list items。community posts 的 item kind 应为 `kindCommunity`，title 格式应为 `[Community] author: content preview`。
3. **Community 地址显示**：在 `tui/event/view.go` 的 `renderHeader()` 或新加 section，当 `e.Kind == 34550` 时，在 header 区域下方突出显示 community 地址（格式 `34550:hexpubkey:d`）

## Open Questions

* (无)

## Acceptance Criteria

* [ ] `r` 在 community timeline 中触发刷新
* [ ] community timeline 渲染帖子列表（author + content preview），不是只显示数量
* [ ] EventView 展示 kind:34550 时，顶部有明显的 community 地址行：`Address: 34550:hexpubkey:d`
* [ ] `go build ./...` 通过
* [ ] `go vet ./...` 通过

## Out of Scope

* 修改 community timeline 的订阅逻辑
* 修改 community posts 的过滤条件

## Technical Notes

* `tui/timeline/model.go` — listKeyMap.refresh, fetchMsg handler, View()
* `tui/timeline/model.go` — community filter branch fetchTimeline()
* `tui/event/view.go` — renderHeader(), 改 kind:34550 分支
* `utils/community.go:ParseCommunityAddr` — 可复用解析 communityAddr
