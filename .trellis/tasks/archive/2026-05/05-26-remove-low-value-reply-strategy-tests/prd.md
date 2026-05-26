# remove low-value reply strategy tests

## Goal

Remove two brittle low-value tests in `utils/reply_strategy_test.go` that assert encoding/parsing details not worth maintaining here.

## What I already know

* `TestBuildAddressableTag` is failing and depends on low-level `nostr.Event` field encoding details.
* `TestExtractPubKeyFromATag` is failing and exercises a brittle helper-level parsing path.
* User explicitly asked to delete these tests instead of preserving them.

## Requirements

* Delete `TestBuildAddressableTag`.
* Delete `TestExtractPubKeyFromATag`.
* Leave the rest of `utils/reply_strategy_test.go` intact.

## Acceptance Criteria

* [ ] The two named tests are removed.
* [ ] Targeted `utils` tests still run successfully for the remaining reply strategy coverage.

## Out of Scope

* Fixing `buildAddressableTag`.
* Fixing `extractPubKeyFromATag`.
* Refactoring reply strategy behavior.

## Technical Notes

* File in scope: `utils/reply_strategy_test.go`
