# migrate-sdk-profilemetadata

## Goal

Migrate from custom `utils.ProfileMetadata` to `sdk.ProfileMetadata` from `fiatjaf.com/nostr/sdk`.

## What I Already Know

**Current**: `utils.ProfileMetadata` — manual JSON unmarshal, no NIP05 validation, no Npub/ShortName helpers
**SDK**: `sdk.ProfileMetadata` — has `NIP05Valid()`, `Npub()`, `ShortName()`, `ParseMetadata()` helper

## Requirements

* Replace `utils.ProfileMetadata` usage with `sdk.ProfileMetadata`
* Use `sdk.ParseMetadata(event)` instead of manual `json.Unmarshal`
* Add `nip05.Valid()` call where needed
* Remove duplicate ProfileMetadata struct from `utils/profile.go`

## Acceptance Criteria

* [ ] All code using `utils.ProfileMetadata` migrates to `sdk.ProfileMetadata`
* [ ] `go build ./...` succeeds
* [ ] `go vet ./...` passes

## Definition of Done

* Build passes, no more duplicate ProfileMetadata