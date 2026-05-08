package timeline

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/jerry-harm/nosmec/config"
)

func RunTimeline(app *config.AppContext, filter string, hashtags []string, limit int) error {
	m := NewModel(app, filter, hashtags, limit)
	_, err := tea.NewProgram(m).Run()
	if err != nil {
		fmt.Println("Error running timeline:", err)
		os.Exit(1)
	}
	return nil
}