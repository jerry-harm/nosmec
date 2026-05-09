# brainstorm: 统一 sync/publish 配置并审查 NIP 支持

## Goal

梳理 config 中所有与 relay sync/publish 相关的配置，统一命名和职责，并审查 NIPs 补充缺失的 relay 配置。

## What I already know

### 现有 Relay 配置（config/types.go）

| 字段 | 类型 | 用途 |
|------|------|------|
| `RelayList` | `[]Relay{URL, Read, Write}` | 显式读写标记的 relay 列表 |
| `KnownRelays` | `[]string` | 发现 relay 的已知列表 |
| `PrivateRelays` | `[]string` | 私有 relay（用于缓存/发布） |
| `DMRelays` | `[]string` | DM 专用 relay |
| `SearchRelays` | `[]string` | 搜索专用 relay |
| `LocalRelay` | `LocalRelayConfig` | 内嵌本地 relay |

### 现有 Helper 方法（config/context.go）

- `AllReadableRelays()` = `ReadableRelays()` + localRelay
- `AllWritableRelays()` = `WritableRelays()` + localRelay
- `PrivateRelays()` = localRelay + `cfg.PrivateRelays`
- `ReadableRelays()` / `WritableRelays()` 从 `RelayList` 提取

### NIP 对应关系

| NIP | Event Kind | Config 字段 | 状态 |
|-----|------------|-------------|------|
| NIP-65 | kind:10002 | `RelayList` (Read/Write) | **未发布** |
| NIP-17 | kind:10050 | `DMRelays` | **未发布** |
| NIP-26 | delegation tag | — | 未支持 |
| NIP-29 | relay groups | `Subscriptions` (community) | 部分支持 |

## Problems

1. `RelayList` 有 Read/Write 标记但从未作为 kind:10002 发布
2. `DMRelays` 从未作为 kind:10050 发布
3. `PrivateRelays` 语义模糊——到底是"写入私有数据"还是"只读私有内容"？
4. `KnownRelays` 只是硬编码 fallback，没有动态更新机制
5. `SearchRelays` 独立存在，和 `KnownRelays` 可能重叠

## Open Questions

* `PrivateRelays` 的语义是否需要明确为"私有数据写入 relay"？
* 是否需要实现 kind:10002 / kind:10050 的发布逻辑？
* NIP-26  delegated signing 是否需要支持？

## Requirements

* 为 `RelayList` 实现 kind:10002 发布（NIP-65）—— 初始启动时和配置变更时发布
* 为 `DMRelays` 实现 kind:10050 发布（NIP-17）—— 初始启动时和配置变更时发布
* 统一 relay 配置的命名和职责（文档层面）
* 明确 `PrivateRelays` 职责（只写私有敏感数据 relay）
* 当用户通过 CLI 修改 relay 配置时，自动重新发布 kind:10002 / kind:10050

## Out of Scope (explicit)

* NIP-26 委托签名（暂不支持）
* Relay 发现机制优化（作为独立任务）

## Technical Notes

### NIP-65 kind:10002 发布
```go
// kind:10002 内容
{
  "kind": 10002,
  "tags": [
    ["r", "wss://relay1.com"],
    ["r", "wss://relay2.com", "write"],
    ["r", "wss://relay3.com", "read"]
  ],
  "content": ""
}
```

### NIP-17 kind:10050 发布
```go
// kind:10050 内容
{
  "kind": 10050,
  "tags": [
    ["relay", "wss://inbox.nostr.wine"],
    ["relay", "wss://myrelay.nostr1.com"]
  ],
  "content": ""
}
```

### 统一后的 Relay 配置提案

```
RelayList     - 用户配置的 relay 列表（NIP-65 kind:10002）
PrivateRelays  - 私有数据写入 relay（kind:4 DMs, kind:3 follows 等敏感数据）
DMRelays       - DM 接收 relay（NIP-17 kind:10050）
SearchRelays   - 搜索专用 relay（保持独立）
KnownRelays    - 已知 relay 列表（发现用，保持硬编码 fallback）
```

### Files to inspect
* `config/types.go` — Config struct
* `config/context.go` — AppContext helpers
* `utils/get.go` — relay 使用处
* `utils/post.go` — publish relay 使用处
* `utils/dm.go` — DM relay 使用处
* `tui/timeline/model.go` — timeline relay 使用处
* NIPs: 01, 17, 26, 29, 65
