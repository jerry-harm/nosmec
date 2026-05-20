# AppContext — Dependency Injection Container

> `AppContext` is the single dependency injection container for all core dependencies.

---

## Struct

```go
type AppContext struct {
    pool        *nostr.Pool      // Nostr connection pool (lazy relay connections)
    store       StoreInterface   // BoltDB event store (fiatjaf.com/nostr/eventstore/boltdb) — now unused, kept for compat
    cfg         Config           // Config snapshot (read-only via Config(), mutable via setters)
    mu          sync.RWMutex     // Protects cfg writes
    viper       *viper.Viper     // Config persistence (WriteConfig on mutations)
    hints       sdk_hints.HintsDB // Relay→pubkey scoring from every incoming event
    sys         *sdk.System      // sdk.System (Pool, Hints, KVStore, RelayStreams, Store)
}
```

**StoreInterface** was `eventstore.Store` from `fiatjaf.com/nostr/eventstore/boltdb` — now superseded by `sdk.System.Store` which stacks BoltDB + Bleve.

---

## Construction

```go
func NewAppContext(pool *nostr.Pool, store StoreInterface, cfg Config, v *viper.Viper) *AppContext
```

Created in `config/config.go` during app init. Pool injected; hints initialized via `GlobalHints()`; system via `GlobalSystem` (sdk.NewSystem with BoltDB+Bleve Store).

---

## Key Methods

### Pool / Store / System

```go
func (a *AppContext) Pool() *nostr.Pool   // Returns the nostr connection pool
func (a *AppContext) Store() StoreInterface // Returns the BoltDB store (legacy, returns nil)
func (a *AppContext) Hints() sdk_hints.HintsDB // Returns relay→pubkey scoring DB
func (a *AppContext) System() *sdk.System // Returns the sdk.System
```

### Identity

```go
func (a *AppContext) GetMyPubKey() (nostr.PubKey, error)
func (a *AppContext) GetMySecretKey() (nostr.SecretKey, error)
func (a *AppContext) GetPrivateKey() string  // Returns nsec bech32 string
```

Secret key is stored in config as NIP-19 `nsec1...` format, decoded on first access.

### Relays

```go
func (a *AppContext) ReadableRelays() []string  // Configured read relays
func (a *AppContext) WritableRelays() []string  // Configured write relays
func (a *AppContext) AllReadableRelays() []string // Same as ReadableRelays (no local relay)
func (a *AppContext) AllWritableRelays() []string // Same as WritableRelays

func (a *AppContext) ListRelays() []Relay       // Full Relay list from config
func (a *AppContext) AddRelay(url string, read, write bool) error
func (a *AppContext) RemoveRelay(url string) error
func (a *AppContext) SetRelayRead(url string, read bool) error
func (a *AppContext) SetRelayWrite(url string, write bool) error
func (a *AppContext) SyncRelayList(relays []Relay)  // Replace all relays

func (a *AppContext) ListDMRelays() []string
func (a *AppContext) AddDMRelay(url string) error
func (a *AppContext) RemoveDMRelay(url string) error
func (a *AppContext) SyncDMRelays(relays []string)

func (a *AppContext) ListSearchRelays() []string
func (a *AppContext) AddSearchRelay(url string) error
func (a *AppContext) RemoveSearchRelay(url string) error

func (a *AppContext) QueryTimeout() time.Duration  // Configurable timeout (default 5s)
```

### Profile

```go
func (a *AppContext) GetProfile() ProfileConfig
func (a *AppContext) SetProfile(profile ProfileConfig) error
```

### Subscriptions

```go
func (a *AppContext) ListSubscriptions(subType string) []Subscription
func (a *AppContext) AddSubscription(sub Subscription) error
func (a *AppContext) RemoveSubscription(subType, subID string) error
func (a *AppContext) ReplaceAllSubscriptions(subscriptions []Subscription) error
```

### Aliases

```go
func (a *AppContext) AddAlias(k, v string)  // Persists immediately
```

### Shutdown

```go
func (a *AppContext) Close() error // Flushes SDK/system resources on shutdown
```

---

## Config Mutation Rule

Every mutation method calls `a.viper.WriteConfig()` to persist changes immediately.

```go
func (a *AppContext) AddRelay(url string, read, write bool) error {
    a.mu.Lock()
    defer a.mu.Unlock()
    a.cfg.RelayList = append(a.cfg.RelayList, Relay{URL: url, Read: &read, Write: &write})
    a.viper.Set("relay_list", a.cfg.RelayList)
    return a.viper.WriteConfig()
}
```

---

## sdk.System Integration

`AppContext.sys` is a `*sdk.System` (from `fiatjaf.com/nostr/sdk`):

- **`sys.Store`** — BoltDB (+ Bleve) event store, handles persistence
- **`sys.KVStore`** — BoltDB-backed KVStore for event→relay hints and profile fetch timestamps
- **`sys.Hints`** — HintsDB (bbolth) for relay→pubkey scoring
- **`sys.MetadataCache`** — In-memory LRU cache for profile metadata (6h TTL)
- **`sys.Pool`** — Points to the same `*nostr.Pool` as `AppContext.pool`

Created via `sdk.NewSystem()` in `GlobalPool()` (config/config.go).

---

## Close() Behavior

`AppContext.Close()` is called on app shutdown:

1. Closes the `sdk.System` (KVStore + Pool)
2. Closes the SDK-backed KVStore cleanly

```go
func (a *AppContext) Close() error {
    a.mu.Lock()
    defer a.mu.Unlock()
    var errs []error
    if a.sys != nil {
        if err := a.sys.Close(); err != nil {
            errs = append(errs, err)
        }
    }
    if len(errs) > 0 {
        return errors.Join(errs...)
    }
    return nil
}
```

---

## Thread Safety

- `mu sync.RWMutex` protects `cfg` writes (all setters acquire write lock)
- Read methods (`Config()`, `ReadableRelays()`, etc.) acquire read lock, release immediately
- `Hints()` and `Pool()` are always safe (no internal mutation after construction)
