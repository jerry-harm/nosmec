# Database Guidelines

> LMDB/Bleve storage patterns and conventions.

---

## Overview

**LMDB** (`github.com/PowerDNS/lmdb-go`) and **Bleve** for local event storage, via `fiatjaf.com/nostr/eventstore`.

- **LMDB event store path**: `~/.cache/nosmec/events/` (LMDB environment directory)
- **Bleve index path**: `~/.cache/nosmec/search_index/`
- **Lock**: `~/.cache/nosmec/nosmec.lock` (PID file, prevents secondary instances)
- **Backend**: `&lmdb.LMDBBackend{Path: eventsPath}` initialized via `lmdbStore.Init()`
- **HintsDB path**: `~/.cache/nosmec/hints/` (LMDB environment directory)
- **KVStore path**: `~/.cache/nosmec/kvstore/` (LMDB environment directory)

Store interface: `eventstore.Store` (type alias `StoreInterface` in `config/interfaces.go`).

**Note**: All persistent backends were switched from BoltDB to LMDB on 2026-05-20. Old BoltDB data (previously `hints.db`, `kvstore.db`, `nosmec_events.db`) is not migrated — LMDB stores start fresh from version 2.

---

## Store Initialization

`GlobalPool()` in `config/config.go` initializes the store stack:

```go
// config/config.go:GlobalPool
eventsPath := filepath.Join(dataDir, "events")
lmdbStore := &lmdb.LMDBBackend{Path: eventsPath}
if err := lmdbStore.Init(); err != nil {
    logger.Warn("failed to create LMDB event store, local cache disabled", "error", err.Error())
} else {
    searchIndexPath := filepath.Join(dataDir, "search_index")
    bleveStore := &bleve.BleveBackend{Path: searchIndexPath, RawEventStore: lmdbStore}
    if err := bleveStore.Init(); err != nil {
        logger.Warn("failed to create Bleve search index, search disabled", "error", err.Error())
        GlobalSystem.Store = lmdbStore
    } else {
        GlobalSystem.Store = bleveStore
    }
}
```

Layered store:
- **Bleve** on top for full-text search (kind 0, 1 events)
- **LMDB** underneath for raw event persistence
- `sdk.System.Store` is set to the topmost available layer

---

## sdk.System KVStore

The `sdk.System.KVStore` (`nostr_sdk/kvstore/lmdb`) stores:

- **Event→relay mappings** — which relay an event was first fetched from (for NIP-10 e-tag relay hints)
- **Profile fetch timestamps** — last time we refreshed a profile from network (7-day debounce)

```go
// Key format
'r' + first 8 bytes of event ID → compact relay list bytes
"prof:{kind}:{pubkeyHex}"      → Unix timestamp (last network fetch)
```

KVStore is accessed via `GlobalSystem.KVStore.Get/Set/Update`.

---

## HintsDB

The `sdk.System.Hints` DB (`nostr_sdk/hints/lmdbh`) stores relay→pubkey scoring data:

- **Path**: `~/.cache/nosmec/hints/` (LMDB environment directory)
- **DBI name**: `"hints"`
- **Key format**: `pubkey[32] + relayURL`
- **Value format**: 4 timestamps (16 bytes) for scoring categories

---

## NIP-65 Relay List Caching (In-Memory Only)

User relay lists (from NIP-65 Kind 10002) are **not stored in the local DB**. Relay inspection for `nosmec relay list` reads only SDK-managed databases (`hints/` and `kvstore/`).

The `DiscoverUserRelays` function queries the network on-demand; there's no `user-relays` bucket.

---

## Profile Caching (sdk.System.MetadataCache)

Profile metadata (Kind 0) is cached in `sdk.System.MetadataCache` (in-memory LRU, 8000 entry cap, 6h TTL).

`FetchProfileMetadata` handles the full cache → store → network fetch pipeline automatically:

1. Check MetadataCache (in-memory)
2. If miss, query Store (LMDB/Bleve persisted events)
3. If stale (>7 days) or miss, fetch from network via replaceable event loaders
4. Save to MetadataCache and Store

---

## Close() Error Handling

`System.Close()` is the single place that closes all backend resources. Errors are accumulated and returned from `System.Close()` → `AppContext.Close()`:

```go
func (sys *System) Close() error {
    var errs []error
    if sys.Pool != nil {
        sys.Pool.Close("sdk.System closed")
    }
    if sys.Store != nil {
        if err := sys.Store.Close(); err != nil {
            errs = append(errs, err)
        }
    }
    if sys.Hints != nil {
        if cl, ok := sys.Hints.(interface{ Close() }); ok {
            cl.Close()
        }
    }
    if sys.KVStore != nil {
        if err := sys.KVStore.Close(); err != nil {
            errs = append(errs, err)
        }
    }
    return errors.Join(errs...)
}
```

**Why**: `Close()` errors indicate failure to flush/sync data. Ignoring them risks data loss. Closing in the correct order (Pool before Store) prevents dangling references.

---

## Migration from BoltDB (v1 → v2)

On 2026-05-20, all persistent stores switched from BoltDB to LMDB:

| Backend | v1 path (BoltDB) | v2 path (LMDB) |
|---------|-----------------|----------------|
| HintsDB | `~/.cache/nosmec/hints.db` | `~/.cache/nosmec/hints/` |
| KVStore | `~/.cache/nosmec/kvstore.db` | `~/.cache/nosmec/kvstore/` |
| Event store | `~/.cache/nosmec/nosmec_events.db` | `~/.cache/nosmec/events/` |
| Search index | `~/.cache/nosmec/search_index/` | unchanged |

Old BoltDB data is not migrated. LMDB stores rebuild from network fetches automatically.
