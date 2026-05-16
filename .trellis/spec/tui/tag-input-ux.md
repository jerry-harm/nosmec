# TUI Tag Input UX Specification

> Tag input patterns for compose window in this project.

---

## Overview

The compose window tag input uses a JSON list format for tag values. Tags are stored as `[]Tag` where `type Tag = []string`.

---

## State Machine

### editingIndex States

| Value | Meaning |
|-------|---------|
| -2 | Not in tag mode (focused on kind or content) |
| -1 | Empty slot (ready to add new tag) |
| >= 0 | Editing tags[editingIndex] |

---

## Tag Input Format

**Input format**: JSON array string, e.g. `["e","event123"]` or `["p","pubkey1","pubkey2"]`

**Placeholder**: `["tag1","tag2"]`

---

## Navigation Map

### Tab (nextField)

```
kindInput → tag[0] → tag[1] → ... → tag[last] → empty slot → contentInput
```

### Shift+Tab (prevField)

```
contentInput → tag[last] → ... → tag[1] → tag[0] → empty slot → kindInput
```

---

## Key Behaviors

### Enter (addTag)

| Condition | Action |
|-----------|--------|
| tagValue != "" | Parse as JSON array, add/replace tag, set editingIndex=-1, clear input |
| tagValue == "" | Blur tagInput, focus contentInput, set editingIndex=-2 |

### Backspace on empty input

| Condition | Action |
|-----------|--------|
| editingIndex < 0 && len(tags) > 0 | Set editingIndex = len(tags)-1, load tag into input |
| editingIndex >= 0 | Delete tags[editingIndex], set editingIndex to new last |

### Tab/Shift+Tab with tags

| Context | Action |
|---------|--------|
| From kindInput (Tab) | Focus tag[0] |
| From kindInput (Shift+Tab) | Focus empty slot |
| From contentInput (Shift+Tab) | Focus tag[last] |
| From tag[i] (Tab) | Focus tag[i+1] or empty slot |
| From tag[i] (Shift+Tab) | Focus tag[i-1] or kindInput (if i==0) |

---

## Tag Display in View

### Normal display (not focused)
```
  [e] event1, pubkey1
  [p] pubkey2
```

### When editingIndex >= 0 and tagInput focused
```
  > ["e","event1","pubkey1"]
```

### When editingIndex == -1 and tagInput focused
```
  > (empty with placeholder)
```

---

## Helper Functions

### tagToListString

Converts a Tag to JSON array string for display in input.

```go
func tagToListString(tag Tag) string
// Example: ["e","event123"]
```

### parseTagListInput

Parses JSON array string input into a Tag.

```go
func parseTagListInput(s string) (Tag, error)
// Input: `["e","event123"]`
// Output: Tag{"e", "event123"}
```

---

## Common Mistakes

### Confusion: editingIndex vs array index

`editingIndex` is NOT the tag's first element. It's an index into the `tags` array.

- `tags[editingIndex]` = the whole tag (all its values)
- `tags[editingIndex][0]` = the tag type (e.g., "e", "p")

### Confusion: empty slot vs no tags

When `len(tags) == 0`:
- Tab from kindInput goes to empty slot (editingIndex = -1)
- Backspace does nothing

### Wrong: Assuming tab always goes to content

Tab from last tag goes to empty slot (editingIndex = -1), NOT to contentInput.
Only Tab from empty slot goes to contentInput.

---

## Implementation Notes

- Single `editingIndex` replaces old dual `editingTagIndex` + `editingItemIndex`
- No inline editing of individual tag values within a tag
- Tag replacement happens on Enter (parses full JSON array)
- Navigation is linear and consistent with visual order