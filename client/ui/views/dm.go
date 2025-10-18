package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DMView 私信视图
type DMView struct {
	app    interface{}
	width  int
	height int
}

// NewDMView 创建新的私信视图
func NewDMView(app interface{}) tea.Model {
	return &DMView{
		app: app,
	}
}

// Init 初始化视图
func (v *DMView) Init() tea.Cmd {
	return nil
}

// Update 更新视图状态
func (v *DMView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return v, nil
}

// View 渲染视图
func (v *DMView) View() string {
	content := "Direct Messages View\n\n"
	content += "DM functionality coming soon..."

	return lipgloss.NewStyle().
		Width(v.width).
		Height(v.height).
		Padding(1, 2).
		Render(content)
}

// SetSize 设置视图尺寸
func (v *DMView) SetSize(width, height int) {
	v.width = width
	v.height = height
}
