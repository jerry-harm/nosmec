# PRD: Refactor community parsing into `nip72` and `nostr_sdk`

## Summary

Refactor community-related logic so that:

1. pure NIP-72 community parsing leaves `utils/` and lives in `nip72/`
2. `nostr_sdk/` gains a thicker, generic filter-based event fetch entrypoint that owns relay selection and retrieval policy
3. `utils/` shrinks toward app-specific publish/orchestration code only

The refactor should align with upstream `fiatjaf.com/nostr` style:

- small NIP helper functions
- pointer-first parsing helpers
- typed small return values when a tag carries multiple semantic fields
- generic SDK fetch APIs that mostly return raw `nostr.Event` values

## Motivation

Current community logic is split awkwardly:

- `utils/community.go` mixes pure protocol parsing, retrieval, and publish orchestration
- `nostr_sdk` already owns read-side retrieval concerns, but community definition fetch is not expressed in the same layer
- `nip72` exists but is still too centered on stricter post classification than on minimal tag extraction

This creates three problems:

1. `utils/` remains too heavy and contains reusable protocol logic
2. community parsing is less reusable from `tui/`, `cmd/`, and `nostr_sdk/`
3. retrieval code keeps growing in specialized helpers instead of converging on a generic filter-based fetch API

## Goals

- Move community pure parsing out of `utils/community.go` into `nip72`
- Keep `nip72` small-grained and stylistically consistent with upstream NIP helper packages
- Add a generic `nostr_sdk` read-side fetch API named `FetchEventsByFilter`
- Let `nostr_sdk` choose relays internally using existing relay list and hint infrastructure
- Reduce `utils/community.go` to write/publish orchestration plus app-specific glue
- Preserve `GetParentPointer(top-level) == nil`

## Non-Goals

- Do not redesign write-side SDK architecture in this task
- Do not move `CreateCommunity` / `PostToCommunity` into `nostr_sdk.System`
- Do not introduce a large `ParseCommunityDefinition()` object API as the primary `nip72` surface
- Do not add a broad set of specialized `FetchCommunity*` SDK methods
- Do not implement strong full-event validation beyond required tag extraction semantics

## Design Decisions

### 1. `nip72` stays small-grained and extraction-oriented

`nip72` should only parse single-event, tag-local semantics.

It should not:

- query relays
- inspect caches/stores
- depend on `System` or `AppContext`
- encode app policy
- require every recommended NIP-72 tag to exist before returning useful data

The parser should extract what it can from the required tags and otherwise return zero values / `nil`.

### 2. Community definition parsing should not use `map[string]string`

Relay tags in community definitions carry structured fields and may repeat. Returning a map loses ordering and can collapse distinct entries.

`nip72` should instead expose a small typed relay structure, for example:

```go
type CommunityRelay struct {
    URL     string
    Purpose string
}
```

### 3. Top-level post parent is `nil`

For a top-level community post, `nip72.GetParentPointer` must return `nil`.

The community definition pointer is community scope, not thread parentage.

### 4. Prefer generic SDK retrieval over community-specific fetch entrypoints

Instead of introducing `FetchCommunityDefinitions()` as the primary read API, `nostr_sdk` should grow a generic thicker fetch method:

```go
func (sys *System) FetchEventsByFilter(ctx context.Context, filter nostr.Filter, opts FetchEventsOptions) ([]nostr.Event, error)
```

This fetch API should:

- accept a standard `nostr.Filter`
- let the SDK choose relays by default
- use relay list streams, hints DB, store-local-first behavior, and fallback relays as appropriate
- deduplicate events
- record event-relay knowledge when fetching from the network
- optionally allow caller-supplied overrides through options, without making relays a required primary argument

### 5. Community definitions are treated as raw events at SDK fetch level

`nostr_sdk` should primarily fetch raw `nostr.Event` values for generic filter-driven reads.

Community definition interpretation belongs to `nip72` getters called after fetch, not to a first-class `CommunityDefinition` domain fetch API in this task.

## Proposed API Shape

### `nip72`

Keep existing community/thread helpers aligned with minimal extraction semantics and add definition getters:

```go
func GetCommunityPointer(event *nostr.Event) nostr.Pointer
func GetRootPointer(event *nostr.Event) nostr.Pointer
func GetParentPointer(event *nostr.Event) nostr.Pointer

func IsCommunityDefinition(event *nostr.Event) bool
func GetDefinitionIdentifier(event *nostr.Event) string
func GetDefinitionName(event *nostr.Event) string
func GetDefinitionDescription(event *nostr.Event) string
func GetDefinitionImage(event *nostr.Event) string
func GetDefinitionModerators(event *nostr.Event) []nostr.PubKey
func GetDefinitionRelays(event *nostr.Event) []CommunityRelay
```

Notes:

- `GetParentPointer(top-level) == nil`
- no `ClassifyPostRole` as a required central API for this refactor
- kind checks should exist where necessary, but the getters should remain lightweight and useful rather than enforcing full-event conformance

### `nostr_sdk`

Add a generic filter-based fetch entrypoint, likely alongside or informed by existing `FetchEventByFilter`, `FetchFeedPage`, scoped fetch, and local-first query behavior:

```go
type FetchEventsOptions struct {
    // shape to be kept minimal; may include timeout/relay override/local-first tuning
}

func (sys *System) FetchEventsByFilter(ctx context.Context, filter nostr.Filter, opts FetchEventsOptions) ([]nostr.Event, error)
```

Behavior expectations:

- default relay choice comes from SDK-owned infrastructure
- explicit relays, if supported, are optional overrides in options
- local store should be used when it helps satisfy the query without losing expected network behavior
- result ordering and dedup rules should be explicit in code/tests

### `utils`

After the refactor, `utils/community.go` should trend toward:

- `CreateCommunity`
- `PostToCommunity`
- app-facing publish/orchestration helpers only

Pure parsing and generic retrieval should no longer live there.

## Data Flow

### Community definition read

1. caller builds a `nostr.Filter` for `kind:34550`
2. caller uses `sys.FetchEventsByFilter(...)`
3. caller or higher-level code iterates returned events
4. `nip72` getters extract `d`, `name`, `description`, moderators, relays, image, etc.

### Community post thread/scope use

1. caller fetches raw events through existing SDK APIs or the new generic fetch path
2. `nip72.GetCommunityPointer` determines community scope source
3. `nostr_sdk.ExtractCommunityScope` / `MatchesCommunityScope` continue handling retrieval-side scope logic
4. `nip72.GetParentPointer` returning `nil` for top-level posts keeps thread traversal semantics correct

## Migration Plan

### Step 1
Move community definition parsing helpers from `utils/community.go` into `nip72`.

### Step 2
Update any community retrieval code to consume `nip72` getters instead of ad hoc tag parsing.

### Step 3
Introduce `nostr_sdk.System.FetchEventsByFilter` as the generic read-side API.

### Step 4
Refactor current community retrieval call sites away from `utils` and onto `nostr_sdk` + `nip72` composition.

### Step 5
Leave publish/write orchestration in `utils` for now.

## Testing Strategy

- Add unit tests for each new `nip72` definition getter
- Test missing/partial tags according to the lightweight extraction rule
- Test that relay tag parsing preserves both URL and purpose
- Test that top-level community posts still return `nil` parent pointers
- Add SDK tests for `FetchEventsByFilter` covering:
  - relay selection defaults
  - deduplication
  - local-store-first behavior where applicable
  - successful raw event retrieval for `kind:34550` filters
- Keep tests behavior-oriented; do not rely only on compile/build checks

## Risks

- generic `FetchEventsByFilter` could overlap awkwardly with existing `FetchEventByFilter` unless its purpose and return shape are clearly differentiated
- if relay selection policy is under-specified, call sites may still bypass SDK and keep custom relay logic
- if `nip72` getters are too strict, they will recreate the over-validation problem we are trying to avoid
- if getters are too loose, malformed events may slip through unexpectedly; tests must document intended tolerance

## Acceptance Criteria

- `utils/community.go` no longer owns pure community parsing
- `nip72` exposes small-grained community definition getters and structured relay extraction
- `GetParentPointer` returns `nil` for top-level community posts
- `nostr_sdk` exposes `FetchEventsByFilter` as a generic thicker filter-based fetch API
- the new SDK fetch path defaults relay choice internally using existing SDK state/hints infrastructure
- community read paths can be composed from `FetchEventsByFilter` + `nip72` getters without needing a new specialized `FetchCommunityDefinitions` API
