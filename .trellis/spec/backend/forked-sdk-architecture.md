# Forked nostr SDK Architecture

> The project forks `fiatjaf.com/nostr/sdk/` into `nosmec/nostr_sdk/` to open up internal APIs and add community feed functions as first-class SDK methods.

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