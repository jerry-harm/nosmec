# fix GetNote ID parsing and add npub/nsec/nevet format support

## Goal

Fix `GetNote`/`GetNoteAsync` bug where `nostr.ID` is incorrectly built from hex string. Also ensure all user-facing outputs (logs, config displays, CLI output) use npub/nsec/nevet format instead of raw hex for pubkeys, seckeys, and event IDs.

## What I already know

### Bug: GetNote/GetNoteAsync ID parsing is wrong

Location: `utils/get.go:136-140` and `150-154`

Current code:
```go
var id nostr.ID
if len(noteID) != 64 {
    return nil
}
copy(id[:], noteID)  // WRONG: copies 64 ASCII chars into 32 bytes
```

`nostr.ID` is `[32]byte` but `noteID` is a 64-character hex string. `copy(id[:], noteID)` copies ASCII bytes, not the decoded hex.

Correct way:
```go
id, err := nostr.IDFromHex(noteID)
if err != nil {
    return nil
}
```

### NIP-19 format requirement

All user-facing outputs should use:
- `npub1...` for pubkeys (Kind 0 metadata)
- `nsec1...` for private keys  
- `nevent1...` for event IDs (not just note ID - should include relay hints and author)

## Requirements

1. **Fix GetNote/GetNoteAsync**: Use `nostr.IDFromHex` instead of `copy`
2. **Fix GetProfile/GetProfileAsync**: Similar issue for pubkey parsing - check if same bug exists
3. **Add npub/nsec/nevet encoding utility**: Create helper to convert hex to NIP-19 format
4. **Update all user-facing outputs**: 
   - Logs that show event IDs or pubkeys
   - Config display commands
   - Any `fmt.Printf` or logger output
5. **Update config file format**: When storing/displaying relay list or other configs, use NIP-19 where applicable

## Acceptance Criteria

- [ ] `GetNote` and `GetNoteAsync` correctly parse 64-char hex event ID
- [ ] All CLI output for pubkeys uses `npub1...` format
- [ ] All CLI output for event IDs uses `nevent1...` or `note1...` format
- [ ] Private keys never logged or displayed (only in config file)
- [ ] `nosmec event <id>` correctly fetches and displays the event

## Definition of Done

* `GetNote("<64-char-hex>")` returns the correct event
* All user-facing hex pubkeys replaced with npub format
* Lint/typecheck passes

## Out of Scope

* Changing internal storage format (keep hex in DB)
* Supporting NIP-57 (Zaps) or other NIPs

## Technical Notes

* `nostr.IDFromHex(hex string)` returns `(ID, error)`
* `nostr.PubKeyFromHex(hex string)` returns `(PubKey, error)` for profile lookups
* `nostr.Encodeeken()` or similar to encode to npub format — need to check SDK for NIP-19 encoding
* Check `fiatjaf.com/nostr/nip19` package for encoding functions