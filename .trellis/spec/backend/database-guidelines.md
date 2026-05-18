# Database Guidelines

> BoltDB storage patterns and conventions.

---

## Overview

**BoltDB** (`go.etcd.io/bbolt`) for local event storage via `fiatjaf.com/nostr/eventstore/boltdb`.

- **Path**: `~/.cache/nosmec/nosmec.db`
- **Lock**: `~/.cache/nosmec/nosmec.lock` (PID file, prevents secondary instances)
- **Backend**: `&boltdb.BoltBackend{Path: dbPath}` initialized via `boltStore.Init()`

Store interface: `eventstore.Store` (type alias `StoreInterface` in `config/interfaces.go`).

---

## Store Initialization

```go
// config/config.go:NewLMDB
dbPath := filepath.Join(dataDir, "nosmec.db")
boltStore := &boltdb.BoltBackend{Path: dbPath}
if err := boltStore.Init(); err != nil {
    return nil, fmt.Errorf("failed to initialize BoltDB: %w", err)
}
return boltStore, nil
```

---

## NIP-65 Relay List Caching (In-Memory Only)

User relay lists (from NIP-65 Kind 10002) are **not stored in BoltDB**. They're kept in-memory in `AppContext.knownRelays` and persisted to the config file (`known_relays` in `nosmec.yaml`) on `Close()`.

The `DiscoverUserRelays` function queries the network on-demand; there's no `user-relays` bucket in BoltDB.

---

## BoltDB as Nostr Event Store

The `eventstore.Store` interface (`fiatjaf.com/nostr/eventstore`) handles:
- Storing and retrieving nostr events by ID/kind/author
- Subscriptions over local event set

**Local relay** (`StartLocalRelay` in `config/config.go`) uses this store as its backend — events published to the local relay are persisted here and served to subsequent queries.

---

## CacheEvent — Local Relay Publishing

```go
func CacheEvent(event *nostr.Event, app *config.AppContext) {
    if !shouldCache(event, app) { return }
    go func() {
        app.Pool().PublishMany(context.Background(), []string{localRelayURL}, *event)
    }()
}
```

Triggered by `shouldCache()` which matches against `Config.CacheFilters`. Only publishes to local relay (`ws://localhost:PORT`).

---

## Close() Error Handling

When closing the store, errors must be accumulated and returned:

```go
func (a *AppContext) Close() error {
    a.mu.Lock()
    defer a.mu.Unlock()
    var errs []error
    if a.store != nil {
        if err := a.store.Close(); err != nil {
            errs = append(errs, err)
        }
    }
    // ... relay persistence ...
    if len(errs) > 0 {
        return errors.Join(errs...)
    }
    return nil
}
```

**Why**: `Close()` errors indicate failure to flush/sync data. Ignoring them risks data loss.