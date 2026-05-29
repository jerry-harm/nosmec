# Research: fyne-collapse

- **Query**: Fyne patterns for collapsible sidebar/community lists relevant to the current GUI sidebar implementation.
- **Scope**: mixed
- **Date**: 2026-05-29

## Findings

### Files Found

| File Path | Description |
|---|---|
| `gui/app.go` | Current sidebar collapse logic built from toggle buttons and conditional `VBox` rebuilding. |

### Code Patterns

The current sidebar uses three independent button-driven collapsed booleans and reconstructs sections conditionally with `container.NewVBox(...)`. The key sites are `myFeedToggle`, `globalToggle`, and `communitiesToggle` at `gui/app.go:328-339`, followed by conditional rebuilding at `gui/app.go:357-373`.

Fyne has an official `widget.Accordion` for disclosure-style sections. The API describes it as “a list of AccordionItems” where each item is “represented by a button that reveals a detailed view when tapped,” with `Open`, `Close`, `OpenAll`, and `MultiOpen` support. Source: <https://docs.fyne.io/api/v2/widget/accordion/>.

Fyne also has an official `widget.Tree` for hierarchical data. Its API exposes `ChildUIDs`, `IsBranch`, `CreateNode`, `UpdateNode`, plus branch lifecycle hooks `OnBranchOpened` and `OnBranchClosed`, and open/close helpers like `ToggleBranch`, `OpenBranch`, and `CloseBranch`. This is the official built-in pattern for expandable nested collections rather than manual button stacks. Source: <https://docs.fyne.io/api/v2/widget/tree/>.

If the UI keeps a manual pattern, the container APIs still matter. `container.NewVBox` keeps each child at vertical minimum size, and `container.NewScroll` / `container.Scroll` are the built-in scrolling wrappers. Sources: <https://docs.fyne.io/container/box/> and <https://pkg.go.dev/fyne.io/fyne/v2/container>.

### External References

- [widget.Accordion docs](https://docs.fyne.io/api/v2/widget/accordion/) — official disclosure/collapse widget with per-item open/close state.
- [widget.Tree docs](https://docs.fyne.io/api/v2/widget/tree/) — official hierarchical expandable collection widget with branch open/close callbacks.
- [Box layout docs](https://docs.fyne.io/container/box/) — explains how `VBox`/`HBox` size children by minimum size.
- [container package docs](https://pkg.go.dev/fyne.io/fyne/v2/container) — documents `NewVBox`, `NewScroll`, `NewStack`, `NewPadded`, and notes `NewMax` is deprecated in favor of `NewStack`.

### Related Specs

- Not found in this research pass.

## Caveats / Not Found

No official “Disclosure” widget separate from `Accordion` was found in the fetched Fyne docs. Web search API was unavailable in this session, so findings rely on direct fetches of official Fyne documentation and local code inspection.
