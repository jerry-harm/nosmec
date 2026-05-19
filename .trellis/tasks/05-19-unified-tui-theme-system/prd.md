# Unified TUI Theme System

## Goal

Create a centralized `tui/theme` package that provides a single source of truth for all TUI color tokens, replacing 8 scattered style definitions with a consistent semantic naming system. Phase 1 focuses on centralized defaults; Phase 2 adds config/env var support.

## What I already know

### Existing style files (8 files, 32+ hardcoded colors)

| File | Struct | darkBG param |
|------|--------|-------------|
| `tui/timeline/model.go:65` | `styles` | yes |
| `tui/event/styles.go:18` | `eventStyles` | yes |
| `tui/compose/styles.go:25` | `styles` | no |
| `tui/dm/styles.go:20` | `styles` | no |
| `tui/community/discover/model.go:64` | `styles` | no |
| `tui/dm/list/model.go:60` | `styles` | yes |
| `tui/thread/thread.go:36` | inline | no |
| `tui/component/label/model.go:186` | inline | no |

### Color inventory

| Hex | Usages | Proposed Token |
|-----|--------|----------------|
| `#25A065` / `#00875A` | Borders, headers, backgrounds | `Primary` / `PrimaryDark` |
| `#00FF00` / `#00875A` | Bright text, input, success | `TextBright` / `TextBrightDark` |
| `#FFFF00` | Selected item highlight | `Selection` |
| `#04B575` | Status messages | `StatusText` |
| `#00AA00` / `#008800` | Author text | `AuthorText` |
| `#AAAAAA` / `#6B6B6B` / `#888888` | Muted text, labels, help | `TextMuted` |
| `#FFFFFF` / `#333333` | Body text | `TextPrimary` / `TextPrimaryDark` |
| `#FF4444` / `#FF6B6B` | Error, confirm | `Error` |
| `#00AAFF` | Tag items (blue) | `TagColor` |
| `#FFD700` | Community address | `CommunityAddr` |
| `#666666` | Tags, help text | `TextMuted` |
| `#FFFDF5` | Title text | `TitleText` |
| `#333333` | Overlay backgrounds | `OverlayBg` |

## Requirements

### Phase 1: Centralized Theme (this task)

- [ ] Create `tui/theme/theme.go` with `Theme` struct containing all color tokens
- [ ] Provide `DefaultTheme(darkBG bool) *Theme` constructor
- [ ] Each view's `newStyles()` accepts `*Theme` instead of `bool darkBG`
- [ ] `newStyles()` calls become `newStyles(theme.Get())`
- [ ] Phase 1: no config integration — just centralized hardcoded defaults

### Phase 2: Config Integration (follow-up)

- [ ] Add `ThemeConfig` to `config.Config`
- [ ] Viper bindings for env var overrides
- [ ] `GetTheme()` reads from config or returns default

## Technical Approach

### Theme struct

```go
type Theme struct {
    Primary         lipgloss.Color  // #25A065
    PrimaryDark     lipgloss.Color  // #00875A
    TextBright      lipgloss.Color  // #00FF00
    TextBrightDark  lipgloss.Color  // #00875A
    Text            lipgloss.Color  // #FFFFFF
    TextDark        lipgloss.Color  // #333333
    TextMuted       lipgloss.Color  // #888888
    TextMutedDark   lipgloss.Color  // #6B6B6B
    Selection       lipgloss.Color  // #FFFF00
    SelectionDark   lipgloss.Color  // #FFFF00
    StatusText      lipgloss.Color  // #04B575
    AuthorText      lipgloss.Color  // #00AA00
    AuthorTextDark  lipgloss.Color  // #008800
    Error           lipgloss.Color  // #FF4444
    ErrorDark       lipgloss.Color  // #FF6B6B
    TagColor        lipgloss.Color  // #00AAFF
    CommunityAddr   lipgloss.Color  // #FFD700
    OverlayBg       lipgloss.Color  // #333333
    TitleText       lipgloss.Color  // #FFFDF5
    TitleBg         lipgloss.Color  // #25A065
    Border          lipgloss.Color  // #25A065
    BorderDark      lipgloss.Color  // #00875A
}
```

### Package structure

```
tui/theme/
  theme.go    — Theme struct, DefaultTheme, global getter
```

### Migration steps per view

1. Add `theme` field to model struct
2. Change `newStyles(bool darkBG)` → `newStyles(*Theme)`
3. Replace all `lipgloss.Color("#XXXXXX")` with `theme.Token`
4. Remove local color constants and `lightDark()` calls
5. Call site: `m.styles = newStyles(theme.Get())`

### Files to modify

- `tui/theme/theme.go` (new)
- `tui/timeline/model.go` — styles struct + newStyles
- `tui/event/styles.go` → `tui/event/styles.go` (update newStyles signature)
- `tui/compose/styles.go` (update newStyles)
- `tui/dm/styles.go` (update newStyles)
- `tui/community/discover/model.go` — styles struct + newStyles
- `tui/dm/list/model.go` — styles + newStyles
- `tui/thread/thread.go` — inline styles
- `tui/component/label/model.go` — inline styles

## Acceptance Criteria

- [ ] `tui/theme/theme.go` exists with `Theme` struct and `DefaultTheme(darkBG bool)`
- [ ] All 8 style files use centralized theme tokens
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] No hardcoded hex colors remain in view style definitions (only in theme.go)

## Out of Scope (Phase 1)

- Config integration (Phase 2)
- Runtime theme switching without app restart
- Light mode support in all views

## Definition of Done

- All colors sourced from `Theme` struct
- Build + tests pass
- No duplicate color definitions across files