# 添加黑箱 PTY 测试流程约束

## Goal

在开发流程中添加黑箱测试要求: 使用 pty_spawn/pty_read 进行黑箱测试,检查命令实际输出效果.

## What I already know

**opencode-pty 插件**:
- `pty_spawn` - 创建 PTY session
- `pty_read` - 读取输出缓冲
- `pty_write` - 发送输入
- `pty_kill` - 终止 session
- Web UI: http://[::1]:37593

**使用场景**:
- CLI 命令测试 (nosmec gossip, nosmec event, etc.)
- TUI 程序测试 (timeline, compose, thread view)
- 交互式程序测试

## Requirements

* 所有新功能在验收时需要通过 PTY 黑箱测试
* 使用 `pty_spawn` 运行程序,用 `pty_read` 检查输出
* 测试命令: `nosmec gossip`, thread view 等

## Acceptance Criteria

* [ ] 新功能添加 PTY 黑箱测试
* [ ] 测试检查实际输出,不只是代码逻辑

## Out of Scope

* 自动化 PTY 测试框架
* 回归测试套件

## Technical Approach

1. 新功能实现后,用 PTY 启动程序验证
2. 检查输出内容是否正确
3. 例如: `nosmec gossip` 运行后检查 "Discovered X relays" 输出