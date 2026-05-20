# Research: nip22-semantics

- **Query**: Re-read NIP-22 carefully and extract its exact semantics for when a client should create kind:1111 comments, how root scope vs parent item are represented, and whether NIP-22 is intended as the general reply mechanism for arbitrary non-kind:1 content or only select cases. Also note any explicit exclusions or scope limits.
- **Scope**: external
- **Date**: 2026-05-20

## Findings

### Files Found

| File Path | Description |
|---|---|
| `https://raw.githubusercontent.com/nostr-protocol/nips/master/22.md` | Canonical NIP-22 spec text used for this reading |

### Code Patterns

No internal code inspection was required for this topic.

### External References

- [NIP-22](https://github.com/nostr-protocol/nips/blob/master/22.md) — authoritative comment semantics for `kind:1111`
- [NIP-73](https://github.com/nostr-protocol/nips/blob/master/73.md) — defines `i`/`I` external-identifier value families referenced by NIP-22
- [NIP-10](https://github.com/nostr-protocol/nips/blob/master/10.md) — explicit exclusion target for `kind:1` note replies

### Related Specs

- `.trellis/tasks/05-20-improve-event-detail-display-and-interaction/prd.md` — current task assumptions and evolving requirements
- `.trellis/tasks/05-20-improve-event-detail-display-and-interaction/research/reply-semantics-standard-events.md` — prior task research that this note refines

## Findings

### Exact NIP-22 semantics

NIP-22 defines a **comment** as a threading note "always scoped to a root event or an `I`-tag" and says it **uses `kind:1111`** with plaintext content.

It draws a strict distinction between:

- **Root scope**: represented with **uppercase** tags: `K` plus one of `E` / `A` / `I`
- **Parent item**: represented with **lowercase** tags: `k` plus one of `e` / `a` / `i`

The spec text says comments:

> MUST point to the root scope using uppercase tag names (e.g. `K`, `E`, `A` or `I`)

and:

> MUST point to the parent item with lowercase ones (e.g. `k`, `e`, `a` or `i`)

It also says:

> Tags `K` and `k` MUST be present to define the event kind of the root and the parent items.

For authored nostr events, author tags are also required:

> Comments MUST point to the authors when one is available (i.e. tagging a nostr event). `P` for the root scope and `p` for the author of the parent item.

### Top-level comment vs reply-to-comment

For a **top-level comment**, the parent item is the root item, so both root and parent references are present and usually refer to the same underlying object.

For a **reply to a comment**, the root uppercase tags still point to the original root scope, while the lowercase parent tags point to the immediate parent comment (typically `e` + `k=1111` + `p`).

This means NIP-22 is not just "comment on event X"; it is a **two-level threading model**:

- root = overall discussion scope
- parent = immediate reply target

### What objects NIP-22 covers

NIP-22 clearly supports comments on:

- event ids via `E` / `e`
- addressable events via `A` / `a`
- external identifiers via `I` / `i`

The examples include comments on:

- a blog post (`kind:30023`)
- a NIP-94 file (`kind:1063`)
- a website URL (`K=web`, `I`/`i` tags)
- a podcast item using a NIP-73 external identifier

So NIP-22 is **not limited to a tiny special-case subset** like only long-form articles or only communities. It is a general comment format for many non-`kind:1` scopes, including external resources.

### Important scope limit / correction to over-broad reading

NIP-22 does **not** say "all arbitrary non-`kind:1` content should always be replied to with `kind:1111`".

What it does say is that `kind:1111` is the format for comments scoped to a root event or `I`-tag, with explicit examples across several targets. That is broader than a narrow interpretation, but still **bounded by representability**:

- the root must be representable as uppercase root tags (`E`/`A`/`I` + `K`)
- the parent must be representable as lowercase parent tags (`e`/`a`/`i` + `k`)
- author tags are required when the target is a nostr event with an author

So the correction is:

- **too narrow**: "NIP-22 is only for a few select cases like communities or articles" → false
- **too broad**: "NIP-22 is automatically the reply mechanism for every non-`kind:1` event whatsoever" → not stated by the spec

The spec supports NIP-22 as a **general-purpose non-note comment model where the root/parent can be expressed in NIP-22's tag system**, not as an explicit blanket rule for every possible event kind.

### Explicit exclusion

NIP-22 contains one direct exclusion:

> Comments MUST NOT be used to reply to kind 1 notes. NIP-10 should instead be followed.

That is the clearest hard boundary in the document.

## Caveats / Not Found

- NIP-22 is marked `draft` and `optional` in the header.
- The document does not provide a universal decision table for every Nostr event kind; it defines the `kind:1111` mechanism and examples, plus the explicit `kind:1` exclusion.
- The spec does not state that wrapper/moderation/repost-like events should always be treated as direct NIP-22 roots; that requires separate protocol or product interpretation outside the exact NIP-22 text.
