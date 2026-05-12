# brainstorm: compose-form-ui

## Goal

重新设计 compose TUI，使用结构化 form 界面（Kind selector、multi-value Tags、Content textarea），通过 windowmanager 统一管理全屏显示，支持独立运行和嵌入 timeline。

## What I already know

**架构要求**：
- 所有 TUI 通过 `windowmanager` 管理显示（参考 timeline、event）
- 全屏：View() 返回时设置 `v.AltScreen = true`
- 独立运行和嵌入 timeline 调用方式一致

**windowmanager 模式**：
```go
wm := windowmanager.New()
wm.Open(composeModel)
tea.NewProgram(wm).Run()
```

**Bubble Tea 全屏模式**：
```go
func (m *model) View() tea.View {
    v := tea.NewView(m.windowManager.View())
    v.AltScreen = true
    return v
}
```

**NIP Tag 类型**：
- `e` — Event ID（回复、引用）
- `p` — Pubkey（提及、收件人）
- `a` — Address（community）
- `t` — Hashtag
- `r` — Relay hint
- `q` — Quote（引用）

**Reply Convention (NIP-10)**：
```json
{"kind": 1, "tags": [["e", "<parent-id>", "<relay>", "reply"], ["p", "<author-pubkey>"]}
```

**Kind 常量**：
- Kind 1 — TextNote
- Kind 1111 — CommunityPost
- Kind 34550 — CommunityDefinition

## Requirements

- [ ] 实现 `window.Window` 接口（Init, Update, View, ID, Close）
- [ ] 通过 windowmanager 全屏显示
- [ ] Kind selector：支持 1, 1111, 34550
- [ ] Multi-value Tags：e, p, a, t, r, q
- [ ] Content textarea
- [ ] Tab 焦点切换：Kind ↔ Content ↔ Tag
- [ ] Pre-fill 支持（reply: e+p, quote: q+e）
- [ ] 通过 windowmanager.Open() 调用，方式与 timeline/event 一致

## Technical Approach

### 1. Model 作为 window.Window
```go
type model struct {
    wm *windowmanager.WindowManager  // 嵌入 windowmanager 时不为 nil
    // ... 其他字段
}

func (m *model) View() tea.View {
    if m.wm != nil {
        v := tea.NewView(m.wm.View())
        v.AltScreen = true
        return v
    }
    // 降级：直接渲染
}
```

### 2. main.go 使用 windowmanager
```go
func RunNoteCompose(app *config.AppContext) error {
    m := NewNoteCompose(app)
    m.wm = windowmanager.New()
    m.wm.Open(m)
    _, err := tea.NewProgram(m).Run()
    return err
}
```

### 3. 焦点管理
- 用 `textinput.Model` + `Focus()`/`Blur()` 管理焦点
- Tab 键循环切换

## Open Questions

1. **Tab 退出 textarea**：在 textarea 中按 Tab 退出输入，切换到下一个字段
2. **Ctrl+Enter 发布**：在 textarea 中按 Ctrl+Enter 发送内容

## Decision (ADR-lite)

**Context**: 需要在多行 textarea 中切换焦点，同时支持发送

**Decision**:
- 在 textarea 中按 Tab：退出输入，切换到下一个字段（不插入 tab 字符）
- 在 textarea 中按 Ctrl+Enter：发送内容
- Tab 切换顺序：Kind → Content → Tag → Kind（循环）

## Design Principles

1. **ESC = Quit/Close**：所有 TUI 视图统一使用 ESC 退出/关闭。（2026-05-12）
   - compose: ESC 退出
   - timeline detail: ESC 关闭详情
   - 通用：ESC 用于取消操作或返回上一级

2. **Tag 输入流程**：
   - 在 Tag 字段：按 e/p/a/t/r/q 选择类型（不显示在输入框）
   - 输入标签值
   - 按 Enter 添加标签
   - 在 Tag 输入为空时按 Backspace 删除最后一个标签

## Acceptance Criteria

- [ ] `note compose` 全屏显示 form
- [ ] 可以切换 Kind
- [ ] 可以添加/删除 Tags
- [ ] Tab 在字段间切换焦点
- [ ] reply/quote 场景 pre-fill 正确
- [ ] 独立运行和嵌入 timeline 行为一致

## Definition of Done

- go test ./... 通过
- 全屏显示正常
- 焦点切换正常
- 发送功能正常

## Out of Scope

- Preview mode
- Tag autocomplete
- Rich text editor

## Technical Notes

- 现有实现：`tui/compose/model.go`, `tui/compose/main.go`
- 参考：`tui/timeline/model.go`, `tui/window/event/event.go`, `tui/windowmanager/windowmanager.go`