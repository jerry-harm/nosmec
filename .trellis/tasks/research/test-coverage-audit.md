# Test Coverage Audit Report

- **Query**: Audit test suite for coverage gaps, integration tests, goroutine leaks
- **Scope**: internal
- **Date**: 2026-05-15

## Findings

### 1. Test Files Found

| File Path | Description |
|---|---|
| `utils/search_test.go` | Tests for NIP-50 search filter parsing |
| `utils/post_test.go` | Tests for PostNote, ReplyToNote, QuoteNote, DeleteNote |
| `utils/dm_test.go` | Tests for SendDM, ListDMConversations, QueryDMHistory |
| `config/config_test.go` | Tests for config loading |
| `config/relay_test.go` | Tests for relay configuration |

**Total: 5 test files**

### 2. Goroutine Leak Detection

**Status: NOT FOUND**

No `TestMain` function exists in the codebase. No `goleak` usage detected. This means:
- Tests may leak goroutines without detection
- Background subscriptions in `utils/get.go`, `utils/dm.go`, `tui/timeline/model.go`, `tui/dm/model.go` could leak on test exit

### 3. Integration Test Build Tags

**Status: NOT FOUND**

No `//go:build integration` tags found. No integration test infrastructure exists.

### 4. Unit Test Coverage Gaps

#### Critical Gaps — TUI Components (Array Bounds Risk)

| File | Risk |
|---|---|
| `tui/timeline/model.go` | List slicing at lines 326, 534, 654, 688 — no bounds check before `items[len(items)-1]` or `currentItems[len(currentItems)-1]` |
| `tui/compose/model.go` | Tag slice mutations at lines 194, 204, 357, 372, 377 — `append(m.tags[:m.editingTagIndex], m.tags[m.editingTagIndex+1:]...)` with unchecked indices |
| `tui/dm/model.go` | Message slice at line 308 — `append(m.messages, ...)` with potential nil viewport |
| `tui/window/event/thread.go` | Reply indexing at line 130 — `m.replies[i]` with no bounds check |

#### Critical Gaps — Nostr Operations (Validity Checks)

| File | Function | Gap |
|---|---|---|
| `utils/post.go` | `ReplyToNote` | No nil check on `parentEvent` before accessing `parentEvent.PubKey` at line 57 |
| `utils/get.go` | `GetParentEvent` | No nil check on `event.Tags` access pattern |
| `utils/get.go` | `GetFollowedTimeline` | `hashtags` parameter mutated (line 545) — append to function argument |
| `utils/relay_list.go` | `publishRelayListMetadata` | No nil check on `relay.Read`, `relay.Write` pointer dereference at lines 152-153 |
| `utils/community.go` | `ParseCommunityAddr` | No nil checks on `parts[1]` before `nostr.PubKeyFromHex` at line 92 |
| `utils/community.go` | `GetParentPostInfo` | `authorPubKey[:]` copy at line 309 — no length validation |
| `utils/search.go` | `ParseSearchFilter` | `kindsRegex`, `authorsRegex`, `tagRegex` compiled per call — no cached compiled regex |
| `utils/profile.go` | `FetchRecipientReadRelays` | No nil check on `result.Event.Tags` before iteration |
| `utils/dm.go` | `ListDMConversations` | `tag[1]` accessed at line 137 without checking `len(tag) >= 2` |
| `utils/subscription.go` | `syncUsersFromNetwork` | `tag[1]`, `tag[2]`, `tag[3]` accessed without length checks at lines 167-175 |
| `utils/user_relays.go` | `VerifyRelaysConnectivity` | Context timeout created but may not apply to individual relay checks |

#### Untested Utility Functions

| File | Functions |
|---|---|
| `utils/show.go` | `PrintEvent` (entire file) |
| `utils/alias.go` | Entire file (not reviewed) |
| `utils/sync.go` | Entire file (not reviewed) |
| `utils/proxy.go` | Entire file (not reviewed) |
| `config/types.go` | `BoolPtr`, pointer helpers (not reviewed) |
| `config/context.go` | Entire file (not reviewed) |
| `cmd/` | All command files (not reviewed) |

### 5. Summary of Risks

| Category | Count |
|---|---|
| TUI files with array/slice access risk | 4 |
| Nostr operation functions without nil checks | 10+ |
| Entire utility files untested | 5+ |
| Goroutine leak detection | MISSING |
| Integration test infrastructure | MISSING |

## Caveats / Not Found

- `utils/alias.go`, `utils/sync.go`, `utils/proxy.go` exist but were not read — may contain additional gaps
- `config/types.go` and `config/context.go` exist but were not read
- All `cmd/` package files exist but were not reviewed
- No test helper files found (no `testutil/` or `testing/` utilities)