# NIP-50 Research

## NIP-50 协议理解

**核心机制**:
- Client 发送 REQ，filter 中带 `search` 字段（人类可读字符串，如 "best nostr apps"）
- Relay 返回匹配事件，按搜索质量排序（不是 `.created_at`）
- 支持 `key:value` 扩展（`kinds:`, `authors:`, `#t:` 等）
- NIP-50 搜索字符串可以包含多个过滤器组合

**协议规范** (`/home/jerry/code/nips/50.md`):
```jsonc
{
  "search": "best nostr apps",
  "kinds": [1, 3],
  "authors": ["npub1xxx"]
}
```

扩展选项：
- `include:spam` — 关闭 spam 过滤
- `domain:<domain>` — NIP-05 域名过滤
- `language:<code>` — 语言过滤
- `sentiment:<negative/neutral/positive>` — 情感过滤
- `nsfw:<true/false>` — NSFW 过滤

## SDK 支持情况

### fiatjaf.com/nostr Filter

**已有 `Search` 字段** (`filter.go:18`):
```go
type Filter struct {
    IDs     []ID
    Kinds   []Kind
    Authors []PubKey
    Tags    TagMap
    Since   Timestamp
    Until   Timestamp
    Limit   int
    Search  string  // ← NIP-50 已支持
}
```

`Filter.Matches()` 不处理 `Search` 字段（客户端过滤），但 `Pool.QuerySingle()` 会将 `Search` 字段发送给 relay。

### BoltDB Store

`boltdb.BoltBackend` **不支持** NIP-50 search（类似 cookbook 中 LMDB 的行为，会被 `NoSearchQueries` 策略拒绝）。

## khatru Local Relay

### 默认策略

khatru 默认 `NoSearchQueries` 策略**拒绝**搜索查询 (`khatru/policies/filters.go:49-54`):
```go
func NoSearchQueries(ctx context.Context, filter nostr.Filter) (reject bool, msg string) {
    if filter.Search != "" {
        return true, "search is not supported"
    }
    return false, ""
}
```

### 启用 NIP-50 搜索

khatru cookbook 提供了完整方案 (`khatru/docs/cookbook/search.md`):

**核心思路**: BoltDB 做普通查询 + Bleve 做全文搜索

```go
relay.UseEventstore(boltDB, 500)  // BoltDB 不支持 search

// 替换 QueryStored，当有 search 时使用 Bleve
relay.QueryStored = func(ctx context.Context, filter nostr.Filter) iter.Seq[nostr.Event] {
    if filter.Search != "" {
        return bleve.QueryEvents(filter)  // Bleve 支持 search
    }
    return boltDB.QueryEvents(filter)     // 普通查询
}
```

### Bleve Backend

`nostr/eventstore/bleve` 原生支持 NIP-50 搜索：
- `BleveBackend` 要求 filter 有 `Search` 字段
- `RawEventStore` 可以包装另一个 store（BoltDB）做底层存储

## 实现方案

### 两部分实现

#### 1. Search Relay 查询（客户端 → 外部 relay）

```go
// utils/search.go
filter := nostr.Filter{
    Kinds:  []nostr.Kind{1},          // 从命令行解析
    Search: "search term kinds:1",
    Limit:  50,
}
result := app.Pool().QuerySingle(ctx, searchRelays, filter, opts)
```

SDK Pool 会将 `Search` 字段发送给 relay。

#### 2. 本地 Relay NIP-50 支持

需要修改 `config/config.go` 的 `StartLocalRelay()`:
1. 引入 `bleve.BleveBackend`
2. 配置 `relay.QueryStored` 区分普通查询和搜索查询
3. 添加 NIP-50 到 relay 的 `SupportedNIPs`

```go
// 本地 relay 使用 BoltDB + Bleve 双存储
normal := &boltdb.BoltBackend{Path: dbPath}
search := &bleve.BleveBackend{Path: searchIndexPath, RawEventStore: normal}

relay.QueryStored = func(ctx context.Context, filter nostr.Filter) iter.Seq[nostr.Event] {
    if filter.Search != "" {
        return search.QueryEvents(filter)
    }
    return normal.QueryEvents(filter)
}
relay.StoreEvent = func(ctx context.Context, event nostr.Event) error {
    if err := normal.SaveEvent(event); err != nil {
        return err
    }
    return search.SaveEvent(event)
}
relay.DeleteEvent = func(ctx context.Context, id nostr.ID) error {
    if err := normal.DeleteEvent(id); err != nil {
        return err
    }
    return search.DeleteEvent(id)
}
```

### 搜索命令 CLI

```bash
nosmec search "keyword"                    # 全文搜索
nosmec search "keyword" --kinds 1,3     # 按 kind 过滤
nosmec search "keyword" --limit 50       # 限制结果
```

解析搜索字符串中的 `key:value` 过滤器。

## 技术约束

1. **Bleve 索引路径**: 需要单独的目录存储 search index
2. **NoSearchQueries 策略**: 本地 relay 需要移除此策略
3. **Relay SupportedNIPs**: 需要添加 NIP-50 (50) 到支持列表
4. **NIP-50 扩展**: 至少支持 `kinds:`, `authors:`, `#t:` 基础过滤器
