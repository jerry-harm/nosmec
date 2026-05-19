# Fork nostr SDK into nosmec and Extend

## Goal

Fork `fiatjaf.com/nostr/sdk/` into `nosmec/nostr_sdk/`, merge sdkplus into it, and open up internal APIs (replaceable/addressable dataloaders) so new event kinds (34550, community feeds) can be registered as first-class citizens — not as second-class wrappers.

## What Gets Forked vs Kept

### Forked into `nostr_sdk/`

```
fiatjaf.com/nostr/sdk/*.go          →  nosmec/nostr_sdk/*.go
fiatjaf.com/nostr/sdk/cache/        →  nosmec/nostr_sdk/cache/
fiatjaf.com/nostr/sdk/dataloader/   →  nosmec/nostr_sdk/dataloader/
fiatjaf.com/nostr/sdk/hints/        →  nosmec/nostr_sdk/hints/
fiatjaf.com/nostr/sdk/kvstore/      →  nosmec/nostr_sdk/kvstore/
```

### Stays as external dependency

```
fiatjaf.com/nostr                   →  core types: Event, Filter, Pool, PubKey, Kind...
fiatjaf.com/nostr/eventstore/       →  Store interface + BoltDB + Bleve + wrappers
fiatjaf.com/nostr/nip*/             →  NIP helpers
```

### Goes away

```
nosmec/sdkplus/                     →  函数合并进 nostr_sdk/
```

## Modifications to Forked Code

1. **`replaceable_loader.go`**:
   - `replaceableIndex` → 改为 map-based 而非 fixed const enum
   - `replaceableLoaders` → 首字母大写，外部可注册
   - `initializeReplaceableDataloaders()` → 导出，加 kind 34550
   - `createReplaceableDataloader()` → 导出

2. **`addressable_loader.go`**:
   - 同上，加 kind 34550（NIP-72 community definition 是 parameterized replaceable）
   - 实际 kind 34550 的 `d` tag 使其类似 addressable，但 SDK 的分类是 replaceable

3. **新增 community feed 方法**（来自原 sdkplus）:
   - `FetchCommunityTimelinePage(ctx, communityAddrs, limit, until)` — 本地优先，tag-based
   - `FetchCommunityDefinitions(ctx)` — FetchProfileMetadata 模式（cache + store + TTL）

4. **`FetchGlobalTimelinePage`** — 保留，加 local-first

5. **`FetchRepliesToRoot`** — 加本地优先

6. **`FetchMyTimelinePage`** — 后续改为调用 `FetchFeedPage`

## Key Decisions

| 决策 | 理由 |
|------|------|
| 不 fork eventstore | `fiatjaf.com/nostr/eventstore` 已暴露 Store 接口，BoltDB/Bleve 后端直接用 |
| 合并 sdkplus 入 nostr_sdk | 不再需要 wrapper 层，自定义函数直接是 SDK 方法 |
| namespace: `nostr_sdk` | 避免与 Go stdlib `sdk` 冲突 |
| kind 34550 放 addressable loader | 它用 `d` tag，语义上是 parameterized replaceable（addressable），与 30000 系列一致 |

## Acceptance Criteria

- [ ] SDK fork 编译通过
- [ ] `go build ./...` 全项目通过
- [ ] `go test ./...` 通过
- [ ] `replaceableLoaders` / `addressableLoaders` 可外部注册
- [ ] `FetchCommunityDefinitions` 遵循 `FetchProfileMetadata` 模式
- [ ] `FetchGlobalTimelinePage` 本地优先
- [ ] `FetchCommunityTimelinePage` 本地优先
- [ ] 原 `sdkplus/system.go` 函数完全迁移

## Out of Scope

- Fork eventstore / core nostr types
- 重构 TUI 中直接调 Pool 的地方（留在后续 task）