# Design Decisions — 2026-05-10

> **Superseded by**: `docs/DESIGN_PRINCIPLES.md`
> Most of these decisions have been implemented. See the new design principles doc for current state.

## 1. Relay查询架构简化

## 1. Relay查询架构简化

### 现状问题
- 本地relay和远程relay分开查询(先本地2s, 再远程10s)
- 私有relay(privateRelays)和本地relay混在一起

### 决定
- **不区分本地relay和远程relay**, 所有relay统一pool查询
- **删除私有relay实现**, 只保留本地relay
- 不做本地优先,所有relay同时查询

## 2. 统一超时设置

### 现状问题
- 本地relay: 2秒超时
- 远程relay: 10秒超时
- 超时值硬编码不可配置

### 决定
- 统一超时设置, 去掉本地/远程的区别
- 超时值应该可配置(通过config/env)

## 3. CLI命令职责划分

### 现状问题
- `note_commands.go` 和 `event_commands.go` 职责不清

### 决定
- `event_command` 处理通用event(所有kind)
- `note` 只处理kind 1类型的note

## 4. Context和超时统一

### 现状问题
- `GetEventAsync` 用自定义channel而非 `context.WithTimeout`
- `SubscribeMany` 没有超时context

### 决定
- 统一使用 `context.WithTimeout`
- `SubscribeMany` 也要使用带超时的context

## 5. 数据库

- **boltdb** (`go.etcd.io/bbolt v1.4.2`)
- 用于本地relay的event存储
- 路径: `~/.cache/nosmec/nosmec.db`
- Journal记录: "存储从lmdb切换到boltdb+bleve支持全文搜索"

## 6. 需要修改的代码

### 删除
- `PrivateRelays()` 方法
- 私有relay相关的config和publish逻辑

### 重构
- `GetEvent`: 删除本地/远程分离查询,统一超时
- `GetEventAsync`: 改用 `context.WithTimeout`
- `SubscribeMany`: 加超时context
- `CacheEvent`: 简化,去掉PublishMany到私有relay

### 配置调整
- 添加 `query.timeout` 配置项
- 删除 `private_relays` 相关配置

## 7. 当前Query模式总结

| 函数 | 模式 | 问题 |
|---|---|---|
| `GetEvent` | 同步,本地2s+远程10s分离 | 超时太长,分离逻辑 |
| `GetEventAsync` | goroutine+channel | 应该用context.WithTimeout |
| `GetNoteAsync` | 调用GetEventAsync | 同上 |
| `GetProfileAsync` | 调用GetEventAsync | 同上 |
| `GetMyTimeline` | 返回chan, 用FetchMany | OK |
| `GetGlobalTimeline` | 返回chan, 用FetchMany | OK |
| `SubscribeMany` | 无超时context | 需要加超时 |
| `CacheEvent` | PublishMany到私有relay | 私有relay要删除 |