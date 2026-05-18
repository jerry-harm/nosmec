# 文档更新与约束整理

## Goal

整理项目文档（README、docs/）、消除重复内容、将重要规则和反模式固化到约束文件。

## What I Already Know

### 过时的文档

| 文件 | 问题 |
|------|------|
| `README.md:142-184` | Project Structure 列出的文件名过时：`note.go`/`relay.go`/`subscribe.go` → 实际是 `note_commands.go`/`event_commands.go`/`gossip_commands.go`/`search_commands.go`/`registry.go`；TUI 结构写 `timeline/`+`common/` → 实际是 `timeline/`/`compose/`/`dm/`/`thread/`/`event/`/`community/`/`bubblon/` |
| `docs/ARCHITECTURE.md:5-16` | 命令文件名同上，缺少 `event_commands.go`/`gossip_commands.go`/`search_commands.go`/`config_commands.go`；且提到 `subscription_commands.go`（不存在） |
| `docs/DEV.md:38-46` | "TUI timeline broken" 已过时（TUI 已完整实现）；"NIP-50 search not implemented" 已实现（`search_commands.go`） |

### 重复内容

| 来源 | 重复目标 | 说明 |
|------|---------|------|
| `docs/DESIGN_PRINCIPLES.md` | `.trellis/spec/backend/` | 架构决策、NIP 表、反模式、查询模式、event 缓存策略 — 内容与 spec 文件大量重叠 |
| `quality-guidelines.md` 的 `copy()` 反模式 | `nostr-sdk-usage.md`、设计原则 | `nostr.IDFromHex` 规则在多处重复 |
| `relay-guidelines.md` NIP-10 部分 | `docs/NIP.md` | 都有 NIP-10 描述，但侧重点不同（spec 偏实现，docs 偏协议说明） |

### 重要规则（应固化到约束）

| 规则 | 当前位置 | 建议归属 |
|------|---------|---------|
| NIP-19 输出约定 | `quality-guidelines.md`、DESIGN_PRINCIPLES | `backend/nip-conventions.md`（新建） |
| `copy()` 反模式（导致 ID/PK 解析失败） | `quality-guidelines.md`、DESIGN_PRINCIPLES | → `error-handling.md` 或 `quality-guidelines.md`（已有） |
| NIP spec 阅读规则 | `quality-guidelines.md` | `quality-guidelines.md`（已有） |
| TUI key 事件（`tea.KeyPressMsg` vs `tea.KeyMsg`） | `quality-guidelines.md` | `quality-guidelines.md`（已有） |
| BubbleTea `View() tea.View` 返回类型 | `quality-guidelines.md` | `quality-guidelines.md`（已有） |
| NIP-10 5-field e tag 格式 + event→relay 追踪 | `relay-guidelines.md` | `relay-guidelines.md`（已有，刚更新） |

### docs/ 下的文件

```
docs/
├── ARCHITECTURE.md   # 架构图（过时）
├── CONFIG.md         # 未审阅
├── DESIGN_PRINCIPLES.md  # 与 spec 重复
├── DEV.md            # 项目状态（过时）
├── NIP.md            # NIP 参考（较全，但 NIP-10 部分需更新为 5-field）
└── README.md         # 主读我（过时）
```

## Assumptions

- `docs/` 下的文档是为了给人类开发者阅读的快速参考
- `.trellis/spec/` 是给 AI 辅助的代码规范，AI 每次实现前会加载
- 设计原则文档（DESIGN_PRINCIPLES.md）可以保留作为高级架构决策的单一真相源，但需要与 spec 同步

## Open Questions

1. **DESIGN_PRINCIPLES.md 保留还是合并？** 是保留为独立的架构决策文档，还是把内容拆分到 spec 目录下然后删除？
2. **约束文件命名** — 是否建立专门的 `conventions.md` 或 `nip-conventions.md`，还是把 NIP 相关规则保持在现有文件中？
3. **docs/ARCHITECTURE.md** — 是精简更新还是删除（README.md 已经描述了项目结构）？
4. **docs/NIP.md** — NIP-10 部分是否更新为 5-field format 描述？

## Requirements

### A. 文档更新

1. 更新 `README.md` 的 Project Structure（匹配实际 cmd 文件名 + tui 子目录）
2. 更新 `docs/DEV.md`（TUI 已完成、NIP-50 已实现）
3. 更新 `docs/NIP.md` 的 NIP-10 部分（5-field format: `["e", id, relay, marker, pubkey]`）
4. 删除 `docs/ARCHITECTURE.md`（内容与 README 重复）

### B. DESIGN_PRINCIPLES 内容拆分到 spec

把 `docs/DESIGN_PRINCIPLES.md` 的内容按 topic 分散到 spec 对应文件中：

| 来源内容 | 目标 spec 文件 |
|---------|--------------|
| 模块目录 (cmd/config/utils/tui/logger) | `backend/directory-structure.md` |
| AppContext DI 说明 | `backend/app-context.md`（新建） |
| CLI 命令分离 (note/event/profile/community/dm) | `backend/directory-structure.md` |
| 统一超时 (`QueryTimeout()`, `context.WithTimeout`) | `backend/query-patterns.md`（新建） |
| BoltDB + Bleve 存储、UserRelayList bucket | `backend/database-guidelines.md` |
| Query patterns (sync/async/streaming/replaceable) | `backend/query-patterns.md` |
| Error handling (`handleError()`, error return 规范) | `backend/error-handling.md` |
| Hardcoded timeout 反模式 | `backend/quality-guidelines.md` |
| File naming conventions | `backend/directory-structure.md` |
| NIP-19 输出约定 | `backend/nip-conventions.md`（新建，集中 NIP 相关约定） |

### C. 删除 docs/ARCHITECTURE.md + 更新相关引用

- 删除 `docs/ARCHITECTURE.md`
- `docs/DESIGN_PRINCIPLES.md` 本身也删除（内容已拆分到 spec）

### D. 确认关键规则已在 spec

（已确认无需新增）

- `copy()` 反模式 → `quality-guidelines.md` ✅
- NIP spec 阅读规则 → `quality-guidelines.md` ✅
- TUI key 事件 / View 返回类型 → `quality-guidelines.md` ✅
- NIP-10 5-field + event→relay 追踪 → `relay-guidelines.md` ✅

## Acceptance Criteria

- [x] README.md Project Structure 与实际文件列表一致
- [x] docs/DEV.md 反映当前 TUI 实现状态（timeline/compose/dm/thread 完成，NIP-50 实现）
- [x] docs/NIP.md NIP-10 部分描述 5-field format
- [x] `docs/ARCHITECTURE.md` 已删除
- [x] `docs/DESIGN_PRINCIPLES.md` 已删除（内容已拆分到 spec）
- [x] `backend/directory-structure.md` 已填写（实际目录 + 命令命名）
- [x] `backend/database-guidelines.md` 已填写（BoltDB path + bucket 结构）
- [x] `backend/error-handling.md` 已填写（handleError + error return）
- [x] `backend/app-context.md` 已新建并填写
- [x] `backend/query-patterns.md` 已新建并填写
- [x] `backend/nip-conventions.md` 已新建并填写
- [x] `backend/index.md` 已更新（Quick Reference 表格）

## Out of Scope

- **不修** relay-guidelines 的差异（留待后续大重构）
- 不改代码，只改文档
- 不新建大量文件（一个 topic 一个文件）