package ui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"fiatjaf.com/nostr"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// EventDetailQuitMsg 退出EventDetail视图的消息
type EventDetailQuitMsg struct{}

func eventDetailQuit() tea.Msg {
	return EventDetailQuitMsg{}
}

// EventDetailKeyMap 定义EventDetail模型的按键绑定
type EventDetailKeyMap struct {
	Back   key.Binding
	Toggle key.Binding // 切换文本/JSON显示
	Reply  key.Binding
	Repost key.Binding
	CopyID key.Binding
	Help   key.Binding
}

// DefaultEventDetailKeyMap 默认EventDetail模型按键绑定
var DefaultEventDetailKeyMap = EventDetailKeyMap{
	Back: key.NewBinding(
		key.WithKeys("q", "esc"),
		key.WithHelp("q", "back"),
	),
	Toggle: key.NewBinding(
		key.WithKeys("j"),
		key.WithHelp("j", "toggle JSON/text"),
	),
	Reply: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "reply"),
	),
	Repost: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "repost"),
	),
	CopyID: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "copy event ID"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
}

// ShortHelp 返回短帮助（显示在状态栏）
func (k EventDetailKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Back, k.Toggle, k.Reply, k.Repost, k.CopyID}
}

// FullHelp 返回完整帮助（显示在帮助面板）
func (k EventDetailKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Back, k.Toggle},
		{k.Reply, k.Repost, k.CopyID},
		{k.Help},
	}
}

// DisplayMode 显示模式枚举
type DisplayMode int

const (
	DisplayText DisplayMode = iota
	DisplayJSON
)

// EventDetailModel 显示事件详情的模型
type EventDetailModel struct {
	event    nostr.Event
	mode     DisplayMode
	viewport viewport.Model
	keys     EventDetailKeyMap
	help     help.Model
	width    int
	height   int
	showHelp bool
}

// NewEventDetailModel 创建新的事件详情模型
func NewEventDetailModel(event nostr.Event, width, height int) EventDetailModel {
	vp := viewport.New(width, height)
	vp.Style = lipgloss.NewStyle().Padding(0, 1)

	m := EventDetailModel{
		event:    event,
		mode:     DisplayText,
		viewport: vp,
		keys:     DefaultEventDetailKeyMap,
		help:     help.New(),
		width:    width,
		height:   height,
		showHelp: false,
	}

	m.updateViewportContent()
	return m
}

// Init 初始化
func (m EventDetailModel) Init() tea.Cmd {
	return nil
}

// Update 处理消息和更新状态
func (m EventDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 5
		m.updateViewportContent()
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Back):
			return m, eventDetailQuit
		case key.Matches(msg, m.keys.Toggle):
			if m.mode == DisplayText {
				m.mode = DisplayJSON
			} else {
				m.mode = DisplayText
			}
			m.updateViewportContent()
			return m, nil
		case key.Matches(msg, m.keys.Reply):
			// TODO: 实现回复功能
			return m, nil
		case key.Matches(msg, m.keys.Repost):
			// TODO: 实现转发功能
			return m, nil
		case key.Matches(msg, m.keys.CopyID):
			// TODO: 实现复制event ID功能
			return m, nil
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil
		}
	}

	// 处理viewport的滚动
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// updateViewportContent 更新viewport的内容
func (m *EventDetailModel) updateViewportContent() {
	var content string
	if m.mode == DisplayText {
		content = m.renderTextMode()
	} else {
		content = m.renderJSONMode()
	}
	m.viewport.SetContent(content)
}

// renderTextMode 渲染文本模式的内容
func (m *EventDetailModel) renderTextMode() string {
	var builder strings.Builder

	// 标题
	title := titleStyle.Render("Event Details")
	builder.WriteString(title)
	builder.WriteString("\n\n")

	// 基本信息
	builder.WriteString(subTitleStyle.Render("Basic Information"))
	builder.WriteString("\n")

	// 作者信息
	author := fmt.Sprintf("Author: %s", m.event.PubKey.String())
	if len(author) > 60 {
		author = author[:57] + "..."
	}
	builder.WriteString(fmt.Sprintf("  %s\n", author))

	// 时间
	timeStr := time.Unix(int64(m.event.CreatedAt), 0).Format("2006-01-02 15:04:05")
	builder.WriteString(fmt.Sprintf("  Time: %s\n", timeStr))

	// 事件ID
	eventID := fmt.Sprintf("Event ID: %s", m.event.ID.String())
	if len(eventID) > 60 {
		eventID = eventID[:57] + "..."
	}
	builder.WriteString(fmt.Sprintf("  %s\n", eventID))

	// 类型
	builder.WriteString(fmt.Sprintf("  Kind: %d\n", m.event.Kind))

	// 内容
	builder.WriteString("\n")
	builder.WriteString(subTitleStyle.Render("Content"))
	builder.WriteString("\n")
	content := strings.TrimSpace(m.event.Content)
	if content == "" {
		content = "(empty)"
	}
	// 添加缩进并保持换行
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		builder.WriteString(fmt.Sprintf("  %s\n", line))
	}

	// 标签
	if len(m.event.Tags) > 0 {
		builder.WriteString("\n")
		builder.WriteString(subTitleStyle.Render("Tags"))
		builder.WriteString("\n")
		for i, tag := range m.event.Tags {
			builder.WriteString(fmt.Sprintf("  [%d] %v\n", i, tag))
		}
	}

	return builder.String()
}

// renderJSONMode 渲染JSON模式的内容
func (m *EventDetailModel) renderJSONMode() string {
	// 创建JSON结构
	jsonEvent := map[string]interface{}{
		"id":         m.event.ID.String(),
		"pubkey":     m.event.PubKey.String(),
		"created_at": m.event.CreatedAt,
		"kind":       m.event.Kind,
		"tags":       m.event.Tags,
		"content":    m.event.Content,
		"sig":        fmt.Sprintf("%x", m.event.Sig),
	}

	// 格式化JSON
	jsonBytes, err := json.MarshalIndent(jsonEvent, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting JSON: %v", err)
	}

	var builder strings.Builder
	builder.WriteString(titleStyle.Render("Event Details (JSON Mode)"))
	builder.WriteString("\n\n")
	builder.WriteString(string(jsonBytes))
	return builder.String()
}

// View 渲染UI
func (m EventDetailModel) View() string {
	return docStyle.Render(m.viewport.View() + "\n\n" + m.help.View(m.keys))
}
