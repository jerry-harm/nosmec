# DM TUI

## Goal

实现完整的 DM TUI 界面：`dm <npub>` 直接打开与指定用户的加密私信对话，支持查看历史消息和发送新消息。

## What I already know

### 现有实现

**utils/dm.go（已有）:**
- `SendDM(ctx, app, recipientPubKey, content)` — 发送加密 DM（NIP-17 GiftWrap）
- `ListDMConversations(ctx, app, limit)` — 获取 DM 会话列表（按最新时间排序）
- `QueryDMHistory(ctx, app, recipientPubKey, limit)` — 查询与某人的 DM 历史

**utils/dm.go 的实现细节:**
- 用 `nip59.GiftWrap/GiftUnwrap` 处理加密
- DM relay list（NIP-17）优先，找不到则用 read relays
- 消息通过 `nip17.PublishMessage` 发送到双方 relay
- 返回 `[]DMMessage{Content, FromMe, Timestamp}`

**CLI 命令（已有）:**
- `dm list` — 列出所有 DM 会话
- `dm send <npub> <msg>` — 发送 DM
- `dm recv` — 接收 DMs（轮询）

### 参考实现

Bubbletea chat example (`/home/jerry/code/bubbletea/examples/chat/main.go`):
- `viewport.Model` 显示消息历史
- `textarea.Model` 编写消息
- `Enter` 发送消息
- `ctrl+c/esc` 退出
- `Viewport + Textarea` 垂直布局

## Requirements

### 功能

- [ ] `dm <npub>` 命令打开 DM TUI（不经过会话列表）
- [ ] TUI 布局：上方 viewport（消息历史）+ 下方 textarea（编写区）
- [ ] 消息显示：`npub` 格式化显示（npub1xxx）、时间戳、内容
- [ ] 自己发的消息和对方的消息用不同样式区分
- [ ] 发送：回车发送，等网络确认后消息出现在历史中（本地 relay 加速）
- [ ] 退出：`ctrl+c` 或 `esc` 返回 CLI

### 消息显示格式

```
[2026-05-11 14:30] npub1abc...def: 你好，这是一条消息
[2026-05-11 14:31] npub1xyz...uvw: 收到！这是回复
```

自己发的消息用不同颜色/样式区分。

### 数据流

```
用户输入消息
    ↓
SendDM() → nip17.PublishMessage → 双方 relay
    ↓
网络确认成功
    ↓
QueryDMHistory() 刷新历史
    ↓
viewport 更新显示
```

### 错误处理

- 发送失败：显示错误提示，消息保留在 textarea 中供重发
- 网络超时：使用 `app.QueryTimeout()`
- 对方无 DM relay：显示警告，继续尝试 read relay

## Out of Scope

- 独立的 DM 会话列表 TUI（`dm list` CLI 够用）
- DM 通知/实时推送（`dm recv` 轮询够用）
- 加密附件/媒体

## Technical Approach

### 新文件

```
tui/dm/
  ├── model.go       # DM TUI model (viewport + textarea)
  ├── view.go        # 渲染逻辑
  ├── styles.go      # lipgloss 样式
  └── main.go        # RunDM(app, npub) 入口
```

### 命令改动

`cmd/dm_commands.go` 或新建 `cmd/dm_tui.go`：
```go
dmCmd.AddCommand(&cobra.Command{
    Use:   "dm <npub>",
    Short: "Open DM chat with user",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        npub := args[0]
        // 解析 npub → PubKey
        // 打开 TUI
    },
})
```

### 关键依赖

- `charm.land/bubbles/v2/textarea` — 文本输入
- `charm.land/bubbles/v2/viewport` — 消息滚动
- `charm.land/lipgloss/v2` — 样式
- `fiatjaf.com/nostr/nip17` — DM 发送（NIP-17）

## Acceptance Criteria

- [ ] `dm <npub>` 能正确解析 npub 并打开 TUI
- [ ] 历史消息正确显示（自己的/对方的样式区分）
- [ ] 发送消息后等网络确认，确认后消息出现在历史中
- [ ] `ctrl+c/esc` 能正确退出
- [ ] 发送失败时错误提示清晰，消息不丢失
