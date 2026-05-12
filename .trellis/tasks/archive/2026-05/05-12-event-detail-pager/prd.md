# event详情用pager替换glamour

## Goal

将 event 详情的内容渲染从 glamour markdown 切换为纯文本显示，消除 `.i2p` 等特殊 TLD URL 的渲染 bug，同时简化渲染逻辑。

## What I already know

* Event detail view at `tui/window/event/event.go` uses `viewport.Model` for scrolling
* `glamour.TermRenderer` renders markdown content via `m.glamour.Render(content)` in `view.go:66`
* glamour's goldmark linkify 存在 bug：`http://x.i2p` 被识别为 `.i 2p`（TLD regex `[a-z]+` 在数字处断裂）
* NIP 对 event.content 没有 markdown 格式要求，content 就是纯 string
* 未来如果要支持 kind 30023/30024（markdown 长文章）会单独讨论
* 没有专门的 `Pager` 组件，viewport.Model 本身就有 pager 行为

## Requirements

* event.content 当**纯文本**渲染，不走 glamour markdown
* 移除所有 glamour 相关代码（`glamour` import、`glamour.TermRenderer` 字段、`glamour.Render` 调用）
* URL 以纯文本显示，不加颜色/样式/括号等额外处理
* 保留 `viewport.Model` 做滚动
* 保留现有的 header、tags、signature 显示部分不变
* 保留 `j` 键切换 Raw JSON 显示功能

## Acceptance Criteria

* [ ] 移除 glamour，content 以纯文本显示
* [ ] `.i2p` / `.onion` 等特殊 TLD 的 URL 正确完整显示（不再是 `.i 2p`）
* [ ] viewport 滚动行为保持不变
* [ ] Raw JSON 切换功能保持不变
* [ ] lint / typecheck / tests pass

## Definition of Done

* event detail 内容以纯文本渲染，URL 不再有识别 bug
* 现有功能（滚动、header、tags、signature、raw json）保持不变
* glamour 依赖可移除（如果 go.mod 里只有这一处用到）

## Out of Scope

* markdown 渲染支持（kind 30023/30024 会单独讨论）
* NIP-27 `nostr:` URI 的特殊处理
* 其他视图的改动

## Technical Approach

### 渲染流程变化

**Before**:
```
content (string)
  → glamour.Render(content)  [markdown + linkify]
  → viewport.SetContent(rendered)
```

**After**:
```
content (string)
  → viewport.SetContent(content)  [纯文本，不做任何转换]
```

### 具体改动

1. **删除 glamour 相关代码**：
   - `event.go`: 移除 `glamour` 和 `styles` imports，移除 `glamour *glamour.TermRenderer` 字段
   - `view.go`: 移除 `m.glamour.Render()` 调用，改为直接显示 content
   - `event.go`: 移除 `View()` 中的 glamour 初始化逻辑

2. **纯文本显示**：
   - content 直接 SetContent，不做任何 markdown 解析
   - URL 就是普通字符，不会被误识别

3. **保留的功能**：
   - header（author、time、kind）
   - tags 显示
   - signature 显示
   - Raw JSON 切换（`j` 键）
   - viewport 滚动

### Key Files

* `tui/window/event/event.go` — 移除 glamour 字段和初始化
* `tui/window/event/view.go` — 移除 glamour.Render 调用，改为纯文本

## Research References

* `research/glamour-url-handling.md` — glamour URL 渲染研究，goldmark linkify bug 详解
* `research/nip-url-formats.md` — NIP-19/21/27 URL 格式研究