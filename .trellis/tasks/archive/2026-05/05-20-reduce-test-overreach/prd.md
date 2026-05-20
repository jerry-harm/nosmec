# reduce test overreach

## Goal

Reduce test overreach in the current codebase so the test suite is better scoped, faster to run, and less coupled to implementation details or incidental network behavior, without losing coverage for behavior that actually matters.

## What I already know

* The user wants to address a current problem described as "测试过度".
* This is implementation work, so a Trellis task was created at `.trellis/tasks/05-20-reduce-test-overreach`.
* The repository has a broad test surface, with many tests concentrated under `nostr_sdk`, `utils`, and `tui`.
* Recent work added new focused tests under `nostr_sdk/fetch_events_test.go`, `nostr_sdk/community_thread_test.go`, `nip72/nip72_test.go`, and `utils/community_test.go`.
* There is at least one explicit skipped integration-style test in `utils/community_test.go` (`TestGetParentPostInfo_RequiresFullIntegration`) labeled as complex integration coverage.
* Earlier verification showed that running some whole-package test combinations can hit long runtimes/timeouts, especially around `nostr_sdk`.
* The current codebase includes tests around relay/network-oriented logic and fetch behavior, which makes it plausible that some tests are too integration-heavy for their package level.
* Running `go test -json ./nostr_sdk` confirmed that `nostr_sdk` is the current hot spot for very slow tests.
* The strongest slow-test signal is `TestLoadWoTManyPeople`, which caused the package run to fail at `600.233s` with many goroutines blocked in `fiatjaf.com/nostr.NewRelay` / websocket timeout loops.
* Another clearly slow test is `TestLoadWoT` at about `42.13s` before failure.
* Additional non-trivial slow tests observed in the same package run include:
  * `TestConcurrentMetadata` at about `21.36s`
  * `TestMetadataAndEvents` at about `14.22s`
  * `TestFetchSpecificEventInScope_NilWhenScopeMismatches` at about `14s`
  * `TestFetchRepliesBreadthFirstInScope_UsesLocalStoreAndFiltersScope` at about `14s`
  * `TestPrepareNoteEvent` sub-suite at about `7.12s`
  * `TestFollowListRecursion` at about `7.12s`
  * `TestFetchEventsByFilter_UsesFallbackRelaysWhenNoOverrideProvided` at about `7s`
* `TestStreamLiveFeed` also failed quickly at about `5.24s`, suggesting another network/live dependency, though not the worst runtime offender.

## Assumptions (temporary)

* "Test overreach" may mean one or more of: too many low-value assertions, tests coupled to internals, package tests doing integration/network work, duplicated coverage across layers, or tests that make refactors expensive.
* The likely first target area is `nostr_sdk`, because it has both the densest recent additions and prior timeout symptoms.

## Open Questions

* Which kind of overreach do we want to optimize first?

## Requirements (evolving)

* Identify the concrete forms of test overreach present in the repo.
* Choose a narrow first slice to fix in this task.
* Preserve meaningful behavioral coverage while reducing test cost/coupling.
* First identify the very slow tests with measured evidence before changing them.
* Treat old `nostr_sdk` tests as out of default verification scope when we are not modifying that code area.
* Stop running WoT tests as part of normal verification.

## Acceptance Criteria (evolving)

* [ ] The targeted overreach pattern is explicitly defined.
* [ ] The task scope names the package(s) or test file(s) to change.
* [ ] The resulting test strategy is narrower and easier to maintain than before.
* [ ] Default verification no longer runs WoT tests.
* [ ] Default verification no longer runs unrelated legacy `nostr_sdk` tests when that package is outside the current code-change scope.

## Technical Approach

Update the verification guidance rather than changing test implementation first.

1. Narrow default verification guidance so it is scope-driven instead of package-wide.
2. For `nostr_sdk`, document that full-package `go test ./nostr_sdk` is not a default verification command.
3. Require targeted `-run` selections for `nostr_sdk` only when the current change actually touches that package.
4. Explicitly exclude WoT tests from default verification guidance.
5. Record the rule both in this task PRD and in long-lived backend spec documentation so future tasks inherit it.

## Decision (ADR-lite)

**Context**: `nostr_sdk` contains legacy tests that are slow and appear to rely on live network or relay behavior. The user does not expect to modify that code often, so paying this verification cost on unrelated tasks is wasteful.

**Decision**: Exclude WoT tests from routine verification, and avoid running legacy `nostr_sdk` tests by default when current work does not modify that area.

**Consequences**: We reduce routine test cost and avoid unrelated failures/timeouts, but we also accept that legacy `nostr_sdk` coverage is no longer part of every normal verification pass. If `nostr_sdk` is modified in the future, that area should use a more targeted test strategy.

## Definition of Done (team quality bar)

* Tests added/updated (unit/integration where appropriate)
* Lint / typecheck / CI green
* Docs/notes updated if behavior changes
* Rollout/rollback considered if risky

## Out of Scope (explicit)

* Broad repository-wide test redesign unless we explicitly expand scope.
* Changing production behavior unrelated to test structure.
* Making WoT tests fast enough to keep in the default suite.
* Preserving full-package `nostr_sdk` verification for unrelated tasks.

## Technical Notes

* Candidate areas from quick inspection:
  * `nostr_sdk/fetch_events_test.go`
  * `nostr_sdk/community_thread_test.go`
  * `nostr_sdk/community_scope_test.go`
  * `nostr_sdk/thread_refs_test.go`
  * `utils/community_test.go`
  * `tui/thread/thread_test.go`
* Prior session context already noted a whole-package `go test ./nostr_sdk` runtime problem, though focused test subsets were green.
* Evidence source: `go test -json ./nostr_sdk`.
* Current likely category: package tests that depend on live network / relay behavior rather than deterministic local fixtures.
* Requested delivery: update verification-command documentation, not test code behavior.
