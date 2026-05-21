# Reorganize Relay Usage Across Fetch Functions

## Goal

重新组织 `nostr_sdk` 里不同 `Fetch*` 函数的 relay 使用方式，明确哪些函数应该本地优先、哪些函数应该由 SDK 内部统一选 relay、哪些函数把 caller 提供的 relay 当作强覆盖还是弱提示，减少策略漂移。

目标不是发明一套和 upstream 完全不同的新模型，而是**在 upstream `nostr/sdk` 的 hints/outbox/fallback 框架下，把 nosmec 自己新增的 fetch 函数也统一到同样的分层思路里**。

## What I already know

### 当前 relay 来源

`System` 目前有多类 relay 来源：

* `FallbackRelays` — 通用兜底
* `JustIDRelays` — ID 查询
* `RelayListRelays` — kind `34550` / kind `10002` 一类 list/indexer 查询
* `FetchOutboxRelays` — 单 author / timeline / feed 查询
* caller-provided relays — thread / scoped fetch / filter override
* pointer-embedded relays — `EventPointer` / `EntityPointer`

### 当前 fetch 行为并不一致

本地优先：

* `FetchMyTimelinePage`
* `FetchFollowedTimelinePage` 的 author 分支
* `FetchEventsByFilter`
* `FetchSpecificEvent`
* scoped thread 查询
* `FetchProfileMetadata`

直连网络或网络优先：

* `FetchGlobalTimelinePage`
* `FetchRepliesToRoot`
* `StreamLiveFeed`
* `FetchFollowedTimelinePage` 的 community-address 分支

### 当前 relay 选择职责是分散的

* `FetchEventsByFilter` 通过 `defaultRelaysForFilter()` 统一了一部分策略
* `FetchSpecificEvent` 自己拼了 pointer relays + fallback + author outbox 的两段式策略
* timeline/feed 系列自己直接调用 `FetchOutboxRelays` 或 `FallbackRelays`
* replaceable loader 又有一套自己的 `determineRelaysToQuery`

### 现有 spec 已经表达了两个重要方向

* generic filter read 应该优先走 `FetchEventsByFilter`，让 SDK 统一 relay 选择和 local-first 语义
* scoped/community thread 查询应该 local-first，再用 relay 补齐

## Research References

* `research/relay-usage-architecture.md` — 全量梳理当前 fetch/relay 组织方式和 3 个架构选项
* `.trellis/spec/backend/query-patterns.md` — generic filter / relay defaults / local-first 契约
* `.trellis/spec/backend/forked-sdk-architecture.md` — SDK 角色边界与 scoped-thread local-first 契约

## Feasible Approaches

### Approach A: 保持每类 fetch 各自管 relay，但按“家族”明确边界

怎么做：

* 保留 `filter` / `specific-event` / `timeline-feed` / `replaceable-loader` 四套 relay 选择逻辑
* 只做文档化和边界收敛，不强推共用 selector

优点：

* 改动最小
* 保留现有语义差异

缺点：

* relay 策略依旧分散
* 后续容易再次漂移

### Approach B: 引入统一 relay selector，按“query intent” 选 relay

怎么做：

* 增加一个中心层，按 `id` / `single-author` / `global` / `community-definition` / `thread` / `replaceable-kind-*` 之类 intent 返回 relay 集
* 各 `Fetch*` 负责 orchestration，但 relay 集构造统一交给 selector

优点：

* 策略最清晰
* 最符合“SDK 统一管理 relay 选择”的方向

缺点：

* 需要较多梳理
* 要设计 intent 分类，避免过度抽象

### Approach C: 两层模型，先收集 hint，再做 SDK 扩展/兜底

怎么做：

* 第一层只收集 candidate relay hints：caller、pointer、author、filter、community scope
* 第二层再决定这些 hints 是“强覆盖”还是“追加扩展”，并补上 `FallbackRelays` / `JustIDRelays` / `RelayListRelays` / `FetchOutboxRelays`

优点：

* 最符合当前代码现实
* 能显式表达“caller relays 是 override 还是 hint”这个关键差异

缺点：

* 比 A 复杂
* 仍要定义每个 fetch family 的扩展规则

## Decision (ADR-lite)

**Context**

upstream `nostr/sdk` 本身并不是“一个总 selector 统一所有查询”，而是：

* 共享的 hints 学习层（query attempts + event middleware）
* 共享的 author-read 入口（`FetchOutboxRelays`）
* 各 fetch family 在其上叠加自己的 relay 扩展策略

nosmec 当前的问题不是缺少一个绝对统一的 selector，而是 fork 后新增的 fetch 函数（timeline/community/filter/scoped helpers）有些偏离了 upstream 的分层方式，导致 relay 责任不一致。

**Decision**

采用 **Approach C**，并且以 **upstream-compatible** 为约束：

* 保留 hints / outbox / fallback / pointer-relays 这些 upstream 核心层次
* 不强推“所有 fetch 全部改成一个万能 selector”
* 把 nosmec 自己新增的 fetch 函数重新归类到 upstream 风格的 family 中
* 在 family 内统一“caller relays 是 hint 还是 override”“是否 local-first”“何时追加 fallback streams”

**Consequences**

* 更容易和 upstream 行为对齐，减少 future drift
* 改动会集中在我们 fork 后新增的 read-side fetch API，而不是重写整个 SDK
* 仍然需要明确各 family 的边界，否则只是把不一致换个地方继续存在

## Assumptions (temporary)

* 这次目标是先统一组织方式，不一定一次把所有 `Fetch*` 都完全重写
* 用户希望优先解决“不同 fetch 函数 relay 用法不一致”这个设计问题，而不是单独修某一个命令

**Confirmed scope**: 先收敛 `system.go` 里的新增 fetch:
`FetchGlobalTimelinePage` / `FetchMyTimelinePage` / `FetchFollowedTimelinePage` / `FetchEventsByFilter` / `FetchRepliesToRoot` / `FetchParent` / `FetchEventByFilter`

## Requirements (evolving)

* 明确不同 fetch family 的 relay 责任边界
* 明确 local-store-first 适用范围
* 明确 caller relays 是 override 还是 hint 的规则
* 给出可执行的重构路线，而不是只做零散修补
* 新增/改造后的组织方式要尽量贴近 upstream `nostr/sdk` 的 hints/outbox/fallback 分层

## Acceptance Criteria (evolving)

* [ ] 能清晰分类当前 `Fetch*` 的 relay 选择策略
* [ ] 能确定一个统一组织方案并说明 tradeoff
* [ ] 能产出后续实现步骤，不会导致 fetch 行为进一步分叉

## Definition of Done

* PRD 明确最终选择的组织方案
* `implement.jsonl` / `check.jsonl` 已补齐供后续 sub-agent 使用
* 任务可进入 `in_progress` 做实现或重构

## Out of Scope

* 本次不处理网络环境本身是否可达
* 本次不要求一次性重写所有 dataloader / list / zap fetch 路径
* 不把“所有 fetch 都收敛到一个函数”作为默认目标

## Technical Notes

* `nostr_sdk/system.go` — timeline / generic filter / replies / default relay policy
* `nostr_sdk/specific_event.go` — specific-event 两段式 relay 策略
* `nostr_sdk/feeds.go` — feed/live path 使用 outbox relay
* `nostr_sdk/outbox.go` — outbox/inbox/write relay heuristic
* upstream `fiatjaf.com/nostr/sdk/tracker.go` — relay hint 的采集入口
* upstream `fiatjaf.com/nostr/sdk/outbox.go` — `Hints.TopN()` 如何驱动 author 读侧 relay 选择
* upstream `fiatjaf.com/nostr/sdk/specific_event.go` — pointer relays + fallback + author outbox 的混合策略
* upstream `fiatjaf.com/nostr/sdk/replaceable_loader.go` — replaceable/addressable loader 的 kind-sensitive relay 补齐逻辑

## Upstream SDK Hint Model

upstream `nostr/sdk` 对 relay hints 的处理大致是 4 层：

1. **查询尝试也记 hint**
   `TrackQueryAttempts()` 会把 `(author, relay, kind)` 的查询尝试写进 `HintsDB`，但会跳过 replaceable list 类 kind（0 / 3 / 10002）和 ephemeral 区间。

2. **收到事件时持续学习 hint**
   `TrackEventHintsAndRelays()` 是 `Pool.EventMiddleware`：
   * 对大多数事件，记录 `event -> relay`
   * 对事件作者记录 `MostRecentEventFetched`
   * 从 `p` tag 第三个字段里的 relay hint 学习 `LastInHint`
   * 从 NIP-27 内容里的 profile/event/entity pointer relay hint 学习 `LastInHint`
   * 对 kind `10002` 特殊处理：从 `r` tags 学习作者的 `LastInRelayList`
   * 对 kind `3` 特殊处理：从联系人列表上的 relay hint 学习联系人的 `LastInHint`

3. **读 author 内容时，核心入口是 `FetchOutboxRelays()`**
   upstream 先触发一次 kind `10002` 拉取，然后直接用 `sys.Hints.TopN(pubkey, 6)` 选 relay；如果没有 hint，才 fallback 到 `damus` / `nos.lol`。

4. **不同 fetch path 在 outbox 之上再叠加自己的策略**
   * `FetchSpecificEvent()`：pointer 自带 relays 优先，其次加一层 fallback，再按 author 补 `FetchOutboxRelays()`
   * `determineRelaysToQuery()`：replaceable/addressable loader 先看 hint/outbox，再按 kind 补 `MetadataRelays` / `FollowListRelays` / `RelayListRelays` / `FallbackRelays`

结论：upstream 不是“统一一个 selector 解决一切”，而是：

* **HintsDB / Trackers 是共享学习层**
* **FetchOutboxRelays 是 author-read 的核心公共入口**
* **具体 fetch family 再在其上叠加自己的 relay 扩展规则**
