# 本地 relay 重构：统一查询接口 + 异步输出

## Goal

重构 `utils/get.go` 中的查询函数，使用 `iter.Seq[Event]` 异步流式输出，保证 UI 能实时接收事件。

## What I already know

- `utils/get.go` 中有两个主要查询函数：`GetEvent` 和 `QueryEventsCached`
- `GetEvent` 返回 `*nostr.Event`（单事件），无需改签名
- `QueryEventsCached` 当前返回 `([]nostr.Event, error)`，同步阻塞
- `FetchMany` 返回 `chan RelayEvent`，事件异步到达，但函数等所有 relay EOSE 才退出
- `Relay.QueryEvents(filter)` 返回 `iter.Seq[Event]`，可直接用于流式

## Decisions

### 返回类型：`iter.Seq[Event]`

- 批量查询函数改返回 `iter.Seq[Event]`，不 buffer，直接 yield 给调用方
- 调用方（TUI / CLI）用 `for range` 收事件，本地事件先到先处理
- 不对本地 relay 做特殊封装，统一走 `FetchMany`

### 情景处理

| 情景 | 处理方式 |
|---|---|
| 单 event 查（不可变） | `QuerySingle` 本地优先，命中即返回 |
| 单 event 查（可替换） | 用 `FetchManyReplaceable`，不做特殊处理 |
| 批量查（时间线） | 返回 `iter.Seq[Event]`，TUI 用 `for range` 消费 |

## Technical Approach

### `QueryEventsCached` 改为返回 `iter.Seq[Event]`

```go
func QueryEventsCached(...) iter.Seq[nostr.Event] {
    return func(yield func(nostr.Event) bool) {
        pool.FetchMany(ctx, relays, filter, opts)
        // 直接 yield，不 buffer
    }
}
```

### `GetFollowedTimeline` 改为返回 `iter.Seq[TimelineEvent]`

内部 filter 逻辑不变，改为流式 yield，每收到一个 event 做 hashtag 过滤和类型转换。

### `GetMyTimeline` / `GetGlobalTimeline`

保持返回 `iter.Seq[Event]`，TUI 用 `for range` 消费。

## Out of Scope

- 流式订阅（SubscribeMany / dm 等），不改动
- 内存缓存层，后续任务
- 本地 relay 对外暴露，后续任务

## Implementation Plan

1. `QueryEventsCached` → 返回 `iter.Seq[Event]`
2. `GetFollowedTimeline` → 返回 `iter.Seq[TimelineEvent]`
3. `GetMyTimeline` → 返回 `iter.Seq[Event]`
4. `GetGlobalTimeline` → 返回 `iter.Seq[Event]`
5. 调用方（TUI）适配新签名
