# brainstorm: 完善测试覆盖

## Goal

为 nosmec 项目补充测试覆盖，提升代码质量和稳定性。

## What I already know

**当前测试覆盖：**
- `config/config_test.go` — Relay 列表操作测试
- `config/relay_test.go` — Relay 相关测试
- 其余所有包：*no test files*

**项目模块（按代码行数）：**
| 模块 | 文件数 | 代码行数 | 测试状态 |
|------|--------|----------|----------|
| utils | 13 | ~2600 | 无测试 |
| config | 5 | ~500 | 有部分测试 |
| cmd | - | - | 无测试 |
| tui | - | - | 无测试 |

**最近新增但无测试的文件：**
- `utils/search.go` (149行) — NIP-50 搜索
- `utils/dm.go` (254行) — NIP-17 GiftWrap
- `tui/dm/` 目录 — DM TUI

## Assumptions (temporary)

1. 测试以单元测试为主，集成测试暂不优先
2. 使用 Go 标准 `testing` 包
3. mock 方式：使用接口 + mock 实现

## Open Questions

1. ~~**优先级**：应该优先测试哪些模块？~~ → 选择 A: 核心业务逻辑（utils）

## Requirements

**优先级 A：核心业务逻辑（utils）**

| 文件 | 行数 | 测试重点 |
|------|------|----------|
| `utils/get.go` | 434 | 查询逻辑、filter 构建 |
| `utils/dm.go` | 254 | SendDM, ListDMConversations, QueryDMHistory |
| `utils/search.go` | 149 | NIP-50 filter 解析、SearchEvents |
| `utils/post.go` | 149 | PostNote, ReplyToNote, QuoteNote |
| `utils/profile.go` | 407 | GetProfile, UpdateProfile |
| `utils/subscription.go` | 369 | Follow/Unfollow, subscription 管理 |

## Acceptance Criteria

- [x] `utils/search.go` 测试（NIP-50 filter 解析）— 15 passing tests
- [x] `utils/dm.go` 测试（DMMessage/Conversation, filter 构建, sorting）— 9 passing tests
- [x] `utils/post.go` 测试（tag 构建, event kind）— 7 passing tests
- [x] `go test ./...` 通过
- [ ] `utils/get.go` 测试（查询逻辑）— 未实现（需要网络 mock）
- [ ] `utils/profile.go` 测试 — 未实现（需要网络 mock）
- [ ] `utils/subscription.go` 测试 — 未实现（需要网络 mock）

## Definition of Done (team quality bar)

* 测试覆盖核心模块
* Lint / typecheck / 测试通过
* 无 regression

## Out of Scope (explicit)

* 集成测试（需要外部网络/relay）
* E2E 测试

## Technical Notes

* Go testing: `testing.T`, `testify/assert` (已有 testify mock)
* BoltDB mock: 可以用接口隔离
* No external network calls in tests