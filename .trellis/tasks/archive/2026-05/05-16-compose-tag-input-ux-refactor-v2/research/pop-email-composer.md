# Research: pop email composer implementation

- **Query**: How pop implements its email composer (files, key bindings, UX patterns), tag input, focus/blur handling between fields
- **Scope**: external
- **Date**: 2026-05-16

## Findings

### Key Files

| File Path | Description |
|---|---|
| `model.go` | Main TUI model — state machine, field management, focus/blur, Update loop |
| `keymap.go` | Key bindings (Tab, Shift+Tab, Enter, Esc, Ctrl+C) |
| `email.go` | Email sending logic (SMTP, Resend), attachment handling |
| `attachments.go` | Attachment list item type and delegate |
| `style.go` | Visual styling (colors, padding, labels) |
| `main.go` | CLI entry, argument parsing |

### State Machine

pop uses a linear `State` enum for field navigation:

```go
const (
    editingFrom State = iota
    editingTo
    editingCc
    editingBcc
    editingSubject
    editingBody
    editingAttachments
    hoveringSendButton
    pickingFile
    sendingEmail
)
```

Fields are Bubbles `textinput.Model` (From, To, Cc, Bcc, Subject) and `textarea.Model` (Body). Attachments is a `list.Model`.

### Tab/Enter Navigation

**`keymap.go`** defines:
- `NextInput`: Tab → advances state machine forward
- `PrevInput`: Shift+Tab → retreats state machine
- `Send`: Ctrl+D or Enter (disabled until hoveringSendButton)
- `Attach`: Enter (disabled except in editingAttachments state)
- `Back`: Esc (disabled except in pickingFile state)

**Navigation pattern** (`model.go:Update`):
```go
case key.Matches(msg, m.keymap.NextInput):
    m.blurInputs()
    switch m.state {
    case editingFrom:
        m.state = editingTo
        m.To.Focus()
    case editingTo:
        if m.showCc {
            m.state = editingCc
        } else {
            m.state = editingSubject
        }
    // ... linear progression through states
    case hoveringSendButton:
        m.state = editingFrom  // wraps around
    }
    m.focusActiveInput()
```

**No tag input** — pop treats To/Cc/Bcc as plain comma-separated text fields (`textinput.Model`), not tag-token fields. The comma-separated string is split with `strings.Split(m.To.Value(), ToSeparator)` where `ToSeparator = ","`.

### Focus/Blur Pattern

**`blurInputs()`** — resets all fields to blurred state and removes active styling:
```go
func (m *Model) blurInputs() {
    m.From.Blur()
    m.To.Blur()
    m.Subject.Blur()
    m.Body.Blur()
    if m.showCc {
        m.Cc.Blur()
        m.Bcc.Blur()
    }
    // Reset prompt/text styles to inactive labels
    m.From.PromptStyle = labelStyle
    // ...
}
```

**`focusActiveInput()`** — activates only the current state's field:
```go
func (m *Model) focusActiveInput() {
    switch m.state {
    case editingFrom:
        m.From.PromptStyle = activeLabelStyle
        m.From.TextStyle = activeTextStyle
        m.From.Focus()
        m.From.CursorEnd()
    // ... same pattern for each state
    }
}
```

### Key Observations

1. **Linear state machine** — Tab always advances forward in a fixed sequence; no conditional routing based on field content
2. **No tag/token UX** — To/Cc/Bcc are plain text inputs, not tag pickers
3. **All fields always rendered** — Cc/Bcc are rendered but hidden via `showCc` bool (shown when they have values)
4. **Single Focus() call** — only one field receives Focus() at a time
5. **CursorEnd() after Focus()** — ensures cursor is at end of existing content when focusing a field
6. **Dynamic keymap enablement** — `updateKeymap()` toggles `SetEnabled()` on bindings based on current state (Attach only works in editingAttachments, Send only in hoveringSendButton, etc.)
7. **Enter behavior changes by state** — Enter triggers file picker in editingAttachments, sends in hoveringSendButton, otherwise nothing

### Relevant for compose-tag-input-ux-refactor-v2

- pop's state machine could inform how to handle multi-field navigation
- pop does NOT have a tag-input UX — it's all plain text with comma separators
- The focus/blur pattern (blur all → focus one) is a clean approach to exclusive focus