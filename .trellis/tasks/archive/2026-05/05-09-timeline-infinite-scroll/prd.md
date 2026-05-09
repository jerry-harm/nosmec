# brainstorm: timeline infinite scroll

## Goal

将 timeline 从"一次性加载固定数量"改为"无限滚动"：
- 初始加载后启动 Nostr 订阅，持续监听新事件并插入列表头部
- 滚动到倒数第二页时自动向后加载旧事件，追加到列表末尾

## What I already know

* `tui/timeline/model.go` 使用 `charm.land/bubbles/v2/list.Model`
* `utils/get.go` 中 `GetGlobalTimeline`、`GetMyTimeline`、`GetFollowedTimeline` 均用 `limit int` 参数，无分页游标
* 底层用 `nostr.Filter{Limit: limit}` 查询，排序规则 `CreatedAt DESC`
* `QueryEventsCached` 从缓存池批量拉取事件
* `pool.Sub()` 可以建立持久订阅，实时接收新事件

## Architecture

### 初始刷新后双通道

1. **订阅通道**（实时）
   - 初始加载完成后，用 `pool.Sub()` 建立持久订阅
   - Filter: `Since: newestEvent.CreatedAt`（只订阅比最新已知事件更新的事件）
   - 新事件通过 channel 到来 → 去重 → `list.InsertItem(0, ...)` 插入列表开头

2. **查询通道**（按需）
   - 滚动到倒数第二页时触发
   - 调用 `QueryEventsCached(until=oldestEvent.CreatedAt)` 向后分页
   - 去重 → `list.AppendItem()` 追加到列表末尾

### 触发条件

* **底部加载**：`!isLoadingMore && hasMoreOld && !OnLastPage() && Page >= TotalPages-2`
* **订阅启动**：初始 fetchMsg 完成且 hasMoreOld=true 时启动
* **尽头判断**：底部加载返回事件数 < pageSize → hasMoreOld=false

## Requirements

* 初始加载完成后立即建立订阅，持续监听新事件
* 新事件到达时自动插入列表头部（去重）
* 滚动到倒数第二页时自动向后加载旧事件
* 加载过程有视觉反馈（spinner）
* 错误时显示友好的错误提示
* 相同 ID 的事件不重复显示

## Acceptance Criteria

* [ ] 初始加载完成后订阅启动，新事件自动出现在列表开头
* [ ] 滚动到底部自动加载旧事件追加到末尾
* [ ] 加载过程有视觉反馈（spinner）
* [ ] 错误时显示友好的错误提示
* [ ] 已有事件不因订阅推送而重复

## Out of Scope (explicit)

* 订阅断开重连（暂时不考虑）
* 用户主动下拉刷新（订阅已覆盖）

## Technical Notes

### 订阅实现
```go
// 初始加载完成后启动订阅
subCh := pool.Sub(ctx, relays, nostr.Filter{
    Kinds: []nostr.Kind{nostr.KindTextNote},
    Since: newestEvent.CreatedAt,  // 只订阅更新的事件
})
// 订阅事件通过 channel 到来
go func() {
    for relayEvent := range subCh {
        // 去重后插入列表头部
        if !seenEventIDs[relayEvent.Event.ID] {
            prependEvent(relayEvent.Event)
        }
    }
}()
```

### 分页加载
```go
// 滚动到倒数第二页时
filter := nostr.Filter{
    Kinds: []nostr.Kind{nostr.KindTextNote},
    Limit: pageSize,
    UNTIL: oldestEvent.CreatedAt - 1,  // 向后分页
}
events := QueryEventsCached(..., filter, ...)
// 去重后追加
```

### Model 新增字段
* `isLoadingMore bool` — 防止重复加载
* `hasMoreOld bool` — false 时停止向后加载
* `seenEventIDs map[nostr.ID]bool` — 去重
* `subscriptionCh chan nostr.Event` — 订阅通道

### Files to modify
* `utils/get.go` — `GetGlobalTimeline`/`GetMyTimeline`/`GetFollowedTimeline` 加 `until` 参数
* `tui/timeline/model.go` — 订阅启动 + 订阅事件处理 + 底部加载逻辑
