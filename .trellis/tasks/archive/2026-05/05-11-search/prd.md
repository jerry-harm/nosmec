# Search 功能

## Goal

实现完整的 NIP-50 搜索功能：以 `search` 命令查询事件，支持 search relay list 和本地 relay 双源搜索，本地 relay 也启用 NIP-50 支持。

## Research Findings

**完整研究结论见**: `research/nip50.md`

### 关键发现

1. **SDK**: `nostr.Filter` 已有 `Search string` 字段（line 18），SDK 原生支持 NIP-50
2. **khatru**: 默认 `NoSearchQueries` 策略**拒绝**搜索查询，需移除
3. **BoltDB**: 不支持 NIP-50 search（Bleve 支持）
4. **方案**: BoltDB 做普通查询 + Bleve 做全文搜索（khatru cookbook 方案）

### NIP-50 协议

- Client 发送 REQ，filter 带 `search` 字段
- Relay 返回按质量排序（不是 `.created_at`）
- 支持扩展：`kinds:`, `authors:`, `#t:`, `#m:`, `domain:`, `language:`, `sentiment:`, `nsfw:`

## Requirements

### 1. CLI 命令

```bash
nosmec search "keyword"                    # 全文搜索
nosmec search "keyword" --kinds 1,3     # 按 kind 过滤
nosmec search "keyword" --limit 50       # 限制结果数
```

### 2. 搜索源

- **search_relays**：配置的搜索 relay 列表
- **本地 relay**：如果 `local_relay.enabled=true`，也加入搜索源
- 所有 relay 同时查询，结果合并去重

### 3. NIP-50 过滤解析

客户端需要解析搜索字符串中的 `key:value` 对：
- `kinds:1,3` → `filter.Kinds`
- `authors:npub1xxx` → `filter.Authors`
- `#t:nostr` → `filter.Tags["t"]`
- 剩余字符串作为全文搜索词

### 4. 本地 relay NIP-50 支持

修改 `config/config.go` 的 `StartLocalRelay()`:
1. 添加 `bleve.BleveBackend` 索引
2. 配置 `relay.QueryStored` 区分普通查询和搜索查询
3. 添加 NIP-50 到 relay 的 `SupportedNIPs`
4. **不添加** `NoSearchQueries` 策略

## Technical Approach

### 架构

```
search 命令
    ↓
解析搜索字符串 → nostr.Filter (Search + Kinds/Authors/Tags)
    ↓
构建搜索源列表：search_relays + [本地 relay]
    ↓
Pool.QuerySingle() / SubscribeMany() 发送 NIP-50 查询
    ↓
合并结果（去重，按 quality 或时间排序）
    ↓
格式化输出
```

### 本地 Relay NIP-50 实现

```go
// config/config.go - StartLocalRelay()

// BoltDB 做普通存储
boltStore := &boltdb.BoltBackend{Path: dbPath}

// Bleve 做全文搜索（包装 BoltDB）
searchIndex := bleve.NewBleveSearchIndex(boltStore)

// 配置 relay.QueryStored
relay.QueryStored = func(ctx context.Context, filter nostr.Filter) iter.Seq[nostr.Event] {
    if filter.Search != "" {
        return searchIndex.QueryEvents(filter)
    }
    return boltStore.QueryEvents(filter)
}

// 写入时同时更新两个 store
relay.StoreEvent = func(ctx context.Context, event nostr.Event) error {
    if err := boltStore.SaveEvent(event); err != nil {
        return err
    }
    return searchIndex.SaveEvent(event)
}

// 添加 NIP-50 到支持列表
relay.Info.SupportedNIPs = append(relay.Info.SupportedNIPs, 50)
```

### 新文件

```
cmd/search_commands.go   # search 命令定义
utils/search.go         # NIP-50 搜索逻辑
```

### 修改文件

```
config/config.go         # 添加 Bleve 索引，修改 StartLocalRelay
```

## Out of Scope

- 搜索结果的分页加载
- 搜索历史
- 搜索结果缓存
- NIP-50 高级扩展（`sentiment:`, `language:`, `domain:`）

## Acceptance Criteria

- [ ] `search "keyword"` 能返回匹配事件
- [ ] `search "keyword" --kinds 1` 过滤器正常工作
- [ ] search_relays 和本地 relay 同时作为搜索源
- [ ] 本地 relay 启用 NIP-50 支持后能响应搜索请求
- [ ] 结果去重、按时序排列
- [ ] 本地 relay 添加 NIP-50 (50) 到 SupportedNIPs
