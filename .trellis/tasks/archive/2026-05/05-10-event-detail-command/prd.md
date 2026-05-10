# event详情命令

## Goal

实现一个一级 `event` 命令,可以通过 eventid 调起 event 详情界面,同时修改现有 event 详情组件的 UI 显示规范。

## What I already know

### 现有代码结构

1. **event 详情组件**位于 `tui/window/event/`
   - `event.go`: EventView 主结构,处理按键、初始化异步获取用户名
   - `view.go`: renderHeader() 和 renderContent() 负责渲染
   - `styles.go`: 样式定义

2. **当前 renderHeader() 布局** (view.go:8-29):
   - 显示 `@authorName (pubkey[:8])… | time | kind`
   - pubkey 只显示前8字符,被截断

3. **windowmanager**: `tui/windowmanager/windowmanager.go` 管理窗口栈,`Open()` 方法调用 `win.Init()`

4. **现有命令** 在 `cmd/` 下,使用 cobra 框架,注册到 `RegisterCommandGroup`

5. **GetNote** (`utils/get.go:154-166`): 根据 noteID 获取 event

6. **timeline 如何打开 event 详情**:
   - `showDetailMsg` → `event.New(&msg.event.Event, m.app, m.width, m.height, msg.authorName)` → `m.windowManager.Open(ev)`

## Requirements

### 1. 新增 `event` 命令

- 一级命令:`nosmec event <event-id>`
- 异步加载 event:先显示 loading,加载完成后显示详情
- 打开后自动聚焦到 event 详情窗口

### 2. 修改 event 详情 header 显示

- **第一行**:显示完整 pubkey (不截断)
- **第二行**:显示 时间、用户名、kind
- 所有字符串不再截断,完全显示

### 3. 新增快捷键查看 raw JSON

- 按 `j` 键切换显示格式化的 raw message JSON
- 再次按 `j` 切换回正常内容视图

## Technical Approach

### 命令实现

1. 在 `cmd/` 下创建 `event_commands.go`
2. 实现 `RunEventDetail(app, eventID)` 函数启动独立 TUI
3. 使用 `tea.NewProgram` 运行 event 详情 window

### 异步加载

1. EventView 新增状态字段:
   - `loading bool`
   - `eventID string`
2. Init() 时发起异步获取,返回 `fetchEventMsg`
3. GetNote 支持异步返回 channel 或使用现有 QuerySingle

### Header 修改

- `renderHeader()` 改为两行输出:
  - Line 1: `PubKey: <full hex>`
  - Line 2: `@username | time | kind`

### Raw JSON 快捷键

- 新增 `showRawJSON bool` 状态
- `j` 键切换该状态
- `renderContent()` 根据状态返回不同内容:
  - 正常:渲染后的 content
  - JSON:格式化的 `nostr.Event` JSON

## Open Questions

- [x] event 命令的 TUI 模式:独立 tea.Program,EventView 可作为独立组件运行
- [x] 如何共享 app context:通过命令行 getApp() 获取

## Out of Scope

- 修改 timeline 中打开 event 详情的方式(仍使用现有方式)
- 修改 event 详情内部的其他逻辑(仅修改显示)

## Acceptance Criteria

- [ ] `nosmec event <64-char-event-id>` 可以打开详情界面
- [ ] 界面上方完整显示 pubkey (64字符不截断)
- [ ] pubkey 下方显示时间、用户名、kind
- [ ] 所有字符串内容不再被截断
- [ ] 按 `j` 可以查看格式化的 raw JSON,再按 `j` 返回
- [ ] loading 状态正确显示