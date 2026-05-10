# unify-query-timeout

## Goal

统一所有查询的超时配置,移除硬编码超时值,改用可配置的统一超时。

## What I already know

- GetEvent 已经在 Task 1 中改为统一 5s 超时
- profile.go 有 10s 超时
- subscription.go 有 5s 超时
- GetEventAsync 目前用 goroutine + channel, 没有超时概念
- SubscribeMany 是 channel-based API,不支持直接超时

## Requirements

1. 添加 `query.timeout` 配置项 (默认 5s)
2. GetEventAsync 改用 context.WithTimeout
3. 所有硬编码超时值改为使用配置的 query.timeout
4. SubscribeMany 暂时保持原样 (channel-based, 需要额外处理)

## Technical Approach

### 配置变更

```go
// config/types.go
type Config struct {
    ...
    Query struct {
        Timeout time.Duration `mapstructure:"timeout"`
    }
}
```

### GetEventAsync 改造

当前使用 goroutine + channel:
```go
found := make(chan struct{})
go func() {
    result := opts.App.Pool().QuerySingle(ctx, relays, filter, ...)
    close(found)
}()
select { case <-found: ... case <-ctx.Done(): ... }
```

改为 context.WithTimeout:
```go
ctxTimeout, cancel := context.WithTimeout(ctx, queryTimeout)
defer cancel()
result := opts.App.Pool().QuerySingle(ctxTimeout, relays, filter, ...)
```

## Acceptance Criteria

- [ ] Build passes
- [ ] GetEventAsync uses context.WithTimeout
- [ ] No hardcoded 2s/5s/10s timeouts in query paths
- [ ] Timeout configurable via config

## Definition of Done

- `go build ./...` passes
- `go vet ./...` passes
- All query functions use configurable timeout

## Out of Scope

- SubscribeMany 超时处理 (复杂,需要单独研究)

## Technical Notes

- GetEvent 已经用 5s timeout (Task 1)
- profile.go: FetchFullProfile 和 FetchRecipientDMRelays 用 10s
- subscription.go: 用 5s