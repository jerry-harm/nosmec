# compose 发送反馈 UI

## Goal

在 compose 界面实现发送状态反馈：ctrl+p 按下后显示"发送中..."提示，发送完成后显示"发送成功"然后自动退出。

## What I already know

* compose 在 `tui/compose/model.go`
* ctrl+p 绑定在 `m.keys.send`（line 76-78）
* `m.sending` 字段已存在（line 61）但目前没有对应的 UI 反馈
* `sendContent()` 是异步执行的（返回 `tea.Cmd`），发送中时 `Update` 里会 `return m, nil` 忽略按键（line 259-261）
* `sendSuccessMsg` 设置 `m.success = true` 后调用 `bubblon.Close()` 自动关闭（line 250-254）
* `sendErrorMsg` 设置 `m.errMsg`（line 245-248）
* renderView 在 line 545-548 显示 `m.errMsg`，line 550-553 显示 `m.success` 成功消息

## Assumptions (temporary)

* 发送中需要显示 "Sending..." 或类似提示
* 发送中应该禁用输入框（视觉上已通过 `m.sending` 屏蔽按键，但 UI 上没有提示）
* 发送成功后显示 "Posted successfully!" 然后自动退出

## Open Questions

* 发送中提示应该放在哪里？（现有字段标签区域 / 内容区域上方 / 单独一行）
* 发送中是否需要显示进度（如 relay 数量）？
* 发送失败后是否自动退出，还是留在界面让用户重试？

## Requirements (evolving)

* ctrl+p 按下后立即显示全屏遮罩（overlay）显示 "Sending..."
* 异步执行发送，不阻塞 UI
* 发送完成后显示 "Posted successfully!" 停留 1-2 秒，然后自动关闭 compose 窗口并清空内容
* 发送失败显示错误消息，遮罩消失，用户可重试或退出

## Acceptance Criteria

* [ ] 按 ctrl+p 后立即显示全屏遮罩 "Sending..."
* [ ] 发送成功后在遮罩上显示 "Posted successfully!" 停留 1.5 秒，然后自动关闭 compose 并清空
* [ ] 发送失败在遮罩上显示错误消息，遮罩消失，compose 恢复编辑状态
* [ ] 发送过程中 backspace/ctrl+c 无法取消发送（保持简单）

## Definition of Done

* 单元测试（如有）
* Lint / typecheck 通过

## Out of Scope

* 发送进度（relay 数量）
* 多个 relay 时的差异化反馈

## Technical Notes

* `m.sending` 在 `Update` 里已用于屏蔽按键，但 renderView 没有对应的状态显示
* 需要在 `renderView()` 里增加 sending 状态的 UI 渲染
* 发送状态在 `sendContent()` 开始时应该设置 `m.sending = true`（目前没有设置）