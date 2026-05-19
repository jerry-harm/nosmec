# Complete Theme System Migration & Code Cleanup

## Goal

Finish migrating all remaining hardcoded `lipgloss.Color("#...")` values to centralized theme tokens, and remove dead code from the label component.

## Requirements

1. **Add viewport border token** — Add `ViewportBorder` / `ViewportBorderDark` to `theme.Theme` matching the existing `#25A065` / `#00875A` split used in `event/event.go:initViewport`
2. **Add input placeholder token** — Add `InputPlaceholder` (`#666666`) for `compose/model.go:tagInput`
3. **Add spinner token** — Add `Spinner` (`#00FF00`) for `compose/model.go:spinner`
4. **Migrate event/view.go hardcoded borderColor** — Use `t.Theme` access in `initViewport`
5. **Migrate compose/model.go hardcoded colors** — Use `m.styles.t` theme access
6. **Delete dead label code** — `component/label/model.go` static vars and `View()` method on `Model` are unreachable; remove them

## Acceptance Criteria

- [ ] `go build ./...` passes
- [ ] `go test ./...` passes
- [ ] No `lipgloss.Color("#...")` calls remain outside `tui/theme/theme.go`
- [ ] Dead label static vars and Model.View removed

## Technical Approach

- Add `ViewportBorder color.Color`, `InputPlaceholder color.Color`, `Spinner color.Color` to `Theme` struct
- Update `DefaultTheme()` light/dark variants to populate new fields
- Replace inline hardcoded colors with theme token accessors
- Delete unreachable code path in label component

## Out of Scope

- Light/dark runtime theme switching
- Config-driven theme values
- Changes to label.RenderLabel (already uses theme)