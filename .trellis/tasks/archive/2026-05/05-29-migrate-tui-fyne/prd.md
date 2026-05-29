# Migrate TUI to Fyne

## Goal

Replace the project's TUI direction with a desktop GUI built on `fyne.io/fyne/v2@latest`, and treat existing `.trellis/archive` history as disposable during the migration.

## What I already know

* The user wants a complete restructuring of the project content.
* The old TUI approach should be fully abandoned.
* The new UI stack target is `fyne.io/fyne/v2@latest`.
* `.trellis/archive` records do not need to be preserved.
* The current `gui` branch appears to contain only Trellis and agent metadata; no Go source files or `go.mod` are present yet.
* Local branches `main`, `gui`, and `old` exist.
* The user wants to migrate code from `main` into `gui` as the starting point.
* The user does not want any TUI content migrated.
* The user does not want old `.trellis` archive content migrated.
* The user delegated the exact migration mechanics as long as the resulting structure matches the goal.
* The user wants the existing code migrated first.
* The initial GUI layout direction is a left sidebar plus a top bar.
* Detailed page/UI behavior can be decided after the migration baseline is in place.
* The user only wants to keep CLI pieces that are unrelated to TUI.
* The likely long-term CLI survivors are testing helpers, environment-variable/configuration support, and similar non-UI utilities.
* The exact long-term retained CLI surface can be refined later.
* For the first migration pass, the user wants the minimal CLI retention strategy.
* GUI should be the default startup path.
* Minimal CLI helpers should remain available only as auxiliary entrypoints.
* The first migration pass should bring over only code required for the GUI shell plus config/env/test support.

## Assumptions (temporary)

* We are defining the migration before writing implementation code.
* The legacy code to migrate lives on the `main` branch.
* The migration should selectively carry over reusable Go application layers from `main`, while excluding `tui/` and legacy `.trellis/tasks/archive/` content.
* The first migration slice should establish a runnable Fyne shell with left sidebar and top bar layout, after the reusable code is brought over.

## Open Questions

* For the first runnable milestone, the GUI shell should use mostly placeholder pages around real app bootstrap/config wiring, not real data-backed screens.

## Requirements (evolving)

* Establish the project around Fyne instead of TUI.
* Migrate code from `main` into `gui` as the basis for the rewrite.
* Exclude `tui/` code from the migration.
* Exclude old `.trellis/tasks/archive/` content from the migration.
* Preserve reusable non-UI packages from `main`, including CLI/core app layers that can be repurposed behind the Fyne UI.
* Establish an initial Fyne shell using a left sidebar and top bar layout.
* Retain only non-TUI CLI entrypoints during migration, primarily for testing, configuration, and environment handling.
* Use the minimal CLI retention strategy in the first migration pass.
* Make the Fyne GUI the default startup path for the migrated application.
* Keep minimal CLI helpers as auxiliary entrypoints instead of the primary app mode.
* In the first migration pass, migrate only the code required for the GUI shell plus config/env/test support.
* Allow disposal of `.trellis/archive` records as part of the migration.

## Acceptance Criteria (evolving)

* [ ] The migration scope is defined clearly enough to split into implementation phases.
* [ ] We know the migration starts from a selective import of reusable layers from `main`.
* [ ] We know the migration keeps only non-TUI CLI entrypoints during the transition.
* [ ] We know GUI is the default startup mode and minimal CLI helpers remain auxiliary.
* [ ] We know whether the first runnable GUI milestone can use placeholder pages (yes, it should).

## Definition of Done (team quality bar)

* Tests added/updated (unit/integration where appropriate)
* Lint / typecheck / CI green
* Docs/notes updated if behavior changes
* Rollout/rollback considered if risky

## Out of Scope (explicit)

* Preserving `.trellis/archive` history.
* Migrating `tui/` source files into `gui`.
* Implementing feature parity before the migration scope is agreed.
* Finalizing detailed Fyne page interactions before the migration baseline exists.

## Technical Notes

* Repository inspection on `gui` found only: `.claude/`, `.cursor/`, `.git/`, `.opencode/`, `.trellis/`, `.gitignore`, `AGENTS.md`.
* `git branch --list` shows local branches: `gui`, `main`, `old`.
* `**/*.go` search on current `gui` working tree returned no files.
* Search for a Go module declaration on current `gui` working tree returned no results.
* `main` contains the Go project including `go.mod`, `main.go`, `cmd/`, `config/`, `logger/`, `nip72/`, `nostr_sdk/`, `utils/`, `docs/`, and `tui/`.
* `main:go.mod` shows current app dependencies include Cobra/Viper plus Bubble Tea / Bubbles / Lip Gloss; those TUI dependencies should be removed or replaced during the GUI migration.
* `main:main.go` enters through `cmd.Execute()`, and `main:cmd/root.go` builds a Cobra CLI around a shared `config.AppContext`; that app context is likely reusable behind a Fyne UI shell.
* TUI-bound CLI commands are in `community_commands.go`, `dm_commands.go`, `event_commands.go`, and `note_commands.go`.
* Non-TUI CLI candidates still present in `main` include `config`, `gossip`, `profile`, `relay`, `search`, and root initialization logic, but the first migration pass should keep only the minimum subset needed for config/env/test workflows.
