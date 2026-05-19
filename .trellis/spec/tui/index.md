# TUI Development Guidelines

> Guidelines for TUI (Terminal UI) development in this project.

---

## Overview

This directory contains guidelines for TUI component development, specifically for Bubble Tea v2 based applications.

---

## TUI Directory Structure

```
tui/
├── component/          # Reusable UI components (tea.Model implementations)
│   ├── bubblon/        # Stack-based navigation controller
│   └── label/          # Pubkey/username chip (clickable, async profile fetch)
├── cmd/                # TUI entry point / command wrappers
├── community/          # Community discovery view
├── compose/            # Note/reply/quote composer
├── dm/                 # Direct messages
├── event/              # Single event detail view
├── thread/             # Thread tree view
└── timeline/           # Main timeline (list-based)
```

**Principles:**
- `component/` — self-contained `tea.Model` with Init/Update/View, no parent dependencies
- Views — application-level screens that compose components and manage global state
- Components emit custom `tea.Msg` types; parents handle navigation/navigation

| Guide | Description | Status |
|-------|-------------|--------|
| [Window Size Management](./window-size-management.md) | Bubblon child view dimensions + frame size subtraction | Complete |
| [TUI Theme System](./theme-system.md) | Centralized color tokens via tui/theme package | Complete |
| [Tag Input UX](./tag-input-ux.md) | Compose window tag input patterns | Complete |
| [Label Component](./label-component.md) | Pubkey/username chip with async profile fetch | Complete |

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