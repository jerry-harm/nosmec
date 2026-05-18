# 清理 utils 过度封装的函数

## Goal

删除 `utils/get.go` 和 `utils/community.go` 中所有对 sdkplus 的封装函数，逻辑全部 inline 到 tui/cmd 调用方。

## Decisions

- `utils/get.go` → **文件全删**，`GetOptions`、`TimelineEvent` 也删
- `utils/community.go` → **删所有 wrapper 函数**（GetCommunity, GetCommunityRelays, GetCommunityPosts, GetPost, GetParentPostInfo, GetPostedCommunities, GetEvent）
- 所有函数逻辑 inline 到 tui/cmd 调用方
- 调用方直接用 sdkplus 或直接网络请求

## 调用方更新清单

| 文件 | 当前调用 | 替换为 |
|---|---|---|
| `tui/timeline/model.go` | GetGlobalTimeline, GetMyTimeline, GetFollowedTimeline, GetCommunityPosts, GetProfileNames | sdkplus.Fetch*Page 直接调用 |
| `tui/dm/model.go` | GetProfileName | sdkplus.FetchProfileMetadata |
| `tui/event/event.go` | GetNote, GetProfileName | sdkplus.FetchNote, FetchProfileMetadata |
| `tui/thread/thread.go` | GetProfileNames | sdkplus.FetchProfilesBatch |
| `tui/compose/model.go` | FindRootEvent, GetNote | nip10.GetThreadRoot inline, sdkplus.FetchNote |
| `tui/component/label/model.go` | GetProfileName | sdkplus.FetchProfileMetadata |
| `tui/dm/list/model.go` | GetProfileName | sdkplus.FetchProfileMetadata |
| `cmd/note_commands.go` | GetNote, GetProfileName | sdkplus |
| `cmd/dm_commands.go` | GetProfileName | sdkplus |
| `cmd/community_commands.go` | GetCommunity, GetPostedCommunities | sdkplus |
| `cmd/profile_commands.go` | GetFullProfile, GetProfile | sdkplus |
| `cmd/search_commands.go` | GetProfileName | sdkplus |
| `cmd/config_commands.go` | GetFollowedTimeline | sdkplus |

## Acceptance Criteria

- [ ] `utils/get.go` 文件已删除
- [ ] `utils/community.go` 只保留无 wrapper 的工具函数（如有）
- [ ] 所有调用方无残留 `utils.` 调用（get/community 相关）
- [ ] `go build ./... && go vet ./... && go test ./...` 通过

## Out of Scope

- sdkplus 本身不改
- utils/profile.go 等暂不动