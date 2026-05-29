# Research: fyne-nested-card

- **Query**: Compact nested reply card styling patterns in Fyne using built-in widgets and containers.
- **Scope**: mixed
- **Date**: 2026-05-29

## Findings

### Files Found

| File Path | Description |
|---|---|
| `gui/app.go` | Current replies are rendered as plain `Label` rows appended into the main card. |

### Code Patterns

The compact card currently renders replies as simple `widget.Label(replyAuthor + ": " + replyBody)` rows inside a nested `container.NewVBox(...)` at `gui/app.go:471-491`. The full thread card repeats the same plain-label pattern at `gui/app.go:528-541`.

Fyne’s built-in `widget.Card` is the official grouping primitive for card-like presentation. Its API exposes `Title`, `Subtitle`, optional `Image`, and a `Content fyne.CanvasObject`, with `NewCard(title, subtitle, content)` and `SetContent`. Source: <https://docs.fyne.io/api/v2/widget/card/>.

For compact composition, Fyne’s box layout docs say vertical boxes keep children stacked at minimum height, which matches dense reply-list rendering. Source: <https://docs.fyne.io/container/box/>. The container package also exposes `NewPadded` for standard inset spacing and `NewStack` / `NewBorder` for layered or framed arrangements. Source: <https://pkg.go.dev/fyne.io/fyne/v2/container>.

The theme package exposes reusable spacing and visual constants, including `theme.Padding()`, `theme.InnerPadding()`, `theme.SeparatorThicknessSize()`, and themed colors/icons. These are the built-in styling hooks available without creating a custom theme. Source: <https://pkg.go.dev/fyne.io/fyne/v2/theme>.

The custom widget guide also shows that a composite list/card item can be wrapped as a custom widget using `widget.BaseWidget` plus `widget.NewSimpleRenderer(...)` when built from a single composite container. Source: <https://docs.fyne.io/extend/custom-widget/>.

### External References

- [widget.Card docs](https://docs.fyne.io/api/v2/widget/card/) — official card/grouping widget.
- [Box layout docs](https://docs.fyne.io/container/box/) — explains dense vertical stacking behavior for compact reply rows.
- [container package docs](https://pkg.go.dev/fyne.io/fyne/v2/container) — documents `NewPadded`, `NewBorder`, `NewStack`, and `NewVBox`.
- [theme package docs](https://pkg.go.dev/fyne.io/fyne/v2/theme) — exposes built-in padding, separator, text-size, and color APIs for styling.
- [Writing a Custom Widget](https://docs.fyne.io/extend/custom-widget/) — documents `widget.NewSimpleRenderer` for composite custom items.

### Related Specs

- Not found in this research pass.

## Caveats / Not Found

No official “nested card reply pattern” example was found in the fetched docs. The available official material is compositional: `widget.Card` for grouping, box/border/padded containers for layout, and theme helpers for spacing and visual consistency.
