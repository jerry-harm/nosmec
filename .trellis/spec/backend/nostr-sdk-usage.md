# nostr 库能力速查

## Key 操作 (nostr.PubKey / nostr.SecretKey)

```go
// PubKey - 32字节
pk := nostr.GetPublicKey(sk)                        // 从 SecKey 派生
pk := nostr.MustPubKeyFromHex("hex...")            // hex 解析 (panic)
pk := nostr.PubKeyFromHex("hex...")                 // hex 解析 (返回 error)
pk := nostr.PubKeyFromHexCheap("hex...")           // 快速解析，不验证有效性

// SecretKey - 32字节
sk := nostr.Generate()                             // 生成新密钥
sk := nostr.MustSecretKeyFromHex("hex...")         // hex 解析 (panic)
sk := nostr.SecretKeyFromHex("hex...")             // hex 解析 (返回 error)

// 已有工具
nostr.IsValid32ByteHex(hex string) bool
nostr.HexEncodeToString(src []byte) string
nostr.HexDecodeString(s string) ([]byte, error)
```

## NIP-19 编码解码

```go
// Decode - 通用解码器，自动识别 prefix
prefix, value, err := nip19.Decode("npub1...")
// prefix: "npub"/"nsec"/"note"/"naddr"/"nevent"/"nprofile"/...
// value:  nostr.PubKey / nostr.SecretKey / nostr.EventID / nostr.ProfilePointer / ...

// 编码
nip19.EncodeNpub(pk)                               // → "npub1..."
nip19.EncodeNsec(sk)                               // → "nsec1..."
nip19.EncodeNevent(id, relays, author)               // → "nevent1..."
nip19.EncodeNaddr(pk, kind, identifier, relays)     // → "naddr1..."
nip19.EncodeNprofile(pk, relays)                    // → "nprofile1..."
nip19.EncodePointer(ptr)                          // → 通用指针编码

// 解析指针 (npub/naddr/nevent/nprofile 统一)
ptr, err := nip19.ToPointer("npub1...")            // → nostr.Pointer
// Pointer 是接口，可为 PubKey / EventID / ProfilePointer / EntityPointer
```

## Pointer 接口体系

```go
// Pointer 接口 - 统一抽象
type Pointer interface {
    GetPointer()
}

// 三种具体 Pointer
type ProfilePointer struct {
    PublicKey PubKey
    Relays    []string
}
type EventPointer struct {
    ID     ID
    Relays []string
    Author PubKey
    Kind   Kind
}
type EntityPointer struct {
    PublicKey  PubKey
    Kind       Kind
    Identifier string
    Relays     []string
}

// 从 Tag 解析
ProfilePointerFromTag(tag Tag) (ProfilePointer, error)
EventPointerFromTag(tag Tag) (EventPointer, error)
EntityPointerFromTag(refTag Tag) (EntityPointer, error)

// Pointer → Filter 转换
pp.AsFilter()  // → Filter{Authors: []PubKey{pp.PublicKey}, ...}

// Pointer → Tag 转换
pp.AsTag()    // → Tag{"p", pk.Hex(), relays...}
pp.AsTagReference() // → "nostr:..." URI
```

## Event 结构

```go
type Event struct {
    ID        ID
    PubKey    PubKey
    CreatedAt Timestamp
    Kind      Kind
    Tags      Tags       // 不是 TagMap！
    Content   string
    Sig       [64]byte
}

// 重要方法
evt.CheckID()     // 验证 ID 正确性
evt.VerifySignature() bool
evt.Sign(secretKey [32]byte) error
evt.Serialize() []byte

// Kind 常量
KindSetMetadata      Kind = 0
KindTextNote         Kind = 1
KindDMRelayList      Kind = 2263
KindRelayListMetadata Kind = 10002
// 10000+: replaceable events
// 30000+: parameterized replaceable events
```

## Tags 结构 (注意不是 TagMap)

```go
type Tags []Tag       // Tag 是 []string

// 查找方法
tags.Find("p")                    // → Tag (第一个匹配的)
tags.FindWithValue("p", value)    // → Tag (带特定值的)
tags.FindLast("e")                // → Tag (最后一个匹配的)
tags.FindAll("e")                // → iter.Seq[Tag] (所有匹配的)
tags.Has("p")                     // → bool
tags.ContainsAny("e", values)     // → bool

// 获取值
tags.GetD()                      // → string (d tag 的值)

// Tag 是 []string，tag[0] 是 name，tag[1:] 是 values
// 例如: Tag{"p", "hexpubkey", "relay.url", "petname"}
```

## TagMap 结构

```go
type TagMap map[string][]string

// 直接用 key 获取值（Content JSON 解析时用）
tags.GetFirst("p")               // → string
tags.Get("p")                    // → []string
tags.Has("e")                    // → bool
```

## Filter 结构

```go
type Filter struct {
    IDs     []ID
    Kinds   []Kind
    Authors []PubKey
    Tags    TagMap     // 注意是 TagMap，不是 Tags！
    Since   Timestamp
    Until   Timestamp
    Limit   int
    Search  string
    LimitZero bool    // json:"-" 当 "limit":0 时设置
}

// 重要方法
filter.Matches(event Event) bool
filter.MatchesIgnoringTimestampConstraints(event Event) bool
filter.GetTheoreticalLimit() int
```

## Pool 查询

```go
pool := nostr.NewPool(opts)

// 单事件查询 (用于 replaceable events)
result := pool.QuerySingle(ctx, relays, filter, opts)
// 返回 *RelayEvent，ID 为零值表示未找到

// 多事件查询
events := pool.FetchMany(ctx, relays, filter, opts)  // chan RelayEvent
events, closed := pool.FetchManyNotifyClosed(...)    // + chan RelayClosed

// 订阅
events := pool.SubscribeMany(ctx, relays, filter, opts)  // chan RelayEvent
events, closed := pool.SubscribeManyNotifyClosed(...)
events, eose := pool.SubscribeManyNotifyEOSE(...)       // + EOSE 通知

// 带批量
events := pool.BatchedSubscribeMany(ctx, dfs, opts)  // 多个 Filter
```

## Subscription 结构

```go
type Subscription struct {
    Relay  *Relay
    Filter Filter
    Events chan Event           // 收到的事件
    EndOfStoredEvents chan struct{}  // EOSE 到达时关闭
    ClosedReason chan string   // CLOSED 消息
    Context context.Context     // subscription 结束时 Done()
}

// 方法
sub.Unsub()              // 取消订阅
sub.GetID() string       // 获取订阅 ID
```

## RelayEvent 结构

```go
type RelayEvent struct {
    Relay  *Relay
    Event Event
    GotEarlier bool  // 是否早于当前订阅收到
}
```

## nostr/sdk 包 (高级封装)

```go
// System - 全局核心对象，管理缓存、relay、dataloader
sys := sdk.NewSystem()
sys.Close()

// Profile 获取
pm := sys.FetchProfileMetadata(ctx, pubkey)           // 获取单用户 metadata
profiles := sys.SearchUsers(ctx, query)                // 搜索用户
pm := sys.FetchProfileFromInput(ctx, "npub1...")      // 从 npub/nip05 输入获取 profile
pm.Nprofile(ctx, sys, 3)                              // 生成 nprofile URI
pm.Npub() string                                       // npub 字符串
pm.NpubShort() string                                  // 缩短 npub
pm.NIP05Valid(ctx) bool                                // 验证 NIP-05

// Event 获取
evt, relays, err := sys.FetchSpecificEvent(ctx, pointer, params)
evt, relays, err := sys.FetchSpecificEventFromInput(ctx, input, params)

// 输入转换
pp := sdk.InputToProfile(ctx, "npub1...")            // npub / nip05 → ProfilePointer
ep := sdk.InputToEventPointer("nevent1...")           // nevent / note → EventPointer
pm, _ := sys.FetchProfileFromInput(ctx, "npub1...")  // 直接获取 profile metadata

// Feed / Stream
events, err := sys.FetchFeedPage(ctx, pubkeys, kinds, ...)
events, err := sys.StreamLiveFeed(ctx, pubkeys, kinds, ...)

// Relay List
relays := sys.FetchRelayList(ctx, pubkey)              // GenericList[string, Relay]
relays := sys.FetchInboxRelays(ctx, pubkey, n)
relays := sys.FetchOutboxRelays(ctx, pubkey, n)
relays := sys.FetchWriteRelays(ctx, pubkey)

// Follow / Mute List
follows := sys.FetchFollowList(ctx, pubkey)             // GenericList[nostr.PubKey, ProfileRef]
mutes := sys.FetchMuteList(ctx, pubkey)

// Event hints tracking
sys.TrackEventHints(ie nostr.RelayEvent)
sys.TrackEventHintsAndRelays(ie nostr.RelayEvent)

// ProfileMetadata - kind 0 event 解析结果
type ProfileMetadata struct {
    PubKey     nostr.PubKey   // 来源 PubKey
    Event     *nostr.Event    // 原始 event
    Name      string
    DisplayName string
    About     string
    Website   string
    Picture   string
    Banner    string
    NIP05     string
    LUD16     string
}
pm, _ := sdk.ParseMetadata(event)                       // 从 kind 0 event 解析

// ProfileRef - 联系人引用
type ProfileRef struct {
    Pubkey  nostr.PubKey
    Relay   string
    Petname string
}

// EventRef - 事件引用
type EventRef struct{ nostr.Pointer }

// Relay
type Relay struct {
    URL    string
    Inbox  bool
    Outbox bool
}

// RelayStream - 轮询 relay URL
rs := sdk.NewRelayStream(urls...)
url := rs.Next()                                       // 轮询获取下一个 URL
```

## nostr/keyer 包 (签名)

```go
// PlainKeySigner - 直接用私钥签名
kr := keyer.NewPlainKeySigner(secretKey)

// ReadOnlyUser / ReadOnlySigner - 只读用户
ru := keyer.NewReadOnlyUser(pk)
rs := keyer.NewReadOnlySigner(pk)

// BunkerSigner - NIP-46 bunker
// EncryptedKeySigner - 加密密钥签名
```

## 重要提醒

**不要自行实现 hex→PubKey 解析**。用 `nostr.PubKeyFromHex` / `nostr.MustPubKeyFromHex`。

**nip19.Decode 是通用解码器**，能自动识别 npub/nsec/note/naddr/nevent/nprofile 等，返回 (prefix, value, err)。不需要先判断 prefix 再手动解析。值类型是 `any`，需要类型断言。

**Tags vs TagMap**：Event.Tags 是 `Tags`（`[]Tag`），Filter.Tags 是 `TagMap`（`map[string][]string`）。这是历史设计差异，查找时用 `Tags.FindWithValue`，过滤时用 `Filter`。

**Pointer 接口**已经统一了各种 NIP-19 指针类型的解析和存储。

**优先使用 sdk.System**：如果需要 profile/feed/relay list 等功能，先看 System 是否有现成方法。