# Research: NIP-10 Complete Reply Format

- **Query**: Full NIP-10 specification for kind:1 reply/thread tags
- **Scope**: external (NIP specs)
- **Date**: 2026-05-16

## Source

- **NIP-10**: https://raw.githubusercontent.com/nostr-protocol/nips/master/10.md (fetched 2026-05-16)
- **NIP-01**: https://raw.githubusercontent.com/nostr-protocol/nips/master/01.md (fetched 2026-05-16)

---

## Findings

### 1. Basic Tag Format (from NIP-01)

The `e` tag per NIP-01:
```
["e", <32-bytes lowercase hex of the id of another event>, <recommended relay URL, optional>, <32-bytes lowercase hex of the author's pubkey, optional>]
```

- Position 0: tag key `"e"`
- Position 1: event ID (32-byte lowercase hex)
- Position 2: recommended relay URL (optional, may be `""`)
- Position 3: author pubkey (optional, 32-byte lowercase hex)

The `p` tag per NIP-01:
```
["p", <32-bytes lowercase hex of a pubkey>, <recommended relay URL, optional>]
```

---

### 2. Marked "e" Tags (PREFERRED / current standard)

Per NIP-10, marked e tags are the **preferred** (current) way to tag thread relationships:

```
["e", <event-id>, <relay-url>, <marker>, <pubkey>]
```

| Position | Meaning | Required |
|----------|---------|----------|
| 0 | Tag key: `"e"` | yes |
| 1 | Event ID (32-byte hex) | yes |
| 2 | Recommended relay URL | SHOULD add valid URL, may be `""` |
| 3 | Marker: `"root"`, `"reply"`, or `"mention"` | optional |
| 4 | Author pubkey of referenced event | SHOULD be set (used in outbox model) |

**Markers:**

| Marker | Meaning |
|--------|---------|
| `"root"` | The root event of the thread (top-most ancestor). |
| `"reply"` | The direct parent event being replied to. |
| `"mention"` | A referenced event that is NOT part of the reply chain (a citation or mention). |

**Sorting rule**: e tags SHOULD be sorted by the reply stack from root to the direct parent.

---

### 3. Direct Reply to Root

A **direct reply** to the root of a thread should have a **single** marked e tag with marker `"root"`:

```json
{
  "tags": [
    ["e", "<root-event-id>", "<relay-url>", "root", "<root-pubkey>"]
  ]
}
```

**Key rule**: For top-level replies (replying directly to the root event), only the `"root"` marker should be used. There should be NO `"reply"` tag since the root IS the direct parent.

---

### 4. Nested Reply (Reply to a Reply)

A **nested reply** (reply to a reply) should have **two or more** marked e tags:

```json
{
  "tags": [
    ["e", "<root-event-id>", "<relay-url>", "root", "<root-pubkey>"],
    ["e", "<parent-event-id>", "<relay-url>", "reply", "<parent-pubkey>"],
    // optional mentions:
    ["e", "<mentioned-event-id>", "", "mention"],
    ...
  ]
}
```

**Tag order**: root → (mentions, in any order) → reply (direct parent) MUST be last.

---

### 5. Root Event

A root event (the original note that starts a thread) should have NO e tags:

```json
{
  "tags": [
    // no e tags at all
    // may have p tags for mentions
  ]
}
```

> If a root event has e tags, they should be `"mention"` markers only — NOT `"root"` or `"reply"`.

---

### 6. Mentions

A mention tag (`"mention"` marker) represents an event that is referenced but is NOT part of the reply chain:

```json
["e", "<mentioned-event-id>", "<relay>", "mention"]
```

Multiple mention tags are allowed. They can appear between the root and reply tags.

---

### 7. The "q" Tag (Quotes)

Quotes use the `q` tag per NIP-21:

```
["q", "<event-id> or <event-address>", "<relay-url>", "<pubkey-if-a-regular-event>"]
```

- Unlike e tags, q tags do NOT have markers.
- `q` tags are for citing events in the `.content` field (inline quotes).
- Authors of q tags SHOULD be added as `p` tags.

---

### 8. The "p" Tag Rules for Replies

NIP-10 specifies rules for p tags in reply events:

**When replying to event E authored by `a1` with p tags [`p1`, `p2`, `p3`]:**
- The reply's p tags should contain: [`a1`, `p1`, `p2`, `p3`]
- In no particular order

**Effectively**: When replying, inherit ALL p tags from the parent event AND add the parent's author as a new p tag.

**Purpose**: p tags are used to notify pubkeys that they have been involved in a thread. This is the notification mechanism for nostr.

**Additional rule from NIP-10**: "Authors of the `e` and `q` tags SHOULD be added as `p` tags to notify of a new reply or quote."

---

### 9. Deprecated Positional "e" Tags (LEGACY)

The old scheme (deprecated, but still encountered on the network):

```
["e", <event-id>, <relay-url>]   // no marker at position 3
```

Positional semantics (by index in the e tags array, NOT by marker):

| # e tags | Tag[0] | Tag[1] | Tag[2...N-2] | Tag[N-1] |
|----------|--------|--------|--------------|----------|
| 0 | — | — | — | — | (not a reply) |
| 1 | reply-id | — | — | — | (reply) |
| 2 | root-id | reply-id | — | — |
| 3+ | root-id | mention-id | ...mentions | reply-id |

**Why deprecated**: "Positional e tags create ambiguities that are difficult, or impossible to resolve when an event references another but is not a reply."

**How to handle in code**: When parsing e tags, if position 3 is empty/missing, treat as positional. Recommended approach: use `"root"` marker for position 0, `"reply"` marker for last position, `"mention"` for middle positions.

---

### 10. Constraints / Edge Cases

| Rule | Source |
|------|--------|
| Kind 1 replies MUST NOT be used to reply to other kinds — use NIP-22 (kind:1111) instead | NIP-10 |
| `e` tags should be sorted root → ... → reply | NIP-10 |
| Markup languages (markdown, HTML) SHOULD NOT be used in `.content` | NIP-10 |
| p tags should inherit from parent + add parent author | NIP-10 |
| Only the first value in any given tag is indexed by relays (NIP-01) | NIP-01 |
| Single-letter tags (a-z, A-Z) are expected to be indexed by relays | NIP-01 |

---

### 11. Comparison with Current nosmec Implementation

The current nosmec implementation (`tui/window/event/thread_treeview.go`) handles NIP-10 as follows:

**`extractParentID` (line 22-50)**:
- Looks for e tags with marker `"reply"` → returns that value as parent
- Falls back to `"root"` marker (if root != self) → treats as direct reply
- Does NOT handle `"mention"` markers — ignores them (correct behavior)

**`extractRootEvent` (line 54-107)**:
- No e tags → event IS root ✓
- Has `"reply"` + `"root"` → returns root from `"root"` marker ✓
- Has `"reply"` but no `"root"` → treats event as root (questionable — should this instead return an error or treat the first e tag as root?)
- Has `"root"` but no `"reply"` → event IS root (questionable — this is a DIRECT reply per NIP-10, the event is NOT root)
- Has e tags but no markers → treats as root

**Issues in current implementation**:
1. `extractRootEvent` incorrectly treats events with `"root"` marker (no `"reply"`) as root events. Per NIP-10, a single `"root"` marker means this is a **direct reply** — the event IS NOT root.
2. The `extractParentID` fallback (root as parent for direct replies) is correct per NIP-10.

**Note**: There is also `utils.FindRootEvent()` in `utils/get.go:223-272` with slightly different logic — see the duplicate logic issue documented in `thread-implementation.md`.

---

## Caveats / Not Found

- The NIP-10 spec mentions `"mention"` marker exists but doesn't elaborate extensively on its usage. The positional section is the primary source for mention semantics.
- `FindAll("e")` in the nostr library returns raw tags — the implementation must check tag length before accessing index 3 to avoid panics.
