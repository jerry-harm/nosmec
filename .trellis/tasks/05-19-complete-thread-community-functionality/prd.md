# Complete thread community functionality

## Goal

Finish NIP-compliant community-aware behavior in the thread view so community posts/replies behave correctly when opened as a thread, instead of being treated only as generic NIP-10 note threads.

## What I already know

* `tui/thread/thread.go` currently fetches thread data with generic root/parent/reply traversal.
* `fetchEventByID()` uses `app.Pool().QuerySingle(...)` directly, not a `nostr_sdk.System` local-first helper.
* `fetchThreadReplies()` recursively fetches replies by `#e` tags only.
* Community posting/reply already exists in `utils/community.go` and `cmd/community_commands.go`.
* Community timeline support already exists via `app.System().FetchFollowedTimelinePage(..., communityAddrs, ...)`.

## Assumptions (temporary)

* The missing functionality is in thread rendering/fetching behavior, not community creation or timeline listing.
* We should likely preserve current generic thread behavior for non-community posts.

## Open Questions

* None.

## Requirements (evolving)

* Thread view must work correctly for community-associated posts.
* MVP scope is community-aware thread fetching/context only.
* In-thread reply/compose/moderation affordances are out of scope for this task.
* Community thread scope must be determined by the community root scope tag, i.e. `A=34550:<community-author-pubkey>:<community-d-identifier>`.
* Reply/root/parent traversal for community comments must follow NIP-22 / NIP-10 semantics, not a custom "same lowercase a-tag" rule.
* When filtering candidate events for a community thread, prefer events whose community root scope matches the active event's community `A` tag; old-style compatibility may inspect lowercase `a` only as fallback for legacy events.

## Acceptance Criteria (evolving)

* [ ] Community thread behavior is explicitly defined for fetch scope and context.
* [ ] Community thread support does not regress normal note threads.
* [ ] No new thread-level compose or moderation UI is introduced.
* [ ] Community thread scope is determined by the same community `A` root scope.
* [ ] Community thread traversal stays NIP-compliant for parent/root/reply semantics.
* [ ] Legacy/compat events do not break normal thread rendering.

## Technical Approach

* Detect community context from the active event's community root scope, preferring `A=34550:...` and only falling back to legacy lowercase `a` when needed.
* Keep normal thread behavior unchanged for non-community events.
* For community comments (`kind:1111`), traverse parent/root using NIP-22 semantics and legacy NIP-10 compatibility where needed.
* Filter candidate thread events by matching community root scope instead of requiring identical lowercase `a` parent tags.
* Reuse existing thread UI; only change fetch/filter behavior and any minimal context plumbing required.

## Decision (ADR-lite)

**Context**: Current thread traversal is generic NIP-10 and may mix community and non-community discussion when IDs overlap through reply chains. Community comments, however, are defined by NIP-72 on top of NIP-22, where root scope is expressed through uppercase tags.

**Decision**: Treat community thread traversal as a scope defined by the active event's community root scope (`A` tag), while preserving NIP-22/NIP-10 parent/root semantics. Do not define scope via identical lowercase `a` tags.

**Consequences**: Community threads stay protocol-aligned and semantically clean, while still tolerating some legacy events. Naive "same lowercase a-tag" filtering is intentionally avoided because it would reject valid nested community replies.

## Definition of Done

* Tests added/updated where appropriate
* Build/tests/verification pass
* Specs/notes updated if behavior changes

## Out of Scope (explicit)

* In-thread reply / compose entrypoints
* Moderator / approval actions in thread
* Community create/list/info CLI changes
* Broad timeline refactors unless required for thread behavior

## Technical Notes

* Thread code: `tui/thread/thread.go`
* Community post/reply helpers: `utils/community.go`
* Community CLI entry points: `cmd/community_commands.go`
* Protocol references: NIP-72 (community definition and posting), NIP-22 (comment root/parent semantics), NIP-10 (legacy/regular note thread semantics)
