# Research: fyne-i18n

- **Query**: Official i18n/localization support patterns in Fyne relevant to the current GUI app.
- **Scope**: mixed
- **Date**: 2026-05-29

## Findings

### Files Found

| File Path | Description |
|---|---|
| `gui/app.go` | Current UI already wraps visible strings with `T(...)` in many places. |

### Code Patterns

The current GUI already routes many visible strings through `T(...)`, for example in top-bar buttons at `gui/app.go:276-299`, sidebar labels at `gui/app.go:307-375`, empty-state text at `gui/app.go:420-422`, and post/reply labels at `gui/app.go:459-487` and `gui/app.go:516-538`.

Fyne has official app translation support since v2.5. The docs say to use the `fyne.io/fyne/v2/lang` package and mark strings with `lang.L(...)` / `lang.Localize(...)`, or use keyed variants `lang.X(...)` / `lang.LocalizeKey(...)` when a stable ID is preferred over the literal fallback string. Source: <https://docs.fyne.io/explore/translations/>.

The same guide documents translation files as JSON, typically embedded with Go `embed`, then loaded once with `lang.AddTranslationsFS(translations, "translation")` between `app.New()` and `Run()`. It also documents file naming like `en.json`, `fr.json`, `zh.json`. Source: <https://docs.fyne.io/explore/translations/>.

The package API confirms the runtime helpers available in `fyne.io/fyne/v2/lang`: `AddTranslations`, `AddTranslationsFS`, `AddTranslationsForLocale`, `Localize`, `LocalizeKey`, `LocalizePlural`, `LocalizePluralKey`, and `SystemLocale`. It also exposes aliases `L`, `N`, `X`, and `XN`. Source: <https://pkg.go.dev/fyne.io/fyne/v2/lang>.

Pluralization is officially supported. The docs show `lang.N(...)` and JSON plural forms using `one` and `other`. Source: <https://docs.fyne.io/explore/translations/>.

### External References

- [Adding app translations](https://docs.fyne.io/explore/translations/) — official high-level guide for marking strings, embedding JSON translation files, loading them, and plural handling.
- [lang package docs](https://pkg.go.dev/fyne.io/fyne/v2/lang) — official API reference for localization helpers and loading functions.

### Related Specs

- Not found in this research pass.

## Caveats / Not Found

The fetched docs show official localization support in `fyne.io/fyne/v2/lang`, but no separate higher-level Fyne “i18n framework” beyond this package was found.
