package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SearchView 搜索视图
type SearchView struct {
	app    interface{}
	width  int
	height int
}

// NewSearchView 创建新的搜索视图
func NewSearchView(app interface{}) tea.Model {
	return &SearchView{
		app: app,
	}
}

// Init 初始化视图
func (v *SearchView) Init() tea.Cmd {
	return nil
}

// Update 更新视图状态
func (v *SearchView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return v, nil
}

// View 渲染视图
func (v *SearchView) View() string {
	content := "Search View\n\n"
	content += "Search functionality coming soon..."

	return lipgloss.NewStyle().
		Width(v.width).
		Height(v.height).
		Padding(1, 2).
		Render(content)
}

// SetSize 设置视图尺寸
func (v *SearchView) SetSize(width, height int) {
	v.width = width
	v.height = height
}
