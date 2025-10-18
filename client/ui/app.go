package ui

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jerry-harm/nosmec/client/nostr"
	"github.com/jerry-harm/nosmec/client/ui/views"
	"github.com/jerry-harm/nosmec/pkg/config"
)

// App 主应用
type App struct {
	config    *config.Config
	nostr     *nostr.Client
	activeTab int
	tabs      []string
	views     map[string]tea.Model
	width     int
	height    int
}

// NewApp 创建新应用
func NewApp(cfg *config.Config) *App {
	// 创建 Nostr 客户端
	nclient, err := nostr.NewClientWithoutI2P(cfg)
	if err != nil {
		log.Printf("Failed to create nostr client: %v", err)
	}

	app := &App{
		config:    cfg,
		nostr:     nclient,
		activeTab: 0,
		tabs:      []string{"Timeline", "Post", "DM", "Search"},
		views:     make(map[string]tea.Model),
	}

	// 初始化视图
	app.views["timeline"] = views.NewTimelineView(app)
	app.views["post"] = views.NewPostView(app)
	app.views["dm"] = views.NewDMView(app)
	app.views["search"] = views.NewSearchView(app)

	return app
}

// Init 初始化应用
func (a *App) Init() tea.Cmd {
	// 启动 Nostr 客户端连接
	if a.nostr != nil {
		go func() {
			if err := a.nostr.Connect(); err != nil {
				log.Printf("Failed to connect to relays: %v", err)
			}
		}()
	}
	return nil
}

// Update 更新应用状态
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return a, tea.Quit
		case "right", "l", "n", "tab":
			a.activeTab = min(a.activeTab+1, len(a.tabs)-1)
			return a, nil
		case "left", "h", "p", "shift+tab":
			a.activeTab = max(a.activeTab-1, 0)
			return a, nil
		}
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// 计算视图可用尺寸（减去标签栏高度）
		tabHeight := lipgloss.Height(a.renderTabs())
		viewHeight := msg.Height - tabHeight
		viewWidth := msg.Width

		// 更新所有视图的尺寸
		for _, view := range a.views {
			if updater, ok := view.(interface {
				SetSize(width, height int)
			}); ok {
				updater.SetSize(viewWidth, viewHeight)
			}
		}
	case tea.MouseMsg:
		if msg.Type == tea.MouseLeft {
			// 处理鼠标点击标签
			tabWidth := (a.width - 4) / len(a.tabs)
			clickedTab := (msg.X - 2) / tabWidth
			if clickedTab >= 0 && clickedTab < len(a.tabs) {
				a.activeTab = clickedTab
			}
		}
	}

	// 委托给当前视图
	currentViewName := a.getCurrentViewName()
	if view, exists := a.views[currentViewName]; exists {
		updatedView, cmd := view.Update(msg)
		a.views[currentViewName] = updatedView
		return a, cmd
	}

	return a, nil
}

// View 渲染应用界面
func (a *App) View() string {
	// 渲染标签栏
	tabs := a.renderTabs()

	// 渲染当前视图
	var content string
	currentViewName := a.getCurrentViewName()
	if view, exists := a.views[currentViewName]; exists {
		content = view.View()
	} else {
		content = "View not found"
	}

	// 组合所有元素 - 简化布局，去除边框
	return lipgloss.JoinVertical(
		lipgloss.Left,
		tabs,
		content,
	)
}

// 样式定义 - 简化样式，去除边框
var (
	highlightColor   = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666")).
				Padding(0, 2)
	activeTabStyle = inactiveTabStyle.
			Foreground(highlightColor).
			Bold(true)
)

// renderTabs 渲染标签栏
func (a *App) renderTabs() string {
	var renderedTabs []string

	for i, tab := range a.tabs {
		var style lipgloss.Style
		isActive := i == a.activeTab
		if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}
		renderedTabs = append(renderedTabs, style.Render(tab))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	return row
}

// getCurrentViewName 获取当前视图名称
func (a *App) getCurrentViewName() string {
	switch a.activeTab {
	case 0:
		return "timeline"
	case 1:
		return "post"
	case 2:
		return "dm"
	case 3:
		return "search"
	default:
		return "timeline"
	}
}

// GetNostrClient 获取 Nostr 客户端
func (a *App) GetNostrClient() *nostr.Client {
	return a.nostr
}

// GetConfig 获取配置
func (a *App) GetConfig() *config.Config {
	return a.config
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max 辅助函数
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
