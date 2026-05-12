# Research: glamour URL Handling + Bubble Tea Pager

- **Query**: How glamour handles URL rendering in markdown, TLD support, and bubbletea Pager component
- **Scope**: internal (code) + external (web docs)
- **Date**: 2026-05-12

## Findings

### Files Found

| File Path | Description |
|---|---|
| `tui/window/event/event.go` | EventView using glamour v0.8.0 with `DarkStyleConfig` |
| `tui/window/event/view.go` | Content rendering calling `m.glamour.Render(content)` |
| `go/pkg/mod/github.com/charmbracelet/glamour@v0.8.0/ansi/link.go` | Link rendering (URL display after link text) |
| `go/pkg/mod/github.com/charmbracelet/glamour@v0.8.0/ansi/elements.go` | Link/AutoLink element creation from markdown AST |
| `go/pkg/mod/github.com/yuin/goldmark@v1.7.8/extension/linkify.go` | URL detection regex patterns |
| `go/pkg/mod/github.com/charmbracelet/bubbletea@v1.3.10/standard_renderer.go` | Deprecated pager mention in renderer |
| `go/pkg/mod/github.com/charmbracelet/bubbles@v1.0.0/viewport/viewport.go` | `SetContent` for pager-like viewport scrolling |

### Code Patterns

#### URL Detection (goldmark Linkify Extension)

**File**: `goldmark@v1.7.8/extension/linkify.go:14-16`

```go
var wwwURLRegxp = regexp.MustCompile(`^www\.[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-z]+(?:[/#?][-a-zA-Z0-9@:%_\+.~#!?&/=\(\);,'">\^{}\[\]` + "`" + `]*)?`)

var urlRegexp = regexp.MustCompile(`^(?:http|https|ftp)://[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-z]+(?::\d+)?(?:[/#?][-a-zA-Z0-9@:%_+.~#$!?&/=\(\);,'">\^{}\[\]` + "`" + `]*)?`)
```

**TLD pattern**: `[a-z]+` — matches any sequence of lowercase letters.

- `.i2p` — matches (`i2p` is all lowercase letters)
- `.onion` — matches (`onion` is all lowercase letters)
- `.local` — matches

**Protocol requirement**: `urlRegexp` requires `http://`, `https://`, or `ftp://` prefix. Naked URLs (e.g. `www.example.i2p`) go through `wwwURLRegxp` which requires `www.` prefix.

**AutoLink (bracket-style `<>`)**: `goldmark@v1.7.8/parser/auto_link.go` uses `util.FindURLIndex()` and `util.FindEmailIndex()` for URL detection within angle brackets.

#### Link Rendering in Glamour

**File**: `glamour@v0.8.0/ansi/link.go:38-48`

```go
u, err := url.Parse(e.URL)
if err == nil && "#"+u.Fragment != e.URL {
    el := &BaseElement{
        Token:  resolveRelativeURL(e.BaseURL, e.URL),
        Prefix: " ",
        Style:  ctx.options.Styles.Link,
    }
    if err := el.Render(w, ctx); err != nil {
        return err
    }
}
```

Glamour renders links with the URL shown after the link text in parentheses, styled with `Link` style. For example: `\[link text](http://example.i2p) → link text (http://example.i2p)`

#### AutoLink Detection in Elements

**File**: `glamour@v0.8.0/ansi/elements.go:235-262`

```go
case ast.KindAutoLink:
    n := node.(*ast.AutoLink)
    u := string(n.URL(source))
    // ... children setup ...
    return Element{
        Renderer: &LinkElement{
            Children: children,
            BaseURL:  ctx.options.BaseURL,
            URL:      u,
        },
    }
```

The `AutoLink` node comes from goldmark's parser for `<http://example.i2p>` style links.

#### Bubble Tea Viewport as Pager

**File**: `bubbles@v1.0.0/viewport/viewport.go:124-133`

```go
// SetContent set the pager's text content.
func (m *Model) SetContent(s string) {
    s = strings.ReplaceAll(s, "\r\n", "\n")
    m.lines = strings.Split(s, "\n")
    m.longestLineWidth = findLongestLineWidth(m.lines)
    if m.YOffset > len(m.lines)-1 {
        m.GotoBottom()
    }
}
```

The `viewport.Model` from `charm.land/bubbles/v2` already has pager-like behavior with `SetContent`, scroll position tracking, and key-based navigation.

### External References

- [glamour GitHub](https://github.com/charmbracelet/glamour) — Stylesheet-based markdown rendering
- [goldmark Linkify extension](https://pkg.go.dev/github.com/yuin/goldmark@v1.7.8/extension#Linkify) — URL detection configuration options (`WithLinkifyURLRegexp`, `WithLinkifyWWWRegexp`)
- [bubbletea v2](https://pkg.go.dev/charm.land/bubbletea/v2) — TUI framework
- [bubbles viewport](https://pkg.go.dev/charm.land/bubbles/v2/viewport) — Scrollable viewport component

### Related Specs

- `.trellis/spec/backend/quality-guidelines.md` — relevant for lint/style guidelines

## Caveats / Not Found

### No Built-in Pager Component in Bubble Tea v2
There is **no dedicated `Pager` component** in `charm.land/bubbletea/v2` or `bubbles/v2`. The `viewport.Model` from `bubbles/v2` provides scrollable content but is not called a "pager". The "pager" reference in `standard_renderer.go:567` is about using the standard renderer for pager-like use cases.

### TLD Compatibility for `.i2p` / `.onion`
The goldmark `linkify` regex `[a-z]+` will match `i2p` and `onion` since they are purely alphabetic. However:
- **`.onion`**: Special-use TLD reserved for Tor hidden services. Browsers handle these specially; standard URL parsers may not recognize them without special configuration.
- **No known issues** with `i2p` — it should work in the current regex.
- **Not configurable**: The TLD regex is hardcoded in goldmark's `linkify.go`. To extend TLD support (e.g., multi-script IDN support), you would need to either:
  1. Fork goldmark and modify the regex
  2. Use `WithLinkifyURLRegexp` with a custom regex to replace the default

### Glamour Version
The project uses `glamour@v0.8.0`. The newer `v0.10.0` is available but has not been reviewed for changes to URL handling.