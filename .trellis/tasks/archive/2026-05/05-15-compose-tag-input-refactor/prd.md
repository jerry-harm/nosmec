# compose tag input UX refactor

## Goal

Redesign compose tag input UX: no more `parseTagInput` style `type:value` strings. Instead, tag values are entered directly, with clear add/edit/delete navigation consistent with pop's page design principles.

## What We Already Know

* Reference: charmbracelet/pop email composer — Prompt as label, focused field gets accent color, no borders, state-driven styling
* Current UX: `parseTagInput` parses `type:value` strings in a single input, confusing and breaks URLs with `:`
* pop's design: label (Prompt) inline with input, focused field highlighted, simple linear flow

## Page Design (from pop reference)

pop email composer — fields stack vertically, Prompt as label:

```
From [input field with gray label]
To [input field]
Cc [input field]
Subject [input field]

[textarea body]

Send [button]

[help bar]
```

Key patterns adopted for compose:
- **Prompt = label** — field label shown inline (e.g., "r: " for relay tag type)
- **Focused = accent color** — active tag input highlighted
- **Minimal decoration** — no borders, color differentiates
- **Empty input always visible** — at bottom of tag list, clear call to action

## Requirements

1. **No `parseTagInput`** — remove or disable it; tag management uses direct value entry
2. **Tag input flow**:
   - Focus tag input → show empty input appended after last tag
   - Backspace on empty input → go back to edit previous tag's value
   - Enter with content → add new tag item
   - Enter empty → blur tag input, focus contentInput
3. **Tag editing**: backspace deletes characters; if already empty, moves to previous tag
4. **Tag display**: `[type] value` per line (type shown as badge, not part of input)
5. **Tag input Prompt**: when editing a tag, show its type as Prompt prefix (e.g., "r: ")
6. **Hint line** below tag input: "enter: add tag | tab: next field | backspace: delete"

## Interaction Design

### Empty tag input (editingTagIndex == -1, no tags yet)

```
Tags:
  [+]  (empty input with placeholder "add tag...")
  enter: add tag | tab: next field | backspace: delete
```

### Tags exist + focus tag input

```
Tags:
  [e] abc123def456...
  [p] def456...
  [r] wss://relay.example
  r:  (empty input showing "r: " prompt — editing relay type)
  enter: add tag | tab: next field | backspace: delete
```

### Enter with content → add tag

```
Tags:
  [e] abc123...
  [p] def456...
  [r] wss://relay.example
  e:  (new empty input after adding relay tag — now editing event type)
```

User types `abc123`, presses Enter → `e:abc123` tag added, focus stays on tag input (new empty slot).

### Backspace on empty → edit previous

User is on `[r]` tag, input is empty, presses Backspace:
- `[r]` tag deleted
- Focus moves to `[p]` tag, input shows its value for editing

### Enter empty → next field

User presses Enter with empty input:
- Tag input blurs, contentInput focuses

## State Design

- `tags []TagValue` — existing tags
- `editingTagIndex int` — index of tag being edited (-1 = empty input mode at end of list)
- `currentTagType string` — type of tag being edited (used to set Prompt prefix)
- `draftValue string` — current input content (for display in input prompt)

When focused on tag input:
- If `editingTagIndex >= 0 && editingTagIndex < len(tags)`: editing existing tag
  - `tagInput.Prompt = tag.Type + ": "` with `m.styles.fieldLabel`
  - `tagInput.SetValue(tag.Values[0])` (or joined if multiple)
- If `editingTagIndex == -1`: empty input mode (adding new tag)
  - `tagInput.Prompt = "e: "` (default type, or last used type)
  - `tagInput.SetValue("")`

## Out of Scope

* Network send behavior
* Bubble Tea Update tests (blocked by textinput.Focus panic)
* Changing tag storage format

## Technical Approach

### Remove parseTagInput dependency

Tags are added/edited by directly constructing `TagValue{Type, Values: []string{value}}`. No string parsing needed.

### Update flow changes

**When tagInput gains focus:**
- If tags exist: set `editingTagIndex = len(tags) - 1` (edit last tag), show its value + type as prompt
- If no tags: `editingTagIndex = -1` (empty mode), default prompt "e: "

**Enter key (addTag binding):**
- If `tagValue != ""`: create new `TagValue{currentTagType, []string{tagValue}}`, append to tags
- If `tagValue == ""`: blur tagInput, focus contentInput
- In both cases: `editingTagIndex = -1` after operation

**Backspace key:**
- If `tagValue == "" && editingTagIndex >= 0`: delete current tag, set `editingTagIndex--`, show previous tag's value
- If `tagValue == "" && editingTagIndex < 0 && len(tags) > 0`: set `editingTagIndex = len(tags) - 1`, show last tag's value
- If `tagValue == "" && editingTagIndex < 0 && len(tags) == 0`: no-op

**Tab key (nextField):**
- If `editingTagIndex < 0`: blur tagInput, focus contentInput
- If `editingTagIndex >= 0`: save current edit, `editingTagIndex++`, show next tag

**Shift+Tab (prevField):**
- Similar to backspace navigation but saves current edit first

### View changes (renderView)

```go
b.WriteString(m.styles.fieldLabel.Render("Tags:"))
b.WriteString("\n")
for i, tag := range m.tags {
    if i == m.editingTagIndex && m.tagInput.Focused() {
        // Show input inline for editing
        b.WriteString("  ")
        b.WriteString(m.styles.fieldLabel.Render(tag.Type + ": "))
        b.WriteString(m.tagInput.View())
        b.WriteString("\n")
    } else {
        b.WriteString(fmt.Sprintf("  [%s] %s\n", tag.Type, strings.Join(tag.Values, ", ")))
    }
}
if m.tagInput.Focused() {
    if m.editingTagIndex < 0 {
        b.WriteString("  ")
        b.WriteString(m.styles.fieldLabel.Render("e: "))
        b.WriteString(m.styles.inputArea.Render(m.tagInput.View()))
        b.WriteString("\n")
    }
    b.WriteString("  enter: add tag | tab: next field | backspace: delete\n")
} else {
    // When not focused, show blurred input
    b.WriteString("  " + m.styles.inputArea.Render(m.tagInput.View()))
    b.WriteString(" | format: e:eventId p:pubkey t:hashtag r:relay q:eventId\n")
}
```

## Key Files

* `tui/compose/model.go` — Update, View, tag management
* `tui/compose/model_test.go` — 23 tests for pure functions

## Research References

* `.trellis/tasks/05-15-compose-tag-input-refactor/research/pop-tag-ux.md`
* `.trellis/tasks/05-15-tddsystematic/research/parseTagInput-separator.md`
* `.trellis/tasks/05-15-tddsystematic/research/tag-input-redesign.md`