# 审计整改：资源生命周期与 Close 语义修复

## Goal

统一资源生命周期管理，修复 `AppContext.Close()` 的不完整实现，使 reload/signal handler 路径可预测，不出现"半关闭后继续使用"或"资源未释放"的问题。

---

## What I already know

### 当前 Close 路径梳理

**`cmd/root.go`**:
- `Execute()`: signal goroutine 调 `app.Close()` 后直接 `os.Exit(0)` — 关闭后立即退出，风险有限但不够干净
- `reloadApp()`: 先 `app.Close()` 再调用 `config.GlobalPool()` 创建新实例
- `reloadApp()` 只有两个调用点，都在 `cmd/config_commands.go`：`config set` 和 `config alias remove`

**用户新增决策（2026-05-26）**:
- 打算删除整个 `config set` 功能
- 打算删除整个 `config alias` 功能组（`list` / `add` / `remove`）
- 这意味着 `reloadApp()` 的现有调用点可能会一起消失

**`config/context.go:415-431` — `AppContext.Close()`**:
```go
func (a *AppContext) Close() error {
    a.mu.Lock()
    defer a.mu.Unlock()
    var errs []error
    // 只关了 KVStore
    if a.sys != nil && a.sys.KVStore != nil {
        if err := a.sys.KVStore.Close(); err != nil {
            errs = append(errs, err)
        }
    }
    if len(errs) > 0 {
        return errors.Join(errs...)
    }
    return nil
}
```
**缺失**：Pool、Store、Hints 都没关。

**`nostr_sdk/system.go:201-208` — `System.Close()`**:
```go
func (sys *System) Close() {
    if sys.KVStore != nil {
        sys.KVStore.Close()
    }
    if sys.Pool != nil {
        sys.Pool.Close("sdk.System closed")
    }
}
```
**缺失**：Store、Hints 没关。

**`config/config.go` 全局 backend 初始化**:
- `GlobalHints()`、`GlobalKVStore()`、`GlobalStore()`、`GlobalPool()` — 都是 `if != nil { return } else { create }` 无锁模式
- 这四个在首次并发调用时可能重复初始化

**`HintsDB` 接口**（`nostr_sdk/hints/interface.go`）: 无 `Close()`，所以即使我们想统一关，也没有统一途径。

**`KVStore` 接口**（`nostr_sdk/kvstore/interface.go`）: 有 `Close() error`。

### 关键观察

1. `AppContext.sys` 和 `GlobalSystem` 指向同一个实例
2. `AppContext.Close()` 调用时 `a.sys` 已存在但只关了 KVStore
3. `System.Close()` 多关了一个 Pool，但没关 Store 和 Hints
4. reload 后 `config.GlobalPool()` 会复用之前的 `GlobalSystem`，如果 `KVStore` 已被关，会拿 nil 或已关闭的 backend 继续工作
5. `reloadApp()` 的真实用途不是“重建整套运行时”，而是“配置变更后刷新当前进程里的 AppContext 配置快照”
6. `reloadApp()` 目前只有两个调用点：`config set` 和 `config alias remove`
7. 用户已决定删除整个 `config set` 和整个 `config alias` 命令组，因此 `reloadApp()` 将变成无调用点的死代码

---

## Requirements

### R1: `AppContext.Close()` 必须完整关闭 System 持有的所有资源

当前：只关 KVStore
目标：关 Pool、KVStore、Store、Hints（按正确顺序）

### R2: `System.Close()` 应该完整关闭所有资源

当前：关 KVStore、Pool
缺失：Store、Hints
目标：关 Pool、Store、Hints、KVStore（按依赖顺序）

### R3: reloadApp() 必须重建 System，不能复用被关闭的实例

当前：reloadApp() 先调 app.Close() 关 KVStore，然后调用 GlobalPool() 复用同一个 GlobalSystem
问题：GlobalSystem 里的 KVStore 已被关，但 GlobalPool() 继续用它
目标：reloadApp() 应该完全重建 System，或完全不复用

### R3a: 如果删除 `config set` 和 `config alias` 后 `reloadApp()` 不再有调用点，应直接删除该函数

当前：`reloadApp()` 只服务于配置命令路径
目标：如果调用点消失，则删除 `reloadApp()`，避免保留无用且误导性的生命周期逻辑

### R3b: 删除整个 `config alias` 命令组

删除：
- `config alias list`
- `config alias add`
- `config alias remove`

目标：移除一整组不再需要的配置能力，同时清理相关 completion / 调用路径 / 文案

### R3c: 删除整个 `config set` 命令

删除：
- `config set <key> <value>`

目标：移除通用“任意 key 写入”入口，避免继续依赖 `reloadApp()` 刷新配置快照

### R3d: 删除 `reloadApp()` 函数本体

前提：`config set` 与整个 `config alias` 命令组删除后，`reloadApp()` 无剩余调用点。

目标：
- 从 `cmd/root.go` 删除 `reloadApp()`
- 清理其所有调用点
- 不再为“配置快照刷新”保留一条单独的生命周期路径

### R4: 全局 backend 初始化必须并发安全

当前：`GlobalHints/GlobalKVStore/GlobalStore/GlobalPool` 都是无锁 if-nil-create
目标：每个加 sync.Once，或统一走 AppContext 注入

### R5: HintsDB 接口应该可关闭（通过类型断言）

当前：HintsDB 接口无 Close()，生产代码无法统一关闭
目标：在不修改接口的前提下，确保调用 Close() 时能正确关闭（类型断言）

---

## Acceptance Criteria

- [ ] `AppContext.Close()` 调用后，Pool、KVStore、Store、Hints 均被关闭
- [ ] `System.Close()` 调用后，Pool、Store、Hints、KVStore 均被关闭（按依赖顺序）
- [ ] reloadApp() 调用后，新的 AppContext 持有新的 System 实例，旧实例不被继续使用
- [ ] 如果 `config set` 和 `config alias` 被删除，`reloadApp()` 及其调用点一并删除
- [ ] `reloadApp()` 从代码中完全删除
- [ ] `config set` 命令从 CLI 中移除
- [ ] `config alias` 命令组及其子命令从 CLI 中完全移除
- [ ] `GlobalHints/GlobalKVStore/GlobalStore/GlobalPool` 在并发首次调用下不重复初始化
- [ ] 所有 backend 的 Close() 错误被收集并返回（或记录到日志）
- [ ] 行为变更不影响现有功能（timeline、feed、post、DM、relay list 等命令正常）

---

## Definition of Done

- 所有资源（Pool、Store、Hints、KVStore）有明确 owner 和关闭路径
- reloadApp() 和 signal handler 路径行为可预测
- 全局初始化并发安全
- lint / typecheck / 测试通过
- 不引入新的资源泄漏

---

## Technical Approach

### 1. 修复 `System.Close()` — 在 nostr_sdk/system.go

添加 Store 和 Hints 的关闭调用：

```go
func (sys *System) Close() {
    if sys.Pool != nil {
        sys.Pool.Close("sdk.System closed")
    }
    if sys.Store != nil {
        sys.Store.Close()
    }
    if sys.Hints != nil {
        if cl, ok := sys.Hints.(interface{ Close() }); ok {
            cl.Close()
        }
    }
    if sys.KVStore != nil {
        sys.KVStore.Close()
    }
}
```

注意：我们自己的 `config` 层如果调用 `System.Close()`，必须在 AppContext.Close() 之前关 Pool（因为 Pool 可能在关闭后还尝试写 Store）。

### 2. 修复 `AppContext.Close()` — 在 config/context.go

改为调用 `System.Close()`，统一走完整关闭路径：

```go
func (a *AppContext) Close() error {
    a.mu.Lock()
    defer a.mu.Unlock()

    if a.sys == nil {
        return nil
    }

    a.sys.Close() // 完整关闭 System 持有的所有资源
    return nil
}
```

这样 `AppContext.Close()` 成为 `System.Close()` 的代理，不重复资源管理逻辑。

### 3. 处理 `reloadApp()` — 在 cmd/root.go

两种方案：

**方案 A（推荐）：完全重建 System**（最干净）
```go
func reloadApp() {
    if app != nil {
        app.Close()
    }
    // 重置全局状态，强制重建
    config.ResetGlobalState()
    cfg := config.InitConfig()
    config.SetProxyConfig(config.ProxyConfig{
        Socks:    cfg.Proxy.Socks,
        I2PSocks: cfg.Proxy.I2PSocks,
    })
    pool := config.GlobalPool()
    app = config.NewAppContext(pool, cfg, config.GetViper())
    completion.SetApp(app)
}
```

**方案 B：不关，只重建 AppContext**（更保守，改动小）
```go
func reloadApp() {
    cfg := config.InitConfig()
    config.SetProxyConfig(config.ProxyConfig{
        Socks:    cfg.Proxy.Socks,
        I2PSocks: cfg.Proxy.I2PSocks,
    })
    pool := config.GlobalPool()
    if app != nil {
        // 不调 app.Close()，避免关闭全局 backend
        // 只清掉旧引用，让 GC 回收
    }
    app = config.NewAppContext(pool, cfg, config.GetViper())
    completion.SetApp(app)
}
```

**如果删除 `config set` / `config alias` 后没有剩余调用点**：
- 直接删除 `reloadApp()`
- 同时删除相关命令实现
- 这样这次任务里不再需要为 `reloadApp()` 设计新的生命周期语义

### 3a. 删除 `config alias` 命令组 — 在 `cmd/config_commands.go`

- 删除 `configAliasCmd`
- 删除 `configAliasListCmd`
- 删除 `configAliasAddCmd`
- 删除 `configAliasRemoveCmd`
- 清理相关 completion / 输出文案 / 可能的帮助文本引用

### 3b. 删除 `config set` 命令 — 在 `cmd/config_commands.go`

- 删除 `configSetCmd`
- 清理相关 completion / 输出文案 / 帮助文本引用
- 如果 `reloadApp()` 无剩余调用点，则删除 `reloadApp()` 函数本体

### 3c. 删除 `reloadApp()` — 在 `cmd/root.go`

- 删除 `reloadApp()` 函数本体
- 删除其调用点（随着 `config set` / `config alias` 删除一并消失）
- 关闭路径只保留真正的进程退出场景，不再保留“配置快照刷新”这条分支

### 4. 修复全局 backend 初始化并发安全 — 在 config/config.go

给每个全局 backend 加 `sync.Once`：

```go
var (
    globalPool     *nostr.Pool
    globalHints    hints.HintsDB
    globalKVStore  kvstore.KVStore
    globalStore    eventstore.Store
    globalConfig   Config
    configDir      string
    onceInit       sync.Once
    onceHints      sync.Once
    onceKVStore    sync.Once
    onceStore      sync.Once
    oncePool       sync.Once
    initialized    bool
    proxyConfig    ProxyConfig
    globalViper    *viper.Viper
    GlobalSystem   *nostr_sdk.System
)

func GlobalHints() hints.HintsDB {
    onceHints.Do(func() {
        // 初始化逻辑...
    })
    return globalHints
}
// 其他三个同理
```

或者更简单：**去掉 GlobalPool/GlobalHints/GlobalKVStore/GlobalStore 这四个 runtime 全局**，统一通过 `AppContext` 持有和访问，`config/config.go` 只保留纯配置相关的初始化。

### 5. 处理 HintsDB 无 Close() 接口问题

通过类型断言调用：

```go
if h, ok := sys.Hints.(interface{ Close() }); ok {
    h.Close()
}
```

这样不影响 `HintsDB` 接口定义（不改 upstream SDK），但调用方能正确关闭 hints backend。

---

## Decision (ADR-lite)

### 上下文
`AppContext.Close()` 只关了 KVStore，而 `System.Close()` 多了 Pool 但仍缺失 Store 和 Hints。两边行为不一致，且 reloadApp 复用被关闭的 backend。

补充：用户倾向于删除 `config set` 和 `config alias`，因此 `reloadApp()` 很可能不再需要保留。

### 决策
- `AppContext.Close()` 改为直接调用 `a.sys.Close()`，放弃"自己列举关哪些"的方式
- `System.Close()` 补齐 Store 和 Hints 的关闭调用
- `HintsDB` 通过类型断言关闭，不改接口
- `reloadApp()` 被认定为“配置快照刷新函数”，不是运行时重载函数
- 由于整个 `config set` 与整个 `config alias` 命令组删除，`reloadApp()` 直接删除
- 整个 `config set` 命令删除
- 整个 `config alias` 功能组删除
- 本次任务同时修复全局初始化并发安全

### 后果
- 所有资源关闭路径统一到 `System.Close()`
- reload 后拿到的 AppContext 持有全新 System，无残留状态
- 全局 runtime 初始化在并发下安全

---

## Out of Scope

- 不修改 `HintsDB` 接口定义（不改 upstream SDK）
- 不重构 `utils/` 职责边界
- 不改动 relay 默认策略
- 不改动 `outboxShortTermCache`

---

## Open Questions

（已清空：当前需求已收敛，可以进入实现阶段）

---

## Technical Notes

### 相关文件
- `cmd/root.go` — Execute/signal handler、reloadApp、app.Close 调用
- `config/context.go` — AppContext 定义、Close 实现
- `config/config.go` — 全局 backend 初始化、wirePersistentBackends
- `nostr_sdk/system.go` — System 结构、Close 实现
- `nostr_sdk/kvstore/interface.go` — KVStore 有 Close()
- `nostr_sdk/hints/interface.go` — HintsDB 无 Close()

### 资源 owner 关系
```
GlobalSystem (config.config.go)
  ├── Pool          (nostr.Pool)        — System.Close() 关
  ├── Store         (eventstore.Store)   — System.Close() 关（本次加）
  ├── Hints         (hints.HintsDB)     — System.Close() 关（本次加）
  └── KVStore      (kvstore.KVStore)    — System.Close() 关
```

*最后更新：2026-05-26*
