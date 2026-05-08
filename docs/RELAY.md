# Relay 管理

## 概述

nosmec 支持三类 Relay：

| 类型 | 用途 | NIP | Kind |
|------|------|-----|------|
| Read/Write Relay | 通用读写 | NIP-65 | 10002 |
| DM Relay | 私信专用 | NIP-17 | 10050 |
| Search Relay | 搜索（预留） | - | - |

## 命令

### 基础命令

```bash
# 列出所有 relay
nosmec relay list

# 添加 read/write relay（默认 read=true, write=true）
nosmec relay add wss://relay.example.com

# 添加只读 relay
nosmec relay add wss://readonly.example.com --read=true --write=false

# 移除 relay
nosmec relay remove wss://relay.example.com

# 设置 relay 属性
nosmec relay set wss://relay.example.com --read=false --write=true
```

### 发布与同步

```bash
# 发布本地 relay list 到网络 (Kind 10002)
nosmec relay publish

# 同步 relay list (Kind 10002 + Kind 10050)
nosmec relay sync

# 获取他人 relay list
nosmec relay fetch <npub-or-hex>
```

### DM Relay

```bash
# 添加 DM relay
nosmec relay dm add wss://dm-relay.example.com

# 列出 DM relay
nosmec relay dm list

# 移除 DM relay
nosmec relay dm remove wss://dm-relay.example.com

# 发布 DM relay list (Kind 10050)
nosmec relay dm publish
```

### Search Relay

```bash
# 添加 search relay
nosmec relay search add wss://search.example.com

# 列出 search relay
nosmec relay search list

# 移除 search relay
nosmec relay search remove wss://search.example.com
```

## 配置结构

```yaml
relay_list:
  - url: wss://relay1.example.com
    read: true
    write: true
  - url: wss://relay2.example.com
    read: true
    write: false

dm_relays:
  - wss://dm-relay.example.com

search_relays:
  - wss://search.example.com
```

## NIP-65 协议

NIP-65 定义了读写 relay 列表的发布和发现机制：

### 读写标记

- `read`: 允许从这个 relay 读取事件
- `write`: 允许向这个 relay 发送事件
- 无标记: 等同于 read + write

### Inbox/Outbox 模型

根据 NIP-A4 (Public Messages):

- **Inbox relays**: 你的 read relays（接收别人发给你的消息）
- **Outbox relays**: 你的 write relays（你发送消息的 relay）

### 隐私考虑

- Read relay 可能暴露你关注的内容
- Write relay 可能暴露你发布的内容
- 建议使用不同 relay 做不同用途

## NIP-17 DM Relay List

NIP-17 定义了 DM 专用 relay 列表：

### 发布

Kind 10050 事件，`relay` 标签包含 DM relay URL

### 发现

1. 查找接收者的 Kind 10050 事件
2. 使用其中的 relay 发送加密 DM
3. 如果没有找到，说明接收者不支持 NIP-17

### 隐私

- DM relay 可能被用于关联收发双方
- 建议使用与公开 relay 不同的 DM relay

## 常见问题

### Q: relay add 默认值是什么？

A: `read=true, write=true`

### Q: relay sync 会覆盖本地配置吗？

A: 会。远程 relay list 会完全覆盖本地配置。

### Q: 如何只同步 DM relay？

A: `nosmec relay sync` 会同时同步 Kind 10002 和 Kind 10050。没有单独的 dm sync 命令。

### Q: relay list 为空会发生什么？

A: `relay publish` 会报错，需要至少有一个 relay 才能发布。
