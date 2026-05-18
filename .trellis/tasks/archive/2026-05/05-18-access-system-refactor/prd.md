# 重构访问系统 - 参考 nostr/sdk System 模块

## Goal

创建 `access` 包，借鉴 nostr/sdk System 的设计模式，构建自研的 access.System，嵌入 AppContext，提供统一的 relay 轮选策略和事件→relay 映射持久化。

## Requirements

- [x] 新建 `access/` 包，包含 `System` 结构体
- [x] `System` 嵌入 `AppContext`（渐进迁移，现有调用者改动最小）
- [x] 实现 `RelayStream` - relay 轮选策略（Read / Write / DM / Search / Fallback）
- [x] `KVStore` (BoltDB) 持久化 event→relay 映射（替代内存 map）
- [x] 关联 Pool、Hints、KVStore 到 System
- [x] 保留现有本地 relay 代码不变
- [ ] 现有调用方逐步迁移到 access.System API

## Acceptance Criteria

- [ ] `access.RelayStream.Next()` 正确轮转返回 relay URL
- [ ] `access.System.TrackEventRelay(id, relay)` 写入 KVStore，重启后不丢失
- [ ] `access.System.GetEventRelays(id)` 从 KVStore 读取
- [ ] `access.System` 提供 ReadableRelays / WritableRelays / DMRelays / SearchRelays 的 RelayStream
- [ ] AppContext.Pool() / ReadableRelays() 等接口保持兼容
- [ ] 现有所有功能不受影响
- [ ] Lint / type-check / 测试通过

## Definition of Done

- Tests added/updated (unit/integration)
- Lint / typecheck / CI green
- Docs/notes updated if behavior changes

## Technical Approach

```
access/
├── system.go      -- System struct (Pool, Hints, KVStore, RelayStreams)
├── relay_stream.go -- RelayStream 实现（轮转 relay URL）
├── kvstore.go      -- KVStore 初始化 + event/relay 映射方法
└── system_test.go
```

## Decision (ADR-lite)

**Context**: 需要重构访问层，使其更统一 + 持久化 event→relay 映射
**Decision**: A - 自建 `access.System` 结构嵌入 AppContext，借鉴 sdk.System 的 RelayStream + KVStore 模式
**Consequences**: 
- 完全控制代码走向，不依赖 sdk.System 的内部实现
- AppContext 现有 API 保持兼容，降低迁移风险
- KVStore 用 BoltDB（和 HintDB 相同的存储引擎）

## Out of Scope

- 完整 NIP-51 列表系统（关注/静音/书签等）-- 框架建立后可后续添加
- 多用户隔离
- 缓存层（ProfileMetadata/RelayList 缓存）-- 后续增量
- 本地 relay 改造

## Technical Notes

- `fiatjaf.com/nostr/sdk/kvstore` -- KVStore 接口
- `fiatjaf.com/nostr/sdk/kvstore/bbolt` -- BoltDB 实现
- `fiatjaf.com/nostr/sdk/hints` -- 已有，pubkey→relay hints
- 受影响文件：`config/config.go`, `config/context.go`, `config/types.go`, `utils/relay_list.go`, `utils/dm.go`, `utils/search.go`, `utils/profile.go`

## Implementation Plan

1. 创建 `access/` 包 + `RelayStream` + `System` 骨架
2. 实现 KVStore event→relay 持久化（替代内存 map）
3. 将 System 嵌入 AppContext，委托 Relay/Pool 调用
4. 更新调用方使用新 API
5. 运行测试 + lint 验证
