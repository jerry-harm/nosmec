# NIP Protocol Reference

## Supported NIPs

| NIP | Name | Kind | Status |
|-----|------|------|--------|
| NIP-01 | Basic Protocol | - | ✅ Supported |
| NIP-02 | Follow List | 3 | ✅ Supported |
| NIP-05 | Identifier Lookup | - | ✅ Supported |
| NIP-06 | Key Formats | - | ✅ Supported |
| NIP-10 | Reply Conventions | 1 | ✅ Supported |
| NIP-17 | Relay List for DMs | 10050 | ✅ Supported |
| NIP-19 | Bech32 Encoding | - | ✅ Supported |
| NIP-21 | `nostr:` URL Scheme | - | ✅ Supported |
| NIP-40 | Expiration Timestamp | - | ✅ Supported |
| NIP-44 | Encrypted Payloads v2 | - | ✅ Supported |
| NIP-46 | Remote Signing | 24133 | 🔜 Planned |
| NIP-47 | Nostr Wallet Connect | - | 🔜 Planned |
| NIP-51 | Lists | 10003, 10004, 10015 | ✅ Supported |
| NIP-65 | Relay List Metadata | 10002 | ✅ Supported |
| NIP-72 | Community Boards | 34550, 1111, 4550 | ✅ Supported |

---

# fiatjaf.com/nostr SDK Reference

> This section documents the NIP packages available in the `fiatjaf.com/nostr` SDK and their recommended usage.

## Package Index

| Package | NIP | Purpose |
|---------|-----|---------|
| [nip17](#nip17---direct-messages) | NIP-17 | Complete DM functionality |
| [nip19](#nip19---bech32-encoding) | NIP-19 | Encode/decode npub, nsec, note, nevent |
| [nip40](#nip40---expiration-timestamp) | NIP-40 | Parse expiration tags |
| [nip42](#nip42---authentication) | NIP-42 | Relay authentication |
| [nip44](#nip44---encryption-v2) | NIP-44 | v2 encryption (chacha20-poly1305) |
| [nip04](#nip04---encryption-v1) | NIP-04 | v1 encryption (AES-256-CBC) - **deprecated** |
| [nip57](#nip57---zaps) | NIP-57 | Zap receipt parsing |
| [nip59](#nip59---gift-wrap) | NIP-59 | Gift-wrapped sealed messages |
| [nip65](#nip65---relay-list-metadata) | NIP-65 | Relay list metadata parsing |
| [nip10](#nip10---thread-conventions) | NIP-10 | Thread/reply conventions |

---

## nip17 - Direct Messages

**NIP:** NIP-17  
**Purpose:** Complete DM functionality with gift-wrapped messages.

```go
import "fiatjaf.com/nostr/nip17"

// Get DM relays for a pubkey from the network
func GetDMRelays(ctx context.Context, pubkey nostr.PubKey, pool *nostr.Pool, relaysToQuery []string) []string

// Sends a DM to both our relays and recipient's relays
func PublishMessage(ctx context.Context, content string, tags nostr.Tags, pool *nostr.Pool, ourRelays []string, theirRelays []string, kr nostr.Keyer, recipientPubKey nostr.PubKey, modify func(*nostr.Event)) error

// Prepares gift-wrapped messages (toUs and toThem)
func PrepareMessage(ctx context.Context, content string, tags nostr.Tags, kr nostr.Keyer, recipientPubKey nostr.PubKey, modify func(*nostr.Event)) (toUs nostr.Event, toThem nostr.Event, error)

// Returns a channel with decrypted DMs
func ListenForMessages(ctx context.Context, pool *nostr.Pool, kr nostr.Keyer, ourRelays []string, since nostr.Timestamp) chan nostr.Event
```

**Event Kinds:**
- `KindGiftWrap` (1059) - The wrapped envelope
- `KindSeal` (13) - Inner seal containing encrypted rumor
- `KindDirectMessage` (4) - The actual DM content

**When to Use:**
- When sending or receiving direct messages
- Replaces manual GiftWrap/GiftUnwrap implementations

**Current Status:** ⚠️ Has manual implementation in `utils/dm.go` - should migrate to this package.

---

## nip19 - Bech32 Encoding

**NIP:** NIP-19  
**Purpose:** Encode and decode bech32 formatted Nostr entities.

```go
import "fiatjaf.com/nostr/nip19"

// Decode any bech32 string (nsec, note, npub, nprofile, nevent, naddr)
func Decode(bech32string string) (prefix string, value any, err error)

// Encode functions
func EncodeNsec(sk [32]byte) string
func EncodeNpub(pk nostr.PubKey) string
func EncodeNprofile(pk nostr.PubKey, relays []string) string
func EncodeNevent(id nostr.ID, relays []string, author nostr.PubKey) string
func EncodeNaddr(pk nostr.PubKey, kind nostr.Kind, identifier string, relays []string) string

// Utilities
func ToPointer(code string) (nostr.Pointer, error)
```

**When to Use:**
- Converting hex keys to npub/nsec format for display
- Parsing npub/nsec/nevent/nprofile/naddr strings from user input

**Current Status:** ✅ Already used in `utils/profile.go` via `nip19.EncodeNpub()`.

---

## nip40 - Expiration Timestamp

**NIP:** NIP-40  
**Purpose:** Parse the expiration tag from events.

```go
import "fiatjaf.com/nostr/nip40"

// Returns expiration timestamp or -1 if no valid expiration tag exists
func GetExpiration(tags nostr.Tags) nostr.Timestamp
```

**When to Use:**
- When displaying event expiration times
- When filtering events by expiration

**Current Status:** ❌ Not used.

---

## nip42 - Authentication

**NIP:** NIP-42  
**Purpose:** Handle relay authentication challenges.

```go
import "fiatjaf.com/nostr/nip42"

// Creates an unsigned auth event for a given challenge
func CreateUnsignedAuthEvent(challenge string, pubkey nostr.PubKey, relayURL string) nostr.Event

// Validates an auth event and returns the pubkey if valid
func ValidateAuthEvent(event nostr.Event, challenge string, relayURL string) (nostr.PubKey, error)

// Extracts relay URL from auth event tags
func GetRelayURLFromAuthEvent(event nostr.Event) string
```

**Event Kinds:**
- `KindClientAuthentication` (22242)

**When to Use:**
- When connecting to relays that require authentication
- When handling `auth-required` errors

**Current Status:** ❌ Not used.

---

## nip44 - Encryption (v2)

**NIP:** NIP-44  
**Purpose:** Modern encryption standard using chacha20-poly1305 with HKDF key derivation.

```go
import "fiatjaf.com/nostr/nip44"

// Encrypts plaintext with a conversation key
func Encrypt(plaintext string, conversationKey [32]byte, opts ...func(*encryptOptions)) (string, error)

// Decrypts ciphertext
func Decrypt(b64ciphertextWrapped string, conversationKey [32]byte) (string, error)

// Generates a conversation key from recipient pubkey and sender secret key
func GenerateConversationKey(pub nostr.PubKey, sk nostr.SecretKey) ([32]byte, error)
```

**Security:** NIP-44 uses:
- chacha20-poly1305 for authenticated encryption
- HKDF for key derivation
- Built-in padding to prevent length leakage

**Current Status:** ✅ Used transitively through `nip59`.

---

## nip04 - Encryption (v1) - DEPRECATED

**⚠️ NIP-04 is deprecated.** It has security issues:
- No authentication (no MAC)
- Length leakage through padding
- CBC mode is less secure

```go
import "fiatjaf.com/nostr/nip04"

func ComputeSharedSecret(pub nostr.PubKey, sk [32]byte) ([]byte, error)
func Encrypt(message string, key []byte) (string, error)
func Decrypt(content string, key []byte) (string, error)
```

**When to Use:** Only for interoperability with legacy clients. Default to NIP-44.

**Current Status:** ❌ Not used.

---

## nip57 - Zaps

**NIP:** NIP-57  
**Purpose:** Parse Lightning zap receipts.

```go
import "fiatjaf.com/nostr/nip57"

// Extracts zap amount in millisats from a zap receipt event
func GetAmountFromZap(event nostr.Event) uint64
```

**Event Kinds:**
- `KindZap` (9735)

**Current Status:** ❌ Not used.

---

## nip59 - Gift Wrap

**NIP:** NIP-59  
**Purpose:** Creates sealed, encrypted envelopes for private DMs.

```go
import "fiatjaf.com/nostr/nip59"

// Wraps a rumor event in a seal, then in a gift envelope
func GiftWrap(rumor nostr.Event, recipient nostr.PubKey, encrypt func(plaintext string) (string, error), sign func(*nostr.Event) error, modify func(*nostr.Event)) (nostr.Event, error)

// Unwraps a gift-wrapped event
func GiftUnwrap(gw nostr.Event, decrypt func(otherpubkey nostr.PubKey, ciphertext string) (string, error)) (rumor nostr.Event, err error)
```

**How It Works:**
1. Takes a DM event (the "rumor")
2. Encrypts it with the sender's key (creating a "seal")
3. Encrypts the seal with a nonce key and signs it (creating the "gift-wrap")
4. The gift-wrap is published to relays

**Event Kinds:**
- `KindGiftWrap` (1059) - Published to relays
- `KindSeal` (13) - Internal, encrypted once
- `KindDirectMessage` (4) - The actual content

**Current Status:** ✅ Already used in `utils/dm.go` for GiftWrap/GiftUnwrap.

---

## nip65 - Relay List Metadata

**NIP:** NIP-65  
**Purpose:** Parse relay list metadata from kind 10002 events.

```go
import "fiatjaf.com/nostr/nip65"

// Parses a NIP-65 relay list event and returns separate read/write relay lists
func ParseRelayList(event nostr.Event) (readRelays []string, writeRelays []string)
```

**Implementation Notes:**
- Validates relay URLs
- Normalizes URLs
- Handles `r` tags with optional `read`/`write` markers
- If no marker is present, relay is added to both lists

**Event Kinds:**
- `KindRelayListMetadata` (10002)

**Current Status:** ⚠️ Has manual implementation in `utils/relay_list.go` - should use this package.

---

## nip10 - Thread Conventions

**NIP:** NIP-10  
**Purpose:** Parse thread/reply structure from event tags.

```go
import "fiatjaf.com/nostr/nip10"

// Gets the root event of a thread (the first event in the chain)
func GetThreadRoot(tags nostr.Tags) nostr.Pointer

// Gets the immediate parent event (direct reply target)
func GetImmediateParent(tags nostr.Tags) nostr.Pointer
```

**Tag Markers:**
- `root` - Marks the root event of a thread
- `reply` - Marks direct parent (immediate reply)
- `mention` - Invalidates as parent candidate

**Current Status:** ⚠️ Replies are built manually in `utils/post.go` - should use this package.

---

## Kind Constants Reference

```go
KindProfileMetadata          Kind = 0
KindTextNote                 Kind = 1
KindFollowList               Kind = 3
KindEncryptedDirectMessage   Kind = 4
KindDeletion                 Kind = 5
KindRepost                   Kind = 6
KindReaction                 Kind = 7
KindSeal                    Kind = 13
KindDirectMessage           Kind = 14
KindGiftWrap                Kind = 1059
KindZap                     Kind = 9735
KindClientAuth              Kind = 22242
KindRelayListMetadata       Kind = 10002
KindBookmarks               Kind = 10003
KindCommunities             Kind = 10004
KindInterests               Kind = 10015
KindDMRelayList             Kind = 10050
KindCommunityPostApproval    Kind = 4550
KindCommunityDefinition      Kind = 34550
KindCommunityPost           Kind = 1111
```

---

## Migration Priority

| Priority | File | Current | Recommended |
|----------|------|---------|-------------|
| **High** | `utils/dm.go` | ✅ Using nip59 GiftWrap | Consider `nip17` for full DM lifecycle |
| **High** | `utils/relay_list.go` | Manual NIP-65 parsing | Use `nip65.ParseRelayList` |
| **Medium** | `utils/post.go` | Manual thread tags | Use `nip10` |
| **Low** | Global | - | Add `nip40` for expiration |
| **Low** | Global | - | Add `nip57` for zap receipts |

---

## Additional Available Packages

The SDK includes many more NIP packages:

| Package | NIP | Purpose |
|---------|-----|---------|
| `nip05` | NIP-05 | Identifier resolution |
| `nip06` | NIP-06 | BIP-39 seed phrases |
| `nip11` | NIP-11 | Relay information |
| `nip45` | NIP-45 | Hyper Log Log filters |
| `nip46` | NIP-46 | Nostr Connect (remote signer) |
| `nip47` | NIP-47 | NWC (Nostr Wallet Connect) |
| `nip53` | NIP-53 | Live Activities |
| `nip54` | NIP-54 | Wikis (long-form) |
| `nip58` | NIP-58 | Badges |
| `nip78` | NIP-78 | Application data |
| `nip84` | NIP-84 | Main relay hint |
| `nip94` | NIP-94 | File sharing |
| `nip90` | NIP-90 | Data Vending Machine |

---

## NIP-17 DM Event Format

Kind 10050 (DM Relay List):
```json
{
  "kind": 10050,
  "tags": [
    ["relay", "wss://dm-relay1.example.com"],
    ["relay", "wss://dm-relay2.example.com"]
  ]
}
```

## NIP-65 Relay List Format

Kind 10002 (Relay List Metadata):
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
- `read`: Can receive events
- `write`: Can send events
- No marker: Equivalent to read + write

## NIP-10 Reply Convention

```json
{
  "kind": 1,
  "tags": [
    ["e", "<parent-id>", "<relay>", "reply"],
    ["p", "<author-pubkey>"]
  ]
}
```

## NIP-02 Follow List

Kind 3:
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

## NIP-51 Lists

| Kind | Name | Tags |
|------|------|------|
| 10003 | Bookmarks | `e`, `a` |
| 10004 | Communities | `a` (34550:...) |
| 10015 | Interests | `t` (hashtags) |

## NIP-72 Community Boards

**Community Definition (Kind 34550):**
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

**Community Post (Kind 1 with `a` tag):**
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
