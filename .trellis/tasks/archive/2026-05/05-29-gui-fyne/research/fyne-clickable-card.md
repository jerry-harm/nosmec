# Research: fyne-clickable-card

- **Query**: Correct Fyne pattern for making a composite card clickable without covering or hiding its visible content.
- **Scope**: mixed
- **Date**: 2026-05-29

## Findings

### Files Found

| File Path | Description |
|---|---|
| `gui/app.go` | Current compact post card uses `container.NewStack` with an empty button layered over content. |

### Code Patterns

The current compact card builds visible content into `mainCard`, wraps it in `container.NewBorder`, then overlays an empty `widget.Button` in `container.NewStack(...)` at `gui/app.go:469-504`. The clickable layer is `tapBtn` at `gui/app.go:496-503`.

Fyne’s official input contract is the `fyne.Tappable` interface, which is defined as `Tapped(*PointEvent)`. The docs state that `Tappable` describes any `CanvasObject` that can be tapped and “should be implemented by buttons etc that wish to handle pointer interactions.” Source: <https://docs.fyne.io/api/v2/fyne/tappable/>.

Fyne’s custom widget guide says a widget should separate state/behavior from rendering, normally embed `widget.BaseWidget`, and provide a renderer. For simple composite widgets built from a single container or composite `CanvasObject`, the guide explicitly recommends `widget.NewSimpleRenderer(...)`. It gives an example of a custom list item widget composed from labels and a `container.NewBorder(...)`, then returned via `widget.NewSimpleRenderer(c)`. Source: <https://docs.fyne.io/extend/custom-widget/>.

For layering specifically, `container.NewStack` draws objects in the order passed, with the last one top-most. The layout docs say all items are sized to the container and the last object is drawn above earlier ones. Source: <https://docs.fyne.io/container/stack/>. That means a full-size button layer is expected to sit above the content layer.

### External References

- [fyne.Tappable docs](https://docs.fyne.io/api/v2/fyne/tappable/) — official tap-handling interface for tappable canvas objects.
- [Writing a Custom Widget](https://docs.fyne.io/extend/custom-widget/) — official guidance for composite interactive widgets and `widget.NewSimpleRenderer`.
- [Stack layout docs](https://docs.fyne.io/container/stack/) — confirms the last object is top-most when stacking.

### Related Specs

- Not found in this research pass.

## Caveats / Not Found

No official `widget.Card` click callback exists in the fetched docs; `widget.Card` is documented as a content-grouping widget with title, subtitle, image, and content only. Source: <https://docs.fyne.io/api/v2/widget/card/>.
