# Theme Config Integration

## Goal

Add theme color configuration via viper (config file / env vars) at startup. No runtime switching.

## Requirements

1. **Add theme color keys to viper defaults** in `config/config.go`:
   - `theme.primary`, `theme.primary_dark`, `theme.text_bright`, etc.
   - All Theme struct fields mapped to viper keys
   - Defaults match existing hardcoded values

2. **Create `LoadTheme(*viper.Viper) *Theme`** in `tui/theme/theme.go`:
   - Reads viper values and builds a Theme instance
   - Falls back to defaults if keys absent

3. **Replace `DefaultTheme(darkBG)` calls with config-driven theme**:
   - At startup, create theme via `LoadTheme(config.GetViper())`
   - Pass to all `newStyles()` calls

## Acceptance Criteria

- [ ] `go build ./...` passes
- [ ] `go test ./...` passes
- [ ] Theme colors configurable via config file (YAML) or env vars (`NOSMEC_THEME_PRIMARY=#FF0000`)
- [ ] All existing views display same colors as before (defaults unchanged)

## Technical Approach

### Config keys (viper)

```yaml
theme:
  primary: "#25A065"
  primary_dark: "#00875A"
  text_bright: "#00FF00"
  text_bright_alt: "#00875A"
  ...
```

### LoadTheme function

```go
func LoadTheme(v *viper.Viper) *Theme {
    return &Theme{
        Primary:       lipgloss.Color(v.GetString("theme.primary")),
        PrimaryDark:   lipgloss.Color(v.GetString("theme.primary_dark")),
        ...
    }
}
```

### Viper defaults

Register all theme defaults in `loadConfig()` alongside existing defaults.

## Out of Scope

- Runtime theme switching
- Per-view different themes
- Live config reload without restart