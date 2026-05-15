# compose tag 输入重构设计

## 目标
改进 compose 的 tag 输入体验，基于 Bubble Tea textinput，每次只编辑一个 tag 项。

## 当前问题
- `parseTagInput` 用 `:` split URL，导致 `r:wss://relay.example` 被错误分割
- 用户体验不够直观：所有 tag 在一行显示，用户不清楚如何编辑/删除

## 建议的新设计

### 核心逻辑
1. **保持一个 textinput** 用于 tag 输入
2. **显示**：在 tag 区域下方显示一个空的 input（始终可见）
3. **当前编辑位置**：每次只编辑一个 tag 值（在 input 中）
4. **空 input + Enter** → 切换到下一个 field
5. **有内容 input + Enter** → 添加新的 tag 项
6. **在 input 中按删除键（内容为空）** → 回退编辑前一个 tag 项

### 参考
- https://github.com/charmbracelet/pop

### 状态设计
- `tags []TagValue` — 已有
- 需要新增：`editingTagIndex int` — 当前正在编辑的 tag 索引（-1 表示空 input）
- `draftTagValue string` — 当前 input 正在编辑的内容

### 流程
1. Focus tag input → 显示最后一个 tag 的内容到 input（或空）
2. Enter（有内容）→ 将内容追加到 tags，input 清空
3. Enter（空）→ 切换到 kindInput（或 contentInput）
4. Backspace（空 input）→ 删除最后一个 tag，开始编辑其内容

### 与现有实现的区别
- 当前：`tagInput.SetValue(m.tagToString(m.tags[editingTagIndex]))` 在每次切换时设置整个 tag 字符串（如 `e:abc123`）
- 建议：input 只显示 tag 的**值**（如 `abc123`），type 由当前上下文决定

## 下一步
需要新 task 来实现此设计。