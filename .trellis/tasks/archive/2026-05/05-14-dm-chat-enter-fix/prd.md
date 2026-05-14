# dm-chat-enter-fix

## Goal

Fix the DM chat input so that pressing Enter sends the message, instead of being captured as a newline by the textarea component.

## What I Already Know

* **Problem**: DM chat uses `textarea.Model` from `charm.land/bubbles/v2/textarea` (tui/dm/model.go:37, 98-100). The textarea intercepts Enter as a newline within its content, preventing the send key binding from firing reliably.
* **Fix**: Replace `textarea.Model` with `textinput.Model` from `charm.land/bubbles/v2/textinput`, which passes Enter through to the key handler without consuming it.
* **Precedent**: tui/compose/model.go already uses both textarea and textinput — textinput is used for single-line inputs.
* **Key binding**: send key is already bound to Enter at model.go:53-56 — it just doesn't fire because textarea consumes it first.

## Requirements

* Enter key triggers message send (not newline in textarea)
* Message text is sent via `m.sendDM(content)` on Enter press
* Text input uses `textinput.Model` (single-line, Enter passes through)
* Placeholder "Type a message..." is preserved
* Focus behavior preserved

## Acceptance Criteria

* [ ] Pressing Enter in DM chat input sends the message (not inserting newline)
* [ ] Input field remains single-line (no multi-line support needed for MVP)
* [ ] Placeholder text "Type a message..." displayed when empty
* [ ] Empty messages are not sent (trim + empty check preserved)
* [ ] `go build` succeeds
* [ ] `go vet` passes

## Definition of Done

* `textarea.Model` replaced with `textinput.Model` in tui/dm/model.go
* Send key binding (`enter`) triggers `m.sendDM()` when textinput has focus
* Build and vet pass

## Out of Scope

* Multi-line message support (Enter always sends)
* Other DM UI changes
* Backend DM logic changes

## Technical Notes

* File: `tui/dm/model.go`
* Replacement component: `charm.land/bubbles/v2/textinput`
* Import path: `charm.land/bubbles/v2/textinput`
* Model type: `textinput.Model`
* No multi-line support in textinput — Enter always passes through to key handler
* Placeholder set via `m.ta.Placeholder = "Type a message..."`
* Focus set via `m.ta.Focus()`
* Value retrieved via `m.ta.Value()` (same as textarea)
* Value set via `m.ta.SetValue("")` (same as textarea)