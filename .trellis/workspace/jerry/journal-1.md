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
