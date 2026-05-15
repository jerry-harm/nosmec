# Research: Go Mock Patterns for TDD Testing

- **Query**: testify/mock patterns for Go testing, mock interfaces for Pool/Relay types in nostr apps, testutil structure
- **Scope**: internal + external
- **Date**: 2026-05-15

## Findings

### Project Context

The project uses `fiatjaf.com/nostr` SDK (v0.0.0-20260310013726-4e490879b558). The `nostr.Pool` type is a concrete struct, not an interface. This means mocking requires defining interfaces that capture the subset of Pool methods used by utils functions.

**Key finding**: testify is NOT currently in the project. The `goleak` package is present (indirect dep) for goroutine leak detection in tests, but `stretchr/testify` is not a dependency.

### Files Found

| File Path | Description |
|---|---|
| `config/context.go` | AppContext struct — concrete, holds `*nostr.Pool` via `Pool()` method |
| `utils/get.go` | Uses `opts.App.Pool().QuerySingle()`, `FetchMany()`, `FetchManyReplaceable()`, `PublishMany()`, `SubscribeMany()` |
| `utils/search.go` | Uses `app.Pool().FetchMany()` for search |
| `utils/dm.go` | Uses `app.Pool().SubscribeMany()` for DMs |
| `utils/post.go` | Uses `app.Pool().PublishMany()` for publishing |
| `utils/community.go` | Uses `app.Pool().PublishMany()`, `FetchMany()` |
| `utils/relay_list.go` | Uses `Pool().QuerySingle()`, `PublishMany()` |
| `utils/profile.go` | Uses `Pool().QuerySingle()`, `PublishMany()` |
| `utils/subscription.go` | Uses `Pool().QuerySingle()`, `PublishMany()` |
| `utils/get.go:102` | `SubscribeWithCache` takes `*nostr.Pool` directly — not through AppContext |
| `go.mod` | No testify dependency |

### Pool Methods Used Across Utils

All calls go through `opts.App.Pool()` or `app.Pool()` returning `*nostr.Pool`:

| Method | Used In |
|---|---|
| `QuerySingle(ctx, relays, filter, opts)` | get.go, relay_list.go, profile.go, subscription.go, thread.go |
| `FetchMany(ctx, relays, filter, opts)` | get.go, search.go, dm.go, community.go |
| `FetchManyReplaceable(ctx, relays, filter, opts)` | get.go |
| `PublishMany(ctx, relays, event)` | post.go, relay_list.go, community.go, subscription.go, get.go (CacheEvent) |
| `SubscribeMany(ctx, relays, filter, opts)` | dm.go, get.go (SubscribeWithCache) |
| `EnsureRelay(url)` | user_relays.go |

### Mock Interface Design

To mock `*nostr.Pool`, you need to define an interface that captures the methods used:

```go
// PoolInterface defines the pool methods used by utils package
type PoolInterface interface {
    QuerySingle(ctx context.Context, relays []string, filter nostr.Filter, opts nostr.SubscriptionOptions) *nostr.QueryResult
    FetchMany(ctx context.Context, relays []string, filter nostr.Filter, opts nostr.SubscriptionOptions) chan nostr.RelayEvent
    FetchManyReplaceable(ctx context.Context, relays []string, filter nostr.Filter, opts nostr.SubscriptionOptions) *nostr.ReplaceableResults
    PublishMany(ctx context.Context, relays []string, event nostr.Event) chan nostr.PublishResult
    SubscribeMany(ctx context.Context, relays []string, filter nostr.Filter, opts nostr.SubscriptionOptions) chan nostr.RelayEvent
    EnsureRelay(url string) error
}
```

### AppContext Mock Interface

Functions use `opts.App` (type `*config.AppContext`). To mock this, define an interface:

```go
// AppContextInterface captures methods used by utils functions
type AppContextInterface interface {
    Pool() PoolInterface
    AllReadableRelays() []string
    AllWritableRelays() []string
    ListRelays() []config.Relay
    GetMySecretKey() (nostr.SecretKey, error)
    GetMyPubKey() (nostr.PubKey, error)
    QueryTimeout() time.Duration
    Config() config.Config
    LocalRelayEnabled() bool
    ListDMRelays() []string
    ListSearchRelays() []string
}
```

Note: `GetOptions.App` is currently typed as `*config.AppContext` (concrete). To use mock interfaces, either:
1. Change `GetOptions.App` to `AppContextInterface` (interface) — risk of breaking existing code
2. Create mock `*config.AppContext` that uses real Pool but can be configured for tests

### testify/mock Usage Pattern

For each method you want to mock:

```go
import "github.com/stretchr/testify/mock"

type MockPool struct {
    mock.Mock
}

func (m *MockPool) QuerySingle(ctx context.Context, relays []string, filter nostr.Filter, opts nostr.SubscriptionOptions) *nostr.QueryResult {
    args := m.Called(ctx, relays, filter, opts)
    if args.Get(0) == nil {
        return nil
    }
    return args.Get(0).(*nostr.QueryResult)
}
```

When setting up tests:

```go
mockPool := new(MockPool)
mockPool.On("QuerySingle", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedResult, nil)
```

### Testutil Package Structure

Recommended structure for nosmec:

```
testutil/
├── mock_app.go      # MockAppContext
├── mock_pool.go     # MockPool implementing PoolInterface
├── mock_store.go    # MockStore for StoreInterface
└── mocks.go         # Auto-generated with mockgen (optional)
```

### Key Caveats

1. **testify not in go.mod** — needs to be added: `go get github.com/stretchr/testify@latest`
2. **`*nostr.Pool` is concrete** — PoolInterface must be defined manually (no automatic mock generation)
3. **`GetOptions.App` is concrete `*config.AppContext`** — changing to interface is breaking; prefer creating mock that wraps real Pool with controlled responses
4. **SubscribeMany/SubscribeMany channels** — need to properly close channels in mock to avoid goroutine leaks (use `goleak` to detect)
5. **goroutine leaks** — `go.uber.org/goleak` is already available as indirect dependency; use in tests with `defer goleak.VerifyNone(t)`

### Related Specs

- `.trellis/spec/backend/index.md` — testing guidelines
- `golang-stretchr-testify` skill — for mock patterns
- `golang-testing` skill — for testing principles

## External References

- [testify/mock documentation](https://pkg.go.dev/github.com/stretchr/testify/mock) — mock usage patterns
- [Go interface design for mocking](https://pkg.go.dev/testing#hdr-Mocking) — standard Go testing guidance
- [OurBigBook Go interfaces for testing](https://ourbigbook.com/#heading-go-interfaces-for-mocking) — practical patterns

## Caveats / Not Found

- `testify` is NOT a dependency — needs to be added
- `nostr.Pool` concrete type means manual interface definition required
- No existing mock infrastructure found in project
- `GetOptions` struct uses concrete `*config.AppContext` — interface migration would be a breaking change worth discussing