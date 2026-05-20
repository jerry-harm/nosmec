# Research: readme-special-reply-kinds

- **Query**: Research the NIP README event kinds table and identify event kinds that define their own specialized reply/comment kind instead of using generic NIP-10 or NIP-22 behavior. Focus on kinds visible in the README table such as voice messages, torrents, public chat/channels, live chat, git, communities, and any other explicit reply/comment companion kinds.
- **Scope**: external
- **Date**: 2026-05-20

## Findings

### Files Found

| File Path | Description |
|---|---|
| `README.md` | Event kinds table listing root kinds and companion reply/comment kinds. |
| `A0.md` | Voice messages define dedicated root and reply kinds. |
| `35.md` | Torrents define a dedicated comment kind. |
| `28.md` | Public chat channels use channel message kind plus NIP-10 reply tags within that same kind. |
| `53.md` | Live activities define a dedicated live chat message kind. |
| `34.md` | Git objects use NIP-22 comments for replies; deprecated git reply kind still appears in README. |
| `72.md` | Communities use NIP-22 comments scoped to the community definition. |
| `29.md` | Relay groups expose deprecated threaded-reply kinds in the README table. |
| `C7.md` | Chats use same-kind replies via `q` tags, not a separate companion kind. |
| `A4.md` | Public messages explicitly have no thread model. |

### Code Patterns

No internal code search was required for the normative answer.

### External References

- [NIPs README](https://raw.githubusercontent.com/nostr-protocol/nips/master/README.md) — authoritative event-kinds table naming companion kinds.
- [NIP-A0](https://raw.githubusercontent.com/nostr-protocol/nips/master/A0.md) — `kind:1222` root voice message, `kind:1244` reply voice message, and replies MUST follow NIP-22 structure.
- [NIP-35](https://raw.githubusercontent.com/nostr-protocol/nips/master/35.md) — `kind:2003` torrent, `kind:2004` torrent comment, and comment behaves like `kind:1` with NIP-10 tags.
- [NIP-28](https://raw.githubusercontent.com/nostr-protocol/nips/master/28.md) — `kind:42` channel messages use NIP-10 root/reply tags; no separate reply kind.
- [NIP-53](https://raw.githubusercontent.com/nostr-protocol/nips/master/53.md) — `kind:1311` live chat message attached to live activity via `a` tag; `e` tag denotes direct parent reply.
- [NIP-34](https://raw.githubusercontent.com/nostr-protocol/nips/master/34.md) — replies to issues, patches, and PRs should follow NIP-22 comments; README still lists deprecated `kind:1622` git replies.
- [NIP-72](https://github.com/nostr-protocol/nips/blob/master/72.md) — communities use NIP-22 `kind:1111` comments rooted at `kind:34550` community definitions.
- [NIP-29](https://raw.githubusercontent.com/nostr-protocol/nips/master/29.md) — README shows deprecated relay-group threaded reply kinds `10` and `12`.
- [NIP-C7](https://raw.githubusercontent.com/nostr-protocol/nips/master/C7.md) — `kind:9` chats reply with another `kind:9` quoting parent via `q` tag.
- [NIP-A4](https://raw.githubusercontent.com/nostr-protocol/nips/master/A4.md) — `kind:24` public messages have no threads; `e` tags must not be used.

### Related Specs

- `.trellis/tasks/05-20-improve-event-detail-display-and-interaction/research/reply-semantics-standard-events.md` — baseline NIP-10 vs NIP-22 reply research.
- `.trellis/tasks/05-20-improve-event-detail-display-and-interaction/research/reply-semantics-special-events.md` — earlier special-event reply semantics notes.

## Findings

From the README event-kinds table, the event kinds that clearly define their own companion reply/comment kind are:

| Root kind | Companion reply/comment kind | Notes |
|---|---|---|
| `1222` Voice Message | `1244` Voice Message Comment | Dedicated reply kind from NIP-A0; reply event MUST follow NIP-22 structure. |
| `2003` Torrent | `2004` Torrent Comment | Dedicated comment kind from NIP-35; comment works like `kind:1` and follows NIP-10 tagging. |
| `30311` Live Event | `1311` Live Chat Message | Dedicated chat/message kind from NIP-53; replies use `e` as direct parent inside live chat. |

Other README-visible kinds in the requested scope do **not** define a distinct modern companion kind and instead use same-kind replies, generic comments, deprecated legacy kinds, or no threading at all:

| Root kind | Reply model |
|---|---|
| `40`/`41`/`42` Public chat channels | Channel discussion uses `kind:42`; replies are another `kind:42` with NIP-10 `root`/`reply` `e` tags, so no separate companion reply kind. |
| `9` Chat Message | Reply is another `kind:9` quoting parent with `q` tag per NIP-C7; no separate reply kind. |
| `24` Public Message | No thread model; `e` tags must not be used, so no reply companion kind. |
| `1617` Patch / `1618` Pull Request / `1621` Issue | Modern replies should use generic NIP-22 comments rather than a dedicated per-root reply kind. README still lists `1622` Git Replies as deprecated. |
| `34550` Community Definition | Uses generic NIP-22 `kind:1111` comments/posts scoped to the community definition per NIP-72; no dedicated community-only reply kind in the README table. |
| `4550` Community Post Approval | No dedicated reply kind found; community discussion still falls back to NIP-22 rules around the community/post scope. |
| `9000`-`9030` Group control / `39000-9` Group metadata | README exposes deprecated relay-group thread kinds `10` and `12` from NIP-29, but these are legacy group-thread companions rather than current generic reply behavior. |
| `30023` Long-form Content | Falls back to generic NIP-22 `kind:1111` comments, not a dedicated article-only reply kind. |

### Caveats / Not Found

- `NIP-72` was referenced from prior task research because a direct fetch timed out in this session, but the previously persisted notes already capture the relevant community reply semantics.
- The README table includes deprecated kinds (`10`, `12`, `1622`) alongside active ones, so “specialized reply kind” in the table does not always mean “current recommended mechanism.”
