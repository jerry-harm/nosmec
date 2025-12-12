package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"fiatjaf.com/nostr"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jerry-harm/nosmec/client"
	"github.com/jerry-harm/nosmec/config"
)

// noteItem 实现list.Item接口
type noteItem struct {
	event nostr.Event
}

func (i noteItem) Title() string {
	// 显示作者公钥的前8个字符（十六进制）
	author := fmt.Sprintf("%x", i.event.PubKey.String())
	if len(author) > 8 {
		author = author[:8]
	}
	content := i.event.Content
	if len(content) > 50 {
		content = content[:50] + "..."
	}
	return fmt.Sprintf("%s: %s", author, content)
}

func (i noteItem) Description() string {
	// 显示时间戳
	return time.Unix(int64(i.event.CreatedAt), 0).Format("2006-01-02 15:04:05")
}

func (i noteItem) FilterValue() string { return i.event.Content }

// GlobalModel 显示全局note的模型
type GlobalModel struct {
	list    list.Model
	loading bool
	notes   []noteItem
	err     error
}

// NewGlobalModel 创建新的GlobalModel实例
func NewGlobalModel() GlobalModel {
	// 创建列表
	items := []list.Item{}
	l := list.New(items, list.NewDefaultDelegate(), 20, 20)
	l.Title = "Global Notes (最近100条)"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true)

	return GlobalModel{
		list:    l,
		loading: true,
		notes:   []noteItem{},
	}
}

// Init 初始化，启动异步获取note
func (m GlobalModel) Init() tea.Cmd {
	return m.fetchNotes
}

// fetchNotes 获取note的命令
func (m GlobalModel) fetchNotes() tea.Msg {
	// 获取系统实例
	system := client.GetSystem()
	if system == nil {
		return fmt.Errorf("client system not initialized")
	}

	// 创建filter获取最近100条note
	filter := nostr.Filter{
		Kinds: []nostr.Kind{nostr.KindTextNote},
		Limit: 100,
	}

	// 使用Pool从网络获取事件
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 获取配置的relay列表
	relays := config.Global.Client.Relays
	if len(relays) == 0 {
		return fmt.Errorf("未配置relay")
	}

	// 使用FetchMany从所有relay获取事件
	eventsChan := system.Pool.FetchMany(ctx, relays, filter, nostr.SubscriptionOptions{})
	events := []nostr.Event{}

	// 收集事件，直到通道关闭或超时
	for {
		select {
		case relayEvent, ok := <-eventsChan:
			if !ok {
				// 通道关闭，返回已收集的事件
				return notesLoadedMsg{notes: convertEventsToNoteItems(events)}
			}
			events = append(events, relayEvent.Event)
		case <-ctx.Done():
			// 超时，返回已收集的事件（可能为空）
			if len(events) == 0 {
				return fmt.Errorf("获取超时，未收到任何note")
			}
			return notesLoadedMsg{notes: convertEventsToNoteItems(events)}
		}
	}
}

// convertEventsToNoteItems 将nostr.Event切片转换为noteItem切片
func convertEventsToNoteItems(events []nostr.Event) []noteItem {
	notes := make([]noteItem, len(events))
	for i, event := range events {
		notes[i] = noteItem{event: event}
	}
	return notes
}

// notesLoadedMsg 笔记加载完成的消息
type notesLoadedMsg struct {
	notes []noteItem
}

// Update 处理消息
func (m GlobalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, subQuit
		case "r":
			return m, m.fetchNotes
		}
	case notesLoadedMsg:
		m.loading = false
		m.notes = msg.notes
		// 更新列表项
		items := make([]list.Item, len(m.notes))
		for i, note := range m.notes {
			items[i] = note
		}
		m.list.SetItems(items)

	case error:
		m.err = msg
		m.loading = false
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View 渲染UI
func (m GlobalModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("错误: %v\n\n按 q 返回", m.err)
	}

	if m.loading {
		return "正在从网络获取note...\n\n按 q 返回"
	}

	if len(m.notes) == 0 {
		return "没有找到note\n\n按 q 返回"
	}

	var b strings.Builder
	b.WriteString(m.list.View())

	return globalDocStyle.Render(b.String())
}

// 全局样式（避免与ui.go冲突）
var (
	globalDocStyle = lipgloss.NewStyle().
		Padding(1, 2, 1, 2)
)
