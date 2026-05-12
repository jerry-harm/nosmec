package compose

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/jerry-harm/nosmec/config"
)

func RunNoteCompose(app *config.AppContext) error {
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	m := NewNoteCompose(app)
	_, err := tea.NewProgram(m).Run()
	return err
}

func RunReplyCompose(app *config.AppContext) error {
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	m := NewReplyCompose(app, nil)
	_, err := tea.NewProgram(m).Run()
	return err
}

func RunQuoteCompose(app *config.AppContext) error {
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	m := NewQuoteCompose(app, nil)
	_, err := tea.NewProgram(m).Run()
	return err
}

func RunCommunityCompose(app *config.AppContext, communityAddr string) error {
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	m := NewCommunityCompose(app, communityAddr)
	_, err := tea.NewProgram(m).Run()
	return err
}