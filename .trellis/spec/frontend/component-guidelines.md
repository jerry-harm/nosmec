# Component Guidelines

> How Fyne UI components are built in this project.

---

## Overview

This project uses Fyne desktop UI composition, not a browser framework. Prefer small helpers that return `fyne.CanvasObject` values and keep state in `gui/app.go` unless a reusable widget is clearly needed.

---

## Component Structure

### App shell composition

- Top-level window layout lives in `gui/app.go`.
- Use `container.NewBorder(...)` for app-bar style layout.
- Use `container.NewHSplit(...)` for sidebar + content shell.
- Keep top navigation, sidebar, post card body, and action row as separate builder helpers.

### Post card composition

- Outer posts are plain `widget.Card` containers.
- Do not make the whole outer post card the primary navigation target.
- Nested reply card is the preferred thread-entry affordance in list/preview mode.
- Post actions belong in an explicit action row, currently `widget.Toolbar`.

---

## Props Conventions

For Fyne helper functions, prefer explicit parameters over hidden globals when feasible.

Examples:

```go
func buildReplyCard(replies []Reply, extraCount int, preview bool, onOpen func()) fyne.CanvasObject
func buildPostActionRow(onOpen func()) fyne.CanvasObject
```

Use `nil` callbacks to disable preview-only interactions in full-thread mode instead of building separate duplicate components.

---

## Styling Patterns

- Use built-in Fyne widgets first: `widget.Card`, `widget.Toolbar`, `widget.Button`, `widget.Entry`, `widget.Accordion`.
- Prefer theme-derived colors and spacing over hard-coded values.
- Use a distinct nested background for embedded reply cards so they read as a separate interaction surface.
- Keep nested reply spacing tight; avoid tall stacked padding for dense discussion previews.

---

## Interaction Patterns

### Preferred

```go
content.Add(buildReplyCard(replies, post.ExtraCount, preview, onOpen))
content.Add(buildPostActionRow(onOpen))
```

### Avoid

```go
// Avoid whole-card navigation for outer posts.
return newTappableCard(buildPostCardContent(post, true), onOpen)
```

Why:

- Whole-card tap made the interaction model ambiguous once reply cards and action rows were added.
- Explicit reply entry and action buttons map better to future Android app-bar/action-row layouts.

---

## Common Mistakes

### Common Mistake: Overusing custom widgets

Use a custom Fyne widget only when built-in composition cannot express the desired interaction clearly.

Bad signs:

- wrapping plain cards just to make the whole area tappable
- adding custom renderers where `container.NewBorder`, `widget.Card`, or `widget.Toolbar` would be enough

### Common Mistake: Coupling tests to container internals

Widget tests should prefer user-visible structure and behavior over exact internal layout object counts unless the layout itself is the contract.
