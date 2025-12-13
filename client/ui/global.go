package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"fiatjaf.com/nostr"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/jerry-harm/nosmec/client"
	"github.com/jerry-harm/nosmec/config"
)

// GlobalKeyMap 定义Global模型的按键绑定
type globalKeyMap struct {
	Quit    key.Binding
	Refresh key.Binding
	Enter   key.Binding
	Help    key.Binding
}

// DefaultGlobalKeyMap 默认Global模型按键绑定
var DefaultGlobalKeyMap = globalKeyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit/back"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter", " "),
		key.WithHelp("enter", "select")),
}

// noteItem 实现list.Item接口
type noteItem struct {
	event nostr.Event
}

func (i noteItem) Title() string {

	author := fmt.Sprintf("%x", i.event.PubKey.String())
	author = author[len(author)-8:]

	content := strings.Replace(i.event.Content, "\n", "", -1)

	return fmt.Sprintf("%s: %s", author, content)
}

func (i noteItem) Description() string {
	// 显示时间戳
	return time.Unix(int64(i.event.CreatedAt), 0).Format("2006-01-02 15:04:05")
}

func (i noteItem) FilterValue() string { return i.event.Content }

// notesLoadingCompleteMsg 笔记加载完成的消息
type notesLoadingCompleteMsg struct{}

// GlobalModel 显示全局note的模型
type GlobalModel struct {
	list             list.Model
	notes            []noteItem
	err              error
	keys             globalKeyMap
	eventdetail      *nostr.Event
	eventDetailModel tea.Model
	eventChan        chan nostr.Event // 用于接收事件的channel
	errChan          chan error       // 用于接收错误的channel
}

// NewGlobalModel 创建新的GlobalModel实例
func NewGlobalModel(width, height int) GlobalModel {
	// 创建列表
	items := []list.Item{}

	// 计算列表大小，考虑内边距
	h, v := docStyle.GetFrameSize()

	l := list.New(items, list.NewDefaultDelegate(), width-v, height-h)
	l.Title = "Global Notes"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true) // 显示列表的帮助
	l.Styles.Title = titleStyle

	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			DefaultGlobalKeyMap.Refresh,
		}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			DefaultGlobalKeyMap.Refresh,
		}
	}

	return GlobalModel{
		list:      l,
		notes:     []noteItem{},
		keys:      DefaultGlobalKeyMap,
		eventChan: make(chan nostr.Event, 100),
		errChan:   make(chan error, 1),
	}
}

func (m GlobalModel) Init() tea.Cmd {
	// 启动spinner
	m.list.StartSpinner()

	// 同时启动监听goroutine和等待第一个事件
	return tea.Batch(
		m.listenForEvents(), // 启动监听goroutine
		m.waitForEvent(),    // 开始等待第一个事件
	)
}

// listenForEvents 启动goroutine监听事件
func (m GlobalModel) listenForEvents() tea.Cmd {
	return func() tea.Msg {
		go func() {
			system := client.GetSystem()
			if system == nil {
				m.errChan <- fmt.Errorf("client system not initialized")
				return
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
				m.errChan <- fmt.Errorf("未配置relay")
				return
			}

			// 使用FetchMany从所有relay获取事件
			eventsChan := system.Pool.FetchMany(ctx, relays, filter, nostr.SubscriptionOptions{})

			for {
				select {
				case relayEvent, ok := <-eventsChan:
					if !ok {
						// 通道关闭，发送完成信号
						m.errChan <- nil // nil表示正常完成
						return
					}
					// 发送事件到channel
					m.eventChan <- relayEvent.Event
				case <-ctx.Done():
					// 超时
					m.errChan <- fmt.Errorf("获取超时")
					return
				}
			}
		}()
		return nil
	}
}

// waitForEvent 等待事件或错误
func (m GlobalModel) waitForEvent() tea.Cmd {
	return func() tea.Msg {
		select {
		case event := <-m.eventChan:
			return noteItem{event: event}
		case err := <-m.errChan:
			if err == nil {
				return notesLoadingCompleteMsg{}
			}
			return err
		}
	}
}

// Update 处理消息
func (m GlobalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// 如果当前有活动的eventDetailModel，将消息路由给它
	if m.eventDetailModel != nil {
		var cmd tea.Cmd
		m.eventDetailModel, cmd = m.eventDetailModel.Update(msg)
		// 检查是否返回了EventDetailQuitMsg
		if _, isQuit := msg.(EventDetailQuitMsg); isQuit {
			m.eventDetailModel = nil
			m.eventdetail = nil
			return m, nil
		}
		// 检查是否返回了subQuitMsg（保持向后兼容）
		if _, isSubQuit := msg.(subQuitMsg); isSubQuit {
			m.eventDetailModel = nil
			m.eventdetail = nil
			return m, nil
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		// 如果已经有eventDetailModel，更新其大小
		if m.eventDetailModel != nil {
			m.eventDetailModel, _ = m.eventDetailModel.Update(msg)
		}
	case tea.KeyMsg:
		// 检查是否在详情模式下
		if m.eventdetail != nil {
			// 创建新的EventDetailModel
			h, v := docStyle.GetFrameSize()
			m.eventDetailModel = NewEventDetailModel(*m.eventdetail, m.list.Width()+v, m.list.Height()+h)
			return m, m.eventDetailModel.Init()
		}

		// 列表模式下的按键处理
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, subQuit
		case key.Matches(msg, m.keys.Refresh):
			// 清空当前notes，重新开始异步获取
			m.notes = []noteItem{}
			m.list.SetItems([]list.Item{})
			m.err = nil
			m.list.StartSpinner()
			// 重新初始化channels
			m.eventChan = make(chan nostr.Event, 100)
			m.errChan = make(chan error, 1)
			return m, tea.Batch(
				m.listenForEvents(),
				m.waitForEvent(),
			)
		case key.Matches(msg, m.keys.Help):
			m.list.SetShowHelp(!m.list.ShowHelp())
			return m, nil
		case key.Matches(msg, m.keys.Enter):
			i, ok := m.list.SelectedItem().(noteItem)
			if ok {
				m.eventdetail = &i.event
			}
			return m, nil
		}
	case noteItem:
		// 将单个note添加到notes切片
		m.notes = append(m.notes, msg)
		// 更新列表项
		items := make([]list.Item, len(m.notes))
		for i, note := range m.notes {
			items[i] = note
		}
		m.list.SetItems(items)
		// 继续等待下一个事件
		return m, m.waitForEvent()

	case notesLoadingCompleteMsg:
		// 加载完成，停止spinner
		m.list.StopSpinner()
		return m, nil

	case error:
		m.err = msg
		m.list.StopSpinner()
		return m, nil
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View 渲染UI
func (m GlobalModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("error: %v", m.err)
	}

	if m.eventDetailModel != nil {
		return m.eventDetailModel.View()
	}

	return docStyle.Render(m.list.View())
}

// renderEventDetail 渲染事件详情表格
// TODO 格式化成json用textarea显示
func renderEventDetail(event nostr.Event) string {
	// 创建表格数据
	rows := [][]string{
		{"Field", "Value"},
		{"ID", event.ID.String()},
		{"PubKey", event.PubKey.String()},
		{"Created At", time.Unix(int64(event.CreatedAt), 0).Format("2006-01-02 15:04:05")},
		{"Kind", fmt.Sprintf("%d", event.Kind)},
		{"Content", event.Content},
	}

	// 添加标签
	if len(event.Tags) > 0 {
		for i, tag := range event.Tags {
			rows = append(rows, []string{fmt.Sprintf("Tag[%d]", i), fmt.Sprintf("%v", tag)})
		}
	}

	// 创建表格
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("63"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
			}
			if col == 0 {
				return lipgloss.NewStyle().Foreground(lipgloss.Color("117"))
			}
			return lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
		}).
		Rows(rows...).Wrap(true).Width(30)

	// 获取表格字符串
	tableStr := t.Render()

	return docStyle.Render(tableStr + "\n\n" + helpStyle.Render("Press AnyKey to return"))
}
