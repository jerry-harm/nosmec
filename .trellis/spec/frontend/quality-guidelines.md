# Quality Guidelines

> Code quality standards for Fyne UI development.

---

## Overview

The GUI layer should stay simple, compositional, and testable with Fyne's own test helpers. Prefer widget interaction tests for UI behavior and plain unit tests for pure helpers.

---

## Forbidden Patterns

### Don't: Use whole-card tap as the default interaction surface

```go
return newTappableCard(buildPostCardContent(post, true), onOpen)
```

Why it's bad:

- makes reply entry and action row semantics unclear
- complicates future mobile adaptation
- tends to force unnecessary custom widget code

Instead:

```go
content.Add(buildReplyCard(replies, post.ExtraCount, preview, onOpen))
content.Add(buildPostActionRow(onOpen))
```

### Don't: Add browser-style abstractions to Fyne layout code

Do not introduce framework-like indirection for simple shell layout. Prefer direct `container.NewBorder`, `container.NewVBox`, `container.NewHSplit`, `widget.Card`, and `widget.Toolbar` composition.

---

## Required Patterns

- Use `widget.Toolbar` for compact icon action rows when the control is semantically a row of actions.
- Keep top navigation as app-bar composition, not a toolbar, when it mixes text mode switches, search, and account actions.
- Use one nested reply card for grouped preview replies, not one separate reply card per reply.
- Prefer explicit callbacks (`func()`) over hidden global mutation when wiring component actions.

---

## Testing Requirements

### Required test split

- `gui/app_test.go`: pure logic tests only
  - locale normalization
  - scope filtering
  - subscription/community extraction
- `gui/ui_test.go`: widget composition and interaction tests using `fyne/test`
  - top bar structure
  - sidebar / accordion presence
  - toolbar action presence
  - reply-card thread entry behavior

### Minimum assertions for widget tests

- Verify user-visible structure, not only non-nil object creation.
- Prefer checking interaction outcomes (`currentView`, `currentPost`, collapse state) over exact container internals.
- Only assert internal object shape when that shape is itself part of the contract.

### Verification commands

```bash
go test ./gui -count=1
go test ./...
go vet ./...
go build ./...
```

---

## Code Review Checklist

- Is the UI built from standard Fyne composition before introducing a custom widget?
- Is the main navigation affordance explicit and discoverable?
- Are preview-only interactions disabled in full-thread mode?
- Are tests placed in `ui_test.go` vs `app_test.go` appropriately?
- Do widget tests validate behavior instead of only constructor success?
