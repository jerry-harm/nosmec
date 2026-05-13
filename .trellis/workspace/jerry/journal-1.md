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


## Session 13: community timeline TUI

**Date**: 2026-05-12
**Task**: community timeline TUI
**Branch**: `main`

### Summary

实现 community timeline TUI：cmd/community_commands.go 纯文本输出改为调用 timeline.RunTimeline TUI；timeline/model.go 添加 communityAddr 字段和 community filter case；timeline/main.go RunTimeline 新增第5参数。

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `dd621fa` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 14: event-detail-compose-call 完成 + unify-tui-ops 回到 brainstorm

**Date**: 2026-05-12
**Task**: event-detail-compose-call 完成 + unify-tui-ops 回到 brainstorm
**Branch**: `main`

### Summary

event-detail-compose-call: 修复 RWMutex deadlock、nil panic、overlay 渲染；完成 reply/quote 通过 windowManager 打开 compose 的功能。用户撤销了最终 commit想把 isRootWindow 判断问题留在 unify-tui-ops brainstorm 里讨论。unify-tui-ops 回到 brainstorm 阶段。

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `3e04185` | (see git log) |
| `e994319` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 15: unify-tui-ops brainstorm 完成，结论：当前架构够用

**Date**: 2026-05-12
**Task**: unify-tui-ops brainstorm 完成，结论：当前架构够用
**Branch**: `main`

### Summary

unify-tui-ops brainstorm：研究 neonmodem 的 wm 实现后确认 nosmec 现有逻辑已符合要求。结论：不需要 isRootWindow()（WindowCount==0 够用），不需要 BaseKeyMap（只有一个 Esc），help 始终显示。PRD 更新后归档。

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


## Session 16: bubblon迁移 — 窗口切换修复未完成

**Date**: 2026-05-13
**Task**: bubblon迁移 — 窗口切换修复未完成
**Branch**: `main`

### Summary

尝试修复bubblon窗口切换: 1) 将ctrl字段改为指针类型并在main.go初始化; 2) 修复View()中m.ctrl.Models()>1条件防止递归; 3) 将Enter处理从delegate移到timeline.Update直接处理; 4) 修复compose/model.go中tea.KeyPressMsg类型转换; 5) event.go中esc返回bubblon.Close()。最终仍无法通过Enter进入event detail，用户决定放弃本次修复，丢弃所有更改。

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


## Session 17: 统一 TUI 显示命令代码，使用 bubblon 实现

**Date**: 2026-05-13
**Task**: 统一 TUI 显示命令代码，使用 bubblon 实现
**Branch**: `main`

### Summary

将 event 命令从直接 tea.NewProgram 改为 bubblon.Controller 包装，与 timeline/compose 保持一致。删除 orphaned tui/cmd/cmd.go。

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `33247e5` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 18: 修复 compose ctrl+enter 和 standalone esc 退出问题

**Date**: 2026-05-13
**Task**: 修复 compose ctrl+enter 和 standalone esc 退出问题
**Branch**: `main`

### Summary

修复 compose 页面两个按键问题：1) 将 tea.KeyMsg 改为 tea.KeyPressMsg 修复 ctrl+enter 无法触发 send；2) 添加 isStandalone 模式检测使 standalone compose 可通过 esc 退出。

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `8817aa9` | (see git log) |
| `9c317e2` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 19: 让所有 TUI 窗口全屏显示

**Date**: 2026-05-13
**Task**: 让所有 TUI 窗口全屏显示
**Branch**: `main`

### Summary

为 event 和 dm 两个窗口的 View() 添加 AltScreen=true，使所有 TUI 窗口（timeline/compose/event/dm）都使用全屏模式显示。

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `9916d32` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 20: 修复 compose ctrl+enter 发送问题

**Date**: 2026-05-13
**Task**: 修复 compose ctrl+enter 发送问题
**Branch**: `main`

### Summary

修复 compose 页面 ctrl+enter 无法发送的问题。根因是 key.Matches() 在 v2 中对 ctrl+enter 匹配失败，改用 msg.String() == "ctrl+enter" 直接字符串匹配解决。

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `aae13cf` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete
