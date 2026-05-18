# Error Handling

> How errors are handled in this project.

---

## Overview

| Layer | Pattern |
|-------|---------|
| **Utils functions** | Return `error` as last value; no log on expected failures |
| **CLI commands** | Use `handleError()` which prints to stderr and exits 1 |
| **TUI** | Return error via `bubblon.Fail(err)` or as `error` return from commands |
| **Initialization** | Use `logger.Fatal`/`logger.Error` + appropriate exit |

---

## Utils Functions

```go
// Return error; caller decides logging
func GetEvent(ctx context.Context, filter nostr.Filter, opts *GetOptions) *nostr.Event {
    if opts == nil || opts.App == nil {
        return nil
    }
    // ... network call ...
    return event
}

// Error-returning variant
func ReplyToNote(ctx context.Context, parentID, content string, opts *GetOptions) (*nostr.Event, error) {
    // ... validation, signing, publishing ...
    return event, nil
}
```

**Rule**: Log at DEBUG level for expected failures (relay timeout, event not found). Log at ERROR only for actual exceptions.

---

## CLI Commands — handleError

```go
// cmd/errors.go
type CommandError struct {
    Message string
    Err     error
}

func (e *CommandError) Error() string {
    return e.Message
}

func handleError(err error) {
    if err == nil {
        return
    }
    var cmdErr *CommandError
    if errors.As(err, &cmdErr) {
        fmt.Fprintf(os.Stderr, "Error: %s\n", cmdErr.Message)
    } else {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    }
    os.Exit(1)
}
```

**Usage**:
```go
func runReply(cmd *cobra.Command, args []string) {
    event, err := utils.ReplyToNote(...)
    if err != nil {
        handleError(err)
    }
    fmt.Println("Replied!")
}
```

---

## TUI Error Handling

```go
// Opening an error window
return m, bubblon.Fail(err)  // shows error, returns to parent

// Within a command (compose send)
if err != nil {
    return sendErrorMsg{err: err}, nil
}
```

`bubblon.Fail(err)` renders the error in a window and sends `Closed{}` on dismiss.

---

## Initialization Errors

```go
// Fatal: cannot proceed without this
logger.Fatal("failed to initialize config", "error", err.Error())

// Non-fatal: continue with degraded state
logger.Error("failed to connect to local relay", "error", err.Error())
```

---

## Common Mistakes

### Ignoring error return values from Close()

**Wrong**:
```go
if a.store != nil {
    if closer, ok := a.store.(interface{ Close() }); ok {
        closer.Close()  // error ignored!
    }
}
```

**Correct**:
```go
var errs []error
if a.store != nil {
    if closer, ok := a.store.(interface{ Close() error }); ok {
        if err := closer.Close(); err != nil {
            errs = append(errs, err)
        }
    }
}
if len(errs) > 0 {
    return errors.Join(errs...)
}
return nil
```

**Why**: `Close()` errors indicate failure to flush/sync data. Ignoring them risks data loss.

### Hardcoded timeouts instead of QueryTimeout()

**Wrong**:
```go
ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
```

**Correct**:
```go
ctx, cancel := context.WithTimeout(ctx, app.QueryTimeout())
```

---

## Error Types

Defined in `cmd/errors.go`:

```go
type CommandError struct {
    Message string
    Err     error
}
```

Use wrapping with `%w` for contextual errors:

```go
return nil, fmt.Errorf("failed to reply to %s: %w", parentID, err)
```