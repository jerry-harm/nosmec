package ui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type subQuitMsg struct{}

func subQuit() tea.Msg { return subQuitMsg{} }

// MenuModel 菜单式UI模型
type MenuModel struct {
	choices      []string  // 菜单选项
	cursor       int       // 当前选中的菜单项索引
	selected     string    // 当前选中的功能
	content      string    // 右侧显示的内容
	inSubMenu    bool      // 是否在子菜单/功能中
	currentModel tea.Model // 当前活动的子模块
}

// Init 初始化
func (m MenuModel) Init() tea.Cmd {
	return nil
}

// Update 处理消息和更新状态
func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// 如果当前有活动的子模块，将消息路由给它
	if m.currentModel != nil {
		var cmd tea.Cmd
		m.currentModel, cmd = m.currentModel.Update(msg)

		return m, cmd
	}

	// 原有的主菜单逻辑
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			// 如果在子菜单中，按q返回主菜单
			if m.inSubMenu {
				m.inSubMenu = false
				m.content = ""
				return m, nil
			}
			// 否则退出程序
			return m, tea.Quit

		case "up", "k":
			if !m.inSubMenu {
				m.cursor = max(m.cursor-1, 0)
			}
			return m, nil

		case "down", "j":
			if !m.inSubMenu {
				m.cursor = min(m.cursor+1, len(m.choices)-1)
			}
			return m, nil

		case "enter", " ":
			if !m.inSubMenu {
				// 进入选中的功能
				m.selected = m.choices[m.cursor]
				m.inSubMenu = true
				m.updateContent()
			}
			return m, nil

		}
	case subQuitMsg:
		m.currentModel = nil
		return m, nil
	}

	return m, nil
}

// updateContent 更新右侧内容区域
func (m *MenuModel) updateContent() {
	switch m.selected {
	case "Global":
		// 创建Global子模块
		m.currentModel = NewGlobalModel()
		m.content = "" // 清空内容，因为子模块会自己渲染
	default:
		m.content = "Welcome to NoSMEC UI\n\nSelect a menu item to begin."
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

	if !m.inSubMenu {
		// 主菜单视图
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

	} else {
		// 功能视图
		doc.WriteString(subTitleStyle.Render(m.selected))
		doc.WriteString("\n\n")
		doc.WriteString(m.content)
		doc.WriteString("\n\n")

	}

	// 状态栏 - 显示程序状态
	status := statusStyle.Render("NoSMEC UI - Use arrow keys to navigate, q to quit")
	doc.WriteString("\n\n" + status)

	return docStyle.Render(doc.String())
}

// 样式定义
var (
	docStyle = lipgloss.NewStyle().
			Padding(1, 2, 1, 2)

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
		choices:   choices,
		cursor:    0,
		selected:  "",
		content:   "Welcome to NoSMEC UI\n\nSelect a menu item to begin.",
		inSubMenu: false,
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
