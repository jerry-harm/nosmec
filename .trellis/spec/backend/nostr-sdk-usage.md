# nostr 库能力速查

## Key 操作 (nostr.PubKey / nostr.SecretKey)

```go
// PubKey - 32字节
pk := nostr.GetPublicKey(sk)                        // 从 SecKey 派生
pk := nostr.MustPubKeyFromHex("hex...")              // hex 解析 (panic)
pk := nostr.PubKeyFromHex("hex...")                 // hex 解析 (返回 error)

// SecretKey - 32字节
sk := nostr.Generate()                               // 生成新密钥
sk := nostr.MustSecretKeyFromHex("hex...")         // hex 解析 (panic)
sk := nostr.SecretKeyFromHex("hex...")              // hex 解析 (返回 error)

// 已有 IsValid32ByteHex / HexEncodeToString / HexDecodeString
```

## NIP-19 编码解码

```go
// Decode - 通用解码器，自动识别 prefix
prefix, value, err := nip19.Decode("npub1...")
// prefix: "npub"/"nsec"/"note"/"naddr"/"nevent"/"nprofile"/...
// value:  nostr.PubKey / nostr.SecretKey / nostr.EventID / nostr.ProfilePointer / ...

// 编码
nip19.EncodeNpub(pk)                                // → "npub1..."
nip19.EncodeNsec(sk)                                // → "nsec1..."
nip19.EncodeNevent(id, relays, author)               // → "nevent1..."
nip19.EncodeNaddr(pk, kind, identifier, relays)     // → "naddr1..."

// 解析指针 (npub/naddr/nevent 统一)
ptr, err := nip19.ToPointer("npub1...")             // → nostr.Pointer
// Pointer 是接口，可为 PubKey / EventID / ProfilePointer / EntityPointer
```

## Pointer 接口

```go
type Pointer interface {
    GetPointer() // 返回具体类型
}
// 已有函数从 tag 解析:
ProfilePointerFromTag(tag Tag) (ProfilePointer, error)
EventPointerFromTag(tag Tag) (EventPointer, error)
EntityPointerFromTag(refTag Tag) (EntityPointer, error)
```

## Event 结构

```go
type Event struct {
    ID        [32]byte
    PubKey   [32]byte
    CreatedAt Timestamp
    Kind     Kind
    Tags     TagMap  // tag[0] 是 name，如 "p", "e", "relay"
    Content  string
    Signature string
}
// Kind 常量: KindSetMetadata = 0, KindDMRelayList = 2263 等
```

## 已有验证/工具函数

```go
nostr.IsValid32ByteHex(hex string) bool
nostr.HexEncodeToString(src []byte) string
nostr.HexDecodeString(s string) ([]byte, error)
nostr.IsValidRelayURL(u string) bool
nostr.NormalizeURL(u string) (string, error)
nostr.ContainsPubKey(haystack []PubKey, needle PubKey) bool
nostr.ContainsID(haystack []ID, needle ID) bool
```

## Pool / Relay

```go
pool := nostr.NewPool(opts)
relay := nostr.RelayConnect(ctx, url, opts)
result := pool.QuerySingle(ctx, relays, filter, opts)
```

## Tag 解析

```go
// 从 TagMap 获取
tags.GetFirst("p")                                  // → string
tags.Get("p")                                       // → []string
tags.Has("e")                                       // → bool

// Tag[0] 是 name, Tag[1:] 是 values
```

## 重要提醒

**不要自行实现 hex→PubKey 解析**。用 `nostr.PubKeyFromHex` / `nostr.MustPubKeyFromHex`。

**nip19.Decode 是通用解码器**，能自动识别 npub/nsec/note/naddr/nevent/nprofile 等，返回 (prefix, value, err)。不需要先判断 prefix 再手动解析。值类型是 `any`，需要类型断言。

**Pointer 接口**已经统一了各种 NIP-19 指针类型的解析和存储。