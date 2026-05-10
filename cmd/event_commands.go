package cmd

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/tui/window/event"
	"github.com/spf13/cobra"
)

func registerEventCommands() {
	eventCmd := &cobra.Command{
		Use:   "event <event-id>",
		Short: "Show event details",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			eventID := args[0]

			// Validate event ID length
			if len(eventID) != 64 {
				fmt.Printf("Error: event-id must be 64 characters, got %d\n", len(eventID))
				os.Exit(1)
			}

			app := getApp()

			if err := RunEventDetail(app, eventID); err != nil {
				fmt.Printf("Error running event detail: %v\n", err)
				os.Exit(1)
			}
		},
	}

	RegisterCommandGroup("Events", "Event operations", eventCmd)
}

func RunEventDetail(app *config.AppContext, eventID string) error {
	m := event.NewFromID(eventID, app, 80, 24)
	_, err := tea.NewProgram(m).Run()
	return err
}
