# event详情compose调用

## Goal

在 event 详情里通过 `r` / `q` 等快捷键调用全局的 compose window，实现 reply/quote 的草稿功能。compose 窗口关闭后状态保留，下次打开继续编辑。

## What I already know

* `window.Window` interface: Init/Update/View/ID
* `windowManager` 已有 `Open(win Window)` / `Close(id)` / `WindowCount()` / `View()` 等方法
* `windowManager` 在 `timeline/model.go` 里用 `wm.Open(ev)` 打开 event detail
* `compose/model.go` 已有：KindNote/KindReply/KindQuote/KindCommunity 四种 kind
* `compose/model.go` 有 `NewReplyCompose` / `NewQuoteCompose` / `NewCommunityCompose` 工厂函数
* `compose/main.go` 有 `RunReplyCompose` / `RunQuoteCompose` 等，各自创建独立 tea.Program
* event detail 里 `r` 键触发 `reply()`，`q` 键触发 `quote()`，目前都是 `logger.Debug("not implemented")`
* 草稿需求：compose 关闭后 content/tags/kind 状态保留

## Requirements

* compose.model 改造成可在已有实例上操作（不清空状态）
* compose 通过 windowManager 打开/关闭，不走独立 tea.Program
* compose.model 在 windowManager 里单例维护，关闭时保留状态（不是 tea.Quit，只是隐藏）
* event detail 按键行为：
  * `r` → compose.AddReply(parentEvent) + windowManager.Open(compose)
  * `q` → compose.AddQuote(parentEvent) + windowManager.Open(compose)
  * `esc` → windowManager.Close(compose.ComposeWindowID)
* compose 打开时持续聚焦，esc 关闭但保留草稿
* 如果 compose 已打开，再次触发 r/q 只是更新 tag/kind，不重复打开
* compose 支持手动清空草稿（发送成功后自动清空）

## Acceptance Criteria

* [ ] event detail 按 `r` 能打开 compose 并带入 reply tag（e: + p:）
* [ ] event detail 按 `q` 能打开 compose 并带入 quote tag（q:）
* [ ] compose 关闭后草稿（content、tags）保留
* [ ] 再次按 r/q 能追加 tag 而不覆盖已有草稿
* [ ] 发送成功后草稿清空
* [ ] 如果 compose 已打开，r/q 不重复创建实例

## Definition of Done

* reply / quote 从 event detail 能正确打开 compose 并带入 tag
* 草稿状态在 compose 关闭后保留
* lint / typecheck / tests pass

## Out of Scope

* community compose（后续独立任务）
* compose 的 UI 交互改动（只改调用方式，不改 compose 内部逻辑）
* timeline 之外的 compose 入口

## Technical Approach

### 改动 compose/model.go

移除工厂函数式构造，改为操作已有实例：

```go
// 已有 model 上的操作方法
func (m *model) AddReply(parentEvent *nostr.Event) {
    m.composeKind = KindReply
    m.parentEvent = parentEvent
    m.parentID = parentEvent.ID.Hex()
    m.tags = []TagValue{
        {Type: "e", Values: []string{parentEvent.ID.Hex()}},
        {Type: "p", Values: []string{parentEvent.PubKey.Hex()}},
    }
}

func (m *model) AddQuote(parentEvent *nostr.Event) {
    m.composeKind = KindQuote
    m.parentEvent = parentEvent
    m.quotedID = parentEvent.ID.Hex()
    m.tags = []TagValue{
        {Type: "q", Values: []string{parentEvent.ID.Hex()}},
    }
}

// 清空草稿（发送成功后调用）
func (m *model) ClearDraft() {
    m.contentInput.SetValue("")
    m.kindInput.SetValue("")
    m.tags = nil
    m.parentEvent = nil
    m.parentID = ""
    m.quotedID = ""
    m.communityAddr = ""
    m.composeKind = KindNote
    m.errMsg = ""
    m.success = false
}
```

### 改动 windowManager

windowManager 需要能持有 compose 实例：

```go
type WindowManager struct {
    mu          sync.RWMutex
    windows     map[string]window.Window
    stack       []string
    focused     string
    composeModel *compose.Model  // 全局单例
}
```

### 改动 event detail

`reply()` / `quote()` 调用：

```go
func (m *EventView) reply() tea.Cmd {
    if m.event == nil {
        return nil
    }
    m.windowManager.ComposeModel().AddReply(m.event)
    return m.windowManager.OpenCompose()
}

func (m *EventView) quote() tea.Cmd {
    if m.event == nil {
        return nil
    }
    m.windowManager.ComposeModel().AddQuote(m.event)
    return m.windowManager.OpenCompose()
}
```

### 发送成功后清空草稿

compose 的 `sendSuccessMsg` 需要通知清空状态。

### 关闭行为

compose 的 `esc` 不再是 `tea.Quit`，而是发送 `closeComposeMsg`，由 timeline 的 Update 收到后调用 `windowManager.Close(compose.ComposeWindowID)`，compose 实例保留在 windowManager 里。

### 依赖关系

此任务**依赖** event-detail-pager 先完成，因为 compose 调用会复用 windowManager 里 event detail 的集成方式。