# Research: NIP-22 Comment Replies

- **Query**: How kind:1111 comments work and differ from kind:1 NIP-10 replies
- **Scope**: external (NIP-22 spec)
- **Date**: 2026-05-16

## Source

- **NIP-22**: https://raw.githubusercontent.com/nostr-protocol/nips/master/22.md (fetched 2026-05-16)

---

## Findings

### 1. Kind:1111 Comments

NIP-22 defines `kind:1111` as a "comment" — a threading note that is **always scoped** to a root event or an `I`-tag (NIP-73 external identifier).

Key properties:
- Uses plaintext `.content` (no HTML, Markdown, or other formatting)
- Comments MUST point to the root scope using **uppercase** tag names
- Comments MUST point to the parent item using **lowercase** tag names
- Comments MUST NOT be used to reply to kind:1 notes — NIP-10 should be used instead

### 2. Tag Format

#### Root Scope (UPPERCASE tags)

| Tag | Meaning | Example |
|-----|---------|---------|
| `E` | Root event ID (regular event) | `["E", "<hex-id>", "<relay>", "<pubkey>"]` |
| `A` | Root addressable/replaceable event | `["A", "30023:<pubkey>:<d-tag>", "<relay>"]` |
| `I` | Root external identifier (URL, hashtag, podcast GUID, etc.) | `["I", "https://abc.com/articles/1"]` |
| `K` | Root event kind (MUST be present) | `["K", "<kind-integer>"]` |
| `P` | Root scope author pubkey | `["P", "<pubkey>", "<relay>"]` |

#### Parent Item (lowercase tags)

| Tag | Meaning | Example |
|-----|---------|---------|
| `e` | Parent event ID (regular event) | `["e", "<hex-id>", "<relay>", "<pubkey>"]` |
| `a` | Parent addressable/replaceable event | `["a", "30023:<pubkey>:<d-tag>", "<relay>"]` |
| `i` | Parent external identifier | `["i", "https://abc.com/articles/1"]` |
| `k` | Parent event kind (MUST be present) | `["k", "<kind-integer>"]` |
| `p` | Parent item author pubkey | `["p", "<pubkey>", "<relay>"]` |

### 3. Key Rule: `K` and `k` Tags Are Mandatory

> "Tags `K` and `k` MUST be present to define the event kind of the root and the parent items."

This ensures the client knows what kind of event is being commented on:
- `K` = kind of root scope
- `k` = kind of parent item

### 4. External Identifiers (I/i tags)

`I` and `i` tags create scopes for non-nostr entities:
- Hashtags
- Geohashes
- URLs
- Podcast GUIDs
- Other external identifiers defined in NIP-73

When related to an external identity, `k` tags use the same values as `i` tags (from NIP-73):
- `["k", "web"]` — for URL comments
- `["k", "podcast:item:guid"]` — for podcast comments

### 5. Comment on a Replaceable/Addressable Event

When the parent is a replaceable or addressable event:

```jsonc
{
  "tags": [
    // Root scope with A tag (address)
    ["A", "30023:<pubkey>:<d-tag>", "<relay>"],
    ["K", "30023"],
    // ALSO include an e tag referencing the specific event id
    ["e", "<event-id>", "<relay>"],
    // Parent matches root for top-level comments
    ["a", "30023:<pubkey>:<d-tag>", "<relay>"],
    ["k", "30023"]
  ]
}
```

> "when the parent event is replaceable or addressable, also include an `e` tag referencing its id"

### 6. Nested Comments (Reply to Comment)

A reply to an existing comment uses the same root scope but different parent:

```jsonc
{
  "tags": [
    // Same root as the comment being replied to
    ["E", "<root-event-id>", "<relay>", "<root-pubkey>"],
    ["K", "<root-kind>"],
    // Different parent — the comment being replied to
    ["e", "<parent-comment-id>", "<relay>", "<parent-pubkey>"],
    ["k", "1111"]  // parent is a comment
  ]
}
```

### 7. The "q" Tag (Quotes in Comments)

Same as NIP-10: `q` tags for citing events in `.content`:
```
["q", "<event-id> or <event-address>", "<relay-url>", "<pubkey-if-a-regular-event>"]
```

### 8. How NIP-22 Differs from NIP-10 (Kind:1) Replies

| Aspect | NIP-10 (kind:1 replies) | NIP-22 (kind:1111 comments) |
|--------|------------------------|----------------------------|
| **Event kind** | 1 (TextNote) | 1111 (Comment) |
| **Scope** | Only kind:1 events | Any kind: events, URLs, podcasts, hashtags, etc. |
| **Root marker** | `"root"` marker on e tag | `E`, `A`, or `I` tag + mandatory `K` tag |
| **Parent marker** | `"reply"` marker on e tag | `e`, `a`, or `i` tag + mandatory `k` tag |
| **Author tagging** | `p` tag | `P` (root author, uppercase) + `p` (parent author, lowercase) |
| **Kind knowledge** | Implicit (kind:1) | Explicit via `K` and `k` tags |
| **Cross-kind** | MUST NOT reply to other kinds | Designed for cross-kind comments |
| **External scope** | Not supported | Supported via I/i tags (NIP-73) |
| **Markers on e tags** | Yes (root/reply/mention) | No markers — uppercase/lowercase distinction instead |
| **p tag inheritance** | Required (inherit + add parent) | Not explicitly required (but should tag authors) |

### 9. Cross-Kind Rule (Critical)

**NIP-10**: "Kind 1 replies MUST NOT be used to reply to other kinds, use NIP-22 instead."

**NIP-22**: "Comments MUST NOT be used to reply to kind 1 notes. NIP-10 should instead be followed."

These two rules are complementary:
- Reply to a kind:1 note → use NIP-10 (kind:1 with marked e tags)
- Reply to any other kind → use NIP-22 (kind:1111 with K/k tags)

### 10. Current nosmec Handling of NIP-22

In `thread_treeview.go:276`:
```go
filter := nostr.Filter{
    Kinds: []nostr.Kind{nostr.KindTextNote, nostr.KindComment},
    Tags:  nostr.TagMap{"e": []string{rootID.Hex()}},
    Limit: 100,
}
```

- The filter includes `nostr.KindComment` (kind:1111) alongside kind:1
- This means comments are fetched as part of the thread replies
- However, NIP-22 comments use a **different tag structure** (E/A/I + K/k + e/a/i) than NIP-10 replies (e tags with root/reply markers)
- `extractParentID()` and `extractRootEvent()` only look for e tags with markers — they would NOT correctly resolve NIP-22 comment parent/root relationships
- For NIP-22 comments, the parent is identified by lowercase `e` tag (without marker), and the root by uppercase `E`/`A`/`I` tags

**This is a bug or missing feature**: NIP-22 comments are fetched but their thread relationships are not correctly resolved by the current tag parsing logic.

### 11. Full Tag Reference

A complete kind:1111 comment on a nostr event:

```jsonc
{
  "kind": 1111,
  "content": "<comment text>",
  "tags": [
    // === Root scope ===
    ["E", "<root-event-id>", "<relay-hint>", "<root-pubkey>"],
    // or: ["A", "<address>", "<relay>"],
    // or: ["I", "<external-id>", "<hint>"],
    ["K", "<root-kind>"],
    ["P", "<root-pubkey>", "<relay-hint>"],

    // === Parent item ===
    ["e", "<parent-event-id>", "<relay-hint>", "<parent-pubkey>"],
    // or: ["a", "<address>", "<relay>"],
    // or: ["i", "<external-id>", "<hint>"],
    ["k", "<parent-kind>"],
    ["p", "<parent-pubkey>", "<relay-hint>"]

    // === Optional quotes ===
    // ["q", "<event-id>", "<relay>", "<pubkey>"]
  ]
}
```

---

## Caveats / Not Found

- Full NIP-73 tag values for `i`/`I` and `k`/`K` tags (when related to external identities) were not fetched — see NIP-73 for the complete list.
- The interaction between NIP-22 comments and the outbox model (author relay lookup) was not covered in the NIP-22 spec directly — NIP-65 likely applies.
