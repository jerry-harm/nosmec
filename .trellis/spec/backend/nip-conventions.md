# NIP Conventions

> Project-level conventions for implementing NIP specifications.

---

## NIP Protocol Rule (Critical)

**Before implementing ANY nostr protocol behavior, fetch and read the relevant NIP specification.**

Protocol URL pattern: `https://github.com/nostr-protocol/nips/raw/refs/heads/master/{nip}.md`

```
NIP-01  → https://github.com/nostr-protocol/nips/raw/refs/heads/master/01.md
NIP-10  → https://github.com/nostr-protocol/nips/raw/refs/heads/master/10.md
NIP-65  → https://github.com/nostr-protocol/nips/raw/refs/heads/master/65.md
NIP-17  → https://github.com/nostr-protocol/nips/raw/refs/heads/master/17.md
NIP-50  → https://github.com/nostr-protocol/nips/raw/refs/heads/master/50.md
```

**Why**: "Common sense" assumptions about protocol behavior are frequently wrong. NIP specs are short, authoritative, and definitive.

**Process**:
1. Identify which NIP governs the feature
2. Fetch the spec (WebFetch or Task sub-agent)
3. Read the spec before writing any code
4. Reference spec in PRD/commits

---

## NIP-19 Format Convention

**All user-facing outputs MUST use NIP-19 bech32 format** (CLI, TUI, logs, error messages):

| Entity | Format | Function |
|--------|--------|----------|
| Public Key | `npub1...` | `nip19.EncodeNpub(pk)` |
| Event ID | `nevent1...` | `nip19.EncodeNevent(id, relays, author)` |
| Private Key | `nsec1...` | `nip19.EncodeNsec(sk)` (config only, never in output) |

**Internal storage**: Hex format is OK for DB/internal use.
**CLI output**: Always NIP-19.
**Command input**: Accept both hex (64-char) and NIP-19 formats. Use `nip19.ToPointer()` for NIP-19 decoding.

```go
// Input: accept both formats
pointer, err := nip19.ToPointer(input)
filter := pointer.AsFilter()

// Output: always NIP-19
fmt.Printf("npub: %s\n", nip19.EncodeNpub(pk))
fmt.Printf("note: %s\n", nip19.EncodeNevent(id, nil, pk))
```

---

## NIP-10 Reply Tag Generation (Full 5-Field)

Per [NIP-10](https://github.com/nostr-protocol/nips/blob/master/10.md), marked e tags use the full format:

```
["e", <event-id>, <relay-url>, <marker>, <pubkey>]
```

- `<relay-url>` — **SHOULD** be the relay where the referenced event was found
- `<pubkey>` — **OPTIONAL**; SHOULD be the hex pubkey of the referenced event's author
- Backward compatible: parsers read tag[1] (ID) and tag[3] (marker); tag[2] and tag[4] are additive

> **Warning — tag length safety**: Always check `len(tag) >= N` before accessing `tag[N]`. The 5-field format means tags can be longer than 4 fields. Existing parsers in `extractParentID` (tui/thread/thread.go) and `FindRootEvent` (utils/get.go) only read tag[1] and tag[3], so tag[4] addition does not break parsing.

### BuildReplyTags — Contract

```go
// Located in utils/post.go

// BuildReplyTags creates NIP-10 marked e tags for a reply to a parent event.
// Relay URLs are looked up from the event→relay tracking map.
// Pubkey is taken from parentEvent.PubKey.Hex().
func BuildReplyTags(parentEvent *nostr.Event) nostr.Tags
```

| Scenario | Returns |
|----------|---------|
| Direct reply (parent IS root) | 1 tag: `["e", rootID, relay, "root", pubkey]` |
| Nested reply (parent HAS root marker) | 2 tags: `["e", rootID, relay, "root"]` + `["e", parentID, relay, "reply", pubkey]` |
| Empty parent event | Empty tags |

**Root event pubkey**: For nested replies, the root event object is not available (only its ID from the parent's tags), so the root tag's `<pubkey>` field is left empty.

**Callers**: `ReplyToNote` (utils/post.go), `compose.AddReply` (tui/compose/model.go).

---

## NIP-65 Relay List (Kind 10002)

**Tag Rules** (from [NIP-65](https://github.com/nostr-protocol/nips/raw/refs/heads/master/65.md)):
- `["r", <url>]` — relay is **both read AND write** (no marker = both)
- `["r", <url>, "read"]` — relay is **read only**
- `["r", <url>, "write"]` — relay is **write only**
- **Never** create separate tags for the same relay with read AND write markers

Published via `utils.PublishRelayList(ctx, app)`.

**Parsing**: `nip65.ParseRelayList(event)` from `fiatjaf.com/nostr/nip65`. When reading, if `len(tag)==2` the relay is both read+write; otherwise check for "read"/"write" markers.

---

## NIP-17 DM Relay List (Kind 10050)

Published to advertise user's DM inbox relays:

```json
{
  "kind": 10050,
  "tags": [
    ["relay", "wss://inbox.nostr.wine"],
    ["relay", "wss://myrelay.nostr1.com"]
  ]
}
```

Published via `utils.PublishRelayList(ctx, app)` (same function handles NIP-65 and NIP-17).

---

## Hex String to nostr Type Conversion

**Always use SDK conversion functions** (never `copy()`):

```go
// Correct
id, err := nostr.IDFromHex(hexStr)
pk, err := nostr.PubKeyFromHex(hexStr)
sk := nostr.SecretKeyFromHex(hexStr)

// WRONG — causes garbage data
var id nostr.ID
copy(id[:], hexStr)
```

`nostr.ID` is `[32]byte` but hex strings are 64 characters. `copy(id[:], hexStr)` copies 64 ASCII bytes into 32 bytes, producing garbage.

---

## Filter Builder Validation

`nostr.IDFromHex()` accepts any 64-character string without error, even invalid hex like `"gggg..."`.

```go
var noteIDRegex = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)

func BuildNoteFilter(noteID string) (nostr.Filter, error) {
    if noteID == "" || !noteIDRegex.MatchString(noteID) {
        return nostr.Filter{}, ErrInvalidNoteID
    }
    id, err := nostr.IDFromHex(noteID)
    if err != nil {
        return nostr.Filter{}, ErrInvalidNoteID
    }
    return nostr.Filter{IDs: []nostr.ID{id}, Limit: 1}, nil
}
```

Always validate hex string format (64 chars, only 0-9a-fA-F) before calling `nostr.IDFromHex()`/`nostr.PubKeyFromHex()`.

---

## Supported NIPs

| NIP | Name | Kind | Status |
|-----|------|------|--------|
| 01 | Basic Protocol | - | ✓ |
| 02 | Follow List | 3 | ✓ |
| 05 | NIP-05 Verification | - | ✓ |
| 06 | Key Formats | - | ✓ |
| 10 | Reply Conventions | 1 | ✓ (5-field) |
| 17 | DM Relay List | 10050 | ✓ |
| 19 | Bech32 Entities | - | ✓ |
| 21 | `nostr:` URL Scheme | - | ✓ |
| 40 | Expiration Timestamp | - | ✓ |
| 44 | Encryption | - | ✓ |
| 51 | Lists | 10003, 10004, 10015 | ✓ |
| 65 | Relay List Metadata | 10002 | ✓ |
| 72 | Community Boards | 34550, 1111 | ✓ |
| 46 | Remote Signing | 24133 | Planned |
| 47 | Nostr Wallet Connect | - | Planned |