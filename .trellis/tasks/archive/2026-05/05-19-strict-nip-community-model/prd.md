# strict NIP community model

## Goal

Refine community-related SDK and thread boundaries so community handling follows strict NIP semantics only, without compatibility fallbacks for legacy or malformed events, while simplifying file/module ownership.

## What I already know

- The user wants strict NIP behavior only: accept only clearly valid standard formats.
- The user does not want compatibility logic for legacy or malformed community events.
- The user prefers minimizing SDK surface and, if possible, keeping community-specific code in a single `community.go` file under `nostr_sdk`.
- The user believes generic thread relationship parsing should not live inside `nostr_sdk.System` and instead belongs in TUI or a lighter utility layer.
- The user prefers a design analogous to `nip10`: a dedicated NIP-72 parsing layer, then higher-level operations built around pointers/references instead of ad hoc tag inspection in UI code.
- The user wants to first add a dedicated `nip72` package.
- Existing forked SDK currently contains `community_scope.go` and `community_thread.go`, plus generic thread-fetch helpers on `nostr_sdk.System`.

## Assumptions (temporary)

- Community membership and thread relationship should remain separate concepts even if the code layout is simplified.
- Network-backed event retrieval may stay in SDK, but pure relationship parsing might move out of `System`.

## Open Questions

- None currently.

## Requirements (evolving)

- Community handling must follow strict NIP semantics only.
- No fallback for legacy lowercase/community-misaligned/non-standard event formats.
- The final design should reduce unnecessary split between community-specific files where practical.
- Pure thread relationship parsing should stay in `nostr_sdk`, but not as methods on `nostr_sdk.System`.
- Community logic should stay in `nostr_sdk` as pure helpers plus a small amount of query helper behavior.
- Community semantics should ideally be normalized before query/traversal code, similar to how `nip10` provides reusable parsing helpers.
- Add a dedicated `nip72` package before refactoring higher-level SDK and TUI code.
- The initial `nip72` MVP should parse community references and classify event role (for example, community membership and top-level/comment semantics), but should not perform network or store queries.
- Generic thread retrieval should remain as a few small composable APIs rather than one large all-in-one thread API.
- `nip72` should interpret a single event's tags according to the NIP and return normalized semantics; it should not expose vague cross-event helpers like "do these two events belong to the same community?".

## Technical Approach

- Keep thread/domain parsing in SDK as pure helpers.
- Avoid attaching pure parsing logic to `nostr_sdk.System`.
- Reserve `System` for stateful retrieval/query behavior only.
- Keep community logic in SDK as pure helpers plus minimal community-aware query helpers.
- Prefer normalized pointer/reference parsing over repeated tag scanning in callers.
- Introduce a dedicated `nip72` package that parses strict NIP-72 community references and exposes normalized values for downstream code.
- Make the first `nip72` API surface protocol-only: parsing plus semantic classification, no retrieval.
- Keep thread retrieval APIs small and composable.
- Keep `nip72` focused on single-event protocol interpretation, not multi-event policy helpers.

## Decision (ADR-lite)

**Context**: Community logic and thread parsing were split across SDK `System` methods and community-specific files. The user wants strict NIP semantics and cleaner ownership.

**Decision**: Keep pure thread relationship parsing inside `nostr_sdk`, but move/keep it as pure helper functions rather than `System` methods.

**Consequences**: Nostr domain logic stays near the SDK instead of being pushed into TUI or generic utils, while `System` remains focused on stateful query behavior.

## Decision (ADR-lite) 2

**Context**: Community ownership was unclear between SDK and TUI, especially for filtered thread reconstruction.

**Decision**: Keep community logic in SDK as pure helpers plus a small amount of community-aware query helper behavior.

**Consequences**: Community semantics stay centralized, while generic thread structure can remain reusable and independent from UI code.

## Decision (ADR-lite) 3

**Context**: Existing community handling is encoded as SDK/TUI helpers instead of a reusable protocol parsing layer.

**Decision**: Add a dedicated `nip72` package first, similar in role to `nip10`, and build later community/thread behavior on top of normalized parsed references.

**Consequences**: Protocol semantics become reusable and testable on their own, reducing ad hoc tag scanning across SDK and TUI layers.

## Decision (ADR-lite) 4

**Context**: The new `nip72` package needs enough semantic value to support refactoring, but should not absorb query responsibilities.

**Decision**: The initial `nip72` MVP will parse strict community references and classify event role/semantics, while excluding network/store access and higher-level retrieval behavior.

**Consequences**: The package becomes immediately useful for note/community thread migration without turning into another SDK query layer.

## Decision (ADR-lite) 5

**Context**: There are two competing risks: over-abstracting thread retrieval into one large API, and overloading `nip72` with project-specific policy helpers.

**Decision**: Keep retrieval as a set of small composable APIs, and keep `nip72` strictly focused on interpreting one event's tags into normalized NIP-72 semantics.

**Consequences**: Thread reconstruction stays flexible, while `nip72` remains protocol-facing, precise, and easy to test.

## Acceptance Criteria (evolving)

- [ ] Community scope/filtering accepts only explicitly valid standard community references.
- [ ] Thread relationship parsing does not rely on legacy or malformed tags.
- [ ] Ownership between SDK, TUI, and utilities is explicit and consistent.

## Definition of Done (team quality bar)

- Tests added/updated
- Lint / typecheck / CI green
- Docs/specs updated for behavior and ownership changes

## Out of Scope (explicit)

- Supporting historical malformed community events
- Preserving compatibility behavior solely for existing data quirks

## Technical Notes

- Current community SDK files: `nostr_sdk/community_scope.go`, `nostr_sdk/community_thread.go`
- Current TUI caller: `tui/thread/thread.go`
- Existing generic fetch helpers on `nostr_sdk.System`: `FetchParent`, `FetchRepliesToRoot`, and scoped variants
