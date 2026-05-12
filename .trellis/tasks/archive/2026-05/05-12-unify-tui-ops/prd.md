# 统一TUI操作逻辑

## Goal

统一 TUI 各窗口的 esc 退出行为、help 显示格式、以及 keymap 构建范式，让所有 window.Window 实现都遵循同样的模式。

## What I already know

* `window.Window` interface: Init/Update/View/ID
* `windowManager` 管理多个窗口的 Open/Close/Update/View
* `timeline` 持有 `windowManager`，是根窗口；`compose` 和 `event detail` 是子窗口
* 各模块 keymap 现状：
  * `timeline`: `listKeyMap`（refresh/quit/spinner等）+ `delegateKeyMap`（enter=view）
  * `event detail`: `eventKeyMap`（reply/quote/delete/follow/open/json/quit）
  * `compose`: `keyMap`（send/quit/nextField/addTag/removeTag等）
* esc 行为不统一：compose 是 tea.Quit，event detail 是 CloseMsg，timeline 是 q=quit
* help 格式各不同：timeline 用 list.AdditionalFullHelpKeys，event detail 用 help.New() + eventKeyMap，compose 用 help.New() + keyMap
* 根窗口判断：使用 `windowManager.WindowCount() == 0`，不需要 isRootWindow() 方法

## Requirements

### 1. Esc 统一行为

* **esc = close 当前子窗口**
* 如果 timeline 收到 esc 且 `WindowCount() == 0`（无子窗口），timeline 自己 quit
* timeline 收到 esc 且 `WindowCount() > 0`：路由给 `wm.UpdateFocused()`，由子窗口处理
* 子窗口收到 esc：发送 `CloseMsg` 给 timeline，timeline 调用 `wm.Close(id)`

当前逻辑已符合此要求，不需要新增 isRootWindow()。

### 2. Help 格式统一

* 所有 window.Window 实现统一使用 `charm.land/bubbles/v2/help` 的标准 `help.Model`
* help **始终显示**在 View 中，不需要 toggle 键
* 统一 help 显示格式：`两列布局，按键 + 说明`
* 不再混用各自定义的 FullHelpFunc / ShortHelpFunc 范式

### 3. Keymap 统一范式

各窗口保持自己的 keymap 结构，不强制引入 BaseKeyMap（只有一个 Esc 键时无必要抽象）。

每个窗口的 esc 处理遵循：

```go
case tea.KeyPressMsg:
    if key.Matches(msg, m.keys.esc) {  // m.keys.esc 是窗口自己的 keyBinding
        return m.handleEsc()
    }
```

handleEsc 由各窗口自行实现。

### 4. windowManager 关闭最后一个窗口时

如果 Close 导致 `WindowCount() == 0`，不返回 tea.Quit（timeline 自己 quit）。

## Acceptance Criteria

* [x] esc 在所有子窗口（event detail/compose）触发 close（通过 CloseMsg）
* [x] esc 在根窗口（timeline）且无其他窗口时触发 quit
* [ ] 所有窗口 help 格式统一（两列布局，始终显示）
* [ ] 各窗口代码遵循 New + Init + Update + View 范式
* [ ] lint / typecheck / tests pass

## Out of Scope

* 新增功能（只做架构统一，不做功能改动）
* DM 视图改动（如果 dm 有独立实现的话）

## Technical Approach

### 各窗口改动

#### event detail (`event.go`)
* `esc` 从 CloseMsg 改为 windowManager.Close()
* 移除独立的 quit help item（esc 已经处理）
* help 格式统一为 help.Model，始终显示

#### compose (`compose/model.go`)
* `esc` 不再是 `tea.Quit`，而是 `windowManager.Close("compose")`
* help 格式统一为 help.Model，始终显示

#### timeline (`timeline/model.go`)
* `q` 键改为 `esc`
* esc 如果有子窗口则路由给子窗口处理，否则 quit
* help 格式统一为 help.Model，始终显示

#### delegate (`timeline/delegate.go`)
* delegate 的按键处理保持（enter=view 是 list 的内置行为）

## 依赖关系

此任务依赖 event-detail-compose-call 完成后，因为 compose 会改造成 window.Window 并接入 windowManager。
