# Journal - jerry (Part 2)

> Continuation from `journal-1.md` (archived at ~2000 lines)
> Started: 2026-05-26

---



## Session 60: relay list 改走 System 读取

**Date**: 2026-05-26
**Task**: relay list 改走 System 读取
**Branch**: `main`

### Summary

删除 config 中未使用的 relay codec helper，并将 relay list 从命令层直接读 LMDB 重构为通过 System.ListKnownEventRelays() 读取，补充 KVStore Iterate 接口与相关测试/spec。

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `18a8a0d` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 61: FetchFollowedTimelinePage race 修复

**Date**: 2026-05-26
**Task**: FetchFollowedTimelinePage race 修复
**Branch**: `main`

### Summary

修复 nostr_sdk 中 FetchFollowedTimelinePage 和 FetchFeedPage 的并发数据竞争：用 channel fan-in 替代多个 goroutine 直接写共享 slice，保留并发抓取性能收益的同时消除 race，并增加 seen 去重。

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `62795ef` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete
