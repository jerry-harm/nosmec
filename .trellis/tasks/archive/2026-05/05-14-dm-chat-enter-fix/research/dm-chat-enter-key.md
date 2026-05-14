# Research: DM Chat Enter Key Fix

- **Query**: DM chat UI implementation, Enter key handling, TextInput vs textarea components
- **Scope**: internal
- **Date**: 2026-05-14

## Findings

### Files Found

| File Path | Description |
|---|---|
| `tui/dm/model.go` | Main DM chat UI model — uses `textarea.Model` from `charm.land/bubbles/v2/textarea` |
| `tui/dm/main.go` | Entry point for DM screen via `RunDM()` |
| `tui/dm/styles.go` | Styling for DM components |
| `tui/compose/model.go` | Note compose UI — uses both `textarea.Model` AND `textinput.Model` from `charm.land/bubbles/v2/textinput` |
| `utils/dm.go` | Backend DM utilities (SendDM, ListenForDMs, QueryDMHistory, ListDMConversations) |

### Current Implementation

**DM Chat Input (tui/dm/model.go:37, 98-100)**:
```go
type model struct {
    ...
    ta      textarea.Model  // Line 37
    ...
}

m.ta = textarea.New()
m.ta.Placeholder = "Type a message..."
m.ta.Focus()
```

**Enter Key Handling (tui/dm/model.go:319-334)**:
```go
case tea.KeyPressMsg:
    if m.ta.Focused() {
        switch {
        case key.Matches(msg, m.keys.send):
            content := m.ta.Value()
            m.ta.SetValue("")
            if content = strings.TrimSpace(content); content != "" {
                cmds = append(cmds, m.sendDM(content))
            }
        case key.Matches(msg, m.keys.quit):
            ...
        }
    }
```

The `send` key binding is configured at line 58-60:
```go
send: key.NewBinding(
    key.WithKeys("enter"),
    key.WithHelp("enter", "send"),
),
```

### Problem Identified

The current implementation uses `textarea.Model` from `charm.land/bubbles/v2/textarea`. The `textarea` component by default interprets Enter key within the textarea as a newline character, not as a send action. The `Update()` method at line 346 passes all keypresses to `m.ta.Update(msg)`, which means Enter key presses get consumed by the textarea first, and the `key.Matches(m.keys.send)` at line 322 may never match if the textarea has focus and processes the key first.

The task title says "Fix chat input Enter key conflict using TextInput component" — this suggests replacing `textarea.Model` with `textinput.Model` from `charm.land/bubbles/v2/textinput`, which is a single-line input that does NOT intercept Enter key (Enter passes through to the key handler).

### TextInput Component

The codebase already uses `textinput.Model` in `tui/compose/model.go`:
- Line 12: import `"charm.land/bubbles/v2/textinput"`
- Line 52-54: three textinput fields defined
```go
kindInput    textinput.Model
contentInput textarea.Model
tagInput     textinput.Model
```

The `textinput` component from bubbles is designed for single-line inputs where Enter key is NOT captured by the input itself — it passes through to the key handler. This is the expected behavior for a chat input where Enter sends the message.

### Key Difference

| Component | Enter Key Behavior |
|---|---|
| `textarea.Model` | Captures Enter as newline within the text area |
| `textinput.Model` | Passes Enter through to key handler |

### Code Patterns

**Enter key conflict pattern (tui/dm/model.go:320-327)**:
```go
if m.ta.Focused() {
    switch {
    case key.Matches(msg, m.keys.send):  // Only fires if textarea doesn't consume Enter
        content := m.ta.Value()
        ...
    }
}
taModel, cmd := m.ta.Update(msg)  // textarea also processes the key
```

The fix likely involves switching from `textarea.Model` to `textinput.Model` so Enter is not intercepted by the input field itself, allowing the key binding at line 322 to properly trigger.

### Related Specs

- `.trellis/spec/backend/index.md` — backend spec (may contain TUI layer guidelines)

## Caveats / Not Found

- `textinput.Model` does NOT have multi-line support — if multi-line messages are needed, the fix may involve a different approach (e.g., Ctrl+Enter for newline, Enter only sends)
- The task description mentions "TextInput component" — the fix is likely replacing `textarea` with `textinput`
- No existing `TextInput` component found — the component to use is `textinput.Model` from `charm.land/bubbles/v2/textinput`