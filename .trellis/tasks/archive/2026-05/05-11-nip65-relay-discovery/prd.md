# NIP-65 Relay Discovery

## Goal

Implement relay discovery based on NIP-65 (Kind 10002). When fetching an event or profile for an unknown user, discover their relay list, store in BoltDB, ensure relays in global pool, and use local relay + discovered relays for querying.

## Design Principles

- **Local relay always included** in read relay lists (cache priority)
- **Backup via publish** to local relay (existing CacheEvent pattern)
- **BoltDB persistence** for discovered relay lists
- **Global pool** for relay management (EnsureRelay for lazy connection)

## Data Flow

```
1. GetProfile(pubKey)
       ↓
2. DiscoverUserRelays(pubKey)
       ↓
   ├── Check BoltDB "user-relays:{pubKey}"
   ├── Not found → Discovery using known relays
   │         ↓
   │    Query NIP-65 (Kind 10002)
   │         ↓
   │    Parse with nip65.ParseRelayList()
   │         ↓
   │    Store in BoltDB
   │         ↓
   └── EnsureRelay(userRelays) in global pool
       ↓
3. Build relay list: [localRelay] + [userRelays] + [knownRelays]
       ↓
4. Query with Pool().QuerySingle()
       ↓
5. CacheEvent() to backup to local relay
```

## Storage

**Bucket**: `user-relays`
**Key**: `pubKey` (hex)
**Value**: JSON `UserRelayList`

```go
type UserRelayList struct {
    PubKey     string   `json:"pubkey"`
    ReadRelays []string `json:"read_relays"`
    WriteRelays []string `json:"write_relays"`
    UpdatedAt  int64    `json:"updated_at"`
}
```

## New Methods

### AppContext

```go
func (a *AppContext) DiscoverUserRelays(ctx context.Context, pubKey nostr.PubKey) ([]string, error)
// Returns user's read relays, ensures in pool, caches to BoltDB

func (a *AppContext) GetCachedUserRelays(pubKey string) (*UserRelayList, error)
// Direct BoltDB lookup, no discovery

func (a *AppContext) EnsureRelays(urls []string)
// Wrapper for Pool().EnsureRelay() - lazy connection registration
```

### GetOptions (extend)

```go
type GetOptions struct {
    App    *config.AppContext
    Relays []string  // Optional: explicit relay list override
}
```

### Modified Functions

```go
func GetProfile(ctx context.Context, pubKey nostr.PubKey, opts *GetOptions) *nostr.Event {
    // 1. Discover user relays
    userRelays, _ := opts.App.DiscoverUserRelays(ctx, pubKey)

    // 2. Build relay list: local + user + known
    relays := union(opts.App.AllReadableRelays(), userRelays)

    // 3. Query with relays...
}

func GetEvent(ctx context.Context, filter nostr.Filter, opts *GetOptions) *nostr.Event {
    // Similar pattern for event queries
}
```

## Implementation Steps

1. **Add BoltDB bucket** for user relays (storage migration if needed)
2. **Add AppContext methods**: `DiscoverUserRelays`, `GetCachedUserRelays`, `EnsureRelays`
3. **Extend GetOptions** to support explicit relay override
4. **Modify GetProfile** to use discovery flow
5. **Modify GetEvent** for consistency
6. **Test**: verify local relay + discovery + cache flow

## Acceptance Criteria

- [ ] BoltDB stores and retrieves UserRelayList
- [ ] Discovery fetches NIP-65 from network when cache miss
- [ ] Local relay always in relay list for queries
- [ ] CacheEvent backs up to local relay
- [ ] GetProfile uses discovered user relays
- [ ] GetEvent works with relay discovery

## Out of Scope

- Relay scoring/reputation
- Multi-hop discovery (user's user's relays)
- TTL/expiration of cached relays

## Technical Notes

- `nip65.ParseRelayList(event)` returns `(readRelays, writeRelays)`
- `Pool().EnsureRelay(url)` is lazy - only connects when first used
- Existing `CacheEvent` already publishes to local relay
