# community timeline TUI 改造

## Goal

将 `community timeline <community-addr>` 命令从当前的纯文本输出改为使用和 `note timeline` 一样的 TUI 界面（bubbletea timeline view）。

## What I already know

* `note timeline` → 调用 `timeline.RunTimeline(app, filter, hashtags, limit)` 启动 TUI
* `community timeline <community-addr>` → 当前是纯文本循环打印 events（cmd/community_commands.go:228-271）
* `timeline.RunTimeline` 调用 `NewModel(app, filter, hashtags, limit)` 创建 model
* `timeline/model.go` 的 filter switch：global / mine / default（followed）
* `GetCommunityPosts(ctx, app, authorPubKey, communityID, limit)` 支持单社区查询
* `utils.ParseCommunityAddr(communityAddr)` 解析地址得到 authorPubKey 和 communityID

## Requirements

* `community timeline <community-addr>` 打开 TUI 界面，显示该社区的 posts
* 复用 `note timeline` 的 TUI 代码（model、list、delegate、styles）
* TUI 内显示 `[Community]` 前缀（复用现有 `kindCommunity` 渲染逻辑）
* 点击 post 能打开 event detail
* 无限滚动、分页等功能与 note timeline 一致

## Technical Approach

### 1. timeline/model.go 改动

在 `NewModel` 添加 `communityAddr` 参数（可选，空字符串表示不用）：

```go
func NewModel(app *config.AppContext, filter string, hashtags []string, limit int, communityAddr string) *model {
```

model 结构体添加字段：
```go
communityAddr string
```

在 fetch 逻辑中添加 `case "community"`：
```go
case "community":
    authorPubKey, communityID, err := utils.ParseCommunityAddr(m.communityAddr)
    if err != nil { return errorMsg{err: err} }
    ch = utils.GetCommunityPosts(ctx, m.app, authorPubKey, communityID, m.limit)
```

在 `fetchMoreOld` 的 filter switch 中同样添加 `case "community"`。

### 2. timeline/main.go 改动

`RunTimeline` 添加 `communityAddr` 参数：
```go
func RunTimeline(app *config.AppContext, filter string, hashtags []string, limit int, communityAddr string) error {
    m := NewModel(app, filter, hashtags, limit, communityAddr)
    ...
}
```

### 3. cmd/community_commands.go 改动

将 `communityTimelineCmd` 从纯文本输出改为调用 TUI：
```go
communityTimelineCmd := &cobra.Command{
    Use:   "timeline <community-addr>",
    Short: "Show community timeline with TUI",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        communityAddr := args[0]
        app := getApp()
        if err := timeline.RunTimeline(app, "community", nil, limit, communityAddr); err != nil {
            handleError(err)
        }
    },
}
```

删除原来的纯文本实现代码（lines 228-271 的 fmt.Printf 循环）。

### 4. cmd/note_commands.go 改动

`noteTimelineCmd` 调用处需更新为 5 参数：
```go
timeline.RunTimeline(app, filter, hashtags, limit, "")
```

## Acceptance Criteria

* [x] `nosmec community timeline <community-addr>` 打开 TUI 界面
* [x] TUI 显示该社区的 posts
* [x] 显示 `[Community]` 前缀
* [x] 点击 post 能打开 event detail
* [x] 无限滚动、分页正常
* [x] lint / typecheck / tests pass
