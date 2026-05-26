# Workspace Index - jerry

> Journal tracking for AI development sessions.

---

## Current Status

<!-- @@@auto:current-status -->
- **Active File**: `journal-2.md`
- **Total Sessions**: 60
- **Last Active**: 2026-05-26
<!-- @@@/auto:current-status -->

---

## Active Documents

<!-- @@@auto:active-documents -->
| File | Lines | Status |
|------|-------|--------|
| `journal-2.md` | ~40 | Active |
| `journal-1.md` | ~1991 | Archived |
<!-- @@@/auto:active-documents -->

---

## Session History

<!-- @@@auto:session-history -->
| # | Date | Title | Commits | Branch |
|---|------|-------|---------|--------|
| 60 | 2026-05-26 | relay list 改走 System 读取 | `18a8a0d` | `main` |
| 59 | 2026-05-26 | 移除低价值 reply strategy 测试 | `820189e` | `main` |
| 58 | 2026-05-26 | 审计整改：AppContext 收拢 runtime 所有权 | `9e6559e`, `0ab9a93` | `main` |
| 57 | 2026-05-21 | Unify fetch relay selection to upstream-compatible defaultRelaysForFilter | `8f83805` | `main` |
| 56 | 2026-05-20 | Fix completion blocking by lazy-loading LMDB stores | `1b10568` | `main` |
| 55 | 2026-05-20 | Switch all persistent DB backends from bbolt/BoltDB to LMDB | `34c19d1` | `main` |
| 54 | 2026-05-20 | Add relay list command | `1bba05c` | `main` |
| 53 | 2026-05-20 | Event detail relay source and NIP-driven reply strategy | `59b4e0b` | `main` |
| 52 | 2026-05-20 | Narrow default nostr_sdk verification scope | `128fa0d` | `main` |
| 51 | 2026-05-20 | Refactor community parsing into nip72 and generic sdk fetch | `565b1a5` | `main` |
| 50 | 2026-05-19 | Add nip72 parsing layer and strict NIP community model | `06c5919` | `main` |
| 49 | 2026-05-19 | Support community event threads | `372e706` | `main` |
| 48 | 2026-05-19 | Fork nostr SDK into nostr_sdk | `ae52d3a`, `09e3981` | `main` |
| 47 | 2026-05-19 | community discover help bar 修复 | `1f283af` | `main` |
| 46 | 2026-05-19 | community discover help bar 修复 | `1f283af` | `main` |
| 45 | 2026-05-19 | community discover 刷新+spinner | `60d2bd9` | `main` |
| 44 | 2026-05-19 | community discover Enter timeline size 修复 | `26e00b5` | `main` |
| 43 | 2026-05-19 | community timeline relay fix: FetchFollowedTimelinePage community 分支 | `5d2fec8` | `main` |
| 42 | 2026-05-19 | community timeline 修复: EventView 显示地址栏, 刷新和帖子显示逻辑已验证 | `dc56c54` | `main` |
| 41 | 2026-05-19 | community discover 导航: Ctrl+E 进入 event 详情, Enter 进入 timeline | `570fdfe` | `main` |
| 40 | 2026-05-18 | 清理 utils 过度封装，sdkplus wrapper 全部移除 | `fa45ecd` | `main` |
| 39 | 2026-05-18 | 实现 label component + 目录重组 | `b074f7e`, `0c2d7e0` | `main` |
| 38 | 2026-05-18 | 重构访问系统 - 参考 nostr/sdk System 模块 | `3578ccd`, `3159849` | `main` |
| 37 | 2026-05-17 | implement HintsDB auto-learning and unified GetQueryRelays relay strategy | `e0cf3b1`, `097b632` | `main` |
| 36 | 2026-05-17 | unify event detail across all entry points, thread UX fixes | `12546e8` | `main` |
| 35 | 2026-05-17 | fix NIP-10 reply parsing, multi-level thread fetch, clarify kind:1111 scope | `dfea33a` | `main` |
| 34 | 2026-05-16 | test thread buildTuiModel, Update, View; fix placeholder cycle and hex validation | `939e81f` | `main` |
| 33 | 2026-05-16 | fix gossip npub decode, migrate thread to TuiTreeModel | `b90e209`, `508cedd` | `main` |
| 32 | 2026-05-16 | Migrate utils.ProfileMetadata to sdk.ProfileMetadata | `c9b1744` | `main` |
| 31 | 2026-05-16 | Complete systematic test coverage improvement with TDD | `7f94402`, `23d58b0`, `354c72a`, `ea4556c`, `cf1c9bb` | `main` |
| 30 | 2026-05-16 | compose tag UX + spinner | `dccfdf1`, `ff5cb89`, `deb24a3`, `5f95e9b` | `main` |
| 29 | 2026-05-15 | Compose tag input UX refactor | `1978167`, `3e39400`, `55f1b99`, `52e6ac2`, `d19fde5`, `4455015`, `ca9324d`, `7cbef92`, `56b61f3` | `main` |
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