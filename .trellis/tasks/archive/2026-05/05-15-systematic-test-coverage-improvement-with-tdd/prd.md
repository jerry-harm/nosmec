# Systematic Test Coverage Improvement with TDD

## Goal

Improve test coverage systematically using TDD methodology. Write failing tests first, then minimal code to pass. The goal is to constrain behavior and make failures self-explanatory — not to hit coverage targets.

## What I already know

From `test-coverage-audit.md`:
- 5 test files exist: `utils/search_test.go`, `utils/post_test.go`, `utils/dm_test.go`, `config/config_test.go`, `config/relay_test.go`
- No goroutine leak detection (`TestMain` + `goleak` missing)
- No integration test build tags (`//go:build integration`)
- TUI files with array bounds risk: `timeline/model.go`, `compose/model.go`, `dm/model.go`, `thread.go`
- Nostr operations with nil check gaps: `publishRelayListMetadata`, `ParseCommunityAddr`, `ReplyToNote`, `GetParentPostInfo`, etc.

## Requirements

### Phase 1: Foundation
1. Add `TestMain` with `goleak.VerifyTestMain` to detect goroutine leaks
2. Add `//go:build integration` build tag to config/config_test.go (requires external services)

### Phase 2: Nostr Operation Tests (TDD)
Each function gets a test BEFORE fixing the bug.

| Function | Bug | Test Case |
|----------|-----|-----------|
| `publishRelayListMetadata` | nil pointer on `relay.Read`/`relay.Write` | Provide relay with nil Read/Write pointers |
| `ParseCommunityAddr` | no `parts[1]` length check | Provide malformed address |
| `ReplyToNote` | no nil check on parentEvent | Provide nil parentEvent |
| `GetParentPostInfo` | `authorPubKey[:]` no length validation | Provide invalid hex |
| `FetchRecipientReadRelays` | no nil check on `result.Event.Tags` | Provide empty event |
| `ListDMConversations` | `tag[1]` accessed without len check | Provide malformed gift wrap tag |
| `syncUsersFromNetwork` | `tag[1]`, `tag[2]`, `tag[3]` without len checks | Provide malformed user tag |

### Phase 3: TUI Bounds Tests (TDD)
| File | Risk | Test Case |
|------|------|-----------|
| `timeline/model.go` | `items[len-1]` without bounds | Empty items list, navigate prev |
| `compose/model.go` | `m.tags[m.editingTagIndex+1:]` index out of range | editingTagIndex at end of tags |
| `dm/model.go` | `m.messages` append with nil viewport | Empty messages |
| `thread.go` | `m.replies[i]` without bounds | Empty replies |

### Phase 4: Untested Utility Coverage
| File | Functions to cover |
|------|-------------------|
| `utils/show.go` | `PrintEvent` |
| `utils/alias.go` | `ResolveAliasToPubKey`, `PubKeyToNpub` |
| `utils/community.go` | `ParseCommunityAddr` (already in Phase 2) |
| `utils/profile.go` | `ProfileMetadataFromSDK`, `FetchRecipientReadRelays` (already in Phase 2) |

## TDD Process

For each item:
1. **RED**: Write failing test showing the bug
2. **Verify RED**: Run test, confirm it fails for expected reason
3. **GREEN**: Write minimal code to fix
4. **Verify GREEN**: Run test, confirm it passes
5. **REFACTOR**: Clean up if needed

## Definition of Done

- [ ] All functions in Phase 2 have failing tests before fix
- [ ] All functions in Phase 2 pass after minimal fix
- [ ] All TUI bounds tests written and passing
- [ ] `TestMain` with `goleak` is implemented
- [ ] `go test ./...` passes
- [ ] `go test -race ./...` passes (no race conditions)
- [ ] No new untested code in modified files

## Out of Scope

- Integration tests (require external services, separate task)
- `cmd/` package tests (CLI testing is complex, separate task)
- Fuzzing (separate task)

## Technical Notes

- Use `testify` assert/require — `require` for preconditions, `assert` for verifications
- Table-driven tests with named subtests
- `t.Parallel()` for independent tests