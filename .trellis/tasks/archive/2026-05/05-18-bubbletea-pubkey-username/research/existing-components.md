# Research: Existing BubbleTea Component Patterns

- **Query**: Search codebase for TUI component patterns, profile name resolution, and existing interactive components
- **Scope**: internal
- **Date**: 2026-05-18

## Findings

### Files Found

| File Path | Description |
|---|---|
| `tui/bubblon/controller.go` | Stack-based navigation controller, all TUI models pass through it |
| `tui/timeline/model.go` | Main timeline: list-based view with infinite scroll, pubkey→name async resolution |
| `tui/timeline/delegate.go` | Custom list delegate for timeline items (keyboard-only interactivity) |
| `tui/timeline/main.go` | Timeline launcher: wraps timeline model in bubblon.Controller |
| `tui/event/event.go` | Single event detail view (EventView), async profile name fetch |
| `tui/event/view.go` | Event header/content rendering - shows npub + @username as static text |
| `tui/event/styles.go` | lipgloss styles for event view (author, time, content, etc.) |
| `tui/thread/thread.go` | Thread tree view using treeview library, eventProvider.Name() for node labels |
| `tui/thread/thread_test.go` | Tests for thread model and eventProvider |
| `tui/compose/model.go` | Note/reply/quote composer with textarea/textinput |
| `tui/compose/main.go` | Compose launcher functions |
| `tui/community/discover/model.go` | Community discovery list view |
| `tui/dm/model.go` | Direct message view, async recipient profile name fetch |
| `tui/dm/list/model.go` | DM conversation list |
| `utils/get.go` | Core profile lookup: GetProfileName, GetProfileNameAsync, GetProfileNames, extractProfileName |

### Code Patterns

#### 1. BubbleTea Component Architecture

All TUI models follow the same pattern - implement `tea.Model` with three methods:

```go
// From tui/event/event.go:1-57, tui/timeline/model.go:91-120, tui/thread/thread.go:217-239
type MyModel struct {
    app    *config.AppContext
    styles styles            // lipgloss styles struct
    keys   keyMap            // bubbletea key bindings
    ctrl   *bubblon.Controller // optional: for stack navigation
    width  int
    height int
    // ... model-specific fields
}

func (m *MyModel) Init() tea.Cmd    { return someInitCmd() }
func (m *MyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { ... }
func (m *MyModel) View() tea.View     { return tea.NewView(...) }
```

**bubblon.Controller** (`tui/bubblon/controller.go:45-149`): A stack-based navigation controller where only the top model receives updates. Key commands:
- `bubblon.Open(model)` - push new model onto stack
- `bubblon.Close()` - pop current model, notify parent
- `bubblon.Replace(model)` - swap current model for new one

All top-level TUI apps wrap their model in `bubblon.New(model)` and pass to `tea.NewProgram()`.

#### 2. Message-Based Communication

Custom messages are used for async operations throughout:

```go
// From tui/timeline/model.go:170-194, tui/event/event.go:26-34
type fetchMsg struct { events []utils.TimelineEvent }
type namesMsg struct { names map[string]string }
type ProfileLoadedMsg struct { Name string }
type EventLoadedMsg struct { Event *nostr.Event }
```

Messages are produced by `tea.Cmd` closures (e.g., async profile fetches) and consumed in `Update()` via type switches.

#### 3. Keyboard Interactivity (NO Mouse Support)

All current interactivity is keyboard-only via `tea.KeyPressMsg`:

```go
// From tui/event/event.go:204-245, tui/thread/thread.go:637-687
case tea.KeyPressMsg:
    switch msg.String() {
    case "r":       return m.reply()
    case "enter":   return m.openDetail()
    case "esc":     return closeCmd()
    // ...
    }
```

**Critical finding**: `grep` for `tea.MouseMsg` / `mouseEvent` across all of `tui/` returned **zero results**. The project has no existing mouse click handling anywhere.

#### 4. Existing Components

All currently rendered UI is **static styled text** - no interactive inline components:

| Package | Components Used |
|---------|----------------|
| `timeline/` | `bubbles/v2/list` (DefaultDelegate), lipgloss styles |
| `event/` | `bubbles/v2/viewport`, `bubbles/v2/help`, lipgloss styles |
| `thread/` | `treeview/v2` (TuiTreeModel), lipgloss styles |
| `compose/` | `bubbles/v2/textarea`, `bubbles/v2/textinput`, `bubbles/v2/spinner`, lipgloss styles |
| `dm/` | `bubbles/v2/viewport`, `bubbles/v2/textinput`, lipgloss styles |
| `community/discover/` | `bubbles/v2/list` (DefaultDelegate), lipgloss styles |

**No label, button, tag, or other interactive inline components exist.**

#### 5. Profile Name Resolution Flow

The profile resolution pattern is consistent across all TUI views:

**Step 1 - Placeholder**: When first rendering events, use truncated npub as author name:
```go
// From tui/timeline/model.go:540-541
npubStr := nip19.EncodeNpub(e.Event.PubKey)
authorName := npubStr[:16] // placeholder truncated
```

**Step 2 - Async Fetch**: Dispatch background fetch via `tea.Cmd`:
```go
// From tui/timeline/model.go:286-306
func (m *model) fetchProfileNames(pubkeys []string) tea.Cmd {
    return func() tea.Msg {
        names := make(map[string]string)
        var wg sync.WaitGroup
        for _, pk := range pubkeys {
            wg.Add(1)
            go func(pubkeyStr string) {
                defer wg.Done()
                var pubKey nostr.PubKey
                if err := pubKey.UnmarshalJSON([]byte("\"" + pubkeyStr + "\"")); err == nil {
                    if name := utils.GetProfileName(context.Background(), pubKey, &utils.GetOptions{App: m.app}); name != "" {
                        names[pubkeyStr] = name
                    }
                }
            }(pk)
        }
        wg.Wait()
        return namesMsg{names: names}
    }
}
```

**Step 3 - Update UI**: On `namesMsg` receipt, update all list items:
```go
// From tui/timeline/model.go:573-584
case namesMsg:
    currentItems := m.list.Items()
    for i, listItem := range currentItems {
        if it, ok := listItem.(item); ok {
            pubkeyStr := it.event.Event.PubKey.Hex()
            if name, ok := msg.names[pubkeyStr]; ok {
                it.authorName = name
                currentItems[i] = it
            }
        }
    }
    m.list.SetItems(currentItems)
```

#### 6. Core Profile Resolution (`utils/get.go:198-464`)

The core resolution functions are:

| Function | Line | Signature | Behavior |
|----------|------|-----------|----------|
| `GetProfile` | 168 | `(ctx, pubKey, opts) *nostr.Event` | Queries relays for kind:0 profile event |
| `GetProfileName` | 198 | `(ctx, pubKey, opts) string` | Calls GetProfile → extractProfileName |
| `GetProfileNameAsync` | 357 | `(ctx, pubKey, opts) string` | Calls GetProfileAsync → extractProfileName |
| `GetProfiles` | 409 | `(ctx, pubKeys, opts) map[PubKey]*Event` | Batch profile fetch via FetchManyReplaceable |
| `GetProfileNames` | 439 | `(ctx, pubKeys, opts) map[PubKey]string` | Batch profile fetch → extract names |
| `extractProfileName` | 453 | `(profile *nostr.Event) string` | Parses kind:0 JSON, returns `pm.Name` or `""` |

`extractProfileName` (`utils/get.go:453-464`):
```go
func extractProfileName(profile *nostr.Event) string {
    if profile == nil { return "" }
    pm, err := sdk.ParseMetadata(*profile)
    if err == nil && pm.Name != "" { return pm.Name }
    return ""
}
```

#### 7. Thread's eventProvider.Name() (`tui/thread/thread.go:198-211`)

```go
func (p *eventProvider) Name(event nostr.Event) string {
    content := event.Content
    if len(content) > 50 { content = content[:47] + "..." }
    pubkey := event.PubKey.Hex()
    globalNameCacheMu.RLock()
    name, ok := globalNameCache[pubkey]
    globalNameCacheMu.RUnlock()
    if ok && name != "" {
        return strings.TrimSpace(content) + " (" + name + ")"
    }
    return strings.TrimSpace(content) + " (" + pubkey[:8] + ")"
}
```

Uses `globalNameCache` (a package-level `map[string]string` protected by `sync.RWMutex`) that is populated by `fetchProfileNames()` goroutines (`tui/thread/thread.go:574-612`).

#### 8. Event View Header Rendering (`tui/event/view.go:11-38`)

```go
func (m *EventView) renderHeader() string {
    // Line 1: full pubkey as npub
    npub := nip19.EncodeNpub(e.PubKey)
    line1 := fmt.Sprintf("PubKey: %s", m.styles.author.Render(npub))
    
    // Line 2: @username | time | kind
    var namePart string
    if m.authorName != "" {
        namePart = "@" + m.authorName
    } else {
        namePart = "@" + npub[:12] + "..."
    }
    line2 := fmt.Sprintf("%s | %s | %s",
        m.styles.author.Render(namePart), ...)
    
    return m.styles.header.Render(line1+"\n"+line2)
}
```

### Dependencies

| Package | Version | Usage |
|---------|---------|-------|
| `charm.land/bubbletea/v2` | v2.0.6 | Core TUI framework |
| `charm.land/bubbles/v2` | v2.1.0 | Component library (list, viewport, textarea, textinput, key, help, spinner) |
| `charm.land/lipgloss/v2` | v2.0.3 | Styling/rendering |
| `fiatjaf.com/nostr` | latest | Nostr protocol types |
| `github.com/Digital-Shane/treeview/v2` | - | Thread tree component |
| `fiatjaf.com/nostr/nip19` | - | npub encoding |

### Caveats / Not Found

- **No mouse click handling exists** anywhere in the TUI codebase. BubbleTea v2 supports `tea.MouseMsg` / `tea.MouseButton` but no code uses it.
- **No interactive inline components** (labels, tags, buttons, clickable text spans) exist. All rendering is static text with lipgloss styling.
- **No "profile card" or "pubkey badge" component** exists for displaying clickable pubkey/usernames.
- The `bubbles/v2` library does NOT include a built-in clickable label/tag component — one would need to be custom-built using BubbleTea's raw terminal rendering with mouse event capture.
- Profile name resolution always goes through relay queries (no local/cached profile DB aside from the in-memory event cache in `utils/get.go`).
