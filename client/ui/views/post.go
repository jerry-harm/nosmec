package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PostView 发布视图
type PostView struct {
	app     interface{}
	width   int
	height  int
	content string
}

// NewPostView 创建新的发布视图
func NewPostView(app interface{}) tea.Model {
	return &PostView{
		app:     app,
		content: "",
	}
}

// Init 初始化视图
func (v *PostView) Init() tea.Cmd {
	return nil
}

// Update 更新视图状态
func (v *PostView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// 发布内容
			if v.content != "" {
				// TODO: 实际发布到 Nostr
				fmt.Printf("Posting: %s\n", v.content)
				v.content = ""
			}
		case "backspace":
			if len(v.content) > 0 {
				v.content = v.content[:len(v.content)-1]
			}
		default:
			if len(msg.String()) == 1 {
				v.content += msg.String()
			}
		}
	}
	return v, nil
}

// View 渲染视图
func (v *PostView) View() string {
	content := "Post View\n\n"
	content += "Type your message and press Enter to post:\n\n"
	content += fmt.Sprintf("> %s", v.content)
	content += "\n\nPress Enter to post, Backspace to delete"

	return lipgloss.NewStyle().
		Width(v.width).
		Height(v.height).
		Padding(1, 2).
		Render(content)
}

// SetSize 设置视图尺寸
func (v *PostView) SetSize(width, height int) {
	v.width = width
	v.height = height
}
