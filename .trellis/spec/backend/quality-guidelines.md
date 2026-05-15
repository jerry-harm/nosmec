# Quality Guidelines

> Code quality standards for backend development.

---

## Overview

<!--
Document your project's quality standards here.

Questions to answer:
- What patterns are forbidden?
- What linting rules do you enforce?
- What are your testing requirements?
- What code review standards apply?
-->

(To be filled by the team)

---

## Forbidden Patterns

<!-- Patterns that should never be used and why -->

(To be filled by the team)

---

## NIP Protocol Implementation Rule (Critical)

**Before implementing ANY nostr protocol behavior, you MUST fetch and read the relevant NIP specification.**

Protocol URL pattern: `https://github.com/nostr-protocol/nips/raw/refs/heads/master/{nip}.md`

| NIP | URL |
|-----|-----|
| NIP-01 | https://github.com/nostr-protocol/nips/raw/refs/heads/master/01.md |
| NIP-50 | https://github.com/nostr-protocol/nips/raw/refs/heads/master/50.md |
| NIP-65 | https://github.com/nostr-protocol/nips/raw/refs/heads/master/65.md |
| NIP-17 | https://github.com/nostr-protocol/nips/raw/refs/heads/master/17.md |

Main NIP index: https://github.com/nostr-protocol/nips/raw/refs/heads/master/README.md

**Why**: "Common sense" assumptions about protocol behavior are frequently wrong. The NIP specs are short, authoritative, and definitive.

**Examples of violations**:
- Implementing read/write relay tags as separate tags for the same URL (NIP-65: one tag, optional "read"/"write" marker, no marker = both)
- Assuming JSON tag order or structure without checking the spec
- Guessing NIP number meanings instead of reading the spec

**Process**:
1. Identify which NIP governs the feature
2. Fetch the spec (WebFetch or Task subagent)
3. Read the spec before writing any code
4. Reference spec in PRD/commits

---

## Common Mistakes

### ID/PK Parsing with `copy()` instead of `IDFromHex`/`PubKeyFromHex`

**Symptom**: Event/PubKey queries return nil events even though valid hex IDs are provided.

**Bug Location**: `utils/get.go:136` and similar

**Wrong**:
```go
var id nostr.ID
copy(id[:], noteID)  // BUG: copies ASCII bytes into 32-byte array
```

**Correct**:
```go
id, err := nostr.IDFromHex(noteID)
if err != nil {
    return nil
}
```

**Why it's bad**: `nostr.ID` is `[32]byte` but hex strings are 64 characters. `copy(id[:], noteID)` copies 64 ASCII bytes (not decoded hex) into 32 bytes, resulting in garbage.

**Prevention**: Always use `nostr.IDFromHex()`, `nostr.PubKeyFromHex()`, `nostr.SecretKeyFromHex()` for hex-to-type conversions.

---

## Required Patterns

<!-- Patterns that must always be used -->

### Hex String to nostr Type Conversion

Always use SDK-provided conversion functions:

| From | To | Function |
|------|-----|----------|
| 64-char hex string | `nostr.ID` | `nostr.IDFromHex(s)` |
| 64-char hex string | `nostr.PubKey` | `nostr.PubKeyFromHex(s)` |
| 64-char hex string | `nostr.SecretKey` | `nostr.SecretKeyFromHex(s)` |
| `nostr.ID` | hex string | `id.Hex()` |
| `nostr.PubKey` | hex string | `pk.Hex()` |

---

## Testing & TDD

### Core Principle

Tests are **executable specifications** — they constrain behavior, not coverage targets.

**TDD Cycle (Red-Green-Refactor)**:
1. **RED**: Write failing test first, watch it fail
2. **GREEN**: Write minimal code to pass
3. **REFACTOR**: Clean up while staying green

**Iron Rule**: No production code without a failing test first. If code exists before tests, delete and start over.

### Test Quality Requirements

| Requirement | Description |
|------------|-------------|
| **Named subtests** | Every `t.Run` needs a `name` field describing the behavior |
| **Minimal tests** | One assertion per test. "and" in name = split it |
| **Real code** | Mock only when unavoidable; prefer real behavior |
| **`require` for guards** | Use `require.New(t)` for preconditions (nil checks, error early returns) |
| **`assert` for verification** | Use `assert.New(t)` for final state checks |
| **`t.Parallel()`** | Independent tests should run in parallel |
| **Goroutine leak detection** | Packages with goroutines need `goleak.VerifyTestMain` in `TestMain` |
| **Integration build tags** | Tests needing external services use `//go:build integration` |

### Common Patterns

```go
// Table-driven with named subtests
tests := []struct {
    name string
    input string
    want  string
}{
    {name: "basic case", input: "hello", want: "hello"},
    {name: "empty string", input: "", want: ""},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        is := assert.New(t)
        must := require.New(t)
        // setup with must (stops on failure)
        result, err := process(tt.input)
        must.NoError(err)
        // verify with is
        is.Equal(tt.want, result)
    })
}
```

### Argument Order (testify)

**Always**: `(expected, actual)` — swapping produces backwards diffs.

### Build Tags

```go
//go:build integration
package mypackage_test
```

Run integration tests: `go test -tags=integration ./...`

### Goroutine Leak Detection

```go
func TestMain(m *testing.M) {
    goleak.VerifyTestMain(m)
}
```

### Test File Naming

- `package foo` (white-box) → `foo_test.go`
- `package foo_test` (black-box) → `foo_test.go`

### NIP-50 Extensions Not Blocked by Local Relay

NIP-50 search extensions (`language:`, `domain:`, `nsfw:`, `sentiment:`, `include:spam`) are passed as-is to relays. Our local Bleve only supports basic full-text search — extensions are handled by external relays that support NIP-50.

---

## NIP-19 Format Convention

All user-facing outputs (CLI, TUI, logs) MUST use NIP-19 bech32 format:
- PubKeys: `npub1...` via `nip19.EncodeNpub(pk)`
- Event IDs: `nevent1...` via `nip19.EncodeNevent(id, relays, author)`
- Private Keys: `nsec1...` via `nip19.EncodeNsec(sk)` (config file only, never in output)

**Internal storage**: Hex format is OK for DB/internal use.
**CLI output**: Always NIP-19.
**Command input**: Accept both hex (64-char) and NIP-19 formats. Use `nip19.ToPointer()` for NIP-19 decoding.

---

## Testing Requirements

<!-- What level of testing is expected -->

(To be filled by the team)

---

## Code Review Checklist

<!-- What reviewers should check -->

---

## TUI Development

### BubbleTea tea.Model View() returns tea.View

**Wrong**:
```go
func (m *model) View() string {  // BUG: returns string
    return someString
}
```

**Correct**:
```go
func (m *model) View() tea.View {
    v := tea.NewView(someString)
    v.AltScreen = true  // for full screen
    return v
}
```

**Why**: `tea.Model` interface requires `View() tea.View`, not `string`. The `tea.View` type wraps content with styling options.

---

### TUI Field Navigation Patterns

**Enter key for navigation**:
- In text inputs (Kind, Tag): Enter moves to next field, does NOT send
- Only textarea (Content) uses Enter to send

**Tab navigation order** (compose TUI example):
- Kind → Tag (or first tag if exists) → Content → Kind (loops)
- When Tag input is empty (`editingTagIndex = -1`), Tab exits to Content
- When editing a tag, Tab cycles through tags

**ESC convention**:
- ESC = quit/close for all TUI views
- In nested edit mode (e.g., editing a tag): first ESC cancels edit, second ESC quits

**Quit key bindings**:
All TUI screens MUST support these keys:
- `esc` → quit/close (graceful)
- `ctrl+c` → immediate program exit via `os.Exit(0)` (not graceful)
- Exception: Event view uses `esc` for quit; `q` is used for "quote" action

**Kill handler pattern (ctrl+c)**:
```go
case tea.KeyPressMsg:
    if key.Matches(msg, m.keys.kill) {
        os.Exit(0)  // immediate kill, no cleanup
    }
```

**Quit handler pattern (ESC)**:
```go
case key.Matches(msg, m.keys.quit):
    if m.subCancel != nil {
        m.subCancel()  // cleanup subscriptions before quit
    }
    if m.isStandalone {
        return m, tea.Quit  // standalone mode: exit program
    }
    return m, bubblon.Close()  // embedded mode: close window, notify parent
```

**Help text for quit**:
- Standard quit: `key.WithHelp("esc", "quit")`
- Kill: `key.WithHelp("ctrl+c", "kill")`
- Event view: `key.WithHelp("esc", "close")`

---

### Tag Input UX Design

**Display format**: `[type] value, value` for normal display

**Edit mode**: When Tabbed to a tag, show `>type:value` with `>` prefix indicating edit mode

**Tag format placeholder**: `e:eventId p:pubkey a:addr t:hashtag r:relay:purpose q:eventId`

**Tag parsing**: `type:value1:value2:...` supports multi-value tags like `r:relay:purpose`

---

### Common TUI Mistakes

#### Mistake: Blur then never refocus

**Symptom**: After pressing Enter in a field, focus is lost completely

**Cause**: Calling `m.input.Blur()` without focusing another element

**Fix**: Always explicitly call `Focus()` on the next element after Blur()

#### Mistake: Tab handler consumed by textarea

**Symptom**: Tab key doesn't switch fields when textarea is focused

**Cause**: Textarea internal handlers may consume Tab before it reaches your key handler

**Fix**: Check for Tab specifically before calling `textarea.Update()`:
```go
if msg.String() == "tab" {
    m.nextField()
    return m, nil
}
// only reach textarea.Update for non-Tab keys
```

#### Mistake: Viewport fills entire window, pushing help off-screen

**Symptom**: Help bar not visible in TUI window (rendered but cut off)

**Cause**: `viewport.SetHeight(msg.Height)` uses full window height. If `View()` appends `help.View()` after `viewport.View()`, the help renders outside the visible area.

**Fix**: Reserve space for fixed UI elements (help bar) at the bottom:
```go
const helpLines = 3  // lines reserved for help bar

func (m *MyView) initViewport(width, height int) {
    m.viewport.SetWidth(width)
    m.viewport.SetHeight(height - helpLines)  // leave room for help
}

func (m *MyView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.viewport.SetWidth(msg.Width)
        m.viewport.SetHeight(msg.Height - helpLines)
    }
    // ...
}

func (m *MyView) View() string {
    return m.viewport.View() + "\n" + m.help.View(m.keys)
}
```

**Prevention**: Always account for fixed UI elements (help bars, status lines) when sizing scrollable content areas.

---

### Window Management with bubblon

**Architecture**: Use `github.com/donderom/bubblon` (or local equivalent `tui/bubblon/controller.go`) for model-stack window management in TUI apps.

**Key concepts**:
- `bubblon.Controller` is a `tea.Model` that holds a `[]tea.Model` stack
- Only the top model receives `Update`/`View` calls
- Commands: `Open(model)`, `Close()`, `Replace(model)`, `Fail(err)`
- When a model calls `Close()`, parent receives `Closed{}` message

**Two approaches**:

| Approach | Description | Use Case |
|----------|-------------|----------|
| **Controller-as-root** | `tea.NewProgram(bubblon.New(rootModel))` — bubblon IS the program root | Full-screen overlay windows |
| **Controller-as-field** | `timeline.model` holds `bubblon.Controller ctrl` as field; `Update()` delegates `ctrl.Update(msg)` | When timeline must remain accessible alongside overlay windows |

**Current project uses Controller-as-field**:
```go
// timeline/model.go
type model struct {
    ctrl    bubblon.Controller  // holds the stack
    // ... other fields
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        // handle timeline keys
    default:
        // delegate to bubblon for window stack
        _, cmd := m.ctrl.Update(msg)
        return m, cmd
    }
}

func (m *model) View() string {
    if len(m.ctrl.Models()) > 0 {
        // render top of bubblon stack (overlay window)
        return m.ctrl.View().String()
    }
    // render timeline list
    return m.list.View()
}
```

**Opening a window** (from timeline or child model):
```go
return m, bubblon.Open(eventViewModel)
// or from EventView:
return ev, ctrl.Update(bubblon.Open(composeModel))
```

**Closing a window** (from compose or child):
```go
return m, bubblon.Close()
// bubblon sends Closed{} to parent when stack drains
```

**Why this replaces WindowManager**:
- `wm.stack` → `ctrl.Models()`
- `wm.Open(win)` → `bubblon.Open(win)`
- `wm.Close(id)` → `bubblon.Close()`
- `wm.WindowCount() > 0` → `len(ctrl.Models()) > 0`
- No more manual `Update` delegation for window messages

---

### BubbleTea v2 Key Event Handling

**Required**: Use `tea.KeyPressMsg` instead of `tea.KeyMsg` in BubbleTea v2.

**Wrong**:
```go
case tea.KeyMsg:
    if key.Matches(msg, m.keys.send) {
        // ctrl+enter handling
    }
```

**Correct**:
```go
case tea.KeyPressMsg:
    if key.Matches(msg, m.keys.send) {
        // ctrl+enter handling
    }
```

**Why**: `tea.KeyMsg` does not fire key press events in BubbleTea v2. Key events arrive as `tea.KeyPressMsg`. Using `tea.KeyMsg` silently breaks all key handling in TUI components.

---

### Standalone vs Embedded TUI Modes

TUI components can run in two modes:

| Mode | How started | ESC behavior |
|------|-------------|--------------|
| **Embedded (wm mode)** | Opened via `bubblon.Open()` from timeline | Returns `bubblon.Close()` → parent receives `Closed{}` |
| **Standalone** | Direct `RunNoteCompose()` / `RunEventView()` | Must return `tea.Quit` to exit |

**Standalone detection pattern**:
```go
type model struct {
    isStandalone bool
    // ...
}

func (m *model) SetStandalone() {
    m.isStandalone = true
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        if key.Matches(msg, m.keys.quit) {
            if m.isStandalone {
                return m, tea.Quit  // exit immediately
            }
            return m, bubblon.Close()  // close window, return to parent
        }
    }
    // ...
}
```

**Why it matters**: Without this check, standalone TUIs cannot exit via ESC because `CloseComposeMsg` or similar close messages have no handler in standalone mode.

---

### TUI State Synchronization

When TUI closes and returns to parent, parent must sync its state:

```go
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case bubblon.Closed:
        // Parent regains control - refresh data to reflect changes
        return m, m.fetchEvents()
    }
    // ...
}
```

**Why**: The parent's view may have stale data after a child window (compose, event detail) made changes. Always refresh in `Closed` handler.
