# 统一TUI操作逻辑

## Goal

统一 TUI 各窗口的 esc 退出行为、help 显示格式、以及 keymap 构建范式，让所有 window.Window 实现都遵循同样的模式。

## What I already know

* `window.Window` interface: Init/Update/View/ID
* `windowManager` 管理多个窗口的 Open/Close/Update/View
* 各模块 keymap 现状：
  * `timeline`: `listKeyMap`（refresh/quit/spinner等）+ `delegateKeyMap`（enter=view）
  * `event detail`: `eventKeyMap`（reply/quote/delete/follow/open/json/quit）
  * `compose`: `keyMap`（send/quit/nextField/addTag/removeTag等）
  * `delegate`: `delegateKeyMap`（enter=view）
* esc 行为不统一：compose 是 tea.Quit，event detail 是 CloseMsg，timeline 是 q=quit
* help 格式各不同：timeline 用 list.AdditionalFullHelpKeys，event detail 用 help.New() + eventKeyMap，compose 用 help.New() + keyMap

## Requirements

### 1. Esc 统一行为

* **esc = close 当前窗口**
* 如果当前窗口是 windowManager 里**唯一**的窗口，esc = **quit 程序**
* 所有 window.Window 的 Update 在收到 `tea.KeyMsg("esc")` 时：
  * 如果是根窗口（timeline/dm 等），发送 `windowmanager.CloseAllMsg{}` 并返回 `tea.Quit`
  * 如果是子窗口（event detail/compose 等），发送 close 消息给 windowManager，返回 nil

### 2. Help 格式统一

* 所有 window.Window 实现统一使用 `charm.land/bubbles/v2/help` 的标准 `help.Model`
* 统一 help 显示格式：`两列布局，按键 + 说明`
* 不再混用各自定义的 FullHelpFunc / ShortHelpFunc 范式

### 3. Keymap 统一范式

定义全局 base keymap + per-window extend 的模式：

```go
// 全局 keymap - 所有 window 都有
type BaseKeyMap struct {
    Esc   key.Binding  // esc = close / quit
    Help  key.Binding  // h = toggle help
}

// Per-window keymap
type WindowKeyMap struct {
    Base *BaseKeyMap   // 继承
    Extra []key.Binding // 扩展
}
```

所有 window.Window 的 Update 先检查 BaseKeyMap，再检查 WindowKeyMap。

### 4. Window 构建范式

所有 window.Window 实现遵循：

```go
type MyWindow struct {
    windowManager *windowmanager.WindowManager  // 持有引用
    keys         *MyWindowKeyMap
    help        help.Model
    // ... 其他字段
}

func NewMyWindow(wm *windowmanager.WindowManager, ...) *MyWindow {
    m := &MyWindow{
        windowManager: wm,
        keys:          newMyWindowKeyMap(),
        help:          help.New(),
    }
    m.help = m.keys.Help()
    return m
}

func (m *MyWindow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        // 1. 先检查 BaseKeyMap（esc, help）
        if key.Matches(msg, m.keys.Base.Esc) {
            return m.handleEsc()
        }
        // 2. 再检查本窗口特定 keys
        if key.Matches(msg, m.keys.Local.Action) {
            return m.handleAction()
        }
    }
    // ...
}

func (m *MyWindow) handleEsc() tea.Cmd {
    if m.isRootWindow() {
        return tea.Quit
    }
    return m.windowManager.CloseSelf()
}
```

### 5. windowManager 关闭最后一个窗口时

如果 Close 导致 `WindowCount() == 0`，windowManager 返回 `tea.Quit` 消息，上层收到后程序退出。

## Acceptance Criteria

* [ ] esc 在所有子窗口（event detail/compose）触发 close
* [ ] esc 在根窗口（timeline）且无其他窗口时触发 quit
* [ ] 所有窗口 help 格式统一（两列布局）
* [ ] keymap 遵循 Base + Local 继承模式
* [ ] 各窗口代码遵循 New + Init + Update + View 范式
* [ ] lint / typecheck / tests pass

## Out of Scope

* 新增功能（只做架构统一，不做功能改动）
* DM 视图改动（如果 dm 有独立实现的话）

## Technical Approach

### 新增 BaseKeyMap

在 `tui/toolkit/toolkit.go` 或新文件 `tui/keymap/base.go`：

```go
type BaseKeyMap struct {
    Esc  key.Binding
    Help key.Binding
}

func NewBaseKeyMap() *BaseKeyMap {
    return &BaseKeyMap{
        Esc: key.NewBinding(
            key.WithKeys("esc"),
            key.WithHelp("esc", "close"),
        ),
        Help: key.NewBinding(
            key.WithKeys("h"),
            key.WithHelp("h", "help"),
        ),
    }
}
```

### windowManager 关闭最后一个窗口

```go
func (wm *WindowManager) Close(id string) tea.Cmd {
    wm.mu.Lock()
    delete(wm.windows, id)
    wm.removeFromStack(id)
    wm.mu.Unlock()

    if wm.WindowCount() == 0 {
        return tea.Quit
    }
    return nil
}
```

### 各窗口改动

#### event detail (`event.go`)
* `esc` 从 CloseMsg 改为 windowManager.Close()
* 移除独立的 quit help item（esc 已经处理）
* help 格式统一

#### compose (`compose/model.go`)
* `esc` 不再是 `tea.Quit`，而是 `windowManager.Close("compose")`
* 如果是最后一个窗口，返回 `tea.Quit`
* help 格式统一

#### timeline (`timeline/model.go`)
* `q` 键改为 `esc`
* esc 如果有子窗口则 close 子窗口，否则 quit
* help 格式统一

#### delegate (`timeline/delegate.go`)
* delegate 的按键处理保持（enter=view 是 list 的内置行为）

## 依赖关系

此任务依赖 event-detail-compose-call 完成后，因为 compose 会改造成 window.Window 并接入 windowManager。