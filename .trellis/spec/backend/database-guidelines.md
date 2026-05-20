# Database Guidelines

> BoltDB/Bleve storage patterns and conventions.

---

## Overview

**BoltDB** (`go.etcd.io/bbolt`) and **Bleve** for local event storage, via `fiatjaf.com/nostr/eventstore`.

- **BoltDB path**: `~/.cache/nosmec/nosmec_events.db`
- **Bleve index path**: `~/.cache/nosmec/search_index/`
- **Lock**: `~/.cache/nosmec/nosmec.lock` (PID file, prevents secondary instances)
- **Backend**: `&boltdb.BoltBackend{Path: boltPath}` initialized via `boltStore.Init()`
- **KVStore path**: `~/.cache/nosmec/kvstore.db` (via `fiatjaf.com/nostr/sdk/kvstore/bbolt`)

Store interface: `eventstore.Store` (type alias `StoreInterface` in `config/interfaces.go`).

---

## Store Initialization

`GlobalPool()` in `config/config.go` initializes the store stack:

```go
// config/config.go:GlobalPool
boltPath := filepath.Join(dataDir, "nosmec_events.db")
boltStore := &boltdb.BoltBackend{Path: boltPath}
if err := boltStore.Init(); err != nil {
    logger.Warn("failed to create BoltDB event store, local cache disabled", "error", err.Error())
} else {
    searchIndexPath := filepath.Join(dataDir, "search_index")
    bleveStore := &bleve.BleveBackend{Path: searchIndexPath, RawEventStore: boltStore}
    if err := bleveStore.Init(); err != nil {
        logger.Warn("failed to create Bleve search index, search disabled", "error", err.Error())
        GlobalSystem.Store = boltStore
    } else {
        GlobalSystem.Store = bleveStore
    }
}
```

Layered store:
- **Bleve** on top for full-text search (kind 0, 1 events)
- **BoltDB** underneath for raw event persistence
- `sdk.System.Store` is set to the topmost available layer

---

## sdk.System KVStore

The `sdk.System.KVStore` (`nostr_sdk/kvstore/bbolt`) stores:

- **Event→relay mappings** — which relay an event was first fetched from (for NIP-10 e-tag relay hints)
- **Profile fetch timestamps** — last time we refreshed a profile from network (7-day debounce)

```go
// Key format
'r' + first 8 bytes of event ID → compact relay list bytes
"prof:{kind}:{pubkeyHex}"      → Unix timestamp (last network fetch)
```

KVStore is accessed via `GlobalSystem.KVStore.Get/Set/Update`.

---

## NIP-65 Relay List Caching (In-Memory Only)

User relay lists (from NIP-65 Kind 10002) are **not stored in BoltDB**. Relay inspection for `nosmec relay list` reads only SDK-managed databases (`hints.db` and `kvstore.db`).

The `DiscoverUserRelays` function queries the network on-demand; there's no `user-relays` bucket in BoltDB.

---

## Profile Caching (sdk.System.MetadataCache)

Profile metadata (Kind 0) is cached in `sdk.System.MetadataCache` (in-memory LRU, 8000 entry cap, 6h TTL).

`FetchProfileMetadata` handles the full cache → store → network fetch pipeline automatically:

1. Check MetadataCache (in-memory)
2. If miss, query Store (BoltDB/Bleve persisted events)
3. If stale (>7 days) or miss, fetch from network via replaceable event loaders
4. Save to MetadataCache and Store

---

## Close() Error Handling

When closing, errors must be accumulated and returned:

```go
func (a *AppContext) Close() error {
    a.mu.Lock()
    defer a.mu.Unlock()
    var errs []error
    if a.sys != nil && a.sys.KVStore != nil {
        if err := a.sys.KVStore.Close(); err != nil {
            errs = append(errs, err)
        }
    }
    // ... other shutdown work ...
    if len(errs) > 0 {
        return errors.Join(errs...)
    }
    return nil
}
```

**Why**: `Close()` errors indicate failure to flush/sync data. Ignoring them risks data loss.
