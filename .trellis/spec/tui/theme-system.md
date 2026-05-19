# TUI Theme System

> Centralized color token management for all TUI views.

---

## Overview

All TUI color values are centralized in `tui/theme/theme.go`. The `Theme` struct holds semantic color tokens; each view's `newStyles()` receives a `*Theme` instead of hardcoding hex values.

---

## Theme Struct

```go
type Theme struct {
    Primary       color.Color  // #25A065 — borders, headers, backgrounds
    PrimaryDark   color.Color  // #00875A — dark mode primary
    TextBright    color.Color  // #00FF00 — bright text, input
    TextBrightAlt color.Color  // #00875A — dark mode text bright
    Text          color.Color  // #FFFFFF — primary text
    TextDark      color.Color  // #333333 — text on light backgrounds
    TextMuted     color.Color  // #AAAAAA — muted text (light mode)
    TextMutedDark color.Color // #6B6B6B — muted text (dark mode)
    TextMutedAlt  color.Color  // #888888 — alternative muted
    Selection     color.Color // #FFFF00 — selected item highlight
    StatusText    color.Color // #04B575 — status messages
    AuthorText    color.Color // #00AA00 — author name text (light)
    AuthorTextAlt color.Color // #008800 — author text (dark)
    Error         color.Color // #FF4444 — error text
    ErrorAlt      color.Color // #FF6B6B — error variant (confirm)
    TagColor      color.Color // #00AAFF — tag/item highlights
    CommunityAddr color.Color // #FFD700 — community address display
    OverlayBg     color.Color // #333333 — overlay backgrounds
    TitleText     color.Color // #FFFDF5 — title text
    TitleBg       color.Color // #25A065 — title background
    Border        color.Color // #25A065 — border color
    BorderDark    color.Color // #00875A — dark mode border
}
```

## Usage

```go
// In newStyles()
func newStyles(t *theme.Theme) styles {
    return styles{
        title: lipgloss.NewStyle().
            Foreground(t.TitleText).
            Background(t.TitleBg).
            Padding(0, 1),
    }
}

// Caller
m.styles = newStyles(theme.DefaultTheme(false))
// or for dark mode:
m.styles = newStyles(theme.DefaultTheme(true))
```

## API

- `theme.DefaultTheme(darkBG bool) *Theme` — returns dark or light theme
- `theme.Default() *Theme` — returns light theme (convenience)
- `theme.NewTheme(...) *Theme` — create custom theme (future)

## Migration Notes

- `lipgloss.Color` in v2 is a **function** returning `color.Color`, not a type
- Theme struct fields use `image/color.Color` to match `Style.Foreground(c color.Color)`
- All views use value-type `list.DefaultDelegate` (not pointer) for delegate params

## Files

- `tui/theme/theme.go` — Theme struct, defaults
- All `styles.go` and inline style definitions migrated to use `*theme.Theme`

## Future (Phase 2)

- Config integration via viper/env vars
- Runtime theme switching
- Light mode support in all views