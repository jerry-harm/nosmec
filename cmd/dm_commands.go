package cmd

import (
	"context"
	"fmt"
	"time"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"github.com/jerry-harm/nosmec/tui/dm"
	"github.com/jerry-harm/nosmec/utils"
	"github.com/spf13/cobra"
)

func registerDMCommands() {
	dmCmd := &cobra.Command{
		Use:   "dm [npub]",
		Short: "Direct messages (or open DM TUI with a specific user)",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				return
			}
			npubOrHex := args[0]
			if err := dm.RunDM(getApp(), npubOrHex); err != nil {
				handleError(newError("failed to open DM TUI", err))
			}
		},
	}

	dmSendCmd := &cobra.Command{
		Use:   "send <recipient> <message>",
		Short: "Send a DM",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			recipientStr := args[0]
			content := args[1]

			_, decoded, err := nip19.Decode(recipientStr)
			if err != nil {
				handleError(newError("invalid recipient pubkey", err))
			}
			recipientPubKey, ok := decoded.(nostr.PubKey)
			if !ok {
				handleError(newError("invalid recipient pubkey format", nil))
			}

			ctx := context.Background()
			if err := utils.SendDM(ctx, getApp(), recipientPubKey, content); err != nil {
				handleError(newError("failed to send DM", err))
			}

			fmt.Printf("DM sent to %s\n", nip19.EncodeNpub(recipientPubKey)[:32]+"...")
		},
	}

	dmListCmd := &cobra.Command{
		Use:   "list",
		Short: "List recent conversations",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			limit := 20
			if l, err := cmd.Flags().GetInt("limit"); err == nil && l > 0 {
				limit = l
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			conversations, err := utils.ListDMConversations(ctx, getApp(), limit)
			if err != nil {
				handleError(newError("failed to list conversations", err))
			}

			if len(conversations) == 0 {
				fmt.Println("No DM conversations found.")
				return
			}

			fmt.Println("Recent DM conversations:")
			fmt.Println()
			for _, conv := range conversations {
				prefix := "←"
				if conv.LatestDM.FromMe {
					prefix = "→"
				}
				name := conv.PubKey[:16] + "..."
				if otherPK, err := nostr.PubKeyFromHex(conv.PubKey); err == nil {
					if profileName := utils.GetProfileName(ctx, otherPK, &utils.GetOptions{App: getApp()}); profileName != "" {
						name = profileName
					}
				}
				fmt.Printf("[%s] %s\n", conv.LatestDM.Timestamp.Time().Format("2006-01-02 15:04"), name)
				fmt.Printf("  %s %s\n", prefix, conv.LatestDM.Content)
				fmt.Println()
			}
		},
	}
	dmListCmd.Flags().IntP("limit", "n", 20, "Number of conversations to show")

	dmHistoryCmd := &cobra.Command{
		Use:   "history <recipient>",
		Short: "View DM history with a user",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, decoded, err := nip19.Decode(args[0])
			if err != nil {
				handleError(newError("invalid recipient pubkey", err))
			}
			recipientPubKey, ok := decoded.(nostr.PubKey)
			if !ok {
				handleError(newError("invalid recipient pubkey format", nil))
			}

			limit := 50
			if l, err := cmd.Flags().GetInt("limit"); err == nil && l > 0 {
				limit = l
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			messages, err := utils.QueryDMHistory(ctx, getApp(), recipientPubKey, limit)
			if err != nil {
				handleError(newError("failed to query DM history", err))
			}

			if len(messages) == 0 {
				fmt.Printf("No DM history with %s.\n", nip19.EncodeNpub(recipientPubKey)[:32]+"...")
				return
			}

			recipientName := nip19.EncodeNpub(recipientPubKey)[:32] + "..."
			if profileName := utils.GetProfileName(ctx, recipientPubKey, &utils.GetOptions{App: getApp()}); profileName != "" {
				recipientName = profileName
			}

			fmt.Printf("=== DM History with %s ===\n", recipientName)
			fmt.Println()

			for _, msg := range messages {
				prefix := "←"
				if msg.FromMe {
					prefix = "→"
				}
				fmt.Printf("[%s] %s\n", msg.Timestamp.Time().Format("15:04:05"), prefix)
				fmt.Printf("  %s\n", msg.Content)
				fmt.Println()
			}
		},
	}
	dmHistoryCmd.Flags().IntP("limit", "n", 50, "Number of messages to show")

	dmListenCmd := &cobra.Command{
		Use:   "listen",
		Short: "Listen for incoming DMs",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			app := getApp()
			dmRelays := app.ListDMRelays()
			if len(dmRelays) == 0 {
				dmRelays = app.ReadableRelays()
			}
			if len(dmRelays) == 0 {
				handleError(newError("no DM relays configured", nil))
			}

			fmt.Printf("Listening on %d relays...\n", len(dmRelays))
			fmt.Println("Press Ctrl+C to stop")

			since := nostr.Timestamp(time.Now().Unix() - 86400)
			ch := utils.ListenForDMs(ctx, app, since)

			for rumor := range ch {
				fmt.Printf("\n[%s] %s\n", rumor.CreatedAt.Time().Format("15:04:05"), nip19.EncodeNpub(rumor.PubKey)[:32]+"...")
				fmt.Printf("  %s\n", rumor.Content)
			}
		},
	}

	dmCmd.AddCommand(dmSendCmd)
	dmCmd.AddCommand(dmListCmd)
	dmCmd.AddCommand(dmHistoryCmd)
	dmCmd.AddCommand(dmListenCmd)

	RegisterCommandGroup("DM", "Direct messages", dmCmd)
}
