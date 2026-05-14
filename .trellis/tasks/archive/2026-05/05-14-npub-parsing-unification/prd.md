# npub-parsing-unification

## Goal

Fix `ParsePubKey` to handle all input formats: bech32 `npub1...`, raw 64-char hex, and 66-char compressed pubkey hex (starting with `02`/`03`).

## What I Already Know

* **Error**: `x coordinate 36636636... is not on the secp256k1 curve` — hex string being passed as raw hex but it's actually a compressed pubkey (66 chars starting with `02`)
* **Current `ParsePubKey`** only handles `npub` bech32 and raw 64-char hex
* **Missing**: Compressed pubkey hex (66 chars starting with `02`/`03`) from NIP-19 `encodePack` output
* **Usage sites**: DM main.go, dm_commands.go — all CLI entrypoints that accept npub or hex

## Requirements

* `ParsePubKey` handles all three formats:
  1. `npub1...` bech32 (already works)
  2. 64-char hex (already works)
  3. 66-char compressed pubkey hex starting with `02` or `03` (new)
* Preserve existing behavior for valid inputs
* Return meaningful error for invalid inputs

## Acceptance Criteria

* [ ] `ParsePubKey("npub1...")` works
* [ ] `ParsePubKey("64charhex...")` works
* [ ] `ParsePubKey("02...")` 66-char compressed pubkey works
* [ ] `go build ./...` succeeds
* [ ] `go vet ./...` passes

## Definition of Done

* ParsePubKey handles all three formats
* Build and vet pass

## Out of Scope

* Changes to encoding/output format
* Changes to how pubkeys are stored internally