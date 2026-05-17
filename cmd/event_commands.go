package cmd

import (
	"fmt"
	"os"

	"fiatjaf.com/nostr/nip19"
	"github.com/jerry-harm/nosmec/tui/bubblon"
	tea "charm.land/bubbletea/v2"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/tui/event"
	"github.com/spf13/cobra"
)

func registerEventCommands() {
	eventCmd := &cobra.Command{
		Use:   "event <event-id>",
		Short: "Show event details",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			eventID := args[0]

			// Decode nevent/note format if needed
			actualID := eventID
			if len(eventID) > 64 {
				pointer, err := nip19.ToPointer(eventID)
				if err != nil {
					fmt.Printf("Error: invalid event ID format: %v\n", err)
					os.Exit(1)
				}
				filter := pointer.AsFilter()
				if len(filter.IDs) > 0 {
					actualID = filter.IDs[0].Hex()
				} else {
					fmt.Printf("Error: no event ID found in pointer\n")
					os.Exit(1)
				}
			} else if len(eventID) != 64 {
				fmt.Printf("Error: event-id must be 64 characters or nevent/note format, got %d\n", len(eventID))
				os.Exit(1)
			}

			app := getApp()

			if err := RunEventDetail(app, actualID); err != nil {
				fmt.Printf("Error running event detail: %v\n", err)
				os.Exit(1)
			}
		},
	}

	RegisterCommandGroup("Events", "Event operations", eventCmd)
}

func RunEventDetail(app *config.AppContext, eventID string) error {
	m := event.NewFromID(eventID, app, 80, 24, nil)
	ctrl, err := bubblon.New(m)
	if err != nil {
		return fmt.Errorf("failed to create bubblon controller: %w", err)
	}
	m.SetController(&ctrl)
	_, err = tea.NewProgram(ctrl).Run()
	return err
}
