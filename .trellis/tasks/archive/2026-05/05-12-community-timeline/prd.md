# 添加community timeline

## Goal

在 timeline 的 filter 选项中添加 "community"，专门显示所有订阅社区的 posts（kind 1111），复用一个已有 note timeline 的代码。

## What I already know

* timeline filter 当前有 `global` / `mine` / `default`（followed = users + communities + hashtags 混查）
* `GetFollowedTimeline` 里 users + communities + hashtags 是混合查询的
* `GetCommunityPosts` 支持查询单个社区的 posts（authorPubKey + communityID）
* `app.ListSubscriptions("community")` 返回所有订阅的社区地址
* `TimelineEvent` 结构体有 `CommunityID string` 字段
* timeline 里有 `kindCommunity` 类型检测（kind 1111）

## Requirements

* timeline 添加 filter 选项 `"community"`
* 当 filter 为 `"community"` 时，只查询所有订阅社区的 posts（不混 user posts）
* 调用 `GetCommunityPosts` 获取数据（需改造以支持多个社区）
* 复用 timeline 的现有代码（model、delegate、styles）
* 渲染时显示为 `[Community]` 前缀（复用现有 `kindCommunity` 渲染逻辑）

## Acceptance Criteria

* [ ] `community` filter 选项存在并能切换
* [ ] 显示所有订阅社区的 posts（kind 1111）
* [ ] 每个 post 显示 `[Community]` 前缀
* [ ] 点击 post 能打开 event detail
* [ ] 无限滚动、分页等功能与 global/mine 一致
* [ ] lint / typecheck / tests pass

## Out of Scope

* 社区创建/管理功能
* 社区详情页
* 社区之外的其他 filter 改动

## Technical Approach

### 新增 utils 函数

改造 `GetCommunityPosts` 支持多个社区，或新增 `GetCommunitiesTimeline`：

```go
func GetCommunitiesTimeline(ctx context.Context, communityAddrs []string, limit int, until nostr.Timestamp, opts *GetOptions) chan *nostr.Event {
    // 1. 遍历 communityAddrs，逐个调用 GetCommunityPosts
    // 2. 合并到单个 channel
    // 3. 按时间排序（类似 GetFollowedTimeline）
}
```

### timeline/model.go 改动

```go
switch m.filter {
case "global":
    ch = utils.GetGlobalTimeline(...)
case "mine":
    ch = utils.GetMyTimeline(...)
case "community":
    // 获取所有订阅社区地址，传给新的 GetCommunitiesTimeline
    subs := m.app.ListSubscriptions("community")
    var addrs []string
    for _, sub := range subs {
        addrs = append(addrs, sub.ID)
    }
    ch = utils.GetCommunitiesTimeline(ctx, addrs, m.limit, 0, opts)
default:
    ch = utils.GetFollowedTimeline(...)
}
```

### 新增 showCommunityTimeline 命令

类似现有的 `showGlobalTimeline` / `showMyTimeline`，添加到 TUI 入口。