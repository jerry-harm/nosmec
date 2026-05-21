# Research: relay-usage-architecture

- **Query**: Research how relay selection is currently organized across fetch functions in nosmec/nostr_sdk and summarize a clean reorganization plan space.
- **Scope**: mixed
- **Date**: 2026-05-21

## Findings

### Files Found

| File Path | Description |
|---|---|
| `nostr_sdk/system.go` | Core `System` relay streams plus `FetchGlobalTimelinePage`, `FetchMyTimelinePage`, `FetchFollowedTimelinePage`, `FetchEventByFilter`, `FetchEventsByFilter`, `defaultRelaysForFilter`, `FetchRepliesToRoot`, `FetchParent` |
| `nostr_sdk/outbox.go` | `FetchOutboxRelays`, `FetchInboxRelays`, `FetchWriteRelays` and their heuristics/contracts |
| `nostr_sdk/specific_event.go` | Pointer-based fetch path with mixed relay-hint, author-outbox, and fallback behavior |
| `nostr_sdk/replaceable_loader.go` | Replaceable dataloader relay selection via `determineRelaysToQuery` |
| `nostr_sdk/feeds.go` | Live feed and historical feed functions using author outbox relays internally |
| `nostr_sdk/community_scope.go` | Scope-filtered reply/reference fetches with caller relays plus fallback default |
| `nostr_sdk/community_thread.go` | Thread-scope wrappers that pass caller-provided relays through |
| `nostr_sdk/metadata.go` | Profile metadata / `nprofile` helper depends on outbox relay selection |
| `nostr_sdk/lists_relay.go` | `FetchRelayList` backing source for kind `10002` relay-list data |
| `.trellis/spec/backend/query-patterns.md` | Spec contract for `FetchEventsByFilter` local-first behavior and centralized relay defaults |
| `.trellis/spec/backend/forked-sdk-architecture.md` | Fork architecture, generic filter contract, and relay-choice defaults for the SDK |

### Current Relay Sources

`System` owns multiple relay streams initialized in `NewSystem()` (`nostr_sdk/system.go:123-153`):

- `RelayListRelays` — indexer/list-oriented defaults (`system.go:126-127`)
- `FollowListRelays` — used as fill-in relays for kind `3` replaceable queries (`replaceable_loader.go:165-168`)
- `MetadataRelays` — used as fill-in relays for kind `0` replaceable queries (`replaceable_loader.go:165-167`)
- `FallbackRelays` — generic network fallback across many call sites (`system.go:129-137`)
- `JustIDRelays` — ID-oriented fallback for event-ID retrieval (`system.go:138-141`, `specific_event.go:82`)

Other relay sources:

- `FetchOutboxRelays` — heuristic read-side author relay selection using hints, kind `10002`, past attempts, etc. (`outbox.go:19-50`)
- Caller-provided relays — accepted explicitly by scoped thread helpers and `FetchEventsByFilter` override (`system.go:530-533`, `community_thread.go:11-19`, `community_scope.go:49-61`)
- Pointer-embedded relay hints — `FetchSpecificEvent` starts from `nostr.EventPointer.Relays` / `nostr.EntityPointer.Relays` (`specific_event.go:76-93`)

### Which Fetch Functions Are Local-Store-First vs Direct Network

#### Local-store-first

- `FetchMyTimelinePage` checks `sys.Store.QueryEvents(...)` before network fetch (`nostr_sdk/system.go:265-281`), then supplements from `FetchOutboxRelays` (`system.go:283-315`).
- `FetchFollowedTimelinePage` does per-author local store reads before network fetch (`system.go:357-375`), then network supplement via outbox relays (`system.go:377-409`). The community-address branch in the same function is direct network only (`system.go:412-441`).
- `FetchEventsByFilter` is explicitly local-first unless `SkipLocalStore` (`system.go:516-523`), then optionally network (`system.go:525-549`).
- `FetchSpecificEvent` checks `sys.Store.QueryEvents(filter, 1)` before relay fetch unless `SkipLocalStore` (`specific_event.go:98-103`).
- `FetchEventsReferencingIDsInScope` queries local store first and then network (`community_scope.go:79-108`). This matches the scoped-thread local-first rule from spec (`forked-sdk-architecture.md:159-180`).
- `FetchProfileMetadata` checks in-memory cache, then local store, then network (`metadata.go:107-155`).

#### Direct network / network-primary

- `FetchGlobalTimelinePage` directly fetches from `FallbackRelays` and publishes to local store, with no local pre-read (`system.go:210-242`).
- `StreamLiveFeed` directly subscribes to author outbox relays (`feeds.go:24-94`).
- `FetchFeedPage` is mixed: it may read local first (`feeds.go:126-141`) but still always proceeds to relay fetch unless outbox relays are empty (`feeds.go:143-172`), so behavior is not “local-only when enough found”.
- `FetchRepliesToRoot` directly fetches from `FallbackRelays` and does not query local store first (`system.go:620-638`).
- `FetchEventByFilter` delegates to `FetchSpecificEvent`; once pointer conversion succeeds, behavior becomes local-first through that lower-level function (`system.go:467-487`).

### Which Fetch Functions Choose Relays Internally vs Accept Caller Relays

#### Choose relays internally

- `FetchGlobalTimelinePage` → `FallbackRelays.URLs` (`system.go:225-228`)
- `FetchMyTimelinePage` → `FetchOutboxRelays(ctx, pubkey, 2)` (`system.go:283-286`)
- `FetchFollowedTimelinePage` author branch → `FetchOutboxRelays(ctx, pk, 2)` (`system.go:377-380`)
- `FetchFollowedTimelinePage` community-address branch → `FetchOutboxRelays(ctx, nostr.PubKey{}, 2)` with a hardcoded extra fallback set if empty (`system.go:425-428`)
- `FetchFeedPage` / `StreamLiveFeed` → `FetchOutboxRelays` (`feeds.go:40-46`, `feeds.go:147-151`)
- `FetchSpecificEvent` internally composes pointer relays + `FallbackRelays.Next()` + `JustIDRelays`/extra fallback + author outbox relays (`specific_event.go:76-119`)
- `FetchEventsByFilter` chooses relays internally through `defaultRelaysForFilter` unless `opts.Relays` is set (`system.go:530-533`)
- `FetchRepliesToRoot` → `FallbackRelays.URLs` (`system.go:628-631`)
- `replaceable_loader.determineRelaysToQuery` internally chooses between outbox relays and stream-specific fill-ins (`replaceable_loader.go:138-180`)

#### Accept caller relays directly

- `FetchEventsByFilter` via `FetchEventsOptions.Relays` override (`system.go:530-533`)
- `FetchEventByIDInScope` and `FetchRootEventInScope` accept `relays []string` and inject them into an `EventPointer` (`community_thread.go:11-19`)
- `FetchEventsReferencingIDsInScope` accepts `relays []string`; if empty, it falls back to `FallbackRelays.URLs` (`community_scope.go:49-64`)

#### Hybrid / hint-passing API

- `FetchParent` does not accept relays directly, but `GetThreadParentPointer(event)` may carry pointer relay hints that then influence `FetchSpecificEvent` (`system.go:641-660`, `specific_event.go:76-93`).

### Current Behavior by Relay Source

#### `FallbackRelays`

- Used for global timeline default query surface (`system.go:225-228`)
- Used as the last branch of `defaultRelaysForFilter` (`system.go:582-585`)
- Used in specific-event primary and fallback attempts (`specific_event.go:81-83`, `specific_event.go:91-92`)
- Used for reply fetching and empty scoped-relay fallback (`system.go:628-631`, `community_scope.go:59-64`)

#### `JustIDRelays`

- Used by `defaultRelaysForFilter` when the filter has `IDs` (`system.go:553-562`)
- Used by `FetchSpecificEvent` only for `EventPointer` fallback path (`specific_event.go:76-84`)

#### `RelayListRelays`

- Used by `defaultRelaysForFilter` for `KindCommunityDefinition` queries (`system.go:574-580`)
- Used by replaceable loader fill-in when loading kind `10002` relay lists (`replaceable_loader.go:164-170`)

#### `FetchOutboxRelays`

- Used for single-author generic filter queries (`system.go:564-571`)
- Used in timeline/feed paths (`system.go:283-286`, `system.go:377-380`, `feeds.go:40-46`, `feeds.go:147-151`)
- Used in `FetchSpecificEvent` when author is known (`specific_event.go:105-119`)
- Used inside replaceable loader for most kinds (`replaceable_loader.go:150-155`)

#### Caller-provided relays

- Scoped thread APIs treat caller relays as priority hints, not exclusive ownership, because they flow into `FetchSpecificEvent`, which still appends fallback/author relays (`community_thread.go:11-19`, `specific_event.go:80-83`, `specific_event.go:90-92`, `specific_event.go:117-118`).
- `FetchEventsByFilter` override is stronger: if `opts.Relays` is provided, `defaultRelaysForFilter` is bypassed (`system.go:530-533`).

### Code Patterns

#### 1. Generic filter fetch centralizes one relay-default policy

Spec says generic multi-event queries should use SDK-owned relay selection and local-store-first behavior (`query-patterns.md:27-45`, `forked-sdk-architecture.md:229-256`). Code matches that through:

```go
relays := opts.Relays
if len(relays) == 0 {
    relays = sys.defaultRelaysForFilter(ctx, filter)
}
```

— `nostr_sdk/system.go:530-533`

`defaultRelaysForFilter` implements the documented priority ladder (`system.go:553-585`).

#### 2. Specialized fetchers keep their own relay logic instead of delegating to one shared selector

Examples:

- `FetchGlobalTimelinePage` hardwires `FallbackRelays.URLs` (`system.go:225-228`)
- `FetchRepliesToRoot` also hardwires `FallbackRelays.URLs` (`system.go:628-631`)
- `FetchSpecificEvent` builds a bespoke two-attempt relay plan (`specific_event.go:124-165`)
- `replaceable_loader.determineRelaysToQuery` has a separate kind-based policy (`replaceable_loader.go:138-180`)

#### 3. “Caller relays” mean different things in different APIs

- In `FetchEventsByFilter`, caller relays fully override SDK defaults (`system.go:530-533`).
- In scoped event fetching, caller relays seed pointer relays but `FetchSpecificEvent` still broadens the search with fallback and author outbox relays (`specific_event.go:80-83`, `specific_event.go:90-92`, `specific_event.go:107-118`).

#### 4. Local-first is strong in specs, but uneven across specialized helpers

The spec strongly documents local-first for generic filter reads and scoped thread reads (`query-patterns.md:37-44`, `forked-sdk-architecture.md:159-180`). This is reflected in `FetchEventsByFilter` and `FetchEventsReferencingIDsInScope`, but not in `FetchGlobalTimelinePage` or `FetchRepliesToRoot`.

### Tensions / Inconsistencies Visible from Specs vs Code

1. **Centralized relay ownership vs scattered specialized logic**  
   Specs say relay choice should stay centralized in the SDK, especially through `FetchEventsByFilter` (`query-patterns.md:123-134`, `forked-sdk-architecture.md:243-256`). In code, several specialized functions still encode their own relay behavior (`system.go:225-228`, `system.go:377-380`, `system.go:425-428`, `system.go:628-631`, `specific_event.go:76-119`, `replaceable_loader.go:138-180`).

2. **Local-store-first is a declared contract, but not universal across fetch functions**  
   The generic and scoped-thread contracts emphasize local-first (`query-patterns.md:37-44`, `forked-sdk-architecture.md:159-180`), yet `FetchGlobalTimelinePage` and `FetchRepliesToRoot` are direct-network helpers, and the community-address branch of `FetchFollowedTimelinePage` does not pre-read local store (`system.go:412-441`).

3. **Community-definition relay policy is centralized for filter fetches, but community-related fetches elsewhere do not consistently use it**  
   `defaultRelaysForFilter` gives `kind:34550` special handling via `RelayListRelays + FallbackRelays` (`system.go:574-580`) exactly as documented in spec (`query-patterns.md:127-133`). But other community-oriented paths, such as `FetchFollowedTimelinePage` community-address branch, use `FetchOutboxRelays(ctx, nostr.PubKey{}, 2)` plus hardcoded fallback defaults (`system.go:425-428`), which is a different selection rule.

4. **The role of caller relays is not uniform**  
   Specs mention override behavior for `FetchEventsByFilter` (`query-patterns.md:43`, `forked-sdk-architecture.md:233-247`), but other fetch APIs treat supplied relays as hints that may be expanded internally (`specific_event.go:80-83`, `specific_event.go:90-92`, `specific_event.go:107-118`).

5. **Replaceable-loader relay policy is separate from generic filter policy**  
   `determineRelaysToQuery` uses kind-sensitive filler streams (`MetadataRelays`, `FollowListRelays`, `RelayListRelays`, `FallbackRelays`) and `FetchOutboxRelays` thresholds (`replaceable_loader.go:150-177`). That policy overlaps conceptually with `defaultRelaysForFilter` but is maintained independently.

### Architectural Organization Options (Plan Space Only)

#### Option 1: Keep per-fetch ownership, but classify fetch APIs by relay-policy family

Organization shape:

- **Generic filter family** — `FetchEventsByFilter` + `defaultRelaysForFilter`
- **Pointer/specific-event family** — `FetchSpecificEvent` and scoped wrappers
- **Timeline/feed family** — timeline/feed/live methods centered on `FetchOutboxRelays`
- **Replaceable-loader family** — `determineRelaysToQuery`

Tradeoffs:

- Preserves current semantic differences between “filter fetch”, “specific event”, “timeline”, and “replaceable loader”.
- Keeps specialized behavior close to the fetch function that uses it.
- Leaves multiple relay-selection implementations in place, so policy drift remains possible.

#### Option 2: Introduce one shared relay-selection layer with query-intent categories

Organization shape:

- A central selector maps **query intent** to relay sources, e.g. `id`, `single-author`, `community-definition`, `thread-reply`, `timeline-author`, `global`, `replaceable-kind-0`, `replaceable-kind-3`, `replaceable-kind-10002`.
- Existing fetch functions keep local-store/network orchestration, but delegate relay-set construction to the selector.

Tradeoffs:

- Makes relay responsibility explicit and comparable across fetch functions.
- Eases consistency with spec-documented priority rules.
- Requires encoding more intent categories because current call sites are not all reducible to the same input shape.

#### Option 3: Split responsibility into two layers: caller hint collection vs SDK expansion policy

Organization shape:

- Layer A: gather **candidate relay hints** from caller/pointer/filter/author/community context.
- Layer B: apply an SDK policy for **expansion/fallback** using streams such as `JustIDRelays`, `RelayListRelays`, `FallbackRelays`, or outbox relays.
- Fetch APIs differ mainly in what hints they can supply and whether expansion is strict override vs additive broadening.

Tradeoffs:

- Fits the current hybrid reality where some APIs accept relays but still widen them (`FetchSpecificEvent`) while others use hard override (`FetchEventsByFilter`).
- Makes “caller relays as override” vs “caller relays as hints” an explicit axis.
- Still leaves a need to define which fetch families use strict vs additive expansion.

### Related Specs

- `.trellis/spec/backend/query-patterns.md` — generic query contracts, `FetchEventsByFilter` relay priority, local-first rule
- `.trellis/spec/backend/forked-sdk-architecture.md` — forked SDK ownership boundaries, generic filter contract, scoped-thread local-first rule

## Caveats / Not Found

- This report focuses on fetch functions and relay-selection paths in `nostr_sdk`; it does not inventory every downstream caller in the rest of the repo.
- Some functions with `Fetch*` names (lists, zaps, mint keys) route through dataloaders or helper stacks not fully expanded here unless they were directly relevant to relay selection.
- The plan-space section describes organizational options and tradeoffs only; it does not recommend a code change sequence.
