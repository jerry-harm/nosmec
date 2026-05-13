# Research: TUI display commands and bubblon integration

- **Query**: TUI display commands, bubblon usage, command structure
- **Scope**: mixed (internal + external)
- **Date**: 2026-05-13

## Findings

### Files Found

| File Path | Description |
|---|---|
| `tui/bubblon/controller.go` | Core bubblon stack-based window management |
| `cmd/event_commands.go` | Event display commands (direct, no bubblon) |
| `cmd/note_commands.go` | Note/timeline commands (uses timeline.RunTimeline, compose.RunNoteCompose) |
| `cmd/registry.go` | Command registration via cobra |
| `tui/timeline/main.go` | Timeline entry using bubblon.New() + tea.NewProgram(ctrl) |
| `tui/cmd/cmd.go` | Old window management messages (WinOpen, WinClose, etc.) — likely obsolete |
| `tui/compose/model.go` | Compose model using bubblon.Close() |
| `tui/window/event/event.go` | EventView with bubblon controller |

### bubblon.Controller API (tui/bubblon/controller.go)

The bubblon package provides stack-based window management for Bubble Tea:

- **`Controller`** — holds `[]tea.Model` stack; top model receives updates/renders
- **`New(model tea.Model)`** — creates controller with initial model
- **`Open(model tea.Model) tea.Cmd`** — command to push model onto stack
- **`Close() tea.Msg`** — message to close top model (sends `Closed{}` to parent)
- **`Replace(model tea.Model) tea.Cmd`** — close + open in one command
- **`Closed{}`** — message sent to parent when top model closes
- **`Models() int`** — returns stack depth

### How Commands Are Structured

1. `cmd/registry.go` registers command groups via `RegisterCommandGroup()`
2. `registerDefaultCommands()` inits all command packages
3. Each `*_commands.go` file defines cobra commands and calls `RegisterCommandGroup()`
4. TUI entry points: `timeline.RunTimeline()`, `compose.RunNoteCompose()`, `event.NewFromID()`

### Current Bubblon Usage

| File | Pattern |
|---|---|
| `tui/timeline/main.go` | `bubblon.New(tlModel)` → `tea.NewProgram(ctrl).Run()` |
| `tui/timeline/model.go` | holds `ctrl bubblon.Controller`, calls `m.ctrl.Update(bubblon.Open(ev))` |
| `tui/window/event/event.go` | holds `ctrl *bubblon.Controller`, calls `bubblon.Open(composeModel)` |
| `tui/compose/model.go` | returns `func() tea.Msg { return bubblon.Close() }` on quit/send |

### Old Window Management (tui/cmd/cmd.go)

Contains `WinOpen`, `WinClose`, `WinFocus`, `WinBlur`, `ViewFocus`, `ViewBlur`, `WinFreshData`, `WinRefreshData`, `MsgError` — appears to be from deleted `windowmanager` package (recent commit: "remove(tui): delete windowmanager package, replaced by bubblon"). This file appears orphaned and should be cleaned up or unified.

### Command Entry Points

| Command | File | Pattern |
|---|---|---|
| `event <event-id>` | `cmd/event_commands.go` | `event.NewFromID()` + `tea.NewProgram(m).Run()` — **no bubblon** |
| `note timeline` | `cmd/note_commands.go` | `timeline.RunTimeline()` — **uses bubblon** |
| `note compose` | `cmd/note_commands.go` | `compose.RunNoteCompose()` — **uses bubblon** |
| `note post/reply` | `cmd/note_commands.go` | CLI-only (no TUI) |

## Key Observations

1. **Inconsistent bubblon adoption**: `event_commands.go` uses direct `tea.NewProgram(m)` while `timeline/main.go` wraps with `bubblon.New()` — the task title "unify-tui-with-bubblon" suggests making event_commands also use bubblon
2. **`tui/cmd/cmd.go` is orphaned**: Window management messages from old windowmanager package still exist but windowmanager was removed in recent commit
3. **Multiple TUI entry points**: event, timeline, compose, dm — each may have different window management patterns

## Related Specs

- `.trellis/spec/backend/quality-guidelines.md` — lines 215-277 cover bubblon window management patterns
- `.trellis/tasks/archive/2026-05/05-13-bubblon/` — prior bubblon migration research