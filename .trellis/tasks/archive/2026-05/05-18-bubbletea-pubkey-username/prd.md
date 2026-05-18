# 实现基于 BubbleTea 的 Pubkey/Username 可点击标签组件

## Goal

创建一个可复用的 bubbletea 组件，用于展示 pubkey/username 标签（如 `@jack` 或 `npub1abc...`），支持点击跳转到用户主页。替代当前 thread/timeline 等视图中用纯文本渲染用户名的做法。

## What I already know

* 从对话历史得知，thread 视图刚添加了 `globalNameCache` 用于缓存 profile name 到 pubkey 的映射
* `utils.GetProfileName()` / `utils.GetProfileNameAsync()` 是现有的 profile name 查询接口
* 项目用 BubbleTea v2.0.6 + Bubbles v2.1.0
* TUI 架构: 7 个包（bubblon, cmd, community, compose, dm, event, thread, timeline），均使用 `tea.Model` (Init/Update/View)
* 所有现有交互均为键盘驱动（`tea.KeyPressMsg`），无鼠标事件
* 无现有可点击 inline 组件，需从零构建
* Thread 的 `eventProvider.Name()` 展示 `{content} ({name})`，若无 name 则 fallback 到 `{content} ({pubkey[:8]})`

## Requirements (evolving)

* 新建 `tui/label/` 包，实现可复用的 pubkey 标签组件（tea.Model）
* 输入 pubkey → Init() 异步 fetch profile name → 先显示短 pubkey（`npub1abc...`），resolved 后显示 `@username`
* 组件内部通过 `tea.Cmd` 异步调 `utils.GetProfileNameAsync()`，收到结果后通过 Update 刷新显示
* 交互：鼠标点击 → emit `LabelClickedMsg{Pubkey string}`，父级 Update 接收处理
* 键盘 Tab+Enter：组件暴露 `Focus()`/`Blur()`/`IsFocused()` 方法供父级管理。当前 MVP 父级未启用（timeline 用 list delegate 处理 Enter；thread/event 无子级焦点导航需求），后续有需要再加
* 视觉：chip/tag 风格背景色块 + `@username` 前缀，三态区分 — 加载中（灰色）、普通态（青色背景）、hover（亮色）、focus（边框高亮）
* 初始集成：thread（替代 `eventProvider.Name()` 中 pubkey 展示）、timeline（事件头部作者名）、event 详情（事件作者/被回复者）
* 错误处理：超时或失败时保持灰色 pubkey 前缀显示

## Assumptions (temporary)

* 用户指的"后续可能还要允许通过这个组件进入用户主页" = MVP 阶段不实现用户主页，但预留 callback 接口
* 组件应可嵌入到 time, compose, dm 等多个视图中
* 组件需要作为内联元素使用（其他 bubbletea model 的子组件）

## Acceptance Criteria

* [ ] 组件在 `tui/label/` 包实现，符合 `tea.Model` 接口
* [ ] 输入 pubkey → 初始显示灰色 `npub1abc...` → 异步 fetch → 更新为 `@username` chip 样式
* [ ] 三态视觉：loading(灰)、normal(青底)、hover(亮色)、focus(边框高亮)
* [ ] 鼠标点击触发 `LabelClickedMsg{Pubkey}`
* [ ] 键盘 Tab 切换焦点（父级管理），Enter 触发 `LabelClickedMsg{Pubkey}`
* [ ] 暴露 `Focus()`/`Blur()`/`IsFocused()` 方法供父级焦点管理
* [ ] 集成到 thread view：替代 `eventProvider.Name()` 中的 `(pubkey[:8])` 部分
* [ ] 集成到 timeline view：事件头部作者名使用标签组件
* [ ] 集成到 event 详情 view：作者/被回复者使用标签组件
* [ ] `go build ./...` 成功

## Definition of Done

* Tests added/updated
* Lint / typecheck / CI green
* 在至少一个现有 TUI 视图中使用新组件

## Out of Scope (explicit)

* 用户主页弹窗/侧边栏（只 emit `LabelClickedMsg`，父级决定如何处理）
* 右键菜单 (copy npub, follow, unfollow)
* @mention 自动补全

## Decision (ADR-lite)

| 决策点 | 选择 | 理由 |
|--------|------|------|
| 交互通道 | 鼠标点击 + 键盘 Tab/Enter | 两者都做，可访问性最好 |
| 回调机制 | `LabelClickedMsg{Pubkey}` 自定义 tea.Msg | bubbletea 惯用，解耦组件与导航 |
| Name 解析 | 组件内部异步 fetch | 自包含，使用方便 |
| 视觉风格 | chip/tag 背景色块 + `@` 前缀 | 醒目的交互标识 |
| 焦点管理 | 父级管理 activeIndex，标签暴露 Focus/Blur | 职责清晰，避免多标签冲突 |
| 集成范围 | thread + timeline + event 详情 | 三个核心视图一次性覆盖 |

## Implementation Plan

1. **创建 `tui/label/` 包** — Model struct, Config, 状态机 (idle/loading/resolved/error)
2. **实现核心渲染** — chip/tag 样式（lipgloss），三态视觉（normal/hover/focus）
3. **实现 async fetch** — Init() → tea.Cmd → `utils.GetProfileNameAsync()` → nameResolvedMsg
4. **实现交互** — 鼠标 `tea.MouseMsg` hit testing + 键盘 Focus/Enter
5. **集成到 thread** — `eventProvider.Name()` 中 pubkey 部分替换为 label
6. **集成到 timeline** — 事件头部作者名替换
7. **集成到 event 详情** — 作者/被回复者替换
8. **test + build 验证**

## Technical Notes

* 现有 TUI 组件模式参考: `tui/thread/`, `tui/compose/`, `tui/event/`
* Profile name 解析: `utils.GetProfileName(ctx, pubkey, &utils.GetOptions{App: app})`
* 线程缓存: `globalNameCache` in `tui/thread/thread.go`
* BubbleTea 鼠标事件: `tea.MouseMsg`, `tea.MouseEvent` — 当前项目中无使用先例

## Research References

* [`research/existing-components.md`](research/existing-components.md) — 现有 TUI 组件模式、无鼠标交互前例、profile 解析流程、thread 缓存机制
