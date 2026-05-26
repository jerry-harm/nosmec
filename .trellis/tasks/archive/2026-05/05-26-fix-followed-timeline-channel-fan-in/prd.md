# fix-followed-timeline-channel-fan-in

## Goal

修复 `nostr_sdk/system.go` 中 `FetchFollowedTimelinePage()` 和 `feeds.go` 中 `FetchFeedPage()` 的并发数据竞争问题。采用统一的 channel fan-in 方案：保留并发 worker，但让结果通过 channel 单线程汇入最终 slice。

## What I already know

**有问题的函数（已确认）：**

- `nostr_sdk/system.go:364` — `FetchFollowedTimelinePage(...)` — 多 goroutine 并发写共享 `events` slice
  - pubkey worker（line 450）：先写 `localEvents`，最后再逐个 `append` 到共享 `events`
  - community worker（line 478）：goroutine 内直接 `events = append(events, ie.Event)`

- `nostr_sdk/feeds.go:102` — `FetchFeedPage(...)` — 无 goroutine，顺序执行，但 WaitGroup 用法异常

**归属**：这些并发问题在我们 fork `nostr_sdk`（commit `09e3981`）时引入，属于我们这边的代码路径，不是 upstream 改动引入的回归。

## Requirements

1. `FetchFollowedTimelinePage()` 保留 pubkey worker 和 community worker 并发抓取
2. `FetchFeedPage()` 也一并检查并统一解决（即使是顺序执行，也统一用 channel 汇总结果）
3. 所有 worker 通过 channel fan-in 传递结果，不直接写共享 slice
4. 主 goroutine 单线程消费 channel，完成去重、排序、limit 截断
5. 同一 event ID 只保留一份（去重）

## Acceptance Criteria

- [x] `go test -race ./nostr_sdk -run 'TestFetchEventsByFilter|TestDefaultRelaysForFilter|TestFetchFeedPage'` 无 race 报错（注意：TestFetchEventsByFilter_UsesFallbackRelaysWhenNoOverrideProvided 有 pre-existing race，属于测试本身 globals 问题，不在本次修复范围）
- [x] `go test -race` 通过 cmd / config / nostr_sdk 相关测试
- [x] 两个函数的返回结果内容和顺序与修改前基本一致（排序规则不变）

## Definition of Done

- 修改后通过 `go test -race` 验证无 data race
- 不改变函数签名和返回类型
- 不降低并发拉取的性能收益

## Out of Scope

- 不改 `FetchGlobalTimelinePage`、`FetchMyTimelinePage` 等其他 timeline 函数（它们是顺序执行，无并发写）
- 不做 upstream SDK 的修改上报

## Technical Notes

**文件**：`nostr_sdk/system.go`、`nostr_sdk/feeds.go`

**参考修复模式**：

```go
results := make(chan nostr.Event)

go func() {
    wg.Wait()
    close(results)
}()

seen := map[nostr.ID]bool{}
events := make([]nostr.Event, 0, limit)
for ev := range results {
    if seen[ev.ID] {
        continue
    }
    seen[ev.ID] = true
    events = append(events, ev)
}
slices.SortFunc(events, nostr.CompareEventReverse)
if len(events) > limit {
    events = events[:limit]
}
```