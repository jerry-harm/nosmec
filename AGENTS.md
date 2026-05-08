# AGENTS.md

## Build & Test

```bash
go build -o nosmec .    # Build binary
go test ./...           # Run tests
go mod tidy             # Update dependencies
```

## Key Paths

- Config: `~/.config/nosmec/nosmec.yaml`
- LMDB cache: `~/.cache/nosmec/nosmec.db`
- Binary output: `./nosmec` (gitignored)

## Architecture

- Single Go module (`github.com/jerry-harm/nosmec`), Go 1.25.5
- CLI built with **spf13/cobra**, config with **spf13/viper**
- Nostr SDK: **fiatjaf.com/nostr**
- Local storage: LMDB via `fiatjaf.com/lib`
- TUI: **charm.land** bubbles/lipgloss (incomplete, needs rework)

## Dependency Injection

All core dependencies via `AppContext` (in `config/context.go`):
```go
type AppContext struct {
    pool  PoolInterface   // Nostr connection pool
    store StoreInterface  // LMDB store
    cfg   Config
    config ConfigManager  // Viper wrapper
}
```

Access via `getApp()` or `getAppFromContext(ctx)` in `cmd/root.go`.

## Adding Commands

1. Create `cmd/<name>.go` with cobra command definition
2. Register in `cmd/root.go` `initCommands()` function via `rootCmd.AddCommand()`

## Adding Configuration

1. Add field to `config/types.go` struct
2. Set default in `config/config.go` `loadConfig()`
3. Viper auto-handles `NOSMEC_` env var override

## Error Handling

```go
fmt.Fprintf(os.Stderr, "Error: %v\n", err)
os.Exit(1)
```

## Logging

Use `logger.Info/Warn/Debug/Error` package (not standard log).

## Proxy Support

Only `socks` and `i2p_socks` proxy types are supported. `onion_socks` is not.

## Git Workflow

### Branch Naming
Format: `<type>/[issue-]<description>` — lowercase, hyphens only
```
feat/user-authentication
fix/42-login-race
docs/api-reference
```

### Worktrees
Place under `.claude/worktrees/`, name by replacing `/` with `-`:
```bash
git worktree add .claude/worktrees/feat-user-authentication feat/user-authentication
```

### Commit Messages
```
<type>[scope]: <description>
```
- Types: `feat`, `fix`, `docs`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`
- Subject ≤72 chars, imperative mood, no capital, no trailing period
- Breaking changes: add `BREAKING CHANGE:` footer
- Issue closing: use footer `Closes #<number>`

## Known Issues

- TUI timeline is incomplete and needs redesign
- DM functionality (NIP-17 actual messaging) needs testing
