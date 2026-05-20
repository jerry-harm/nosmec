# Research: reply-semantics-special-events

- **Query**: Research which Nostr event kinds have standardized reply semantics in NIPs for reposts, generic reposts, addressable events, community-definition-related events, and any event kinds where replying is undefined or discouraged.
- **Scope**: external
- **Date**: 2026-05-20

## Findings

### Files Found

| File Path | Description |
|---|---|
| `NIP-10` | Standard reply semantics for `kind:1` text notes only. |
| `NIP-18` | Reposts (`kind:6`) and generic reposts (`kind:16`); quote handling via `q` tags. |
| `NIP-22` | Standard comment / reply semantics for non-`kind:1` roots using `kind:1111`. |
| `NIP-72` | Community posting/reply rules built on NIP-22, rooted in community definition `kind:34550`. |
| `NIP-01` | Base definitions for `a` tags, replaceable, and addressable events. |

### Code Patterns

No internal code search was required for the normative answer. Relevant project docs mention NIP-10 reply tag handling at `.trellis/spec/backend/relay-guidelines.md:67` and `.trellis/spec/backend/nip-conventions.md:57`.

### External References

- [NIP-10](https://github.com/nostr-protocol/nips/blob/master/10.md) — `kind:1` replies only; explicitly says `kind:1` replies MUST NOT be used for other kinds.
- [NIP-18](https://github.com/nostr-protocol/nips/blob/master/18.md) — reposts, generic reposts, and quote-repost `q` tag semantics.
- [NIP-22](https://github.com/nostr-protocol/nips/blob/master/22.md) — standardized comment threading for non-`kind:1` roots via `kind:1111`.
- [NIP-72](https://github.com/nostr-protocol/nips/blob/master/72.md) — communities use NIP-22 comments scoped to `kind:34550` community definitions.
- [NIP-01](https://github.com/nostr-protocol/nips/blob/master/01.md) — defines `a` tags and the replaceable/addressable kind classes that NIP-22 can target.

## Findings

NIP-10 gives standardized reply semantics only for `kind:1` text notes. It says: “Kind 1 events with `e` tags are replies to other kind 1 events” and “Kind 1 replies MUST NOT be used to reply to other kinds, use NIP-22 instead.” This means `kind:1` is the only event kind with standardized NIP-10 reply threading, and using `kind:1` to reply to reposts, addressable events, community definitions, approvals, or other non-`kind:1` roots is explicitly disallowed by spec. Source: NIP-10.

NIP-22 defines the standardized reply/comment mechanism for non-`kind:1` content. It uses `kind:1111` comments, where uppercase tags (`K`, `E`, `A`, `I`, `P`) point at the root scope and lowercase tags (`k`, `e`, `a`, `i`, `p`) point at the direct parent. NIP-22 explicitly supports roots that are event ids, event addresses, or external identifiers, so it is the standardized way to reply to replaceable/addressable events and other non-note content. It also says “Comments MUST NOT be used to reply to kind 1 notes. NIP-10 should instead be followed.” Source: NIP-22.

For addressable and replaceable events, the standardized reply path is therefore NIP-22 `kind:1111`, typically with `A`/`a` tags pointing at the event coordinate and `K`/`k` tags naming the root and parent kinds. NIP-22 gives an explicit example of a top-level comment on an addressable article (`kind:30023`), and notes that when the parent event is replaceable or addressable, clients may also include an `e` tag for the concrete event id of the current version. Combined with NIP-01’s definition of `a` tags and addressable events (`30000–39999`), this is the clearest standard answer for “reply to addressable event.” Sources: NIP-22, NIP-01.

For reposts, there is no separate repost-specific “reply protocol”; the handling is special. NIP-18 defines `kind:6` reposts only for reposting `kind:1` notes and `kind:16` generic reposts for reposting any non-`kind:1` event. NIP-18 also says quote references must be converted into `q` tags so quote reposts “are not pulled and included as replies in threads.” That means quote reposts are explicitly prevented from masquerading as replies. If a client wants to reply to the repost event itself rather than the original content, the standard threading vehicle would still be NIP-22 `kind:1111` because reposts are non-`kind:1` events; but NIP-18 does not define a repost-native reply semantic of its own. Sources: NIP-18, NIP-22.

For communities, NIP-72 gives special standardized handling rooted in the community definition event `kind:34550`. It says posts to a community SHOULD use NIP-22 `kind:1111` events, with the uppercase `A` tag always scoped to the community definition. For top-level community posts, both uppercase and lowercase tags point to the community definition itself. For nested replies, uppercase tags still point to the community definition, while lowercase tags point to the parent post or parent reply. This is a special case of NIP-22 where the “root” is the community definition rather than necessarily the immediate content thread root. NIP-72 also says older `kind:1` + `a`-tag community posting is backward compatibility only and SHOULD NOT be used for new posts. Source: NIP-72.

For community approval events (`kind:4550`), NIP-72 defines moderation/approval semantics, not reply semantics. The approval event must tag one or more community `a` tags plus an `e` or `a` tag for the approved post, and may embed the approved event JSON in `content`. However, NIP-72 does not define a standard way to “reply to an approval” as a threaded conversation primitive. In practice, replying to the underlying approved post has a standard path (NIP-22 within the community), but replying to the approval event itself appears undefined by the cited NIPs. Source: NIP-72.

The kinds or situations that are explicitly undefined or discouraged are therefore: using NIP-10/`kind:1` replies for any non-`kind:1` event (explicitly forbidden); using NIP-22 comments for `kind:1` notes (explicitly forbidden, use NIP-10 instead); treating quote repost references as thread replies (NIP-18 explicitly says `q` tags prevent that); using legacy `kind:1` community posts for new community interactions (allowed only for backwards compatibility, but new posts SHOULD use NIP-22 `kind:1111`); and treating moderation approvals (`kind:4550`) as having their own standardized reply thread model (not defined in these NIPs). Sources: NIP-10, NIP-18, NIP-22, NIP-72.

### Related Specs

- `.trellis/spec/backend/nip-conventions.md` — local project guidance for NIP-10 reply tag generation.
- `.trellis/spec/backend/relay-guidelines.md` — local project guidance for relay hints in reply tags.

## Caveats / Not Found

This research is limited to the cited NIPs and their current draft text on 2026-05-20. No additional NIP was found that gives a separate standardized “reply to repost” or “reply to approval event” model beyond using NIP-22 for non-`kind:1` roots. NIP draft status may change the strength of these conventions over time.
