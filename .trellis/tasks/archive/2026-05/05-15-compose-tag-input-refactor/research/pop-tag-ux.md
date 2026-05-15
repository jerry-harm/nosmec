# Research: pop email composer page design

## Reference

charmbracelet/pop — email composer TUI

## Key Page Design Patterns

### Layout

```
From [input]
To [input]
Cc [input]
Bcc [input]
Subject [input]

[textarea body]

Attachments [list]

Send [button — gray inactive, accent+yellow when active]

[help bar]
```

### Prompt = Label

Each field uses `textinput.Prompt` as the inline label:
- `from.Prompt = "From "` — "From " rendered in gray before the input cursor
- `to.Prompt = "To "`
- When focused: `m.From.PromptStyle = activeLabelStyle` (accent color)
- When blurred: `m.From.PromptStyle = labelStyle` (gray)

### State-Driven Styling

- **Focused field**: Prompt in accent color (yellow), text in white
- **Blurred field**: Prompt in gray, text in light gray
- **No borders** — color differentiates active from inactive
- **Send button**: dark gray when inactive, accent background + yellow text when active (hoveringSendButton state)

### Field Navigation

State machine: `editingFrom → editingTo → editingCc → editingBcc → editingSubject → editingBody → editingAttachments → hoveringSendButton`

Tab/Shift+Tab cycle through states, calling `blurInputs()` then `focusActiveInput()`.

### Cursor End on Focus

```go
func (m *Model) focusActiveInput() {
    case editingFrom:
        m.From.Focus()
        m.From.CursorEnd()  // cursor goes to end when focusing
```

### Minimal Hint System

Help bar at bottom via `m.help.View(m.keymap)` — shows only bindings relevant to current state (keymap updated via `m.updateKeymap()`).

## What This Means for nosmec compose

For the tag input section, we adopt:

1. **Prompt as label prefix**: `tagInput.Prompt = "r: "` when editing a relay tag
2. **State-driven prompt style**: accent when focused, gray when blurred
3. **Hint line below input** instead of hints mixed into tag list
4. **No borders around fields** — lipgloss styles already do this

## Files Examined

| File | Relevant Lines |
|------|----------------|
| `pop/model.go` | 87-211 (NewModel), 229-364 (Update), 428-477 (View) |
| `pop/style.go` | 10-41 (style definitions) |
| `pop/keymap.go` | 1-90 (key bindings) |