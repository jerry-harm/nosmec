# Journal - jerry (Part 1)

> AI development session journal
> Started: 2026-05-09

---



## Session 1: Clean up Manus planning files; commit all WIP changes

**Date**: 2026-05-09
**Task**: Clean up Manus planning files; commit all WIP changes
**Branch**: `main`

### Summary

删除根目录多余的 task_plan/findings/progress 三个 Manus 风格文件；完成并提交 Trellis 基础设施更新、Timeline TUI 重构、cmd 命令重构、logger/config/relay 工具更新；归档 bootstrap-guidelines 任务。

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `dc5a9a3` | (see git log) |
| `a810e6a` | (see git log) |
| `bbc9637` | (see git log) |
| `759282a` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 2: 订阅功能配置化改进

**Date**: 2026-05-09
**Task**: 订阅功能配置化改进
**Branch**: `main`

### Summary

扩展 profile --full 输出，支持从网络读取 Kind 3/10004/10015 获取 follows、communities、hashtags 完整列表

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `021f731` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 3: Inline session: TUI fix, community commands, cache filter, bleve

**Date**: 2026-05-09
**Task**: Inline session: TUI fix, community commands, cache filter, bleve
**Branch**: `main`

### Summary

TUI详情页截断修复; 统一community命令ID格式; 修复CacheFilter初始化逻辑和ToNostr PubKey转换; PostNote改用AllWritableRelays发布到本地relay; 存储从lmdb切换到boltdb+bleve支持全文搜索

### Main Changes

(Add details)

### Git Commits

(No commits - planning session)

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 4: Channel-based async queries + NOTICE suppression + TUI rate limit

**Date**: 2026-05-09
**Task**: Channel-based async queries + NOTICE suppression + TUI rate limit
**Branch**: `main`

### Summary

Refactored all nostr query functions (GetMyTimeline, GetGlobalTimeline, GetFollowedTimeline, GetCommunityPosts, GetMyCreatedCommunities, GetPostedCommunities) to return chan *nostr.Event instead of ([]Event, error), yielding events as they arrive. Added NoticeHandler to config.NewPool() to suppress 'too many concurrent REQs' NOTICE noise (logs at DEBUG instead of stderr). Added 2-second refresh rate limit in TUI fetchTimeline via lastRefresh cooldown. All changes committed.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `b6a3884` | (see git log) |
| `fe8fe19` | (see git log) |
| `0dddf9e` | (see git log) |
| `7917acb` | (see git log) |
| `bc9ced4` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 5: Implement event detail command with async loading

**Date**: 2026-05-10
**Task**: Implement event detail command with async loading
**Branch**: `main`

### Summary

Implemented nosmec event command with async relay queries, QuerySingle for non-replaceable events, FetchManyReplaceable for replaceable kinds, proper TUI with viewport, j key for raw JSON toggle, and bubbles/help

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `e3486a2` | (see git log) |
| `df8891b` | (see git log) |
| `d1178a7` | (see git log) |
| `952a514` | (see git log) |
| `0dda5fc` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 6: Fix GetNote ID parsing and add NIP-19 format output

**Date**: 2026-05-11
**Task**: Fix GetNote ID parsing and add NIP-19 format output
**Branch**: `main`

### Summary

Fixed GetNote/GetNoteAsync using copy() instead of nostr.IDFromHex. Updated all CLI and TUI output to use npub/nevent format. Added nevent input support to event command. Documented the bug and NIP-19 convention in spec.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `e6fe409` | (see git log) |
| `77a1bcc` | (see git log) |
| `3bcf4fb` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 7: NIP-65 relay discovery via relay pool

**Date**: 2026-05-11
**Task**: NIP-65 relay discovery via relay pool
**Branch**: `main`

### Summary

Implemented NIP-65 relay discovery that queries local relay (cache) + remote relays simultaneously using FetchManyReplaceable for Kind 10002. Discovered relays are registered in global pool via EnsureRelay (lazy connection), cached to local relay via CacheEvent, and tracked in known_relays. GetProfile triggers discovery before querying.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `3157560` | (see git log) |
| `0b9be9f` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 8: Implement NIP-50 search and DM TUI

**Date**: 2026-05-11
**Task**: Implement NIP-50 search and DM TUI
**Branch**: `main`

### Summary

Implemented NIP-50 search (search command with kinds:/authors:/#t: filters, Bleve full-text index, local relay + remote relay dual-source) and DM TUI (dm npub command, viewport+textarea, NIP-17 GiftWrap send/receive, network-confirmed messaging).

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `415e892` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 9: Add unit tests for utils modules

**Date**: 2026-05-11
**Task**: Add unit tests for utils modules
**Branch**: `main`

### Summary

Added unit tests for utils modules (search, dm, post) covering NIP-50 filter parsing, DM types and filtering, and post tag construction. Tests use table-driven approach with no network dependencies.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `80c4812` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 10: Add utils tests (search, dm, post)

**Date**: 2026-05-11
**Task**: Add utils tests (search, dm, post)
**Branch**: `main`

### Summary

Added unit tests for utils modules: search (ParseSearchFilter), dm (Conversation/DMMessage/filter), post (tag construction). go test ./... passes.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `80c4812` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 11: clarify proxy README, commit compose-form-ui

**Date**: 2026-05-12
**Task**: clarify proxy README, commit compose-form-ui
**Branch**: `main`

### Summary

Updated proxy docs in README (clarified socks/onion/i2p behavior); committed compose TUI form enhancement

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `5142c20` | (see git log) |
| `15f581e` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 12: event-detail-pager完成，4个任务brainstorm完成

**Date**: 2026-05-12
**Task**: event-detail-pager完成，4个任务brainstorm完成
**Branch**: `main`

### Summary

完成event-detail-pager任务：移除glamour改纯文本渲染，修复help高度bug。完成4个TUI任务的brainstorm规划：event-detail-pager(完成)、event-detail-compose-call、community-timeline、unify-tui-ops。

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `c8bc56c` | (see git log) |
| `66a9aff` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete
