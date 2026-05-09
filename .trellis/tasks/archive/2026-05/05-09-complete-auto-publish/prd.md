# 补全 auto-publish

## Goal

在 subscription 和 dm-relay 变更后自动 publish 到网络。

## What to fix

1. `config subscribe add user/community/hashtag` 后调用 `utils.PublishSubscriptions`
2. `config subscribe remove` 后调用 `utils.PublishSubscriptions`
3. 添加 `config dm-relay sync` 命令（从网络同步 DM relays）
4. `config dm-relay add/remove` 后调用 `utils.PublishRelayList`

## Technical Notes

### cmd/config_commands.go 修改

```go
// FollowUser 后
if err := utils.FollowUser(...); err != nil { ... }
utils.PublishSubscriptions(ctx, getApp())  // 添加

// FollowCommunity 后
if err := utils.FollowCommunity(...); err != nil { ... }
utils.PublishSubscriptions(ctx, getApp())  // 添加

// FollowHashtag 后
if err := utils.FollowHashtag(...); err != nil { ... }
utils.PublishSubscriptions(ctx, getApp())  // 添加

// UnfollowUser/Community/Hashtag 后同理

// 添加 config dm-relay sync 命令
configDMRelaySyncCmd := &cobra.Command{
    Use: "sync",
    Short: "Sync DM relays from network",
    Run: func(cmd *cobra.Command, args []string) {
        ctx := context.Background()
        if err := utils.SyncDMRelaysFromNetwork(ctx, getApp()); err != nil {
            handleError(...)
        }
        fmt.Println("DM relays synced")
    },
}
```

### utils/sync.go 或 relay_list.go

需要添加 `SyncDMRelaysFromNetwork` 函数（目前 `syncDMRelaysFromNetwork` 是小写未导出）。
