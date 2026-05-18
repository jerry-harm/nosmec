# AppContext — Dependency Injection Container

> `AppContext` is the single dependency injection container for all core dependencies.

---

## Struct

```go
type AppContext struct {
    pool        *nostr.Pool      // Nostr connection pool (lazy relay connections)
    store       StoreInterface    // BoltDB event store (fiatjaf.com/nostr/eventstore/boltdb)
    cfg         Config            // Config snapshot (read-only via Config(), mutable via setters)
    mu          sync.RWMutex      // Protects cfg writes
    viper       *viper.Viper      // Config persistence (WriteConfig on mutations)
    knownRelays map[string]struct{} // Discovered relays, persisted on Close()
    hints       sdk_hints.HintsDB // Relay→pubkey scoring from every incoming event
    sys         *access.System    // Access layer (Pool, Hints, KVStore, RelayStreams)
}
```

**StoreInterface** is `eventstore.Store` from `fiatjaf.com/nostr/eventstore/boltdb`.

---

## Construction

```go
func NewAppContext(pool *nostr.Pool, store StoreInterface, cfg Config, v *viper.Viper) *AppContext
```

Created in `config/config.go` during app init. Pool and store are injected; hints is initialized via `GlobalHints()`.

---

## Key Methods

### Pool / Store / System

```go
func (a *AppContext) Pool() *nostr.Pool   // Returns the nostr connection pool
func (a *AppContext) Store() StoreInterface // Returns the BoltDB store
func (a *AppContext) Hints() sdk_hints.HintsDB // Returns relay→pubkey scoring DB
func (a *AppContext) System() *access.System  // Returns the access layer
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
func (a *AppContext) ReadableRelays() []string  // Configured read relays (no local)
func (a *AppContext) WritableRelays() []string  // Configured write relays (no local)
func (a *AppContext) AllReadableRelays() []string // Prepends local relay (ws://localhost:PORT)
func (a *AppContext) AllWritableRelays() []string // Same as WritableRelays (local excluded from write)

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
func (a *AppContext) LocalRelayEnabled() bool
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

### Relay Discovery Tracking

```go
func (a *AppContext) TrackRelays(relays []string)  // Adds to knownRelays map
func (a *AppContext) Close() error                 // Persists knownRelays to config on shutdown
```

---

## Config Mutation Rule

Every mutation method calls `a.viper.WriteConfig()` to persist changes immediately.

```go
// Pattern for all config setters
func (a *AppContext) AddRelay(url string, read, write bool) error {
    a.mu.Lock()
    defer a.mu.Unlock()
    a.cfg.RelayList = append(a.cfg.RelayList, Relay{URL: url, Read: &read, Write: &write})
    a.viper.Set("relay_list", a.cfg.RelayList)
    return a.viper.WriteConfig()
}
```

---

## Local Relay

```go
func (a *AppContext) localRelayURL() string {
    if !a.LocalRelayEnabled() { return "" }
    port := a.cfg.LocalRelay.Port  // default "8989"
    return fmt.Sprintf("ws://localhost:%s", port)
}
```

**Read path**: Local relay prepended to relay list for cache hits.
**Write path**: Local relay NOT included — never the primary write target.

---

## Close() Behavior

`AppContext.Close()` is called on app shutdown:

1. Closes the BoltDB store
2. Closes the KVStore (via `sys.Close()`)
3. Merges `knownRelays` into `cfg.KnownRelays`
4. Persists merged list via `viper.WriteConfig()`

```go
func (a *AppContext) Close() error {
    a.mu.Lock()
    defer a.mu.Unlock()
    var errs []error
    if a.store != nil {
        a.store.Close()
    }
    // ... relay persistence ...
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