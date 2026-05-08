# 配置管理

## 配置文件

位置: `~/.config/nosmec/nosmec.yaml`

## 配置结构

```yaml
server:
  host: localhost
  port: 8080

i2p:
  enabled: true
  sam_address: 127.0.0.1
  sam_port: 7656

private_key: ""  # nsec 格式

relay_list: []    # Read/Write relay 列表

dm_relays: []    # DM relay 列表

search_relays: [] # Search relay 列表

known_relays: [] # 已知 relay 列表（程序结束时自动维护）

private_relays: []  # 私有中继列表（用于自动缓存）

cache_filters: [] # 缓存过滤器列表，默认动态生成

local_relay:
  enabled: true   # 本地 relay 开关
  port: "8989"   # 本地 relay 端口
  data_dir: ~/.cache/nosmec  # 数据目录

proxy:
  i2p_socks: ""
  onion_socks: ""
  socks: ""

alias: {}  # 别名映射
```

## 环境变量

所有配置支持 `NOSMEC_` 前缀的环境变量覆盖：

| 配置项 | 环境变量 | 说明 |
|--------|----------|------|
| `private_key` | `NOSMEC_PRIVATE_KEY` | 私钥 (nsec 格式) |
| `relay_list` | `NOSMEC_RELAY_LIST` | Relay 列表 |
| `dm_relays` | `NOSMEC_DM_RELAYS` | DM relay 列表 |
| `known_relays` | `NOSMEC_KNOWN_RELAYS` | 已知 relay |
| `private_relays` | `NOSMEC_PRIVATE_RELAYS` | 私有中继列表 |
| `local_relay.enabled` | `NOSMEC_LOCAL_RELAY_ENABLED` | 本地 relay 开关 |
| `local_relay.port` | `NOSMEC_LOCAL_RELAY_PORT` | 本地 relay 端口 |
| `server.host` | `NOSMEC_SERVER_HOST` | 服务器地址 |
| `server.port` | `NOSMEC_SERVER_PORT` | 服务器端口 |
| `proxy.i2p_socks` | `NOSMEC_PROXY_I2P_SOCKS` | I2P 代理 |
| `proxy.onion_socks` | `NOSMEC_PROXY_ONION_SOCKS` | Tor 代理 |
| `proxy.socks` | `NOSMEC_PROXY_SOCKS` | 通用 SOCKS 代理 |

## 配置管理函数

在 `config/config.go` 中定义：

### Relay 管理

```go
config.AddRelay(url string, read, write bool) error
config.RemoveRelay(url string) error
config.GetRelay(url string) (Relay, bool)
config.ListRelays() []Relay
config.SetRelayRead(url string, read bool) error
config.SetRelayWrite(url string, write bool) error
```

### DM/Search Relay 管理

```go
config.AddDMRelay(url string) error
config.RemoveDMRelay(url string) error
config.ListDMRelays() []string

config.AddSearchRelay(url string) error
config.RemoveSearchRelay(url string) error
config.ListSearchRelays() []string
```

### 同步

```go
config.SyncRelayList(relays []Relay)    // 从远程同步
config.SyncDMRelays(relays []string)     // 从远程同步 DM relay
```

### 过滤

```go
app.WritableRelays() []string  // 获取可写 relay（通过 AppContext）
app.ReadableRelays() []string   // 获取可读 relay（通过 AppContext）
```

### 缓存过滤器 (CacheFilters)

`CacheFilters` 用于指定哪些事件应该被缓存到本地 LMDB store。如果不设置，程序会自动生成默认过滤器。

```yaml
cache_filters:
  - kinds: [0, 3, 10002, 10050]  # 缓存特定 kind 的所有事件
  - kinds: []                     # 缓存用户自己的所有事件
    authors:
      - "pubkey-hex-of-self"
```

**默认行为：**
- 如果配置中完全没有 `cache_filters` 字段，程序会自动生成两个过滤器：
  1. 缓存特定 kind（0, 3, 10002, 10050）的所有事件
  2. 缓存当前用户的所有事件

**Filter 结构（来自 nostr 库）：**
```go
type Filter struct {
    IDs     []ID
    Kinds   []Kind
    Authors []PubKey
    Tags    TagMap
    Since   Timestamp
    Until   Timestamp
    Limit   int
}
```

**匹配规则（来自 nostr.Filter.Matches）：**
- `Kinds: nil` 或空 slice → 匹配所有 kind
- `Authors: nil` → 不限制作者；非 nil 空 slice `[]` → 不匹配任何事件
- 如果设定了 `Kinds` 或 `Authors`，则只匹配满足条件的事件

**示例 - 只缓存特定作者的所有事件：**
```yaml
cache_filters:
  - kinds: []           # 所有 kind
    authors:
      - "hex-pubkey-1"
      - "hex-pubkey-2"
```

**示例 - 缓存特定 kind 的特定作者：**
```yaml
cache_filters:
  - kinds: [1, 7]      # TextNote 和 Reaction
    authors:
      - "target-author-pubkey"
```

## 数据结构

### Relay

```go
type Relay struct {
    URL   string `mapstructure:"url"`
    Read  *bool  `mapstructure:"read,omitempty"`
    Write *bool  `mapstructure:"write,omitempty"`
}
```

使用 `*bool` 而非 `bool` 以支持三态：
- `nil`: 未设置（默认 read=true, write=false）
- `true`: 已启用
- `false`: 已禁用

### Config

```go
type Config struct {
    ConfigDir      string
    RelayList      []Relay
    DMRelays       []string
    SearchRelays   []string
    PrivateKey     string
    KnownRelays    []string
    PrivateRelays  []string
    LocalRelay     LocalRelayConfig
    Proxy          ProxyConfig
    Alias          map[string]string
}

type LocalRelayConfig struct {
    Enabled bool
    Port    string
    DataDir string
}
```

## 实现细节

### Viper 配置

```go
viper.SetConfigName("nosmec")
viper.SetConfigType("yaml")
viper.AddConfigPath(configDir)      // ~/.config/nosmec/
viper.AddConfigPath("$HOME/.config")
viper.AddConfigPath("./")
viper.AddConfigPath(".")

viper.SetEnvPrefix("NOSMEC")
viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
viper.AutomaticEnv()
```

### 配置文件自动创建

如果配置文件不存在，程序会自动创建：

```go
configFile := filepath.Join(configDir, "nosmec.yaml")
if err := viper.WriteConfigAs(configFile); err != nil {
    log.Printf("Warning: Could not write config file: %v", err)
}
```

## 路径说明

| 路径 | 说明 |
|------|------|
| `~/.config/nosmec/nosmec.yaml` | 配置文件 |
| `~/.cache/nosmec/nosmec.db` | LMDB 数据库 |
