# background nosmec still blocks other processes and completion

## Goal

确认并修复 `GlobalSystem` 持久化 wiring 与 spec 不一致的问题，确保 `Store`、`KVStore`、`Hints` 真正接入 LMDB/Bleve，而不是停留在 `nostr_sdk.NewSystem()` 的默认内存实现。

## What I already know

- 之前已经做过一次 lazy-init 修复：`initApp()` 不再主动调用 `GlobalPool()`，理论上 completion 不应该在启动时打开 LMDB。
- 我本地已复现一部分场景：一个 PTY 中运行 `nosmec note timeline --mine` 时，直接执行 `/tmp/nosmec-debug completion bash` 可以快速返回。
- 代码里发现一个明确问题：
  - `cmd/root.go` 定义了一个 `cmdContextKey`
  - `cmd/completion/completion.go` 又定义了另一个同名但不同类型的 `cmdContextKey`
  - 导致 completion 子包里的 `GetApp(cmd)` 实际总是拿不到 `AppContext`
- 代码里还发现一个更大的可疑点：spec 说 `GlobalPool()` 应负责把 `GlobalSystem.Store/KVStore/Hints` 接到 LMDB/Bleve，但当前 `config/config.go` 只看到了 `GlobalPool()` 创建 `Pool`，没有看到把 `GlobalSystem` 的 persistent store/hints/kvstore 正确 wiring 回去。
- 新复现结果：在一个 PTY 中运行 `nosmec note timeline --mine --limit 5` 时，直接执行 `eval "$(nosmec completion bash)"` 依然立即返回 `0`，说明“任意 timeline 运行就必现 completion 卡住”这个假设不成立。
- 新证据：`cmd/root.go` 和 `cmd/completion/completion.go` 定义了两个不同的 `cmdContextKey` 类型，所以 completion 动态补全函数里的 `GetApp()` 实际总是返回 `nil`。
- 新确认的 root cause 候选：`config/config.go` 中 `GlobalPool()` 只创建并回填了 `Pool`，没有把 persistent `Hints` / `KVStore` / `Store` 接回 `GlobalSystem`。
- `nostr_sdk.NewSystem()` 的默认值明确是：memory `KVStore`、memory `Hints`、`nullstore.NullStore`。
- 全仓库目前没有搜到别处把 LMDB/Bleve `Store`、LMDB `KVStore`、LMDB `Hints` 重新赋值给 `GlobalSystem`。

## Assumptions (temporary)

- 当前代码实际运行状态很可能是“局部打开了持久化资源，但 `GlobalSystem` 仍持有内存实现”。
- 修复 wiring 后，依赖 `GlobalSystem` 的路径会开始真正读写 LMDB/Bleve，需要同步验证不会破坏 lazy-init 和现有调用方。

## Open Questions

- `GlobalSystem.Store` 应该直接挂 `BleveBackend`，还是在 Bleve 初始化失败时回退到 `LMDBBackend`，与 spec 保持一致？

## Requirements (evolving)

- `GlobalSystem.Hints` 必须使用 `GlobalHints()` 返回的 LMDB hints 实例
- `GlobalSystem.KVStore` 必须切到 LMDB-backed KVStore，而不是 memory store
- `GlobalSystem.Store` 必须按 spec 接成 Bleve-over-LMDB，Bleve 失败时回退 LMDB，LMDB 失败时保留 local cache disabled 行为
- wiring 必须尽量保持 lazy-init，不引入无关启动副作用

## Acceptance Criteria (evolving)

- [ ] 有测试能证明 `GlobalSystem` 不再保留默认 memory/null wiring
- [ ] `GlobalSystem.Hints` / `KVStore` / `Store` 在初始化后与 spec 一致
- [ ] 目标测试和相关构建检查通过

## Out of Scope

- 不顺手修 completion 的 `cmdContextKey` 分裂问题
- 不处理与本次 wiring 无关的 shell/job-control 场景

## Technical Notes

- `cmd/root.go` — app 初始化与 lazy-init
- `cmd/completion/completion.go` — completion 的 app 获取路径
- `config/context.go` — `Pool()` / `Hints()` lazy-init
- `config/config.go` — `GlobalPool()` / `GlobalHints()` / `GlobalSystem` wiring
- `nostr_sdk/system.go` — `NewSystem()` 默认 memory/null 实现
- `nostr_sdk/kvstore/lmdb/store.go` — LMDB KVStore backend
- `fiatjaf.com/nostr/eventstore/lmdb` / `bleve` — persistent event store stack
