package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TimelineView 时间线视图
type TimelineView struct {
	app    interface{}
	width  int
	height int
	events []string
}

// NewTimelineView 创建新的时间线视图
func NewTimelineView(app interface{}) tea.Model {
	return &TimelineView{
		app:    app,
		events: []string{},
	}
}

// Init 初始化视图
func (v *TimelineView) Init() tea.Cmd {
	return nil
}

// Update 更新视图状态
func (v *TimelineView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			// 刷新时间线
			v.events = []string{"Event 1", "Event 2", "Event 3"} // 模拟数据
		}
	}
	return v, nil
}

// View 渲染视图
func (v *TimelineView) View() string {
	if len(v.events) == 0 {
		return lipgloss.NewStyle().
			Width(v.width).
			Height(v.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No events. Press 'r' to refresh")
	}

	content := "Timeline View\n\n"
	for i, event := range v.events {
		content += fmt.Sprintf("%d. %s\n", i+1, event)
	}

	return lipgloss.NewStyle().
		Width(v.width).
		Height(v.height).
		Padding(1, 2).
		Render(content)
}

// SetSize 设置视图尺寸
func (v *TimelineView) SetSize(width, height int) {
	v.width = width
	v.height = height
}
