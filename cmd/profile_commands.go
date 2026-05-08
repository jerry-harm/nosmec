package cmd

import (
	"context"
	"os"

	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/cmd/completion"
	"github.com/jerry-harm/nosmec/utils"
	"github.com/spf13/cobra"
)

func registerProfileCommands() {
	profileCmd := &cobra.Command{
		Use:               "profile [npub or pubkey]",
		Short:             "Get user profile",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completion.PubKeyCompletionFunc,
		Run: func(cmd *cobra.Command, args []string) {
			app := getApp()
			ctx := context.Background()

			var pubKey nostr.PubKey
			var err error

			if len(args) == 0 {
				pubKey, err = app.GetMyPubKey()
				if err != nil {
					handleError(newError("failed to get public key", err))
				}
			} else {
				pubKey, err = utils.ResolveAliasToPubKey(app, args[0])
				if err != nil {
					handleError(newError("failed to parse identifier", err))
				}
			}

			full, _ := cmd.Flags().GetBool("full")

			if full {
				fp, err := utils.GetFullProfile(ctx, pubKey, &utils.GetOptions{App: app})
				if err != nil {
					handleError(newError("failed to get full profile", err))
				}

				data, err := utils.SerializeProfile(fp)
				if err != nil {
					handleError(newError("failed to serialize profile", err))
				}
				os.Stdout.Write(data)
				os.Stdout.Write([]byte("\n"))
			} else {
				event := utils.GetProfile(ctx, pubKey, &utils.GetOptions{App: app})
				if event == nil {
					handleError(newError("profile not found", nil))
				}
				utils.PrintEvent(event, false)
			}
		},
	}

	profileCmd.Flags().Bool("full", false, "Show full profile including relays, dm_relays, and follows")

	RegisterCommandGroup("Profile", "Profile operations", profileCmd)
}