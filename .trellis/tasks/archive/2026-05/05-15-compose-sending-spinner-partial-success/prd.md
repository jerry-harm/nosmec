# compose sending spinner + partial success

## Goal

Improve compose sending UX with two changes:
1. Show spinner animation during publish instead of static "Sending..." overlay
2. Change success logic: if any relay succeeds (no error), consider it a success — don't require all relays to succeed

## What I already know

### Current sending flow (model.go:485-527)
- `sendContent()` creates event, calls `Pool().PublishMany()` on writable relays
- PublishMany returns a channel of results
- If ANY result has an error, returns `sendErrorMsg` immediately → entire publish fails
- `renderSendingOverlay()` shows static "Sending..." or statusMsg text

### Current partial success situation
```go
for result := range resultChan {
    if result.Error != nil {
        return sendErrorMsg{err: fmt.Errorf("failed to publish to %s: %w", result.RelayURL, result.Error).Error()}
    }
}
return sendSuccessMsg{eventID: event.ID.Hex()}
```

### Bubble Tea spinner pattern
charm.land/bubbles/v2 has a `spinner` bubble. Need to check if already used elsewhere.

## Assumptions

* spinner bubble is available in charm.land/bubbles/v2
* success message should show event ID even if some relays failed
* errors should still be collected and displayed, but not blocking

## Open Questions

1. ~~Should failed relays be tracked and shown to user after success?~~ → **Yes, show failed relays**
2. ~~Spinner should show during the entire publish phase, including the loop waiting for results?~~ → **Yes**
3. ~~Failure returns to compose?~~ → **No, always quit after send (success or all-fail)**
4. Spinner style: Dot (⠋⠙⠹...) vs Line (|/-\) — which?

## Requirements

1. **Sending state**: Show spinner animation (charm.land/bubbles/v2 spinner.Dot) instead of static overlay
2. **Partial success**: If `event.Sign` succeeds and at least one relay publish succeeds, return `sendSuccessMsg`
3. **Error collection**: Collect ALL relay errors, return them in error message (comma-separated)
4. **Always quit on send**: After `sendSuccessMsg` or `sendErrorMsg` (all relays failed), quit — don't return to compose
5. **Failed relays display**: Show which relays failed in error message (e.g., "Failed: relay1, relay2")

## Acceptance Criteria

- [ ] Spinner displays during publish (animated, not static text)
- [ ] If all relays fail → error message with ALL relay errors, then quit
- [ ] If some relays succeed → success message, then quit (show failed relays info)
- [ ] Tests pass for new partial-success logic
- [ ] No returning to compose state after send attempt

## Acceptance Criteria

- [ ] Spinner displays during publish (animated, not static text)
- [ ] If all relays fail → error message with all relay errors
- [ ] If some relays succeed → success message even if some failed
- [ ] Tests pass for new partial-success logic

## Definition of Done

* Spinner bubble integrated into compose UI during sending
* Partial success logic implemented and tested
* Lint/typecheck green

## Out of Scope

* Retry logic for failed relays
* UI showing which specific relays succeeded/failed post-send (error message is enough)

## Technical Notes

### Files: `tui/compose/model.go`

### Model changes needed:
```go
type model struct {
    // ... existing fields
    spinner     spinner.Model  // add spinner
}

// sendContent() changes:
- Collect ALL relay errors into []string
- Track bool for "at least one success"
- Return sendErrorMsg with formatted error list (comma-separated failed relays)
- OR return sendSuccessMsg if at least one relay succeeded

// sendErrorMsg handling:
- m.errMsg = msg.err
- m.statusMsg = "Error: " + msg.err
- After 3s delay: m.sending = false, then tea.Quit

// sendSuccessMsg handling:
- m.success = true
- m.statusMsg = "Posted successfully!"
- m.ClearDraft()
- tea.Quit immediately (no delay)
```

### Spinner API:
```go
sp := spinner.New()
sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#69"))
```

### Example usage pattern (from bubbletea docs):
```go
func (m model) Init() tea.Cmd {
    return m.spinner.Tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case spinner.Msg:
        newSpinner, cmd := m.spinner.Update(msg)
        m.spinner = newSpinner
        return m, cmd
    // ...
    }
    return m, nil
}
```