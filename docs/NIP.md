# 支持的 NIP 协议

## 完整支持

| NIP | 名称 | Kind | 说明 |
|-----|------|------|------|
| NIP-01 | Basic Protocol | - | 基础协议支持 |
| NIP-02 | Follow List | 3 | 关注列表 |
| NIP-04 | Encrypted Direct Messages | 4 | 加密私信 (已废弃) |
| NIP-05 | Mapping Public Keys to DNS-based Identifiers | - | NIP-05 验证 |
| NIP-06 | Basic Key Formats | - | nsec/npub 格式 |
| NIP-10 | Conventions for Replies | 1 | 笔记回复约定 |
| NIP-17 | Relay List for DMs | 10050 | DM Relay 列表 |
| NIP-19 | Bech32-encoded Entities | - | bech32 编码 |
| NIP-21 | `nostr:` URL Scheme | - | URL 协议 |
| NIP-40 | Expiration Timestamp | - | 过期时间戳 |
| NIP-44 | Encrypted Payloads | - | NIP-44 加密 |
| NIP-46 | Nostr Remote Signing | 24133 | 远程签名 (待支持) |
| NIP-51 | Lists | 10003, 10004, 10015 | 书签、社区、兴趣列表 |
| NIP-65 | Relay List Metadata | 10002 | Read/Write Relay 列表 |
| NIP-72 | Community Boards | 34550, 1111, 4550 | 社区板块 |

## 详细说明

### NIP-01: Basic Protocol

基础协议支持，包括：
- Event 格式
- Relay 通信
- 签名验证

### NIP-10: Conventions for Replies

`note reply` 命令使用标准回复约定：

```json
{
  "kind": 1,
  "tags": [
    ["e", "<parent-id>", "<relay>", "reply"],
    ["p", "<author-pubkey>"]
  ]
}
```

### NIP-17: Relay List for DMs

Kind 10050 事件格式：
```json
{
  "kind": 10050,
  "tags": [
    ["relay", "wss://dm-relay1.example.com"],
    ["relay", "wss://dm-relay2.example.com"]
  ]
}
```

### NIP-65: Relay List Metadata

Kind 10002 事件格式：
```json
{
  "kind": 10002,
  "tags": [
    ["r", "wss://relay.example.com", "read", "write"],
    ["r", "wss://relay2.example.com", "read"]
  ]
}
```

读写标记：
- `read`: 可接收事件
- `write`: 可发送事件
- 无标记: 等同于 read + write

### NIP-02: Follow List

Kind 3 事件用于关注列表：

```json
{
  "kind": 3,
  "tags": [
    ["p", "<pubkey>", "<relay>", "<petname>"],
    ["p", "91cf9..4e5ca", "wss://alicerelay.com/", "alice"]
  ],
  "content": ""
}
```

### NIP-51: Lists

NIP-51 定义了多种列表类型：

| Kind | 名称 | 标签 |
|------|------|------|
| 10003 | Bookmarks | `e`, `a` |
| 10004 | Communities | `a` (34550:...) |
| 10015 | Interests | `t` (hashtags) |

**Kind 10004 (Communities):**
```json
{
  "kind": 10004,
  "tags": [
    ["a", "34550:<author-pubkey>:<community-id>", "<relay>"]
  ],
  "content": ""
}
```

**Kind 10015 (Interests):**
```json
{
  "kind": 10015,
  "tags": [
    ["t", "nostr"],
    ["t", "bitcoin"]
  ],
  "content": ""
}
```

### NIP-72: Community Boards

Kind 34550 事件用于社区：

**Community Definition:**
```json
{
  "kind": 34550,
  "tags": [
    ["d", "<community-id>"],
    ["e", "<community-rules-id>"],
    ["p", "<moderator-pubkey>"]
  ],
  "content": "<name>\n<description>"
}
```

**Community Post:**
```json
{
  "kind": 1,
  "tags": [
    ["a", "34550:<community-id>", "<relay>"],
    ["e", "<root-event-id>", "<relay>", "root"],
    ["e", "<reply-to-id>", "<relay>", "reply"],
    ["p", "<community-author>"]
  ]
}
```

### NIP-46: Nostr Remote Signing (待支持)

支持远程签名者（bunker）连接：

**连接 URL 格式:**
```
nostrconnect://<client-pubkey>?relay=<wss://relay>&secret=<secret>&perms=sign_event,nip44_encrypt
bunker://<remote-signer-pubkey>?relay=<wss://relay>&secret=<optional-secret>
```

**支持的方法:**
- `connect` - 建立连接
- `get_public_key` - 获取用户公钥
- `sign_event` - 签名事件
- `nip44_encrypt` / `nip44_decrypt` - NIP-44 加密

**Kind 24133** 用于请求事件：
```json
{
  "kind": 24133,
  "pubkey": "<client-pubkey>",
  "content": "<NIP-44 encrypted request>",
  "tags": [["p", "<remote-signer-pubkey>"]]
}
```

## 未支持的 NIP

| NIP | 名称 | 说明 |
|-----|------|------|
| NIP-07 | `window.nostr` capability | 浏览器扩展，CLI 不适用 |
| NIP-26 | Delegated Event Signing | 委托签名 |
| NIP-47 | Nostr Wallet Connect | 钱包连接 |
| NIP-57 | Lightning Zaps | Zap 支付 |
| NIP-58 | Badges | 徽章系统 |
| NIP-59 | Gift Wraps | 礼物封装 |
| NIP-75 | Emoji Reactions | Emoji 反应 |
| NIP-78 | Application-specific Data | 应用数据 |
| NIP-89 | Handshake | 应用发现 |

## fiatjaf/nostr 库支持的方法

项目使用的 `fiatjaf.com/nostr` 库提供了以下包：

### nip17 - DM Relay List
```go
nip17.PrepareMessage()
nip17.ListenForMessages()
nip17.PublishMessage()
nip17.GetDMRelays()
```

### nip44 - Encrypted Payloads
```go
nip44.Encrypt(plaintext, recipientPubkey)
nip44.Decrypt(ciphertext, senderPubkey)
nip44.GenerateConversationKey()
```

### keyer - Key Signing
```go
keyer.NewPlainKeySigner(secretKey)
```

## Kind 常量参考

```go
KindProfileMetadata          Kind = 0
KindTextNote                 Kind = 1
KindFollowList               Kind = 3
KindEncryptedDirectMessage   Kind = 4
KindDeletion                 Kind = 5
KindRepost                   Kind = 6
KindReaction                 Kind = 7
KindDirectMessage            Kind = 14
KindCommunityPost            Kind = 1111
KindGiftWrap                 Kind = 1059
KindRelayListMetadata        Kind = 10002
KindBookmarks                Kind = 10003
KindCommunities              Kind = 10004
KindInterests                Kind = 10015
KindDMRelayList              Kind = 10050
KindCommunityPostApproval    Kind = 4550
KindCommunityDefinition      Kind = 34550
```

## Subscription Commands

`nosmec subscribe` 命令管理订阅（支持网络同步）：

```bash
# 添加订阅
nosmec subscribe add community 34550:abc123:dev-chat
nosmec subscribe add user npub1abc... --petname "Alice"
nosmec subscribe add hashtag nostr

# 列出订阅
nosmec subscribe list
nosmec subscribe list community

# 从网络同步（覆盖本地）
nosmec subscribe sync

# 发布到网络
nosmec subscribe publish
```
