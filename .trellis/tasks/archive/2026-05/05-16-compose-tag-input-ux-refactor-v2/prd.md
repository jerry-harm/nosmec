# compose tag input UX refactor v2

## Goal

Simplify compose tag input to use straightforward tab navigation and array-based tag editing, inspired by pop's email composer UX. Remove the complex `editingItemIndex` logic that caused navigation bugs.

## What We Know

- Reference: charmbracelet/pop email composer — label as Prompt, focused field highlighted, minimal decoration
- Current implementation has `editingTagIndex` + `editingItemIndex` two-level indexing for tags, causing confusing tab/backspace behavior
- User wants: simple array format `["a","b"]` per tag, Tab cycles to new slot

## State Design

```go
tags         []Tag         // []string, each tag is a simple string array (e.g., ["eventid"] or ["pubkey1", "pubkey2"])
editingIndex int           // -2 = not in tag-edit mode, -1 = new empty slot, >= 0 = editing tags[editingIndex]
```

### State Semantics

| editingIndex | Meaning |
|--------------|---------|
| -2 | Not in tag-edit mode (focused on kind/content) |
| -1 | Tag input focused, showing new empty slot |
| >= 0 | Editing tags[editingIndex] |

## Tag Structure

Each tag in the `tags` slice is a `Tag` (= `[]string`). This allows multi-value tags like `["p", "pubkey1", "pubkey2"]` for relay lists.

## Core Interactions

### Focus tag input (from kindInput, contentInput, or initial)
- `editingIndex = -1` (empty slot mode)
- `tagInput.SetValue("")`

### Tab (when tagInput focused)
- `editingIndex == -1` (empty slot):
  - If user typed something → add as new tag first
  - Then blur tagInput, focus contentInput
- `editingIndex >= 0`:
  - Save current edit
  - `editingIndex++`
  - If `editingIndex >= len(tags)` → `editingIndex = -1`, blur, go to contentInput
  - Else → load tags[editingIndex] into tagInput

### Shift+Tab (when tagInput focused)
- `editingIndex == -1`:
  - Go to last tag: `editingIndex = len(tags) - 1`, load into tagInput
- `editingIndex >= 0`:
  - Save current edit
  - `editingIndex--`
  - If `editingIndex < 0` → blur, focus contentInput
  - Else → load tags[editingIndex] into tagInput

### Enter (addTag key)
- `tagValue != ""`:
  - `editingIndex == -1` → append new tag, stay in empty slot mode
  - `editingIndex >= 0` → update tags[editingIndex] or insert if desired
- `tagValue == ""`:
  - blur tagInput, focus contentInput

### Backspace (when tagInput focused, value empty)
- `editingIndex == -1` && `len(tags) > 0`:
  - Go to last tag: `editingIndex = len(tags) - 1`
- `editingIndex >= 0`:
  - Delete tags[editingIndex]
  - If empty → `editingIndex = -1` else `editingIndex = min(editingIndex, len(tags)-1)`

## Tab Navigation Flow (key requirement)

```
┌─────────────────────────────────────────────────────┐
│  kindInput ──Tab──> tagInput (empty slot, -1)       │
│                              │                       │
│                              │ Tab with content      │
│                              ↓                       │
│                      add new tag → stay (-1)         │
│                              │                       │
│                              │ Tab on empty          │
│                              ↓                       │
│                      contentInput                    │
│                                                      │
│  contentInput ─Shift+Tab──> tagInput (empty, -1)    │
│  contentInput ────Tab──> kindInput                  │
└─────────────────────────────────────────────────────┘

Tag cycling:
  [-1] empty ─Tab──> [0] tag0 ─Tab──> [1] tag1 ─Tab──> [-1] empty ─Tab──> contentInput
                          ↑
                          │ Shift+Tab
                          └─── ... ←── [last] ← Shift+Tab (empty)
```

## View Display

```
Tags:
  [e] abc123def456...
  [p] def456...
  [r] wss://relay.example
  >  (empty input when editingIndex == -1)
  enter: add tag | tab: next | backspace: delete
```

When editingIndex >= 0, show tag values directly, no inline editing (unless we add it back later).

## Testing Strategy (TDD)

Following tui-testing skill patterns:

1. **Unit tests** for tag parsing/transform logic (if any)
2. **Component tests** for Update state transitions:
   - Tab navigation between tags
   - Enter to add tag
   - Backspace to delete
   - Shift+Tab to go back
3. **Golden file tests** for View rendering:
   - Empty state
   - Single tag
   - Multiple tags
   - With editingIndex states

### Test Pattern (from skill)

```go
func TestComposeTabNavigation(t *testing.T) {
    m := newTestModel()
    m.tagInput.Focus()

    // Start with editingIndex = -1 (empty slot)
    assert.Equal(t, -1, m.editingIndex)

    // Tab with empty input → go to content
    m, _ = m.Update(tea.KeyPressMsg{Text: "tab"})
    assert.False(t, m.tagInput.Focused())
    assert.True(t, m.contentInput.Focused())
}
```

## Key Files

- `tui/compose/model.go` — Update, View, tag management
- `tui/compose/model_test.go` — existing tests to adapt
- `tui/compose/update_test.go` — existing tests to adapt

## Scope

- Simplify tag editing state machine
- Fix Tab navigation to work correctly
- Maintain TDD approach with good test coverage
- Do NOT change tag storage format (keep as `[]Tag` = `[]string`)
- Do NOT add complex multi-value inline editing

## Out of Scope

- Network send behavior (already done)
- Visual redesign (colors, styles)
- Other compose fields (kind, content)