# Design Principles

> Last updated: 2026-05-11

## 1. Core Architecture

### 1.1 Project Type
**Nostr CLI client for power users** — a terminal-first client with optional TUI.

### 1.2 Tech Stack
- **Language**: Go
- **CLI Framework**: Cobra + Viper
- **Nostr SDK**: fiatjaf.com/nostr
- **Local Storage**: BoltDB (`go.etcd.io/bbolt v1.4.2`) + Bleve (full-text search)
- **Proxy Support**: SOCKS, I2P (no Tor)

### 1.3 Module Organization

```
cmd/           # Cobra command definitions (one file per command group)
config/        # Configuration management (Viper + AppContext)
utils/         # Business logic (post, get, profile, community, subscription, dm, relay, alias)
tui/           # Terminal UI (Bubbles/Tea-based)
logger/        # Structured logging (slog)
```

### 1.4 Dependency Injection
All core dependencies managed via `AppContext`:
```go
type AppContext struct {
    pool   *nostr.Pool      // Nostr connection pool
    store  StoreInterface   // BoltDB store
    cfg    Config           // Config snapshot
    viper  *viper.Viper     // Config manager
}
```

---

## 2. Design Decisions

### 2.1 Relay Query Architecture

**Decision**: Unified relay pool, no local/remote distinction.

| Before | After |
|--------|-------|
| Local relay: 2s timeout | All relays queried simultaneously |
| Remote relays: 10s timeout | Single configurable timeout (default 5s) |
| PrivateRelays + LocalRelays mixed | Only LocalRelay (embedded relay for caching) |

**Rationale**: Simplifies code, faster user experience, local relay only used for caching.

**Key methods**:
- `AllReadableRelays()` — prepends local relay to readable relay list
- `AllWritableRelays()` — prepends local relay to writable relay list
- `QueryTimeout()` — returns configurable timeout (default 5s)

### 2.2 CLI Command Separation

| Command | Scope |
|---------|-------|
| `event` | Generic events (all kinds) |
| `note` | Kind 1 (text notes) only |
| `profile` | Kind 0 (profile metadata) |
| `community` | Kind 1111/34550 (NIP-72) |
| `dm` | Direct messages (NIP-17) |

### 2.3 Context and Timeout

**All async operations use `context.WithTimeout`**:
- No custom channel-based timeout patterns
- `SubscribeMany` uses timeout context
- `GetEventAsync` uses `context.WithTimeout`

### 2.4 Database

**BoltDB** for local event storage:
- Path: `~/.cache/nosmec/nosmec.db`
- Replaced LMDB (2026-05-10)
- Bleve for full-text search

---

## 3. Coding Conventions

### 3.1 NIP-19 Format Convention

**All user-facing outputs MUST use NIP-19 bech32 format**:

| Entity | Format | Function |
|--------|--------|----------|
| Public Key | `npub1...` | `nip19.EncodeNpub(pk)` |
| Event ID | `nevent1...` | `nip19.EncodeNevent(id, relays, author)` |
| Private Key | `nsec1...` | `nip19.EncodeNsec(sk)` (config only) |

**Command input**: Accept both hex (64-char) and NIP-19 formats.
```go
pointer, err := nip19.ToPointer(eventID)
filter := pointer.AsFilter()
```

### 3.2 Hex String to nostr Type Conversion

**Always use SDK conversion functions** (never `copy()`):

```go
// Correct
id, err := nostr.IDFromHex(hexStr)
pk, err := nostr.PubKeyFromHex(hexStr)

// Wrong - causes garbage data
var id nostr.ID
copy(id[:], hexStr)  // BUG!
```

### 3.3 Configuration

**Viper-based** with layered precedence:
1. CLI flags
2. Environment variables (`NOSMEC_*` prefix)
3. Config file (`~/.config/nosmec/nosmec.yaml`)
4. Defaults

**Config changes persist via `viper.WriteConfig()`**.

---

## 4. Query Patterns

### 4.1 Synchronous Query
```go
event := GetEvent(ctx, filter, opts)
```
- Uses `Pool().QuerySingle()` with timeout context
- Returns first result or nil

### 4.2 Async Query
```go
event := GetEventAsync(ctx, filter, opts)
```
- Uses `context.WithTimeout`
- Returns `*nostr.Event`

### 4.3 Streaming Query
```go
ch := GetTimeline(ctx, limit, until, opts)
for event := range ch {
    // process event
}
```
- Returns `chan *nostr.Event`
- Events yielded as they arrive (no buffering)

### 4.4 Replaceable Events
```go
results := Pool().FetchManyReplaceable(ctx, relays, filter, opts)
results.Range(func(key nostr.ReplaceableKey, ev nostr.Event) bool {
    // process
})
```

---

## 5. Event Caching

`CacheEvent()` publishes events to local relay for caching:
- Triggered by `shouldCache()` which checks `Config.CacheFilters`
- Runs asynchronously (non-blocking)
- Only publishes to local relay (`ws://localhost:8989`)

---

## 6. Supported NIPs

| NIP | Name | Status |
|-----|------|--------|
| 01 | Basic Protocol | ✓ |
| 02 | Follow List (Kind 3) | ✓ |
| 05 | NIP-05 Verification | ✓ |
| 06 | Key Formats | ✓ |
| 10 | Reply Conventions | ✓ |
| 17 | DM Relay List (Kind 10050) | ✓ |
| 19 | Bech32 Entities | ✓ |
| 21 | `nostr:` URL Scheme | ✓ |
| 40 | Expiration Timestamp | ✓ |
| 44 | Encryption | ✓ |
| 51 | Lists | ✓ |
| 65 | Relay List (Kind 10002) | ✓ |
| 72 | Community Boards | ✓ |
| 46 | Remote Signing | Planned |
| 47 | Nostr Wallet Connect | Planned |

---

## 7. Anti-Patterns

### 7.1 Never use `copy()` for ID/PK conversion
```go
// WRONG
var id nostr.ID
copy(id[:], noteID)

// CORRECT
id, err := nostr.IDFromHex(noteID)
```

### 7.2 Never log private keys or nsec
```go
// WRONG - security issue
logger.Info("private key", "sk", sk.Hex())

// CORRECT - only for internal/config storage
config.PrivateKey = nip19.EncodeNsec(sk)
```

### 7.3 Never use hardcoded timeouts
```go
// WRONG
ctx, cancel := context.WithTimeout(ctx, 10*time.Second)

// CORRECT
timeout := app.QueryTimeout()
ctx, cancel := context.WithTimeout(ctx, timeout)
```

---

## 8. File Naming Conventions

| Pattern | Example |
|---------|---------|
| Commands | `note_commands.go`, `event_commands.go` |
| Utils | `get.go`, `post.go`, `profile.go` |
| Config | `config.go`, `types.go`, `relay.go` |

---

## 9. Error Handling

- Use `handleError()` for CLI commands
- Return errors from utils functions
- Log at appropriate level (DEBUG for expected cases, ERROR for failures)

---

## 10. Future Considerations

- NIP-46 Remote Signing
- NIP-47 Nostr Wallet Connect
- TUI improvements (detail view, interactive posting)
