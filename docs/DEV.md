# Development Guide

## Environment Setup

```bash
# Check Go version (requires 1.23+)
go version

# Install dependencies
go mod tidy

# Build
go build -o nosmec .

# Run
./nosmec --help
```

## Current Project Status

### Completed Features

- [x] Configuration management (Viper + YAML)
- [x] Environment variable support (NOSMEC_*)
- [x] Note posting with NIP-10 conventions
- [x] Reply and quote notes
- [x] Relay management (NIP-65, Kind 10002)
- [x] DM relay management (NIP-17, Kind 10050)
- [x] Profile management (Kind 0)
- [x] Community support (NIP-72, Kind 34550, 1111)
- [x] Subscription/follow system (NIP-02 Kind 3, NIP-51 Kind 10004/10015)
- [x] Alias management
- [x] Shell completion
- [x] I2P support
- [x] Logging with verbosity levels

### In Progress / Needs Rework

- [ ] TUI timeline - current implementation is broken, needs redesign

### Planned Features

- [ ] DM functionality (NIP-17 actual messaging)
- [ ] NIP-46 Remote Signing
- [ ] Search functionality
- [ ] Event cache management
- [ ] Offline mode

## Code Style

- Use `gofmt` for formatting
- Error handling: `fmt.Fprintf(os.Stderr, "Error: %v\n", err)` + `os.Exit(1)`
- Logging: use `logger.Info/Warn/Debug/Error` package
- Config changes: 通过 `AppContext.ConfigManager()` 访问和修改配置
- Use `charm.land/` v2 packages for TUI components

## Adding New Commands

1. Create file in `cmd/`, e.g., `cmd/example.go`
2. Define cobra command
3. Register in `cmd/root.go` `init()`

```go
var exampleCmd = &cobra.Command{
    Use:   "example",
    Short: "Example command",
    Run: func(cmd *cobra.Command, args []string) {
        // implementation
    },
}

func init() {
    rootCmd.AddCommand(exampleCmd)
}
```

## Adding Configuration

1. Add struct field in `config/types.go`
2. Set default in `config/config.go` `loadConfig()`
3. Viper handles `NOSMEC_` env var automatically

## Adding NIP Support

1. Check `fiatjaf.com/nostr` library for relevant packages
2. Reference `utils/post.go` for event creation patterns
3. Publish via `app.Pool().PublishMany()`

## Debugging

```bash
# Verbose output
./nosmec -v --help

# Check config
cat ~/.config/nosmec/nosmec.yaml

# Check database
ls -la ~/.cache/nosmec/
```

## Testing

```bash
go test ./...
go build -o nosmec . && ./nosmec relay list
```
