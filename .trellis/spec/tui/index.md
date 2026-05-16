# TUI Development Guidelines

> Guidelines for TUI (Terminal UI) development in this project.

---

## Overview

This directory contains guidelines for TUI component development, specifically for Bubble Tea v2 based applications.

---

## Guidelines Index

| Guide | Description | Status |
|-------|-------------|--------|
| [Tag Input UX](./tag-input-ux.md) | Compose window tag input patterns | Complete |

---

## TUI Testing Patterns

TUI components use layered testing:
- **Unit tests** - Pure logic, state transformations
- **Component tests** - Update/View behavior with synthetic messages
- **Golden file tests** - Visual regression testing of rendered output

See [tui-testing skill](../skills/tui-testing/SKILL.md) for detailed patterns.

---

## Common Patterns

### Focus Management

Use exclusive focus pattern:
1. `blurInputs()` - Reset all fields
2. `focusActiveInput()` - Activate only current field

### State Machine for Navigation

Use single index for navigation state:
- `-2` = not in component mode
- `-1` = empty/new slot
- `>= 0` = editing existing item

### Sync Helpers for Async Operations

```go
func loadFileSync(m *Model, file *File) {
    cmd := m.SetFile(file)
    if cmd != nil {
        msg := cmd()
        *m, _ = m.Update(msg)
    }
}
```