# Research: reply semantics standard events

- **Query**: Which Nostr event kinds have standardized reply semantics in NIPs, focusing on text notes, comments, long-form/article-like content, and generic event/thread replies?
- **Scope**: external
- **Date**: 2026-05-20

## Findings

### Files Found

| File Path | Description |
|---|---|
| `https://raw.githubusercontent.com/nostr-protocol/nips/master/10.md` | NIP-10 defines threaded replies for `kind:1` text notes. |
| `https://raw.githubusercontent.com/nostr-protocol/nips/master/22.md` | NIP-22 defines `kind:1111` comments for replies scoped to arbitrary root events or external IDs. |
| `https://raw.githubusercontent.com/nostr-protocol/nips/master/23.md` | NIP-23 defines `kind:30023` long-form content and requires NIP-22 comments for replies. |
| `https://raw.githubusercontent.com/nostr-protocol/nips/master/27.md` | NIP-27 clarifies inline references and that reply recognition can optionally use quote tags. |
| `https://raw.githubusercontent.com/nostr-protocol/nips/master/73.md` | NIP-73 defines external `i`/`I` content identifiers used by NIP-22 comment scopes. |

### Code Patterns

The explicit standardized reply models found are:

1. **`kind:1` text notes → NIP-10 reply semantics**  
   NIP-10 says `kind:1` is a plaintext note and that replies to other `kind:1` notes use `e` tags with reply-thread markers. It explicitly states: **“Kind 1 events with `e` tags are replies to other kind 1 events”** and **“Kind 1 replies MUST NOT be used to reply to other kinds, use NIP-22 instead.”** Source: `10.md`.

   Required/preferred tag structure:

   ```json
   ["e", "<event-id>", "<relay-url>", "root", "<pubkey>"]
   ["e", "<event-id>", "<relay-url>", "reply", "<pubkey>"]
   ```

   Notes from NIP-10:
   - top-level reply to root: a single marked `e` tag with marker `"root"`;
   - nested reply: use `"root"` for thread root and `"reply"` for direct parent;
   - `p` tags should include the replied-to author plus inherited participants.

2. **`kind:1111` comments → NIP-22 generic reply/comment semantics**  
   NIP-22 standardizes cross-kind comments using `kind:1111`. It says a comment is **“a threading note always scoped to a root event or an `I`-tag”** and that comments **must** distinguish root vs parent with uppercase vs lowercase tag names. Source: `22.md`.

   Required tag structure for standardized comments:

   ```json
   ["<A|E|I>", "<root-address-or-id-or-I-value>", "<hint>", "<root-pubkey-if-E>"]
   ["K", "<root-kind>"]
   ["P", "<root-pubkey>"]
   ["<a|e|i>", "<parent-address-or-id-or-i-value>", "<hint>", "<parent-pubkey-if-e>"]
   ["k", "<parent-kind>"]
   ["p", "<parent-pubkey>"]
   ```

   Mandatory semantics from NIP-22:
   - uppercase `A`/`E`/`I` + `K`/`P` = root scope;
   - lowercase `a`/`e`/`i` + `k`/`p` = parent item;
   - `K` and `k` **MUST** be present;
   - `P`/`p` are required when the referenced item is a nostr event with an author;
   - comments **MUST NOT** be used to reply to `kind:1`; NIP-10 applies there instead.

3. **`kind:30023` long-form articles → replies must use NIP-22 comments**  
   NIP-23 defines long-form content and has an explicit **“Replies & Comments”** section stating: **“Replies to `kind 30023` MUST use NIP-22 `kind 1111` comments.”** Source: `23.md`.

   So `kind:30023` itself does **not** define its own reply-tag scheme; its standardized reply path is:
   - root article is `kind:30023` (usually addressed by `a` / address form because it is addressable);
   - reply event is `kind:1111` following NIP-22;
   - root tags should typically use `A` + `K=30023` + `P=<article author>`;
   - parent tags use lowercase equivalents, with top-level comments usually mirroring the article as parent.

4. **Generic event/thread replies across non-`kind:1` content → NIP-22 is the standard path**  
   NIP-10 explicitly forbids using `kind:1` replies for other kinds, and NIP-22 fills that gap with a generic comment model over event ids (`E`/`e`), addresses (`A`/`a`), and external identifiers (`I`/`i`). This is the standardized reply mechanism for article-like content and arbitrary threaded comments on other event kinds. Sources: `10.md`, `22.md`, `23.md`, `73.md`.

5. **Inline references / quote tags are related but not the core standardized reply carrier**  
   NIP-10 and NIP-22 both allow `q` tags for cited events in content, and NIP-27 says clients may add corresponding tags when they want mentions recognized or notified. But these are optional citation/mention aids, not the primary required reply-thread structure. Sources: `10.md`, `22.md`, `27.md`.

### External References

- [NIP-10](https://raw.githubusercontent.com/nostr-protocol/nips/master/10.md) — standardized reply threading for `kind:1` text notes via marked `e` tags.
- [NIP-22](https://raw.githubusercontent.com/nostr-protocol/nips/master/22.md) — standardized generic comments/replies using `kind:1111` and root/parent uppercase/lowercase tag pairs.
- [NIP-23](https://raw.githubusercontent.com/nostr-protocol/nips/master/23.md) — long-form `kind:30023`; replies must use NIP-22 comments.
- [NIP-27](https://raw.githubusercontent.com/nostr-protocol/nips/master/27.md) — inline `nostr:` references and optional tags for mentions/recognition.
- [NIP-73](https://raw.githubusercontent.com/nostr-protocol/nips/master/73.md) — external content identifiers used by NIP-22 `I`/`i` scopes.

### Related Specs

- Not found.

## Caveats / Not Found

- The search was focused on NIP-standardized reply semantics for the requested categories, not on every event kind in the NIP repository.
- NIP-18 quote/repost behavior was not needed to answer the core question because it does not define the main reply-thread structure for these event classes.
- The fetched NIP documents are currently marked `draft`/`optional` in the upstream repository, so “standardized” here means documented in NIPs, not necessarily universal network/client adoption.
