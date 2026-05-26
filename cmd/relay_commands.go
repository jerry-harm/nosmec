package cmd

import (
	"fmt"
	"io"

	"github.com/jerry-harm/nosmec/config"
	"github.com/spf13/cobra"
)

func registerRelayCommands() {
	relayCmd := &cobra.Command{
		Use:   "relay",
		Short: "Relay operations",
	}

	relayListCmd := &cobra.Command{
		Use:   "list",
		Short: "List relays discovered from SDK databases",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp()
			if app == nil {
				return newError("app not initialized", nil)
			}

			if err := writeRelayList(cmd.OutOrStdout(), app); err != nil {
				return newError("failed to list relays", err)
			}
			return nil
		},
	}

	relayCmd.AddCommand(relayListCmd)
	RegisterCommandGroup("Relay", "Relay operations", relayCmd)
}

func writeRelayList(w io.Writer, app *config.AppContext) error {
	sys := app.System()
	if sys == nil {
		return nil
	}

	relays, err := sys.ListKnownEventRelays()
	if err != nil {
		return err
	}

	for _, relay := range relays {
		if _, err := fmt.Fprintln(w, relay); err != nil {
			return err
		}
	}

	return nil
}
