# event-detail-actions: EventView 交互操作 + WindowManager 重构

## Goal

为 nosmec TUI 引入完整的 WindowManager 框架，使 Event 详情视图支持键盘交互操作，并为未来更多窗口页面做好准备。

## What I already know

* 项目: nosmec - 基于 Bubble Tea v2 的 Nostr CLI 客户端
* Event 详情视图位于 `tui/window/event/`
* 当前 EventView 只有展示功能，无键盘交互
* Footer 显示 "esc: close" 但未实现实际关闭逻辑
* 已有 utility 函数: `PostNote`, `ReplyToNote`, `QuoteNote`, `FollowUser`, `UnfollowUser`
* 缺少: `DeleteNote` 函数 (Kind 5 deletion event)
* neonmodem 有完整的 WindowManager + ToolKit 框架可参考

## Architecture Decision

**采用方案 C — 完整 WindowManager**

### 组件设计

1. **Window 接口** (`tui/window/window.go` 已存在)
   ```go
   type Window interface {
       Init() tea.Cmd
       Update(msg tea.Msg) (tea.Model, tea.Cmd)
       View() string
       ID() string
   }
   ```

2. **WindowManager** (`tui/windowmanager/`)
   - 栈式管理多个打开的 window
   - 支持 Open/Close/Focus/Update/Resize
   - `UpdateAll` / `UpdateFocused` 分发消息
   - `View` 渲染所有 window（按 z-order 叠加）

3. **ToolKit** (`tui/toolkit/`)
   - 按键注册: `KeymapAdd(id, help, keys...)`
   - `KeymapHelpStrings()` 返回排序后的底部提示
   - Focus/Blur 状态管理
   - View 缓存

4. **cmd 消息系统** (`tui/cmd/`)
   - `WinOpen` / `WinClose` / `WinFocus` / `WinBlur`
   - `ViewFocus` / `ViewBlur`
   - `WinFreshData` / `WinRefreshData`

5. **EventView 改造** (`tui/window/event/`)
   - 实现 ToolKit 集成
   - KeyMap: `reply`, `quote`, `delete`, `follow`, `open`
   - Footer 动态显示按键提示

### 消息流

```
timeline model
    │
    ├── tea.KeyMsg ──→ WindowManager.UpdateFocused()
    │                      │
    │                      └── EventView (通过 ToolKit 处理按键)
    │
    └── tea.WindowSizeMsg ──→ WindowManager.ResizeAll()
                                 │
                                 └── EventView (更新 viewport)
```

## Requirements

* [ ] 实现 `WindowManager` (`tui/windowmanager/windowmanager.go`)
* [ ] 实现 `ToolKit` (`tui/toolkit/toolkit.go`)
* [ ] 实现 `cmd` 消息类型 (`tui/cmd/cmd.go`)
* [ ] 实现 `DeleteNote` 函数 (`utils/post.go`)
* [ ] 改造 `EventView` 集成 ToolKit
* [ ] Timeline 详情模式通过 WindowManager 管理
* [ ] 底部 footer 显示动态按键提示
* [ ] ESC 关闭当前 window

## KeyMap Actions

| 按键 | 操作 | 说明 |
|------|------|------|
| `r` | reply | 回复 author |
| `q` | quote | 引用转发 |
| `d` | delete | 发送 deletion request (Kind 5) |
| `f` | follow | 关注/取消关注 author |
| `o` | open | 在浏览器打开 event URL |
| `esc` | close | 关闭详情 window |

## Out of Scope

* 回复/引用的文本输入 UI（后续 task）
* 多 window 并排显示（后续 task）
* 其他 window 迁移到 WindowManager（后续 task）

## Technical Notes

* 参考 `neonmodem/ui/windowmanager/windowmanager.go`
* 参考 `neonmodem/ui/toolkit/{toolkit,keymap,dialog,msg}.go`
* 参考 `neonmodem/ui/windows/postshow/` 的 ToolKit 使用模式
* Bubble Tea v2 API: `charm.land/bubbletea/v2`
