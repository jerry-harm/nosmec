# simplify-relay-architecture

## Goal

简化 relay 查询架构,删除 PrivateRelays 概念,统一 relay pool 查询,缓存只发到 local relay。

## What I already know

- `PrivateRelays()` in config/context.go:101, returns local + cfg.PrivateRelays
- PrivateRelays used in 28 places: utils/get.go, dm.go, community.go, timeline/model.go
- `Config.PrivateRelays` in config/types.go:94
- Local relay for backup/cache, DM relay for inbox
- All queries should use AllReadableRelays() (includes local already)
- Cache should only publish to local relay

## Requirements

1. 删除 `PrivateRelays()` 方法
2. 删除 `Config.PrivateRelays` 配置字段
3. 删除所有对 `PrivateRelays()` 的调用
4. CacheEvent 只发到 local relay,不是 PrivateRelays
5. 所有查询统一使用 AllReadableRelays() (local + remote combined, no local-first split)

## Technical Approach

### Files to modify:

1. **config/types.go** - 删除 `PrivateRelays []string` 字段
2. **config/context.go** - 删除 `PrivateRelays()` 方法
3. **utils/get.go**:
   - GetEvent: 删除本地/远程分离查询,统一 relay pool
   - GetEventAsync: 同上
   - CacheEvent: 只发到 local relay
   - 所有调用 PrivateRelays() 的地方改掉
4. **utils/dm.go** - 删除 PrivateRelays 调用
5. **utils/community.go** - 删除 PrivateRelays 调用
6. **tui/timeline/model.go** - 删除 PrivateRelays 调用

### 查询逻辑统一:

```
relays := opts.App.AllReadableRelays()  // local + remote combined
// 不再分离本地/远程,统一查询
```

### Cache逻辑统一:

```
localURL := config.GetLocalRelayURL()
if localURL != "" {
    app.Pool().PublishMany(ctx, []string{localURL}, event)
}
```

## Acceptance Criteria

- [ ] Build passes
- [ ] No references to PrivateRelays remain
- [ ] All queries use unified relay pool
- [ ] Cache publishes to local relay only

## Definition of Done

- `go build ./...` passes
- `go vet ./...` passes
- No PrivateRelays in codebase

## Out of Scope

- DM relay 逻辑 (DMRelays 是 inbox, 不是 cache)
- 超时配置 (Task 2)

## Technical Notes

- Local relay URL: `ws://localhost:8989` (configurable via local_relay.port)
- DMRelays 用于 inbox 接收/发送,不在此任务范围