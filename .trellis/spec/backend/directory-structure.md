# Directory Structure

> How backend code is organized in this project.

---

## Module Organization

```
nosmec/
├── cmd/                    # Cobra command definitions (one file per command group)
├── config/                 # Configuration management (Viper + AppContext)
├── utils/                  # Business logic (pure functions + nostr SDK orchestration)
├── tui/                    # Terminal UI (BubbleTea v2)
└── logger/                 # Structured logging (slog)
```

**CLI commands** (`cmd/`): One file per command group. Files named `*_commands.go`.
**Business logic** (`utils/`): Pure filter builders + network orchestration分开。
**TUI** (`tui/`): Independent Bubbletea models, window stack via bubblon.

---

## Command File Naming

| Pattern | Example | Description |
|---------|---------|-------------|
| `*_commands.go` | `note_commands.go`, `event_commands.go` | Command group definitions |
| `root.go` | `root.go` | Root command |
| `errors.go` | `errors.go` | Error type definitions |
| `registry.go` | `registry.go` | Command registration helpers |
| `completion/` | `completion/` | Shell completion subcommand |

**Current cmd files**:
```
cmd/
├── root.go              # Root command + flag definitions
├── note_commands.go     # Note commands (Kind 1): post/reply/quote/timeline
├── event_commands.go     # Generic event commands (all kinds)
├── relay_commands.go     # Relay management: list/add/remove/set/publish/sync/fetch
├── search_commands.go    # NIP-50 search: query
├── gossip_commands.go    # Batch NIP-65 relay discovery from user lists
├── config_commands.go   # Config get/set/unset commands
├── profile_commands.go   # Profile: set/get
├── community_commands.go # NIP-72 community: list/create/join/post
├── dm_commands.go       # NIP-17 DMs: list/send/recv
├── errors.go             # CLI error types
├── registry.go           # Command add helper
└── completion/           # Shell completion subcommand
```

---

## Utils Organization

**Rule**: Separate pure filter-building logic from network orchestration.

```
utils/
├── filters.go           # Pure nostr.Filter builders (testable without mocks)
├── get.go               # Query: GetEvent, GetProfile, GetTimeline, ExtractRelayHints
├── post.go              # Publish: PostNote, ReplyToNote, QuoteNote, BuildReplyTags
├── profile.go            # Profile metadata, name display
├── community.go         # NIP-72 community operations
├── subscription.go      # NIP-02 follow list, NIP-51 lists
├── dm.go                # NIP-17 DM send/recv with nip59 GiftWrap
├── relay_list.go        # Publish/parse NIP-65 (Kind 10002) and NIP-17 (Kind 10050)
├── user_relays.go       # NIP-65 discovery, GetQueryRelays, EnsureRelays
├── search.go            # NIP-50 search (Bleve + relay extensions)
├── alias.go             # Alias management
├── show.go              # Display formatting (NIP-19 bech32 output)
├── sync.go              # Sync from network
├── proxy.go             # SOCKS/I2P proxy support
└── *_test.go            # Tests alongside implementation
```

### Filter Builder Pattern

Filter builders in `filters.go` return `nostr.Filter` with no side effects — unit testable without mocks.

```go
// utils/filters.go — pure, testable
func BuildNoteFilter(noteID string) (nostr.Filter, error) { ... }
func BuildProfileFilter(pubKey nostr.PubKey) nostr.Filter { ... }

// utils/get.go — composes builders + network
func GetNote(ctx context.Context, noteID string, opts *GetOptions) *nostr.Event {
    filter, err := BuildNoteFilter(noteID)  // tested via table-driven tests
    ...
}
```

---

## Config Organization

```
config/
├── config.go         # Viper init, LoadConfig, pool/setup wiring, local relay
├── types.go          # Config struct, Relay, CacheFilter, ProfileConfig, etc.
├── context.go        # AppContext (DI container) — all pool/config access
├── relay.go          # Relay list helpers: GetWritableRelaysFromList, etc.
├── interfaces.go     # StoreInterface = eventstore.Store (type alias)
└── *_test.go         # Tests
```

**AppContext** is the single dependency injection container — holds `pool`, `store`, `cfg`, `viper`, `knownRelays`, `hints`.

---

## TUI Organization

```
tui/
├── timeline/         # Timeline list + window
├── compose/          # Note/DM compose window
├── thread/           # Thread view with treeview
├── event/            # Event detail window
├── dm/               # DM list + chat window
├── community/        # Community view
├── bubblon/          # bubblon.Controller window management
└── cmd/              # TUI command registration
```

**Window management**: `github.com/donderom/bubblon` (Controller-as-field pattern). Timeline holds `bubblon.Controller ctrl` as field; `Update()` delegates non-timeline keys to `ctrl.Update(msg)`.

---

## Naming Conventions

| Item | Convention | Example |
|------|-----------|---------|
| Command files | `*_commands.go` | `note_commands.go` |
| Utils files | `*.go` (noun/verb) | `get.go`, `post.go`, `filters.go` |
| Test files | `*_test.go` alongside impl | `post_test.go` |
| Config structs | `Config`, `Relay`, `ProfileConfig` | PascalCase |
| CLI commands | PascalCase + `Cmd` suffix | `NotePostCmd`, `RelayListCmd` |
| TUI models | `model` struct + `New...()` constructor | `model`, `NewModel()` |

---

## New Feature Organization

1. **CLI** → new `cmd/*_commands.go` file, register in `root.go init()`
2. **Business logic** → `utils/*.go` (pure functions in `filters.go` + orchestration in `get.go`/`post.go`)
3. **Config** → fields in `config/types.go`, defaults in `config/config.go`
4. **TUI** → new subdirectory under `tui/`, window stack via `bubblon.Open(model)` / `bubblon.Close()`