# tui-quit-standardization

## Goal

Audit and standardize quit/exit patterns across all TUI screens. Ensure consistent ESC/q/ctrl+c key bindings and shared quit helper pattern.

## What I Already Know

Research found 4 TUI screens with inconsistencies:

| Screen | Keys | Issue |
|--------|------|-------|
| DM | `q`, `ctrl+c`, `esc` | OK |
| Timeline | `q`, `ctrl+c` | **Missing `esc`** |
| Compose | `q`, `ctrl+c`, `esc` | OK, has dual mode |
| Event view | `esc` only | `q` = quote (intentional) |

**No shared quit helper** — each model implements its own key binding and quit handler.

## Requirements

* **Timeline**: Add `esc` to quit key bindings
* **Event view**: Keep `esc` as quit (no change needed), `q` = quote is intentional
* **Compose**: Already correct with `esc`
* **DM**: Already correct with `esc`
* **Shared quit helper**: Document the standard pattern, not necessarily a shared function

## Standard Quit Pattern

### Key Bindings
- All screens that need to quit: `q`, `ctrl+c`, `esc` (except event view where `q` = quote)
- Event view: `esc` only for quit (special case)

### Quit Handler Pattern
```go
case key.Matches(msg, m.keys.quit):
    if m.subCancel != nil {
        m.subCancel()  // cleanup subscription before quit
    }
    return m, tea.Quit
```

### Help Text
- "q" / "ctrl+c" / "esc" → "quit"
- Event view: "esc" → "close" (because q = quote)

## Acceptance Criteria

* [ ] Timeline has `esc` in quit key bindings
* [ ] All screens with quit use consistent key bindings
* [ ] Standard pattern documented
* [ ] `go build ./...` succeeds
* [ ] `go vet ./...` passes

## Definition of Done

* Timeline ESC added
* Pattern documented
* Build and vet pass

## Out of Scope

* Composing a shared quit helper function (documenting pattern is sufficient)
* Event view `q` = quote change (intentional)
* Window manager (bubblon) changes

## Technical Notes

* File: `tui/timeline/model.go` — add `esc` to key bindings at line 138
* Key binding pattern: `key.WithKeys("q", "ctrl+c", "esc")`
* Help text: `key.WithHelp("q", "quit")` (or update to show esc)