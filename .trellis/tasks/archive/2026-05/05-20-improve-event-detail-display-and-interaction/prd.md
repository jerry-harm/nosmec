# improve event detail display and interaction

## Goal

Improve the event detail view so it shows relay-origin information for the displayed event, and make reply / quote interactions behave correctly for different Nostr event kinds instead of assuming plain note-only behavior.

## What I already know

- The event detail UI lives under `tui/event/`.
- `tui/event/view.go` currently renders author, time, kind, community address for kind `34550`, content, tags, and signature data.
- `tui/event/event.go` exposes `r` reply and `q` quote actions from event detail and opens `tui/compose`.
- `tui/compose/model.go` already has `AddReply` and `AddQuote`, but its tagging and kind selection logic is currently oriented around note/comment flows.
- Event-to-relay tracking already exists in the project via relay tracking middleware and `config.GetEventRelay(...)` / `nostr_sdk.GetEventRelays(...)`-style infrastructure.
- Thread fetching already uses relay-aware query helpers in places like `tui/thread/thread.go` and `utils.GetQueryRelays(...)`.

## Assumptions (temporary)

- Relay-origin display should use already tracked relay knowledge instead of adding a new persistence model.
- Reply / quote behavior should preserve existing note/comment behavior while extending support for other event kinds.
- NIP-22 should be treated as the default standardized reply mechanism for non-`kind:1` content, not as an experimental fallback.
- This task is about event detail UX and compose behavior, not a broad redesign of timeline/thread rendering.

## Open Questions

None currently.

## Requirements (evolving)

- Event detail should surface the relay source information known for the current event.
- Reply behavior must follow NIP-defined semantics by event kind instead of assuming one shared tag model.
- Event kinds with a dedicated reply/comment kind from the NIP README table must use that dedicated reply kind before considering generic fallback rules.
- `kind:1` note replies must use NIP-10 semantics.
- Non-`kind:1` replyable roots must use NIP-22 `kind:1111` comment semantics.
- Community interactions must respect NIP-72's NIP-22-based community-rooted reply structure.
- Quote behavior should distinguish quoting from threaded reply semantics and avoid misclassifying quotes as replies.
- Only events that cannot be represented as a valid NIP-10 or NIP-22 reply target should show a "cannot reply" UX.
- `kind:4550` community approval events should be treated as having no standard reply path, not as a special case with its own reply model.

## Acceptance Criteria (evolving)

- [ ] Event detail shows relay-origin information for the displayed event.
- [ ] Reply behavior is defined for the supported event kinds in MVP with the correct NIP-specific tag structure.
- [ ] Existing note/comment behavior does not regress.
- [ ] Unsupported reply targets, including `kind:4550`, show a clear non-replyable UX instead of inventing thread tags.

## Definition of Done (team quality bar)

- Tests added/updated where appropriate
- Lint / typecheck / CI-relevant checks green
- Docs/notes updated if behavior changes

## Out of Scope (explicit)

- Broad redesign of timeline or thread list UX
- Adding a brand new relay persistence mechanism

## Technical Notes

- Inspected: `tui/event/event.go`, `tui/event/view.go`, `tui/compose/model.go`, `tui/timeline/model.go`, `cmd/event_commands.go`, `utils/community.go`
- Related relay tracking hooks found in `config/config.go`, `nostr_sdk/tracker.go`, and relay query helpers in `utils/user_relays.go`
- Research references:
  - `research/reply-semantics-standard-events.md` — `kind:1` uses NIP-10; non-`kind:1` standardized replies use NIP-22 `kind:1111`; `kind:30023` replies must use NIP-22.
  - `research/reply-semantics-special-events.md` — reposts do not define their own reply protocol, communities use NIP-72 on top of NIP-22, approval events do not define their own reply threading model.

## Research Notes

### What the NIPs say

- `kind:1` replies are standardized only through NIP-10 marked `e` tags and must not be used to reply to other kinds.
- `kind:1111` is the standardized generic comment/reply event for non-`kind:1` roots via NIP-22 and is listed in the main NIP index / event kind table.
- `kind:30023` replies must use NIP-22 comments.
- Community posting/reply under NIP-72 is a special NIP-22 shape rooted at the `kind:34550` community definition.
- Reposts and quote reposts do not define a repost-native threaded reply model.
- `kind:4550` approval/moderation events do not define a standardized threaded reply model.

### Special reply kinds from the README table

- `1222` voice message -> `1244` voice message comment, with NIP-A0 semantics layered on NIP-22-style scoping.
- `2003` torrent -> `2004` torrent comment, with NIP-35 semantics.
- `30311` live event -> `1311` live chat message, with live-event-specific reply structure.
- `40`/`41`/`42` channel discussion stays on `kind:42` with NIP-10-style root/reply threading.
- `9` chat replies with another `kind:9` using `q`-tag semantics, not NIP-22.
- `24` public message has no thread model and should not be treated as generically replyable.
- Git objects like `1617` / `1618` / `1621` fall back to NIP-22; deprecated `1622` is not the modern path.

## Feasible Approaches

**Approach A: Strict NIP-driven replyability** (Recommended)

- How it works:
  - Event detail always shows relay origin data.
  - `reply` is enabled only when the viewed event has a defined reply path under NIPs.
  - First choose any specialized reply kind/model exposed by the NIP README table for that root kind.
  - `kind:1` -> compose `kind:1` with NIP-10 tags.
  - channel/chat/live/torrent/voice-message roots use their own NIP-defined reply model where applicable.
  - otherwise, non-`kind:1` supported roots -> compose `kind:1111` with NIP-22/NIP-72 tags.
  - only events that cannot be expressed as a valid NIP-10/NIP-22 root+parent target show a user-facing unsupported/degraded behavior.
  - `kind:4550` is handled the same way as any other event without a standard reply path: the UI says it cannot be replied to.
  - `quote` remains available where quoting is meaningful and uses quote/citation tags rather than reply-thread tags.
- Pros:
  - Spec-correct and predictable.
  - Maximizes reply coverage while avoiding non-standard thread shapes.
  - Easier to test because the decision boundary is explicit.
- Cons:
  - Requires a reply-strategy decision table instead of one generic implementation path.

## Decision (ADR-lite)

**Context**: Event detail currently assumes one reply model. The NIP index actually has multiple reply shapes, so the UI needs a protocol-driven decision layer.

**Decision**: Introduce a reply-strategy table. Specialized reply kinds from the README table take precedence, then NIP-10 for `kind:1`, then NIP-22/NIP-72 for generic non-`kind:1` replyable roots. Events with no standard reply path, including `kind:4550`, display a non-replyable state.

**Consequences**: More explicit code, but no invented reply semantics and clearer user feedback.

**Approach B: Always-reply fallback through NIP-22 for non-`kind:1`**

- How it works:
  - Treat almost every non-`kind:1` event as replyable through `kind:1111`, including reposts and approvals.
- Pros:
  - More permissive UX.
  - Fewer disabled actions.
- Cons:
  - Ignores specialized reply kinds already defined in other NIPs.
  - Can over-flatten important distinctions between wrapper events and content roots.
  - Risks generating the wrong event kind for voice/torrent/live/chat/channel cases.

**Approach C: Redirect unsupported events to underlying/root content**

- How it works:
  - When viewing repost/approval-like events, `reply` targets the referenced underlying event/community root instead of the current event.
- Pros:
  - Keeps reply available while staying closer to standard semantics.
  - Useful for repost wrappers.
- Cons:
  - More implicit and potentially surprising.
  - Requires reliable extraction of underlying targets per event kind.

## Decision (ADR-lite)

**Context**: Event detail reply/quote actions currently assume a narrow note/comment world, but the NIP index and NIP-22 define a broader matrix: some roots have dedicated reply kinds, some use same-kind threading, some fall back to NIP-22, and some should not be replied to.

**Decision**: Use a strict reply-strategy table driven by NIPs. Prefer dedicated reply kinds/models from the NIP README table, use NIP-10 for `kind:1`, use NIP-22/NIP-72 for generic non-`kind:1` replyable scopes, and show a non-replyable UX for unsupported kinds such as `kind:24` and `kind:4550`.

**Consequences**: Implementation will need an explicit per-kind reply-strategy resolver, but event detail behavior will match protocol semantics instead of inventing one generic fallback.
