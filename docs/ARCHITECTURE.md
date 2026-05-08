# 架构文档

## 项目结构

```
nosmec/
├── cmd/                    # Cobra 命令定义
│   ├── root.go            # 根命令
│   ├── note.go            # 笔记命令
│   ├── relay.go           # Relay 管理命令
│   ├── profile.go         # Profile 命令
│   ├── community.go       # Community 命令
│   ├── alias.go           # 别名命令
│   └── dm.go              # DM 命令 (空，待实现)
│
├── config/                # 配置管理
│   ├── config.go         # Viper 初始化和管理函数
│   ├── types.go          # 配置结构体定义
│   └── relay.go          # Relay 相关配置和过滤器
│
├── utils/                 # 工具函数
│   ├── utils.go          # 通用工具 (GetMyPubKey, GetMySecretKey, ParsePubKey)
│   ├── post.go           # 发布函数 (PostNote, ReplyToNote, QuoteNote)
│   ├── get.go            # 查询函数 (GetEvent, GetNote, GetTimeline)
│   ├── profile.go         # Profile 相关
│   ├── community.go       # Community 相关
│   ├── relay.go          # Relay 服务器相关
│   ├── alias.go          # 别名相关
│   ├── completion.go      # Shell 补全
│   ├── show.go           # 显示函数
│   └── types.go          # 类型定义
│
├── docs/                  # 文档
│   ├── README.md         # 快速开始
│   ├── ARCHITECTURE.md   # 本文档
│   ├── RELAY.md          # Relay 管理详情
│   ├── CONFIG.md         # 配置管理详情
│   └── NIP.md           # 支持的 NIP 协议
│
├── go.mod
└── main.go
```

## 核心依赖

| Package | Version | 用途 |
|---------|---------|------|
| fiatjaf.com/nostr | v0.0.0-20260310013726-4e490879b558 | Nostr SDK |
| github.com/spf13/cobra | v1.10.2 | CLI 框架 |
| github.com/spf13/viper | v1.21.0 | 配置管理 |
| github.com/go-i2p/sam3 | v0.33.92 | I2P 支持 |
| github.com/fatih/color | v1.18.0 | 彩色输出 |

## 配置管理

使用 Viper 进行配置管理，支持：
- YAML 配置文件 (`~/.config/nosmec/nosmec.yaml`)
- 环境变量 (`NOSMEC_*` 前缀)
- 命令行默认值

### 初始化流程

1. `loadConfig()` 创建 viper 实例
2. 设置 config path: `~/.config/nosmec/`
3. 设置 env prefix: `NOSMEC`
4. 读取配置文件，不存在则创建默认配置
5. Unmarshal 到 `Config` 结构体

## 依赖注入

所有核心依赖通过 `AppContext` 管理：

```go
type AppContext struct {
    pool   PoolInterface      // Nostr 连接池
    store  StoreInterface    // LMDB 本地存储
    cfg    Config            // 配置副本
    config ConfigManager      // Viper 配置管理器
}
```

在 `main()` 或应用启动时创建 `AppContext`，通过构造函数注入依赖。

## 数据流

### 发布笔记流程

```
User Input
    ↓
cmd/note.go (command)
    ↓
utils/post.go (PostNote/ReplyToNote/QuoteNote)
    ↓
nostr.Event (构建事件，签名)
    ↓
app.Pool().PublishMany() (发送到多个 relay)
```

### 查询数据流程

```
User Input
    ↓
cmd/* (command)
    ↓
utils/get.go (GetEvent/GetNote/GetTimeline)
    ↓
1. 先查 LMDB 本地缓存
    ↓
2. 缓存无效则查询网络
app.Pool().QuerySingle/SubscribeMany()
    ↓
3. 结果存入 LMDB
app.Store().SaveEvent()
```

## Relay 管理

### 配置结构

```go
type Relay struct {
    URL   string `mapstructure:"url"`
    Read  *bool  `mapstructure:"read,omitempty"`
    Write *bool  `mapstructure:"write,omitempty"`
}

type Config struct {
    RelayList   []Relay  // Read/Write relay 列表
    DMRelays    []string // DM relay 列表 (Kind 10050)
    SearchRelays []string // Search relay 列表
}
```

### NIP-65 Relay List

Kind 10002 事件格式：
```json
{
  "kind": 10002,
  "tags": [
    ["r", "wss://relay1.com", "read", "write"],
    ["r", "wss://relay2.com", "read"]
  ]
}
```

### NIP-17 DM Relay List

Kind 10050 事件格式：
```json
{
  "kind": 10050,
  "tags": [
    ["relay", "wss://dm-relay1.com"],
    ["relay", "wss://dm-relay2.com"]
  ]
}
```

## 数据库

使用 LMDB (Lightning Memory-Mapped Database) 作为本地存储：

- 路径: `~/.cache/nosmec/nosmec.db`
- 存储内容: 事件、Relay 信息缓存

## I2P 支持

通过 `github.com/go-i2p/sam3` 支持 I2P 网络：

- 配置: `app.Config().LocalServer.I2P`
- SAM 地址: 默认 `127.0.0.1:7656`
- 代理: `app.Config().Proxy.I2PSocks`
