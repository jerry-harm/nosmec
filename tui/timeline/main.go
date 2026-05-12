package timeline

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/jerry-harm/nosmec/config"
)

func RunTimeline(app *config.AppContext, filter string, hashtags []string, limit int, communityAddr string) error {
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	m := NewModel(app, filter, hashtags, limit, communityAddr)
	_, err := tea.NewProgram(m).Run()
	if err != nil {
		fmt.Println("Error running timeline:", err)
		os.Exit(1)
	}
	return nil
}
