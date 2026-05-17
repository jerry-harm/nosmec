# 重新审视 NIP-10 回复逻辑和 tag 格式

## Goal

基于 NIP-10/NIP-22 规范仔细审查 nostr 回复的 tag 格式定义，理清 thread 显示的完整逻辑链，修复已知 bug，完善测试覆盖。

## Research References

- [`research/nip10-spec.md`](research/nip10-spec.md) — NIP-10 完整 e tag 格式：positions 0-4，root/reply/mention 标记语义
- [`research/nip22-comments.md`](research/nip22-comments.md) — NIP-22 kind:1111 评论的 E/A/I + K/k 双层大小写 tag 系统
- [`research/client-thread-patterns.md`](research/client-thread-patterns.md) — 主流 Nostr 客户端线程处理模式对比

## What I already know (from research)

### NIP-10 e tag 格式（已确认）

```
["e", <id>, <relay>, <marker>, <pubkey>]
```

| 场景 | tag 结构 | parent ID | root ID |
|------|---------|-----------|---------|
| 根事件 | 无 e tag | — | self |
| 直接回复 | `["e", root, relay, "root"]` | **root** | root |
| 嵌套回复 | `["e", root, relay, "root"]` + `["e", parent, relay, "reply"]` | reply tag 值 | root tag 值 |

**结论：之前的 fix 方向正确** — `extractParentID` 的 root fallback、`extractRootEvent` 的 "root!=self" 判断都是对的。

### NIP-22 kind:1111 评论（新发现）

| 方面 | NIP-10 (kind:1) | NIP-22 (kind:1111) |
|------|----------------|---------------------|
| root 标识 | e tag + "root" marker | `E`/`A`/`I` 大写 tag + `K` tag |
| parent 标识 | e tag + "reply" marker | `e`/`a`/`i` 小写 tag + `k` tag |
| 跨种类 | 禁止 | 专为此设计 |

**当前代码的 bug**: `fetchRepliesToRoot` 包含了 kind:1111，但 `extractParentID`/`extractRootEvent` 只查 e tag marker，不识别 NIP-22 的大小写标签系统。kind:1111 的父子关系无法被正确解析。

### 线程 fetch 策略

nosmec 使用 **Pattern 1**（单次 #e 查询）：只取直接回复 root 的事件，不递归拉孙子节点。这是 PRD 明确的范围（"1 level of direct replies"）。

### 重复逻辑

- `thread_treeview.go`: `extractRootEvent()` + `extractParentID()`
- `utils/get.go`: `FindRootEvent()` (old counterpart with divergent behavior)

## Requirements

* [ ] 巩固 NIP-10 解析（已大部分完成：extractParentID root fallback + extractRootEvent direct reply 判断）
* [ ] **决策：如何处理 kind:1111 评论？** 选项 A: 忽略（从 filter 移除 KindComment） / 选项 B: 支持解析（扩展 extractParentID/extractRootEvent 识别 NIP-22 标签）
* [ ] **决策：thread 深度？** 选项 A: 保持 1 层 / 选项 B: 支持懒加载展开子节点
* [ ] 消除 `extractRootEvent` vs `FindRootEvent` 重复逻辑
* [ ] 完善单元测试：NIP-10 全场景 + NIP-22（如选定）+ 边界条件
* [ ] 补充 PTY 黑盒测试

## Acceptance Criteria

* [ ] 所有 NIP-10 场景（root/direct/nested/mention/legacy positional）有单元测试
* [ ] `go build ./...` + `go test ./...` 通过
* [ ] PTY 验证 thread 能正常加载并显示正确的树结构

## Decision (ADR-lite)

**Context**: kind:1 线程显示只取 1 层直接回复，嵌套回复在树中缺失。
**Decision**: 实现递归多层 fetch（`fetchThreadTree`），深度上限 10。kind:1111 community 后续复用。
**Consequences**: 每层一次 relay 查询，树结构会更完整。

## Out of Scope

* 多层嵌套回复的自动 fetch（保持 Pattern 1 单层）

## Technical Notes

- 关键文件: `tui/window/event/thread_treeview.go`, `utils/get.go` (FindRootEvent)
- 已修复 (上轮): extractParentID 的 root fallback, extractRootEvent 的 direct reply 判断
- 未修复: NIP-22 解析, 重复 root 逻辑
