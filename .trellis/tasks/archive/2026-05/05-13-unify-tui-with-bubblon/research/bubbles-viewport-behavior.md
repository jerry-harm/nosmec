# Research: bubblon/bubbles v2 viewport behavior

- **Query**: Check if `SetContent()` calls `Update()` internally, viewport height sizing with `height - helpLines`, and `tea.NewView()` sizing behavior
- **Scope**: internal (local Go module source)
- **Date**: 2026-05-13

## Findings

### Source Code Verified

The source was read from `charm.land/bubbles/v2@v0.x.x` (via `go mod download`) and `charm.land/bubbletea/v2@v2.0.6`.

### 1. Does `SetContent()` automatically call `viewPort.Update()` internally?

**No.** `SetContent()` and `SetContentLines()` do NOT call `Update()` internally. They only update internal state:

```go
// viewport.go:226
func (m *Model) SetContent(s string) {
    m.SetContentLines(strings.Split(s, "\n"))
}

// viewport.go:233
func (m *Model) SetContentLines(lines []string) {
    m.lines = lines
    // ... normalize line endings ...
    m.longestLineWidth = maxLineWidth(m.lines)
    m.ClearHighlights()

    if m.YOffset() > m.maxYOffset() {
        m.GotoBottom()  // Only adjusts yOffset, does NOT re-render
    }
}
```

After calling `SetContent()`, the caller must still call the viewport's `Update()` method (or let the tea runtime call it) to handle keyboard/mouse scroll input. The `Update()` method handles `PageUp`/`PageDown`/`ScrollDown`/`ScrollUp` key messages.

### 2. Issue with viewport height set to `height - helpLines`

If `viewport.SetHeight(height - helpLines)` where `helpLines = 3` and the content (header + body + help text) is larger than the resulting viewport height, the `View()` method will render only the visible portion (clamped by `yOffset`):

```go
// viewport.go:720 (View method)
contentHeight := h - m.Style.GetVerticalFrameSize()
contents := lipgloss.NewStyle().
    Width(contentWidth).
    Height(contentHeight).
    Render(strings.Join(m.visibleLines(), "\n"))
```

The `visibleLines()` function returns only the lines within the `yOffset` window. If content exceeds the viewport height, the user can scroll with keyboard/mouse — but only if `Update()` is being called to process those messages.

**Potential issue**: If `Update()` is not being called on the viewport (e.g., if the viewport is embedded in a larger model and the parent doesn't forward key/mouse messages), the viewport will be non-scrollable and content beyond the first screenful will be inaccessible.

### 3. `tea.NewView()` sizing behavior

`tea.NewView()` is a helper that creates a `tea.View` struct with the content set:

```go
// tea.go:76
func NewView(s string) View {
    var view View
    view.SetContent(s)
    return view
}
```

The returned `View` has no intrinsic width/height — sizing is determined by:
1. The terminal/window dimensions managed by the Bubble Tea runtime
2. Any `Width()`/`Height()` calls on the `Model` that produces the `View`
3. Lipgloss style sizing when the view is rendered

`tea.NewView()` itself does NOT wrap or size the content specially — it simply stores the string. The content is rendered by `cursedRenderer` which uses the terminal dimensions.

### Key Implication for EventView on bubblon stack

When `EventView` is pushed onto a bubblon stack:
- If the viewport is sized `viewport.SetHeight(height - helpLines)` where `height` is the terminal height, the viewport gets `terminal_height - 3` lines
- If content is larger, content beyond the first page is only accessible if:
  1. The viewport's `Update()` is being called with scroll key/mouse messages
  2. The viewport's `View()` is being rendered in the bubblon layer's View

The most common issue: **the viewport renders fine, but scroll doesn't work** — which happens when the parent model doesn't forward `tea.KeyMsg` or `tea.MouseMsg` to the viewport's `Update()` method.

## Related Specs

- None found directly relevant

## Caveats / Not Found

- The exact version of `charm.land/bubbles/v2` in use was not confirmed (only `go mod download` fetched the latest compatible version)
- Whether `bubblon` framework specifically forwards messages to embedded viewport models was not verified — would require checking bubblon source or EventView implementation