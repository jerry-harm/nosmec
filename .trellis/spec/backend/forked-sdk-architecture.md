# Forked nostr SDK Architecture

> The project forks `fiatjaf.com/nostr/sdk/` into `nosmec/nostr_sdk/` to open up internal APIs and add community feed functions as first-class SDK methods.

---

## Why Fork

SDK's `replaceableLoaders` / `addressableLoaders` are unexported and hardcoded to specific event kinds. Kind 34550 (community definitions) cannot be registered. Forking lets us:

1. Export dataloader registration APIs
2. Add kind 34550 to replaceable/addressable loaders
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
  ├── system.go (System struct + NewSystem)
  ├── feeds.go (FetchFeedPage, StreamLiveFeed)
  ├── replaceable_loader.go   ← MODIFIED: exported, +kind 34550
  ├── addressable_loader.go   ← MODIFIED: exported, +kind 34550
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

### 1. Exported Dataloader APIs

```go
// replaceable_loader.go
var ReplaceableLoaders map[nostr.Kind]*dataloader.Loader[nostr.PubKey, nostr.Event]

// addressable_loader.go
var AddressableLoaders map[nostr.Kind]*dataloader.Loader[nostr.PubKey, []nostr.Event]

// Both initialized in NewSystem() with known kinds, ready for external registration
```

### 2. Kind 34550 Registration

```go
// In NewSystem() or passed as SystemModifier:
sys.RegisterReplaceableDataloader(nostr.KindCommunityDefinition)
```

### 3. Community Feed Methods

```go
// Follow FetchProfileMetadata pattern: cache → store → TTL → network
func (sys *System) FetchCommunityDefinitions(ctx context.Context) ([]CommunityDefinition, error)

// Follow FetchFeedPage pattern: local-first, KVStore boundaries
func (sys *System) FetchCommunityTimelinePage(ctx context.Context, addrs []string, limit int, until nostr.Timestamp) ([]nostr.Event, error)
```

### 4. Generic Replaceable/Addressable Fetch (bonus)

```go
// For kinds not covered by default loaders:
func (sys *System) FetchReplaceableEvent(ctx context.Context, pubkey nostr.PubKey, kind nostr.Kind) *nostr.Event
func (sys *System) FetchAddressableEvents(ctx context.Context, pubkey nostr.PubKey, kind nostr.Kind) []nostr.Event
```

## Migration Path

| Before | After |
|--------|-------|
| `sdkplus.System` wraps `*sdk.System` | Embed `*nostr_sdk.System` directly or use as-is |
| `sdkplus.FetchCommunityTimelinePage` | `nostr_sdk.System.FetchCommunityTimelinePage` |
| `config.GlobalSystem *sdk.System` | `config.GlobalSystem *nostr_sdk.System` |

## Why Not Fork eventstore

- `eventstore.Store` interface is already fully public
- `boltdb.BoltBackend` and `bleve.BleveBackend` are already fully configurable
- `wrappers.DynamicPublisher` wraps Store and is fully functional
- No modifications needed to any eventstore code

## Merge Strategy

1. Copy `fiatjaf.com/nostr/sdk/` → `nosmec/nostr_sdk/`
2. Add our modifications (export dataloaders, add kind 34550)
3. Merge `sdkplus/system.go` functions into `nostr_sdk/system.go`
4. Update all imports: `fiatjaf.com/nostr/sdk` → `nosmec/nostr_sdk`
5. Delete `sdkplus/` package
6. Run `go mod tidy` to clean up unused `fiatjaf.com/nostr/sdk` dependency