package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type subQuitMsg struct{}

func subQuit() tea.Msg { return subQuitMsg{} }

// KeyMap 定义按键绑定
type KeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
	Quit  key.Binding
	Back  key.Binding
	Help  key.Binding
}

// ShortHelp 返回短帮助（显示在状态栏）
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Quit, k.Help}
}

// FullHelp 返回完整帮助（显示在帮助面板）
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.Quit, k.Back, k.Help},
	}
}

// DefaultKeyMap 默认按键绑定
var DefaultKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter", " "),
		key.WithHelp("enter", "select"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Back: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "back"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
}

// MenuModel 菜单式UI模型
type MenuModel struct {
	choices      []string  // 菜单选项
	cursor       int       // 当前选中的菜单项索引
	selected     string    // 当前选中的功能
	currentModel tea.Model // 当前活动的子模块
	keys         KeyMap    // 按键绑定
	help         help.Model
	width        int // 窗口宽度
	height       int // 窗口高度
}

// Init 初始化
func (m MenuModel) Init() tea.Cmd {
	return nil
}

// Update 处理消息和更新状态
func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// 存储窗口大小
		m.width = msg.Width
		m.height = msg.Height

		h, _ := docStyle.GetFrameSize()
		m.help.Width = msg.Width - h

		// 如果当前有活动的子模块，将窗口大小消息传递给它
		if m.currentModel != nil {
			var cmd tea.Cmd
			m.currentModel, cmd = m.currentModel.Update(msg)
			return m, cmd
		}
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			if m.currentModel != nil {
				m.currentModel = nil
				return m, nil
			}
			return m, tea.Quit
		case key.Matches(msg, m.keys.Up):
			if m.currentModel == nil {
				m.cursor = max(m.cursor-1, 0)
			}
		case key.Matches(msg, m.keys.Down):
			if m.currentModel == nil {
				m.cursor = min(m.cursor+1, len(m.choices)-1)
			}
		case key.Matches(msg, m.keys.Enter):
			if m.currentModel == nil {
				m.selected = m.choices[m.cursor]
				m.updateContent()
				return m, m.currentModel.Init()
			}
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		}
	case subQuitMsg:
		m.currentModel = nil
		return m, nil
	}
	// 如果当前有活动的子模块，将其他消息路由给它
	if m.currentModel != nil {
		var cmd tea.Cmd
		m.currentModel, cmd = m.currentModel.Update(msg)
		return m, cmd
	}
	return m, nil
}

// updateContent 更新右侧内容区域
func (m *MenuModel) updateContent() {
	switch m.selected {
	case "Global":
		m.currentModel = NewGlobalModel(m.width, m.height)
	}
}

// View 渲染UI
func (m MenuModel) View() string {
	// 如果当前有活动的子模块，渲染子模块的视图
	if m.currentModel != nil {
		return m.currentModel.View()
	}

	doc := strings.Builder{}

	// 标题
	title := titleStyle.Render("NoSMEC - Nostr Server Management & Control")
	doc.WriteString(title)
	doc.WriteString("\n\n")

	// 主菜单视图（只有在没有活动的子模块时才显示）
	doc.WriteString("Main Menu:\n\n")

	// 渲染菜单选项
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		style := menuItemStyle
		if m.cursor == i {
			style = selectedMenuItemStyle
		}

		doc.WriteString(style.Render(fmt.Sprintf("%s %s", cursor, choice)))
		doc.WriteString("\n")
	}

	// 添加帮助信息
	doc.WriteString("\n")
	doc.WriteString(m.help.View(m.keys))

	return docStyle.Render(doc.String())
}

// 样式定义
var (
	docStyle = lipgloss.NewStyle().
			Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	subTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 2)

	menuItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC"))

	selectedMenuItemStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#7D56F4")).
				Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1)
)

// StartMenu 启动菜单式UI
func StartMenu() {
	choices := []string{"Global"}

	m := MenuModel{
		choices:  choices,
		cursor:   0,
		selected: "",
		keys:     DefaultKeyMap,
		help:     help.New(),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

// 辅助函数
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
