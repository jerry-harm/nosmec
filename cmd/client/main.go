package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jerry-harm/nosmec/client/ui"
	"github.com/jerry-harm/nosmec/pkg/config"
)

func main() {
	// 加载配置
	config := config.LoadConfig()

	log.Printf("Starting nostr client...")
	log.Printf("Default relays: %v", config.Client.DefaultRelays)
	log.Printf("Theme: %s", config.Client.Theme)

	// 创建应用
	app := ui.NewApp(config)

	// 启动 TUI
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
