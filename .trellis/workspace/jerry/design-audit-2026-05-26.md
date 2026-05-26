# 项目架构审计排查计划

> 基于 2026-05-26 全面代码审查整理。
> 原则：优先排查我们自己加的层，尽量不先动上游 nostr_sdk。

---

## 排查顺序与优先级总览

| # | 优先级 | 归属 | 主题 | 重点文件 |
|---|--------|------|------|----------|
| 1 | P0 | ours | 资源生命周期一致性 | `cmd/root.go`, `config/context.go`, `config/config.go` |
| 2 | P0 | ours | 全局单例并发安全 | `config/config.go` |
| 3 | P0 | mixed | 时间线/Feed 并发写共享状态 | `nostr_sdk/system.go`, `nostr_sdk/feeds.go` |
| 4 | P1 | ours | AppContext 和 Global* 双轨并存 | `config/context.go`, `config/config.go` |
| 5 | P1 | ours | utils 职责过载 | `utils/*.go` |
| 6 | P1 | ours | 配置写入语义不一致 | `config/context.go`, `config/relay.go` |
| 7 | P2 | ours | relay 编码逻辑重复 | `config/config.go`, `nostr_sdk/event_relays.go`, `cmd/relay_commands.go` |
| 8 | P2 | ours | relay list 命令越层 | `cmd/relay_commands.go` |
| 9 | P2 | upstream | 默认 relay 分散 | `nostr_sdk/system.go`, `nostr_sdk/outbox.go` |
| 10 | P3 | upstream | HintsDB 接口不完整 | `nostr_sdk/hints/interface.go` |
| 11 | P3 | upstream | outboxShortTermCache 缓存模型 | `nostr_sdk/outbox.go` |

---

## P0 — 必查

### 条目 1
- **名称**：资源生命周期不一致
- **优先级**：P0
- **归属**：ours
- **文件**：`cmd/root.go`, `config/context.go`, `config/config.go`
- **症状**：`app.Close()` 只关了 `KVStore`，没关 `Pool`、`Store`、`Hints`。`reloadApp()` 关闭后再复用全局 `GlobalSystem`/`GlobalPool`。signal handler 直接 `os.Exit(0)`。
- **触发条件**：reload、多次启动、Ctrl+C 退出时
- **当前实现**：
  - `AppContext.Close()` 只调 `a.sys.KVStore.Close()`
  - `System.Close()` 会关 KVStore 和 Pool
  - reloadApp 先 app.Close 再调 GlobalPool()
- **风险**：关闭后又继续使用被关闭的 backend；后台连接/goroutine 残留；LMDB/Bleve 资源泄漏
- **是否必须修改 SDK**：否
- **建议处理方式**：明确 `System` 是 runtime owner；`AppContext.Close()` 代理到 `System.Close()`；`reloadApp()` 不复用半关闭对象
- **结论**：待查

---

### 条目 2
- **名称**：全局单例初始化非并发安全
- **优先级**：P0
- **归属**：ours
- **文件**：`config/config.go`
- **症状**：`GlobalHints()`、`GlobalKVStore()`、`GlobalStore()`、`GlobalPool()` 都是 `if nil then create` 无锁模式
- **触发条件**：并发首次访问（如 shell completion + normal CLI 并发跑）
- **当前实现**：
  ```go
  func GlobalHints() hints.HintsDB {
      if globalHints != nil { return globalHints }
      // create...
  }
  ```
- **风险**：LMDB/Bleve 重复打开、指针覆盖、资源泄漏
- **是否必须修改 SDK**：否
- **建议处理方式**：每个全局 backend 加 `sync.Once`，或彻底取消 runtime global 统一走 `AppContext` 注入
- **结论**：待查

---

### 条目 3
- **名称**：时间线/Feed 并发写共享切片（疑似 data race）
- **优先级**：P0
- **归属**：mixed（可能在 upstream）
- **文件**：`nostr_sdk/system.go:337-446`、`nostr_sdk/feeds.go:112-173`
- **症状**：
  - `FetchFollowedTimelinePage` 多 goroutine 同时 `append(events, ...)`
  - `FetchFeedPage` 也有类似模式
- **触发条件**：多 pubkey 查询时
- **当前实现**：
  ```go
  wg.Add(len(pubkeys))
  for _, pk := range pubkeys {
      go func(pk nostr.PubKey) {
          // append 到共享 events slice
          events = append(events, ie.Event)
      }(pubkey, oldestTimestamp)
  }
  wg.Wait()
  ```
- **风险**：事件丢失、重复、乱序；race detector 报错；排序前数据已损坏
- **是否必须修改 SDK**：待确认（先确认是否我们引入的改动）
- **建议处理方式**：每个 goroutine 写本地 slice 最后汇总；或走 channel 单线程收集；先标记为"疑似 upstream"
- **结论**：待查

---

## P1 — 高价值

### 条目 4
- **名称**：AppContext 和 Global* 双轨并存
- **优先级**：P1
- **归属**：ours
- **文件**：`config/context.go`、`config/config.go`、所有直接调 `config.Global*` 的位置
- **症状**：同时存在 `config.GlobalSystem` 和 `app.System()` 两条访问路径，初始化顺序隐含
- **触发条件**：维护/测试时不知道该走哪条
- **当前实现**：
  - `GlobalPool()` 内部初始化 `GlobalSystem`
  - `NewAppContext` 也依赖 `GlobalSystem`
  - 两边都持有 Pool/Hints/KVStore
- **风险**：边界不清导致逻辑错位；测试需要手工 reset global
- **是否必须修改 SDK**：否
- **建议处理方式**：长期：运行时对象全部通过 `AppContext` 持有；全局函数只保留配置加载；短期：规定业务层禁止直接用 `config.Global*`
- **结论**：待查

---

### 条目 5
- **名称**：utils 职责边界过载
- **优先级**：P1
- **归属**：ours
- **文件**：`utils/post.go`、`utils/dm.go`、`utils/community.go`、`utils/search.go`、`utils/profile.go`
- **症状**：一个函数同时包含网络访问、relay 选择、event 组装、签名、发布逻辑
- **触发条件**：新功能加不进去时容易堆砌到 utils
- **当前实现**：例如 `ReplyToNote` 同时做了：FetchNote、BuildReplyTags、event 构建、签名、PublishMany
- **风险**：SDK 不纯、utils 越来越重、跨层依赖混乱
- **是否必须修改 SDK**：否（但某些逻辑确实应该放 SDK 层）
- **建议处理方式**：
  - `utils` 保留：纯 helper、CLI 装配
  - 移出：网络访问、event 构造/发布、relay 选择 → SDK 或 service 层
  - 先标记每个函数的"职责类型"，不要急着搬家
- **结论**：待查

---

### 条目 6
- **名称**：配置写入语义不一致
- **优先级**：P1
- **归属**：ours
- **文件**：`config/context.go`、`config/relay.go`
- **症状**：
  - 有的方法返回 error
  - 有的吞掉 error
  - 有的带锁有的不带
  - `SyncRelayList`、`SyncDMRelays`、`AddAlias` 完全不返 error
- **触发条件**：配置持久化失败时无法感知
- **当前实现**：
  ```go
  func (a *AppContext) SyncRelayList(relays []Relay) {
      a.mu.Lock()
      a.cfg.RelayList = relays
      a.viper.Set("relay_list", relays)
      a.viper.WriteConfig() // 错误被静默
  }
  ```
- **风险**：配置修改失败用户不知道；难排障
- **是否必须修改 SDK**：否
- **建议处理方式**：
  - 纯数据操作留在 `config/relay.go`
  - 带持久化副作用的统一返回 error
  - 抽统一 config mutation helper
- **结论**：待查

---

## P2 — 冗余与结构

### 条目 7
- **名称**：relay 编码/解码逻辑重复
- **优先级**：P2
- **归属**：ours
- **文件**：`config/config.go`、`nostr_sdk/event_relays.go`、`cmd/relay_commands.go`
- **症状**：三处实现了几乎一样的 relay 列表二进制编码协议
- **触发条件**：底层格式一变，多处不同步
- **当前实现**：
  - `config.encodeRelayListCompat()` / `decodeRelayListCompat()`
  - `nostr_sdk.encodeRelayList()` / `decodeRelayList()`
  - `cmd.decodeKVRelayList()`
- **风险**：格式漂移；CLI 和 SDK 行为不一致
- **是否必须修改 SDK**：否（但协议本身是 SDK 在用）
- **建议处理方式**：
  - 确定"事实标准"是 SDK 那套
  - config 层和 CLI 层全部复用同一份
  - 删除重复实现
- **结论**：待查

---

### 条目 8
- **名称**：relay list 命令越层直接读 LMDB
- **优先级**：P2
- **归属**：ours
- **文件**：`cmd/relay_commands.go:110-193`
- **症状**：`relay list` 命令直接打开 kvstore LMDB 文件并手动解析二进制格式
- **触发条件**：查看已知 relay 时
- **当前实现**：命令层绕过业务接口直接扫存储文件
- **风险**：存储细节泄漏给 CLI；格式改 CLI 就坏；难以单元测试
- **是否必须修改 SDK**：否
- **建议处理方式**：
  - 给 `System` 或 `KVStore` 增加 `ListKnownEventRelays()` 公开接口
  - `relay list` 改走业务层调用
  - 短期：先记账；长期：收敛到统一接口
- **结论**：待查

---

### 条目 9
- **名称**：默认 relay 和 fallback 策略分散
- **优先级**：P2
- **归属**：upstream
- **文件**：`nostr_sdk/system.go`、`nostr_sdk/outbox.go`
- **症状**：硬编码 relay URL 散落在 6+ 处
- **触发条件**：更换/添加默认 relay 时
- **当前实现**：
  - `RelayListRelays`、`FollowListRelays`、`MetadataRelays`、`FallbackRelays`、`JustIDRelays`、`UserSearchRelays`、`NoteSearchRelays` 分别硬编码
  - outbox 里 fallback `[]string{"wss://relay.damus.io", "wss://nos.lol"}`
- **风险**：配置变更困难；fallback 不一致
- **是否必须修改 SDK**：否（上游设计）
- **建议处理方式**：
  - 先记录
  - 如需修改，通过我们自己的 config 层 wrap，不直接改 SDK 内部
  - 长期：抽统一的 relay policy 类型
- **结论**：暂缓

---

## P3 — 暂缓观察

### 条目 10
- **名称**：HintsDB 接口缺少 Close()
- **优先级**：P3
- **归属**：upstream
- **文件**：`nostr_sdk/hints/interface.go`
- **症状**：接口无 `Close()`，导致生产代码无法统一释放；测试只能类型断言手工关
- **触发条件**：退出时 hints backend 资源释放
- **当前实现**：`HintsDB` 只有 `TopN/Save/PrintScores/GetDetailedScores/GetAllKnownRelays`
- **风险**：如果 backend 需要显式关闭（如 LMDB），接口层面无法统一调用
- **是否必须修改 SDK**：仅当阻碍我们生命周期收口时
- **建议处理方式**：先记账；看 AppContext.Close() 能不能绕开；如果不能，再考虑在接口层补 wrapper
- **结论**：暂缓

---

### 条目 11
- **名称**：outboxShortTermCache 简化哈希碰撞问题
- **优先级**：P3
- **归属**：upstream
- **文件**：`nostr_sdk/outbox.go:11-50`
- **症状**：256 格固定槽，用 `pubkey[7]` 单字节索引，两个 pubkey 第8字节相同就互相覆盖缓存
- **触发条件**：高并发下不同 pubkey 缓存互相污染
- **当前实现**：
  ```go
  var outboxShortTermCache = [256]atomic.Pointer[ostcEntry]{}
  ostcIndex := pubkey[7]
  ```
- **风险**：relay 推断结果随机抖动；调试困难（不是必现）
- **是否必须修改 SDK**：先看外层是否已受影响
- **建议处理方式**：先记账；如果出现异常 relay 选择行为，再决定是 patch SDK 还是外层加校验
- **结论**：暂缓

---

## 修改优先级总结

### P0 先改（真 bug / 高风险，我们自己的层）
1. 资源生命周期收口（`app.Close`、`reloadApp`、signal handler）
2. 全局 backend 初始化加并发保护
3. 时间线/Feed 并发问题（先确认归属，再决定是否改 SDK）

### P1 后做（高价值，降低维护成本）
4. 统一 `AppContext` vs `Global*` 的职责边界
5. utils 职责标记（不急着搬，先看清楚）
6. 配置写入统一 error 语义

### P2 中期做（减少冗余）
7. relay 编码逻辑三合一
8. `relay list` 改走业务层接口
9. 默认 relay 策略记录（暂不动 SDK）

### P3 暂缓
10. HintsDB Close() 接口问题
11. outboxShortTermCache 缓存模型

---

## 排查动作检查清单

### 第一轮（我们自己控制的层）

- [ ] `cmd/root.go`：reloadApp 路径、signal handler 退出、app.Close 调用点
- [ ] `config/context.go`：所有方法分类（纯 read / 带写 / 带持久化）、Close 实现
- [ ] `config/config.go`：所有 Global* 函数初始化路径、是否带锁
- [ ] `utils/post.go`、`utils/dm.go`、`utils/community.go`：每个函数的职责类型打标
- [ ] `cmd/relay_commands.go`：relay list 是否绕过业务层、readLMDB 错误处理

### 第二轮（确认归属）

- [ ] `nostr_sdk/system.go`：FetchFollowedTimelinePage 是否我们改过
- [ ] `nostr_sdk/feeds.go`：FetchFeedPage 是否我们改过
- [ ] `nostr_sdk/outbox.go`：outboxShortTermCache 是否我们引入

### 第三轮（决定修改范围）

- [ ] 列出所有必须改 SDK 的条目
- [ ] 列出所有可以外层绕开的条目
- [ ] 确定哪些是 upstream 原样、先不碰

---

## 附：nostr_sdk 上游归属判断方法

1. 检查文件头注释或 git history 是否标注 "forked from"
2. 对比与我们改动前的上游版本是否一致
3. 如果是上游原样，优先：
   - 通过我们自己的 config/wrapper 层规避
   - 或在 SDK 外层做校验/缓存
   - 不直接改 SDK 内部实现
4. 只有当问题明确在外层无法解决、且改动范围可控时，才局部 patch SDK

---

*最后更新：2026-05-26*