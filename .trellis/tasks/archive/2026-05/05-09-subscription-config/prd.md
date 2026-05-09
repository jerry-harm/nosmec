# brainstorm: subscription-config

## Goal

让订阅功能更好地与配置文件集成，并在 `profile` 命令中显示 follow 列表。

## What I already know

- 订阅数据已通过 Viper 持久化到 `nosmec.yaml`（`subscriptions` 字段）
- `AppContext` 已有完整的订阅管理方法
- `profile --full` 目前只显示 nostr profile 事件，不包含本地订阅信息
- CLI 命令 `config subscribe add/remove/list/sync/publish` 都已实现

## Assumptions (temporary)

- 用户希望在 `profile` 命令输出中看到自己 follow 的用户/社区/话题列表

## Open Questions

- `profile` 命令显示订阅列表时，格式偏好？（纯文本列表 / JSON 结构化输出 / 其他）

## Requirements (evolving)

- [ ] `profile --full` 完整显示 follow list（从网络读取）
  - Kind 3: 用户关注 (follows)
  - Kind 10004: 社区列表 (communities)
  - Kind 10015: 话题列表 (hashtags)

## Acceptance Criteria (evolving)

- [ ] `profile --full` 输出包含 follows、communities、hashtags 三个分类
- [ ] 每个分类从对应的 nostr event kind 读取

## Definition of Done (team quality bar)

- Tests added/updated
- Lint / typecheck / CI green

## Out of Scope (explicit)

- 订阅的网络同步功能改动

## Technical Notes

- 相关文件：
  - `config/types.go` - Subscription struct
  - `config/context.go` - AppContext 订阅管理方法
  - `utils/subscription.go` - 订阅工具函数
  - `cmd/profile_commands.go` - profile 命令