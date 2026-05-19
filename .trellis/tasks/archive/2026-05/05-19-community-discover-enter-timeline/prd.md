# community discover Enter timeline - 界面不同且不显示事件

## Goal

修复从 community discover 按 Enter 进入的 timeline 与直接 `nosmec community timeline` 进入的 timeline 行为不一致的问题。

## What I already know

* `nosmec community timeline <addr>` 直接命令进入：显示 6 items，绿色标题栏，界面正常
* community discover 中 Enter 进入：只显示 spinner（"No items"），背景是灰色而非绿色，无 items 显示
* 两者调用相同的 `timeline.NewModel(m.app, "community", nil, 10, communityAddr)`
* 进入后 timeline model 的 Init() 会调用 fetchTimeline() → fetchMsg → items

## Root Cause

从 discover 的 bubblon stack 进入 timeline 时，`timeline.Init()` 可能没有被正确调用，或者 `WindowSizeMsg` 没有被正确处理，或者某些消息路由出了问题。

对比 `cmd/community_commands.go` 的 `timeline.RunTimeline(app, "community", nil, limit, communityAddr)` — 它用 `bubblon.New(tlModel)` 包装然后 `tea.NewProgram(ctrl).Run()`。

从 discover 进入时用 `timeline.NewModel()` + `bubblon.Open(tlModel)`。

**关键差异**: `RunTimeline` 调用 `tlModel.SetBubblonController(&ctrl)`，而 discover 的 Enter handler 也调用了 `tlModel.SetBubblonController(m.ctrl)`。

但更重要的是：从 `RunTimeline` 进入时，program 的初始消息序列（`tea.WindowSizeMsg` 等）会正确到达 timeline model。从 bubblon stack push 进入时，timeline 的 `Init()` 被 `bubblon.Open` 返回的 cmd 调用了吗？

`bubblon.Open` 返回 `msg.model.Init()` — 所以 Init() 会被调用。问题可能在于：
1. `WindowSizeMsg` 没有被处理
2. `fetchTimeline()` 发起的 goroutine 在 discover 被 Open 时没有正确执行
3. 或者 fetchTimeline 的 2s rate limit 阻止了刷新

## Acceptance Criteria

* [ ] 从 discover Enter 进入的 timeline 显示和直接 `community timeline` 命令进入的一样的内容
* [ ] 两者 UI 样式一致（标题栏颜色、背景等）
* [ ] `go build ./...` 通过

## Out of Scope

* 不改 community discover 的导航逻辑（除了必要的修复）
* 不改 timeline model 的 fetch 逻辑