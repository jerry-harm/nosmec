# community timeline 复用: FetchFollowedTimelinePage relay 参数

## Goal

修复从 community discover 进入 timeline 时 community timeline 为空的问题。根本原因：调用 `FetchFollowedTimelinePage(ctx, nil, []string{addr}, limit, 0)` 时 `pubkeys == nil`，但 community posts 不需要 outbox relay 推送，它们通过 `a` tag 过滤直接查询 relay。

## What I already know

* `sdkplus/system.go:FetchFollowedTimelinePage` 的 community 分支（第 250-273 行）：当 `until=0` 时，filter 没有 `until`/`since` 限制，会获取所有历史 community posts
* 但 community 分支调用 `sys.System.FetchOutboxRelays(ctx, nostr.PubKey{}, 2)` — 这对 community 地址不起作用，因为它用空 pubkey 查 outbox relay
* 当 `pubkeys` 和 `communityAddrs` 都非空时，community 分支的 filter 实际上没有 `Limit` 限制导致可能只返回 0 个
* 从 `cmd/community_commands.go` 进的 timeline：用 `community.RunTimeline(app, "community", nil, limit, communityAddr)` — 它也走 `FetchFollowedTimelinePage`，但因为 `cmd/community_commands.go` 不走 discover model，它可能用的是另一条数据路径（或者也用相同 SDK call）

## Root Cause

`FetchFollowedTimelinePage` 中 community 分支在 `until=0` 时：
1. Filter 有 `Tags: nostr.TagMap{"a": []string{communityAddr}}` ✓
2. Filter 有 `Kinds: [TextNote, Comment]` ✓
3. 但 `relays := sys.System.FetchOutboxRelays(ctx, nostr.PubKey{}, 2)` — 空 pubkey 查 outbox relay 返回什么？

community 地址不是 pubkey，所以 outbox relay 逻辑不适用。应该用 `CommunityDefinition.Relays`（community 定义里带的 relay 列表）。

## Decision

community 分支在 `until=0` 时，filter 没问题，但 relay 选择逻辑错误。改为：当 communityAddr 传入时，用 `CommunityDefinition.Relays` 作为查询 relay。如果 Relays 为空，才 fallback 到默认 relay。

## Implementation

这个修复需要改动 `sdkplus/system.go`。但这是一个 shared SDK，修改可能有副作用。更好的方案：让 timeline model 在 fetch 时传入正确的 relays。

实际上，最干净的方案是：在 `FetchFollowedTimelinePage` 的 community 分支，当 `relays` 为空时，查询事件带上 `#a` 标签过滤，relay 列表用系统默认可读 relay（`app.AllReadableRelays()` 或类似）。

## Acceptance Criteria

* [ ] 从 community discover Enter 进的 timeline 显示 community posts
* [ ] 直接 `nosmec community timeline <addr>` 也能显示 posts
* [ ] `go build ./...` 通过
* [ ] `go vet ./...` 通过

## Out of Scope

* 不修改 community discover model 的导航逻辑
* 不修改 timeline model 的 Init/fetchTimeline 逻辑

## Technical Notes

* `sdkplus/system.go:250-274` — community 分支，relay 选择逻辑
* `tui/timeline/model.go:285-290` — community filter fetchTimeline 调用点
* `utils/community.go:CommunityDefinition.Relays` — community 自带的 relay 列表