# TDD发现：parseTagInput separator 问题

## 问题描述

`parseTagInput` 使用 `strings.Split(input, ":")` 导致 URL 中的 `:` 被错误分割。

**示例**：
- 输入 `r:wss://relay.example` → split 后变成 `["r", "wss", "", "//relay.example"]`
- 期望：按**第一个** `:` 分割，得到 `["r", "wss://relay.example"]`

## NIP-01 Tag 格式

NIP-01 定义 tag 是数组：
```json
["e", "<event-id>", "<relay-url>", "<pubkey>"]
["p", "<pubkey>", "<relay-url>"]
["a", "<kind>:<pubkey>:<d>", "<relay-url>"]
```

Tag value（如 relay URL）是**独立数组元素**，不和其他值合并。

## 根因

`utils/tui/compose/model.go:539`:
```go
parts := strings.Split(input, ":")
```

应该用 `strings.SplitN(input, ":", 2)` 只按第一个 `:` 分割。

## 两个修复方案

### 方案 A：修复 split 逻辑（保守修复）
```go
parts := strings.SplitN(input, ":", 2)
if len(parts) < 2 {
    return TagValue{Type: "t", Values: []string{input}}
}
```

优点：最小改动，向后兼容。
缺点：语义上仍然是 `type:value`，只是限制 split 次数。

### 方案 B：换 separator 字符（破坏性修复）
改用其他字符如 ` `（空格）或 `;` 作为 separator。

优点：更符合直觉（tag 格式 `e event-id relay` 用空格分隔）。
缺点：破坏现有输入格式，需要迁移。

## 建议

采用**方案 A**（`SplitN`），因为：
1. 最小改动
2. NIP 本身就是用数组存储，split 只是一种简化的输入格式
3. 现有代码的 placeholder 暗示用 `:` 是预期设计

## 测试覆盖

已添加边界测试：
- `TestParseTagInput_RelayWithColon` - URL 无 `:`
- `TestParseTagInput_RelayWithColonsSplits` - 记录当前行为（URL 中 `:` 会被 split）