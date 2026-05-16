# Unify nip19.Decode usage - remove ParsePubKey wrapper

## Goal

Remove the `utils.ParsePubKey` wrapper and use `nip19.Decode` directly at all call sites. The wrapper was a thin convenience layer that added little value and now blocks us from using the full nip19 API.

## What I already know

* Research confirms 3 call sites of `ParsePubKey`:
  * `utils/search.go:44` - `ParsePubKey(authorStr)` in `ParseSearchFilter`
  * `cmd/dm_commands.go:40,85,103` - DM recipient parsing
  * `tui/dm/main.go:13` - DM UI pubkey lookup
* 4 direct `nip19.Decode` calls already exist in the codebase
* `ResolveAliasToPubKey` in `alias.go` re-implements the same 3-step parse sequence as `ParsePubKey`
* The 66-char compressed key hack (`02`/`03` prefix) exists in both `utils.go` and `alias.go` but research shows **no caller passes this format** - it's dead code

## Requirements

* Replace all `utils.ParsePubKey(...)` calls with direct `nip19.Decode` usage
* Keep `alias.go` `ResolveAliasToPubKey` separate (it handles alias resolution, not just pubkey parsing)
* Remove `ParsePubKey` from `utils/utils.go`
* The 66-char compressed key hack is **not needed** - remove it entirely (no real caller passes 66-char format)

## Technical Approach

For each call site, replace:
```go
pk, err := utils.ParsePubKey(s)
```
with:
```go
_, decoded, err := nip19.Decode(s)
if err != nil { return ..., err }
pk, ok := decoded.(nostr.PubKey)
if !ok { return ..., fmt.Errorf("not a pubkey") }
// use pk
```

**Note**: The `alias.go` has its own duplicate compressed key hack in `ResolveAliasToPubKey`. That function should be kept as-is for now since it's about alias resolution, not pure pubkey parsing - but we should verify if the compressed key hack there is similarly dead code.

## Implementation Plan

1. Update `utils/search.go` - replace ParsePubKey call
2. Update `cmd/dm_commands.go` - replace 3 ParsePubKey calls
3. Update `tui/dm/main.go` - replace ParsePubKey call
4. Remove `ParsePubKey` from `utils/utils.go`

## Acceptance Criteria

* [ ] `go build ./...` passes
* [ ] `go vet ./...` passes
* [ ] `./nosmec config show` works (uses DM commands)

## Out of Scope

* Modifying `alias.go` `ResolveAliasToPubKey` - it has its own logic for alias resolution