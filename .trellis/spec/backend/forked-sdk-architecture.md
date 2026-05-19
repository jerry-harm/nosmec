# Forked nostr SDK Architecture

> The project forks `fiatjaf.com/nostr/sdk/` into `nosmec/nostr_sdk/` to open up internal APIs, while protocol parsing stays in `fiatjaf.com/nostr` NIP packages plus project-local `nosmec/nip72`.

---

## Why Fork

SDK's `replaceableLoaders` / `addressableLoaders` are unexported and hardcoded to specific event kinds. Kind 34550 (community definitions) cannot be registered. Forking lets us:

1. Export dataloader registration APIs
2. Add kind 34550 to addressable loaders (it's addressable: 30000 ≤ 34550 < 40000)
3. Merge `sdkplus` functions directly into the SDK (no wrapper layer)

## Dependency Boundary

```
fiatjaf.com/nostr (core)       ← lib dep, not forked
  ├── Event, Filter, Pool, PubKey, Kind, Tags...
  ├── eventstore/              ← lib dep, not forked
  │   ├── Store interface
  │   ├── boltdb/ (6 indexes)
  │   ├── bleve/ (full-text)
  │   └── wrappers/ (DynamicPublisher)
  └── nip*/                    ← lib dep, not forked

nosmec/nip72/                  ← our code
  └── strict NIP-72 tag interpretation helpers

nosmec/nostr_sdk/ (forked)     ← our code
  ├── system.go (System struct + NewSystem + timeline methods)
  ├── feeds.go (FetchFeedPage, StreamLiveFeed, makePubkeyStreamKey, etc.)
  ├── replaceable_loader.go   ← MODIFIED: map-based + exported RegisterReplaceableDataloader
  ├── addressable_loader.go   ← MODIFIED: map-based + exported RegisterAddressableDataloader + kind 34550
  ├── specific_event.go
  ├── metadata.go
  ├── tracker.go, outbox.go, lists_*.go, ...
  ├── cache/                  ← sub-packages forked as-is
  ├── dataloader/
  ├── hints/
  └── kvstore/

nosmec/sdkplus/                ← DELETED, merged into nostr_sdk
```

## SDK Modifications

### 1. Map-Based Dataloaders

Changed from fixed-size arrays indexed by enum constants to maps keyed by `nostr.Kind`:

```go
// system.go
type System struct {
    ReplaceableLoaders map[nostr.Kind]*dataloader.Loader[nostr.PubKey, nostr.Event]
    AddressableLoaders map[nostr.Kind]*dataloader.Loader[nostr.PubKey, []nostr.Event]
}
```

### 2. Exported Registration Functions

```go
// replaceable_loader.go
func (sys *System) RegisterReplaceableDataloader(kind nostr.Kind)
func (sys *System) createReplaceableDataloader(kind nostr.Kind) *dataloader.Loader[nostr.PubKey, nostr.Event]

// addressable_loader.go
func (sys *System) RegisterAddressableDataloader(kind nostr.Kind)
```

Initialized with known kinds in `NewSystem()`:
- Replaceable: 0, 3, 10000-10007, 10015, 10019, 10030
- Addressable: 30000, 30002, 30015, 30030, **34550** (KindCommunityDefinition)

### 3. Kind 34550 is Addressable

Kind 34550 falls in range 30000-40000, making it addressable (parameterized replaceable). It uses the `d` tag for community identifiers and is registered via `RegisterAddressableDataloader(34550)`.

### 4. Merged Timeline/Feed Methods

Added to `nostr_sdk/system.go` from sdkplus:

```go
func (sys *System) FetchGlobalTimelinePage(ctx context.Context, limit int, until nostr.Timestamp) ([]nostr.Event, error)
func (sys *System) FetchMyTimelinePage(ctx context.Context, pubkey nostr.PubKey, limit int, until nostr.Timestamp) ([]nostr.Event, error)
func (sys *System) FetchFollowedTimelinePage(ctx context.Context, pubkeys []nostr.PubKey, communityAddrs []string, limit int, until nostr.Timestamp) ([]nostr.Event, error)
func (sys *System) FetchProfilesBatch(ctx context.Context, pubkeys []nostr.PubKey) map[nostr.PubKey]*nostr.Event
func (sys *System) FetchEventByFilter(ctx context.Context, filter nostr.Filter, timeoutMs int) *nostr.Event
func (sys *System) FetchNote(ctx context.Context, noteID string, timeoutMs int) *nostr.Event
func (sys *System) FetchRepliesToRoot(ctx context.Context, rootID nostr.ID, limit int) []*nostr.Event
func (sys *System) FetchParent(ctx context.Context, event *nostr.Event, timeoutMs int) *nostr.Event
```

### 5. NIP-72 Parsing Layer

Community protocol parsing is not owned by `nostr_sdk.System`. It lives in a dedicated `nip72` package, shaped similarly to `nip10` / `nip22`:

```go
func GetCommunityPointer(event *nostr.Event) nostr.Pointer
func GetRootPointer(event *nostr.Event) nostr.Pointer
func GetParentPointer(event *nostr.Event) nostr.Pointer
func ClassifyRole(event *nostr.Event) (Role, bool)
```

Rules:
- strict NIP-only interpretation
- no lowercase-only / malformed-event fallback
- pointer-first API (`nostr.EntityPointer`, `nostr.EventPointer`, `nostr.Pointer`)
- no project-owned cross-event policy helpers in `nip72`

### 6. Community Thread Query APIs

Community-thread protocol semantics are split between `nip72` parsing and `nostr_sdk` retrieval. The SDK owns:

1. scope extraction from already-parsed strict NIP-72 community pointers
2. scope-aware event filtering
3. thread-scoped parent-chain and reply traversal
4. local-store-first retrieval before relay fallback
5. pure thread helpers that are not attached to `System`

```go
// community_scope.go
func ExtractCommunityScope(event *nostr.Event) string
func MatchesCommunityScope(event *nostr.Event, scope string) bool

// thread_refs.go
func GetThreadParentPointer(event *nostr.Event) nostr.Pointer
func GetThreadRootID(event *nostr.Event) (rootID nostr.ID, isRoot bool, err error)

// community_scope.go / community_thread.go / system.go
func (sys *System) FetchSpecificEventInScope(ctx context.Context, pointer nostr.Pointer, scope string, params FetchSpecificEventParameters) (*nostr.Event, []string, error)
func (sys *System) FetchEventsReferencingIDsInScope(ctx context.Context, ids []nostr.ID, relays []string, scope string) []*nostr.Event
func (sys *System) FetchEventByIDInScope(ctx context.Context, id nostr.ID, relays []string, scope string) (*nostr.Event, []string, error)
func (sys *System) FetchRootEventInScope(ctx context.Context, rootID nostr.ID, relays []string, scope string) (*nostr.Event, []string, error)
func (sys *System) FetchParentInScope(ctx context.Context, event *nostr.Event, scope string, timeoutMs int) *nostr.Event
func (sys *System) FetchParentChainInScope(ctx context.Context, event *nostr.Event, scope string, timeoutMs int, maxDepth int) []*nostr.Event
func (sys *System) FetchRepliesBreadthFirstInScope(ctx context.Context, rootID nostr.ID, relays []string, scope string, maxDepth int, batchSize int) []*nostr.Event
```

### 7. Local-First Rule for Scoped Thread Queries

Scoped reply fetches should query `sys.Store` first, then use relays only as a supplement. This keeps thread reconstruction aligned with the rest of the forked SDK's local-first design.

Wrong:
```go
for ie := range sys.Pool.FetchMany(ctx, relays, filter, nostr.SubscriptionOptions{}) {
    // filter by scope here
}
```

### 8. Wrong vs Correct: Parsing Ownership

Wrong:
```go
type CommunityRef struct {
    Address string
    Author  nostr.PubKey
}

func ParseCommunity(event *nostr.Event) CommunityRef { ... }
```

Correct:
```go
ptr := nip72.GetCommunityPointer(event)
if ptr == nil {
    return ""
}
scope := ptr.(nostr.EntityPointer).AsTagReference()
```

Why:
- `nip72` should mirror existing pointer-oriented NIP helpers
- caller policy stays outside the parser
- SDK and TUI compare normalized pointer values instead of custom wrappers

Correct:
```go
for evt := range sys.Store.QueryEvents(filter, limit) {
    if MatchesCommunityScope(&evt, scope) {
        // use local event first
    }
}
for ie := range sys.Pool.FetchMany(ctx, relays, filter, nostr.SubscriptionOptions{}) {
    // supplement with network events not already seen
}
```

Helper functions in `feeds.go` (were already in forked SDK):
```go
func makePubkeyStreamKey(prefix byte, pubkey nostr.PubKey) []byte
func decodeTimestamp(data []byte) nostr.Timestamp
func encodeTimestamp(v nostr.Timestamp) []byte
```

## Migration Path

| Before | After |
|--------|-------|
| `sdkplus.System` wraps `*sdk.System` | `config.GlobalSystem *nostr_sdk.System` (direct, no wrapper) |
| `sdkplus.FetchFollowedTimelinePage` | `nostr_sdk.System.FetchFollowedTimelinePage` |
| `sdkplus.Wrap(app.System()).Method()` | `app.System().Method()` |
| `fiatjaf.com/nostr/sdk` import | `github.com/jerry-harm/nosmec/nostr_sdk` import |

## Why Not Fork eventstore

- `eventstore.Store` interface is already fully public
- `boltdb.BoltBackend` and `bleve.BleveBackend` are already fully configurable
- `wrappers.DynamicPublisher` wraps Store and is fully functional
- No modifications needed to any eventstore code

## Build Verification

After any change to nostr_sdk:
```bash
go build ./nostr_sdk/...   # SDK itself
go build ./...             # Full project
go mod tidy                # Clean up dependencies
```

## Common Issues

### Import aliasing when updating old code

Old code used `sdk "fiatjaf.com/nostr/sdk"`. After migration, use:
```go
sdk "github.com/jerry-harm/nosmec/nostr_sdk"
```

This keeps `sdk.ProfileMetadata`, `sdk.ParseMetadata` etc. unchanged in call sites.

### Kind 34550: addressable, not replaceable

Kind 34550 is in range 30000-40000, so it's addressable per `Kind.IsAddressable()`. Do NOT register it with `RegisterReplaceableDataloader`.
