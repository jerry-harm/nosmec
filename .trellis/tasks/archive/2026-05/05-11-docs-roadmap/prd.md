# docs-update-and-roadmap-planning

## Goal

全面审查并更新 docs/ 下所有文档，修正过时内容；整理下一阶段 roadmap。

## What I already know

### 文档现状（扫描结果）

**docs/ARCHITECTURE.md（严重过时）:**
- 提到 LMDB（2026-05-10 已切换为 BoltDB）
- 提到 `config.ConfigManager`（不存在，AppContext 用的是 `viper *viper.Viper`）
- 项目结构列表过时（`cmd/note.go` → 实际是 `cmd/note_commands.go`；列出的文件与实际不符）

**docs/CONFIG.md（过时）:**
- 路径说明提到 LMDB：`~/.cache/nosmec/nosmec.db` → 应为 BoltDB
- 配置示例有 `server.host/port` 字段（Config 结构体中不存在）
- 函数文档 `config.AddRelay()` 等写的是 top-level 函数，实际在 AppContext 方法上
- 提到 `private_relays` 字段（Config 中不存在，删除了）
- CacheFilters 描述提到 LMDB store

**docs/RELAY.md（基本准确）:** 无重大问题。

**docs/DEV.md（过时）:**
- Go 版本要求写 1.23+，go.mod 是 1.25.5
- "DM functionality (NIP-17)" 标记为 Planned，但代码已实现
- 缺少 NIP-47 的规划

**docs/NIP.md（内容有误）:**
- NIP-04 标记为 Deprecated，但实际完全没有使用（代码用的是 NIP-17 GiftWrap）
- Migration Priority 表中 `utils/dm.go` 标记"should migrate to nip17"，但实际已经在用 nip59 GiftWrap
- nip40/nip42/nip57/nip45 等标记 ❌ Not used，但没说明是否需要使用

### 代码现状（对照文档）

- NIP-17 DM：已实现（`utils/dm.go` 用 nip59.GiftWrap/GiftUnwrap）
- NIP-46 Remote Signing：未实现 ✅ 确实是 Planned
- NIP-47 Nostr Wallet Connect：未实现 ✅ 确实是 Planned
- TUI timeline：不完整/需要重构 ✅ 与文档一致
- Search functionality：文档提到预留，实际代码中 `search_relays` 配置存在但没有搜索实现

## Assumptions

- 文档更新是纯文字工作，不需要代码改动
- roadmap 规划不需要立刻出 PRD，可以是结构化的 markdown 列表

## Requirements

### Part 1: 文档更新

- [ ] 更新 `docs/ARCHITECTURE.md`：
  - LMDB → BoltDB
  - 修正 `ConfigManager` → `AppContext.viper`
  - 更新项目结构列表（对照实际文件）
- [ ] 更新 `docs/CONFIG.md`：
  - LMDB → BoltDB
  - 删除 `server.host/port` 配置示例
  - 修正函数文档路径（top-level → AppContext 方法）
  - 删除 `private_relays` 字段
- [ ] 更新 `docs/DEV.md`：
  - Go 版本 → 1.25.5
  - DM functionality 标记为 ✅ 已实现
  - 补充 NIP-47 规划
- [ ] 更新 `docs/NIP.md`：
  - 移除 NIP-04 相关内容（或明确标记为不相关）
  - 修正 Migration Priority 表（当前实现状态）
  - 补充 nip45/nip53/nip54 等可用包说明
- [ ] 检查 `docs/README.md` 是否需要补充

### Part 2: Roadmap 整理

- [ ] 整理下一阶段功能优先级
- [ ] 为每个计划功能补充简要说明

## Acceptance Criteria

- [ ] 所有 docs/ 下文档内容与代码实现一致
- [ ] 无 LMDB/LMBD 相关引用
- [ ] go.mod 版本与 DEV.md 一致
- [ ] Roadmap 文档结构清晰、可执行

## Out of Scope

- 代码实现（仅限文档）
- 不修改根目录 README.md（已相对完整）

## Technical Notes

### 扫描发现的问题文件

| 文件 | 问题数 | 严重程度 |
|------|--------|----------|
| docs/ARCHITECTURE.md | 3 | 高 |
| docs/CONFIG.md | 4 | 中 |
| docs/DEV.md | 2 | 低 |
| docs/NIP.md | 3 | 中 |

### 实际文件结构（ARCHITECTURE.md 修正用）

```
cmd/
  ├── root.go, registry.go
  ├── note_commands.go, event_commands.go
  ├── config_commands.go, profile_commands.go
  ├── dm_commands.go, community_commands.go
  └── completion/

config/
  ├── config.go, types.go, relay.go
  ├── context.go, interfaces.go
  └── relay_test.go, config_test.go

utils/
  ├── get.go, post.go, profile.go
  ├── relay_list.go, subscription.go
  ├── dm.go, community.go
  ├── alias.go, show.go, sync.go
  ├── proxy.go, utils.go, user_relays.go

tui/
  ├── timeline/, window/, toolkit/
  └── cmd/, windowmanager/
```

## Decision (ADR-lite)

**Context**: 用户需要明确下一阶段的实现重点，并要求详细规划（每个功能走 brainstorm → prd）

**Decision**: 本次规划聚焦两个功能：
1. **DM 完善 + DM TUI** — 现有 DM 代码已实现（NIP-17 GiftWrap），但需要测试和完善 TUI 交互
2. **Search 功能** — `search_relays` 配置已存在，但搜索功能未实现

**Out of Scope（本次不规划）**:
- NIP-46 Remote Signing
- NIP-47 Nostr Wallet Connect
- TUI timeline 重构
- Event cache 管理
- Offline 模式

## Subtasks

- `05-11-dm-tui` — DM 完善 + DM TUI
- `05-11-search` — Search 功能实现

## Open Questions

- [x] 规划范围：✅ 已确认（DM+DM TUI, Search）
- [ ] DM TUI 范围：纯查看 DM 列表/历史，还是要支持发送？
- [ ] Search 实现方式：基于什么做搜索（本地 Bleve？远程搜索 API？）
