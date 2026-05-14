# batch-profile-fetch

## Goal

Replace `GetProfile`/`GetProfileName` with list-based APIs that fetch multiple profiles in a single pool query.

## What I Already Know

* Current `GetProfile` takes single `pubKey` and returns single `*nostr.Event`
* Current `GetProfileAsync` same but runs in goroutine
* Current `GetProfileName`/`GetProfileNameAsync` extract name from profile
* Pool has `FetchMany` which returns `chan RelayEvent` for multiple events

## Requirements

* Add `GetProfiles(ctx, pubKeys []nostr.PubKey, opts) map[nostr.PubKey]*nostr.Event`
* Add `GetProfileNames(ctx, pubKeys []nostr.PubKey, opts) map[nostr.PubKey]string`
* Use `FetchManyReplaceable` or single `QuerySingle` with multiple Authors
* Preserve existing single-key functions for backward compatibility

## Acceptance Criteria

* [ ] `GetProfiles` returns map with all found profiles
* [ ] `GetProfileNames` returns map with all found names
* [ ] `go build ./...` succeeds
* [ ] `go vet ./...` passes

## Definition of Done

* Batch APIs exist and build passes