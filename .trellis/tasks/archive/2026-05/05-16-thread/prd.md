# 测试 thread 功能是否正常工作

## Goal

验证 thread treeview 功能在真实环境（relay 网络）下能否正常工作：正确识别 root、拉取回复、渲染树结构、键盘导航可用。

## What I already know

- Thread 实现位于 `tui/window/event/thread_treeview.go` (445 行)
- 使用 `treeview/v2` TuiTreeModel 做树形渲染和键盘导航
- 入口：Timeline → Enter 进入 EventView → 按 `t` 打开 thread
- fetch 策略：先按 relay hint 查 root，再查 direct replies，中途缺父节点用 placeholder
- 已有测试仅覆盖 NIP-10 解析函数（extractParentID, extractRootEvent, NostrEventProvider）
- **模型本身（Init/Update/View/fetch/build）零测试**

## Requirements

* 验证 thread 启动后的 fetch 流程正常（root 取到、replies 取到）
* 验证 TuiTreeModel 渲染正确（树形结构，非平铺列表）
* 验证键盘导航可用（↑↓ 切换节点，←→ 展开折叠）
* 验证 esc 关闭回到 EventView（不退出整个 app）
* 验证 placeholder 逻辑（缺少父节点时不 panic）

## Acceptance Criteria

* [ ] PTY 黑盒测试：启动 TUI → 打开某事件 → 按 `t` → thread 窗口出现
* [ ] 线程树正确渲染（root 在顶部，replies 为子节点）
* [ ] 键盘 ↑↓ 可切换选中节点
* [ ] 键盘 ←→ 可展开/折叠子节点
* [ ] esc 关闭 thread 回到 EventView
* [ ] 追加单元测试：fetch/build/Update 关键路径覆盖

## Definition of Done

* PTY 黑盒测试通过（exit 0，输出包含预期内容）
* 新增单元测试通过
* `go build ./...` + `go test ./...` 全绿

## Decision (ADR-lite)

**Context**: 模型层零测试，需验证 thread 正常工作。
**Decision**: 先补单元测试（fetch/build/Update 关键路径），再 PTY 黑盒测试（端到端验证）。
**Consequences**: 单元测试先行可快速验证逻辑正确性；PTY 补充真实环境验收。

## Out of Scope

* Reply 写入功能（本次只测展示）
* 多层嵌套回复（单层测试即可）
* UI 样式微调

## Technical Notes

- 研究参考: `research/thread-implementation.md`
- 关键文件: `tui/window/event/thread_treeview.go`, `tui/window/event/thread_treeview_test.go`
- 入口: `tui/window/event/event.go:261-270` (thread 方法), `tui/timeline/delegate.go` (Enter 打开 EventView)
- 已有 bug: placeholder 永不 resolve；dead code (`thread.go`)；root 解析在 treeview 和 utils/get.go 逻辑分歧
