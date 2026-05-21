# note timeline --global 等 relay 选择过于局限

## Goal

`FetchGlobalTimelinePage` 只用 `FallbackRelays`，不够。改为使用 `defaultRelaysForFilter`（综合 FallbackRelays + kind 路由），让 global 和社区发现使用所有已知 relay 拉取。

## Diagnostic Results

测试此机器上：
- `Pool.FetchMany` → damus.io, nos.lol, offchain.pub: 全部 **0 results**（TCP timeout，无网络）
- `FetchEventsByFilter` kind:34550: **0 results**
- `FetchGlobalTimelinePage` kind:1: **0 results**
- `FetchMyTimelinePage` 能用是因为它先读 **本地 LMDB store**（有之前缓存的 event）

结论：代码层面 relay 选择过于局限 + 当前机器无网络。修复 relay 选择逻辑后，在有网络的机器上应该能正常工作。

## Root Cause

`FetchGlobalTimelinePage` (system.go:225) 只用 `FallbackRelays`：

```go
relays := sys.FallbackRelays.URLs
if len(relays) == 0 {
    relays = []string{"wss://relay.damus.io", "wss://nos.lol"}
}
```

对比 `FetchEventsByFilter` (system.go:532) 使用 `defaultRelaysForFilter`，后者综合了：
- IDs 查询 → JustIDRelays + FallbackRelays
- 单 author 查询 → FetchOutboxRelays + FallbackRelays  
- kind:34550 → RelayListRelays + FallbackRelays
- 默认 → FallbackRelays

## Requirements

* [ ] `FetchGlobalTimelinePage` 改为使用 `defaultRelaysForFilter`（或等效的综合 relay 选择），不是只用 `FallbackRelays`
* [ ] 检查 `FetchFollowedTimelinePage` 同样是否需要修复
* [ ] 确保 `community discover` 的 `FetchEventsByFilter` 路径 relay 选择正确（已验证正确）

## Technical Approach

**Approach A (推荐)**：让 `FetchGlobalTimelinePage` 内部调用 `defaultRelaysForFilter` 来计算 relay 集合，然后用这些 relay 查询：

```go
relays := sys.defaultRelaysForFilter(ctx, filter)
if len(relays) == 0 {
    relays = []string{"wss://relay.damus.io", "wss://nos.lol"}
}
```

这样 global timeline 的 kind:1 filter 会走到 default 分支（`FallbackRelays`），但未来如果有 kind 路由需求也能自动生效。

**Approach B**：完全重构，让 `FetchGlobalTimelinePage` 调用 `FetchEventsByFilter`。改动更大。

## Out of Scope

* 不添加 `--kind` flag（那是独立功能）
* 不修改 timeline 订阅的 hardcoded kinds
* 不修复本机器网络连接问题

## Definition of Done

* `FetchGlobalTimelinePage` 使用 `defaultRelaysForFilter` 获取 relay
* `FallbackRelays` 仍然作为 fallback
* Get + vet + tests pass

## Files to Modify

* `nostr_sdk/system.go:212-243` — `FetchGlobalTimelinePage` relay 选择逻辑
