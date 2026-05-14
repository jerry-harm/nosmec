# Research: ctrl+c handling in BubbleTea v2

- **Query**: How ctrl+c handling works in BubbleTea v2 with tea.KeyPressMsg
- **Scope**: internal + external
- **Date**: 2026-05-14

## Findings

### How ctrl+c is Represented as a Key Event

In BubbleTea v2, `tea.KeyPressMsg` represents keyboard events. The `ctrl+c` key combination is represented as the string `"ctrl+c"` when calling `msg.String()` on a `tea.KeyPressMsg`.

**Key binding pattern** (from `tui/compose/model.go:82`):
```go
quit: key.NewBinding(
    key.WithKeys("q", "ctrl+c", "esc"),
    key.WithHelp("esc", "quit"),
),
```

**Matching in Update** (from `tui/compose/model.go:272`):
```go
if key.Matches(msg, m.keys.quit) {
    if m.isStandalone {
        return m, tea.Quit
    }
    return m, func() tea.Msg { return bubblon.Close() }
}
```

### Current Behavior in Codebase

| File | ctrl+c behavior | Standalone mode |
|------|----------------|-----------------|
| `tui/compose/model.go` | Sends `tea.Quit` (standalone) or `bubblon.Close()` (embedded) | Has `SetStandalone()` (line 138) |
| `tui/timeline/model.go` | `key.WithKeys("q", "ctrl+c", "esc")` at line 138 — uses same quit binding | N/A (embedded) |
| `tui/dm/model.go` | `key.WithKeys("q", "ctrl+c", "esc")` at line 64 | N/A (embedded) |

### Immediate Exit (os.Exit) vs Graceful Quit (tea.Quit)

The task asks for `ctrl+c` to trigger **immediate program exit** via `os.Exit`, not graceful quit via `tea.Quit`.

**Currently**: All TUI models use `tea.Quit` for quit keys, which performs a graceful shutdown.

**To implement os.Exit on ctrl+c**:
```go
case tea.KeyPressMsg:
    if msg.String() == "ctrl+c" {
        os.Exit(0)  // or os.Exit(1) for error exit
    }
    // ... other key handling
```

### Pattern for Separate ctrl+c Handling

To differentiate ctrl+c from q/esc:

**Option 1: Separate binding for ctrl+c with os.Exit**
```go
// In keyMap
quit: key.NewBinding(
    key.WithKeys("q", "esc"),
    key.WithHelp("esc", "quit"),
),
kill: key.NewBinding(
    key.WithKeys("ctrl+c"),
    key.WithHelp("ctrl+c", "kill"),
),

// In Update
if msg.String() == "ctrl+c" {
    os.Exit(0)
}
if key.Matches(msg, m.keys.quit) {
    // graceful quit logic
}
```

**Option 2: Check string directly in quit handler**
```go
if key.Matches(msg, m.keys.quit) {
    if msg.String() == "ctrl+c" {
        os.Exit(0)
    }
    // ... rest of quit logic
}
```

### Standalone Mode Note

From `tui/compose/main.go:22`, standalone compose calls `m.SetStandalone()`, which sets `m.isStandalone = true`. In standalone mode, `esc` triggers `tea.Quit` (line 280), not `bubblon.Close()`.

The task description says "ctrl+c kills program" — meaning `os.Exit(0)` immediately. This is different from standalone mode where `esc` also uses `tea.Quit`.

### Recommendation

For **immediate exit on ctrl+c** (os.Exit, not tea.Quit):

1. **Add a separate check** in Update for `msg.String() == "ctrl+c"` before the general quit binding
2. **Keep q/esc bound to the existing quit binding** for graceful exit
3. **Use `os.Exit(0)`** for immediate program termination

Example implementation pattern:
```go
case tea.KeyPressMsg:
    // Handle ctrl+c for immediate exit FIRST
    if msg.String() == "ctrl+c" {
        os.Exit(0)
    }
    
    // Then handle other quit keys (q, esc) for graceful quit
    if key.Matches(msg, m.keys.quit) {
        if m.isStandalone {
            return m, tea.Quit
        }
        return m, func() tea.Msg { return bubblon.Close() }
    }
    // ... rest of key handling
```

### Files Found

| File Path | Description |
|---|---|
| `tui/compose/model.go` | Compose model with standalone mode and key bindings (lines 81-84, 267-284) |
| `tui/timeline/model.go` | Timeline model with quit key binding at line 138 |
| `tui/dm/model.go` | DM model with quit key binding at line 64 |
| `tui/compose/main.go` | Standalone compose entry point with SetStandalone() |
| `.trellis/spec/backend/quality-guidelines.md` | Line 145: `q`, `ctrl+c`, `esc` → quit/close pattern |