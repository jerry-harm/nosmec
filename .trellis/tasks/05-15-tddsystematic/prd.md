# Systematic TDD Implementation Plan

## Goal

Achieve comprehensive test coverage across nosmec layers using appropriate testing strategies per layer.

## What We Know

### Testing Strategy by Layer

| Layer | Strategy | Tools |
|-------|----------|-------|
| **utils (pure logic)** | Table-driven unit tests, no mocks | Standard `testing`, `assert`/`require` |
| **TUI** | Golden file output tests | `charmbracelet/x/exp/teatest` |
| **Network/relay** | Integration tests with local relay | nostr SDK's built-in test relay |
| **Config/persistence** | With real DB | `*bolt.DB` in temp dir |

### Teatest for Bubble Tea TUI

`charmbracelet/x/exp/teatest` provides:
- `NewTestModel(t, model)` — wraps tea.Model for testing
- `tm.Output()` — real-time output stream (check intermediate state)
- `tm.Send(msg)` — send messages to model
- `tm.FinalModel(t)` — get final model after program exits
- `tm.FinalOutput(t)` — get final output (golden file comparison)
- `-update` flag — auto-regenerate golden files
- `lipgloss.SetColorProfile(termenv.Ascii)` — CI-safe, no colors

```go
func init() {
    lipgloss.SetColorProfile(termenv.Ascii)
}

func TestSomething(t *testing.T) {
    tm := teatest.NewTestModel(t, initialModel(), teatest.WithInitialTermSize(80, 24))
    defer tm.Stop()

    // Check intermediate output
    teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
        return bytes.Contains(b, []byte("expected text"))
    }, teatest.WithCheckInterval(50*time.Millisecond))

    // Send input
    tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

    // Wait for finish
    tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
```

### Local Relay for Network Tests

nostr SDK includes `nostrserve` or a simple relay implementation for testing. Alternatively:
- Use a real public relay in test mode (filter by test pubkey)
- Implement a minimal in-memory relay for unit tests
- Mock `Pool` interface for pure unit tests

### Test Database

Use `os.MkdirTemp` + `bolt.Open` in temp directory, clean up after tests.

## Out of Scope

- Code coverage % targets (coverage is outcome, not goal)
- Mocking production code that should be refactored to be testable

## Implementation Plan

### Phase 1: Pure Utils (no network, no UI)
1. Identify pure functions in `utils/` that aren't tested
2. Add table-driven tests for filter builders, parsers, etc.
3. Ensure `go test ./utils/...` passes with good coverage

### Phase 2: TUI with Teatest
1. Add `charmbracelet/x/exp/teatest` dependency
2. Write a simple TUI test as proof-of-concept (e.g., timeline list renders)
3. Set up golden files, CI-safe color profile
4. Expand TUI test coverage incrementally

### Phase 3: Integration Tests
1. Set up test relay (nostr SDK's or minimal implementation)
2. Test relay list discovery, event publishing
3. Test GetEvent, GetProfile with real relay

### Phase 4: Database/Persistence Tests
1. Use temp bolt DB for config tests
2. Test relay list persistence, preference loading

## Definition of Done

- [ ] Utils layer: all pure functions have table-driven tests
- [ ] TUI: at least one working teatest example passing
- [ ] Golden files setup with CI-safe color profile
- [ ] Test database pattern documented and used
- [ ] `go test ./...` passes in CI (no color-dependent failures)