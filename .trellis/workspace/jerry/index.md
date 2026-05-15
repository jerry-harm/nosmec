# Workspace Index - jerry

> Journal tracking for AI development sessions.

---

## Current Status

<!-- @@@auto:current-status -->
- **Active File**: `journal-1.md`
- **Total Sessions**: 28
- **Last Active**: 2026-05-15
<!-- @@@/auto:current-status -->

---

## Active Documents

<!-- @@@auto:active-documents -->
| File | Lines | Status |
|------|-------|--------|
| `journal-1.md` | ~947 | Active |
<!-- @@@/auto:active-documents -->

---

## Session History

<!-- @@@auto:session-history -->
| # | Date | Title | Commits | Branch |
|---|------|-------|---------|--------|
| 28 | 2026-05-15 | Extract filter builders from get.go for pure-unit-testability | `def8019`, `0bf9151` | `main` |
| 27 | 2026-05-14 | Relay design doc: event hints, discovery patterns, local relay role | `5208a36` | `main` |
| 26 | 2026-05-14 | Local relay cache-only: exclude from writable relay list | `92a582e` | `main` |
| 25 | 2026-05-14 | Relay discovery fixes: error handling + relay selection consistency | `4b2a7be`, `a311f41` | `main` |
| 24 | 2026-05-14 | Fix compose tag deletion panic + correct delete behavior | `6ed6266` | `main` |
| 23 | 2026-05-13 | 实现 compose 发送遮罩反馈 | `4d86621`, `39674f4`, `0539fe5` | `main` |
| 22 | 2026-05-13 | 实现 compose 发送遮罩反馈 | `0539fe5` | `main` |
| 21 | 2026-05-13 | 修复 compose 发送和超时问题 | `eb58f93` | `main` |
| 20 | 2026-05-13 | 修复 compose ctrl+enter 发送问题 | `aae13cf` | `main` |
| 19 | 2026-05-13 | 让所有 TUI 窗口全屏显示 | `9916d32` | `main` |
| 18 | 2026-05-13 | 修复 compose ctrl+enter 和 standalone esc 退出问题 | `8817aa9`, `9c317e2` | `main` |
| 17 | 2026-05-13 | 统一 TUI 显示命令代码，使用 bubblon 实现 | `33247e5` | `main` |
| 16 | 2026-05-13 | bubblon迁移 — 窗口切换修复未完成 | - | `main` |
| 15 | 2026-05-12 | unify-tui-ops brainstorm 完成，结论：当前架构够用 | - | `main` |
| 14 | 2026-05-12 | event-detail-compose-call 完成 + unify-tui-ops 回到 brainstorm | `3e04185`, `e994319` | `main` |
| 13 | 2026-05-12 | community timeline TUI | `dd621fa` | `main` |
| 12 | 2026-05-12 | event-detail-pager完成，4个任务brainstorm完成 | `c8bc56c`, `66a9aff` | `main` |
| 11 | 2026-05-12 | clarify proxy README, commit compose-form-ui | `5142c20`, `15f581e` | `main` |
| 10 | 2026-05-11 | Add utils tests (search, dm, post) | `80c4812` | `main` |
| 9 | 2026-05-11 | Add unit tests for utils modules | `80c4812` | `main` |
| 8 | 2026-05-11 | Implement NIP-50 search and DM TUI | `415e892` | `main` |
| 7 | 2026-05-11 | NIP-65 relay discovery via relay pool | `3157560`, `0b9be9f` | `main` |
| 6 | 2026-05-11 | Fix GetNote ID parsing and add NIP-19 format output | `e6fe409`, `77a1bcc`, `3bcf4fb` | `main` |
| 5 | 2026-05-10 | Implement event detail command with async loading | `e3486a2`, `df8891b`, `d1178a7`, `952a514`, `0dda5fc` | `main` |
| 4 | 2026-05-09 | Channel-based async queries + NOTICE suppression + TUI rate limit | `b6a3884`, `fe8fe19`, `0dddf9e`, `7917acb`, `bc9ced4` | `main` |
| 3 | 2026-05-09 | Inline session: TUI fix, community commands, cache filter, bleve | - | `main` |
| 2 | 2026-05-09 | 订阅功能配置化改进 | `021f731` | `main` |
| 1 | 2026-05-09 | Clean up Manus planning files; commit all WIP changes | `dc5a9a3`, `a810e6a`, `bbc9637`, `759282a` | `main` |
<!-- @@@/auto:session-history -->

---

## Notes

- Sessions are appended to journal files
- New journal file created when current exceeds 2000 lines
- Use `add_session.py` to record sessions