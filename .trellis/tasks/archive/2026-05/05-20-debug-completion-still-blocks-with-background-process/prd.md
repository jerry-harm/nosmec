# debug completion still blocks with background process

## Goal

Fix shell completion so it does not block when another nosmec process is running in the background. Completion should avoid unnecessary database and runtime initialization when it only needs config-backed data.

## What I already know

* Completion handlers in `cmd/completion/completion.go` only read config-backed app data such as relays, aliases, and subscriptions.
* `cmd/root.go` registers `cobra.OnInitialize(initApp)`.
* `initApp()` always does `config.InitConfig()`, `config.GlobalPool()`, and `config.NewAppContext(...)`.
* `config.GlobalPool()` initializes persistent backends (`HintsDB`, `KVStore`, event store), so completion currently opens LMDB even when it does not need it.
* That means completion can still block behind another running process even after the storage backend switch.

## Assumptions (temporary)

* The blocking is caused by unconditional runtime initialization on CLI startup, not by the completion functions themselves.
* A minimal completion-safe app context can likely be built from config without initializing the pool/system stores.

## Open Questions

* None — confirmed approach: lazy pool initialization.

## Requirements (evolving)

* Shell completion must not initialize `GlobalPool()` or open persistent databases.
* Existing completion results for aliases, relays, subscriptions, and config keys must still work.
* Normal commands that need pool/store/network still get them (lazy open on first access).

## Decision (ADR-lite)

**Context**: `initApp()` unconditionally calls `GlobalPool()` which opens all LMDB backends synchronously. Even though completion functions don't need the pool, they trigger this initialization because `cobra.OnInitialize` fires before every command including completion.

**Decision**: Make `GlobalPool()` lazy — `InitConfig()` and `NewAppContext()` still run (config-only), but `GlobalPool()` opens LMDB files only when first accessed (first real pool operation), not during `initApp()`.

**Consequences**: Completion runs without touching databases. Normal commands transparently trigger lazy open on first network/store operation.

## Acceptance Criteria (evolving)

* [ ] Completion path avoids `GlobalPool()` / persistent DB initialization.
* [ ] Completion functions still return existing suggestions.
* [ ] Build/tests covering the affected command startup/completion path pass.

## Definition of Done (team quality bar)

* Tests added/updated (unit/integration where appropriate)
* Lint / typecheck / CI green
* Docs/notes updated if behavior changes
* Rollout/rollback considered if risky

## Out of Scope (explicit)

* Broad runtime architecture changes beyond what is needed to unblock completion
* Storage backend changes (already handled by prior task)

## Technical Notes

### Root Cause
`cmd/root.go:initApp()` → `config.GlobalPool()` → `nostr_sdk.NewSystem()` → opens LMDB HintsDB + KVStore + event store synchronously. Shell completion triggers `cobra.OnInitialize`, which runs `initApp()` before completion logic executes. So completion opens LMDB files even though the completion functions themselves don't use them. If a background process holds an exclusive lock (possible during certain write patterns or env open), completion blocks.

### Fix: Lazy pool and system initialization
- `initApp()` keeps `config.InitConfig()` + `config.NewAppContext()` (config-only)
- Remove `config.GlobalPool()` call from `initApp()`
- `GlobalPool()` becomes a lazy getter: pool and LMDB stores open on first actual use (first relay network call)
- `GlobalSystem` fields (Hints, KVStore, Store) similarly lazy
- Completion runs without touching any persistent resources
- All existing commands transparently trigger lazy open on first operation that needs the pool

### Implementation approach
1. `config/config.go`: change `GlobalPool()` to lazy initialization, or use a separate lightweight path for `initApp()` that skips pool creation
2. `config/context.go` `NewAppContext()`: accept nil pool gracefully (completion case)
3. All callers of `GlobalPool()` / `app.Pool()` that need real pool must tolerate nil and trigger lazy init
4. Completion functions already handle `app == nil` safely — keep this pattern

### Files to change
- `config/config.go` — lazy pool/system init
- `config/context.go` — NewAppContext with nil pool
- `cmd/root.go` — remove GlobalPool call from initApp
