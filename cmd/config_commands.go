package cmd

import (
	"context"
	"fmt"
	"sort"

	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/cmd/completion"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/utils"
	"github.com/spf13/cobra"
	"fiatjaf.com/nostr/nip19"
)

func registerConfigCommands() {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
	}

	configPubkeyCmd := &cobra.Command{
		Use:   "pubkey",
		Short: "Show your public key",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			app := getApp()
			pubKey, err := app.GetMyPubKey()
			if err != nil {
				handleError(newError("failed to get public key", err))
			}

			fmt.Printf("Public Key:\n")
			fmt.Printf("  Hex:  %s\n", pubKey.Hex())
			fmt.Printf("  NPub: %s\n", nip19.EncodeNpub(pubKey))
		},
	}

	profileCmd := &cobra.Command{
		Use:   "profile",
		Short: "Profile operations (NIP-01)",
	}

	profileSyncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync profile from network to local config",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			app := getApp()
			ctx := context.Background()

			if err := utils.SyncProfile(ctx, app); err != nil {
				handleError(newError("failed to sync profile", err))
			}

			fmt.Println("Profile synced from network")
		},
	}

	profileSetCmd := &cobra.Command{
		Use:   "set",
		Short: "Set your profile (reads current config, merges with flags, publishes)",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			about, _ := cmd.Flags().GetString("about")
			picture, _ := cmd.Flags().GetString("picture")
			displayName, _ := cmd.Flags().GetString("display-name")
			website, _ := cmd.Flags().GetString("website")
			banner, _ := cmd.Flags().GetString("banner")
			bot, _ := cmd.Flags().GetString("bot")
			birthday, _ := cmd.Flags().GetString("birthday")
			nip05, _ := cmd.Flags().GetString("nip05")
			lud06, _ := cmd.Flags().GetString("lud06")
			lud16, _ := cmd.Flags().GetString("lud16")

			ctx := context.Background()
			event, err := utils.SetProfile(ctx, getApp(), false, name, about, picture, displayName, website, banner, bot, birthday, nip05, lud06, lud16)
			if err != nil {
				handleError(newError("failed to set profile", err))
			}

			fmt.Printf("Profile updated!\n")
			fmt.Printf("Event ID: %s\n", event.ID.Hex())
		},
	}

	profileSetCmd.Flags().String("name", "", "Display name (empty to clear)")
	profileSetCmd.Flags().String("about", "", "About text (empty to clear)")
	profileSetCmd.Flags().String("picture", "", "Picture URL (empty to clear)")
	profileSetCmd.Flags().String("display-name", "", "Display name alternative (empty to clear)")
	profileSetCmd.Flags().String("website", "", "Website URL (empty to clear)")
	profileSetCmd.Flags().String("banner", "", "Banner URL (empty to clear)")
	profileSetCmd.Flags().String("bot", "", "Bot flag true/false (empty to clear)")
	profileSetCmd.Flags().String("birthday", "", "Birthday YYYY-MM-DD (empty to clear)")
	profileSetCmd.Flags().String("nip05", "", "NIP-05 identifier (empty to clear)")
	profileSetCmd.Flags().String("lud06", "", "Lightning address lud06 (empty to clear)")
	profileSetCmd.Flags().String("lud16", "", "Lightning address lud16 (empty to clear)")
	profileSetCmd.RegisterFlagCompletionFunc("bot", completion.BotFlagCompletionFunc)

	profileCmd.AddCommand(profileSetCmd)
	profileCmd.AddCommand(profileSyncCmd)

	configShowCmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cfg := getApp().Config()

			fmt.Println("=== nosmec Configuration ===")

				fmt.Printf("Data Directory: %s\n", cfg.LocalRelay.DataDir)
			fmt.Printf("Config Directory: %s\n", cfg.ConfigDir)
			fmt.Printf("Private Key: %s\n", maskString(cfg.PrivateKey, 8))

			fmt.Println("\n--- Relay List ---")
			if len(cfg.RelayList) == 0 {
				fmt.Println("  (none)")
			}
			for _, r := range cfg.RelayList {
				read := "false"
				if r.Read != nil && *r.Read {
					read = "true"
				}
				write := "false"
				if r.Write != nil && *r.Write {
					write = "true"
				}
				fmt.Printf("  %s  read=%s write=%s\n", r.URL, read, write)
			}

			fmt.Println("\n--- DM Relays ---")
			if len(cfg.DMRelays) == 0 {
				fmt.Println("  (none)")
			}
			for _, url := range cfg.DMRelays {
				fmt.Printf("  %s\n", url)
			}

			fmt.Println("\n--- Search Relays ---")
			if len(cfg.SearchRelays) == 0 {
				fmt.Println("  (none)")
			}
			for _, url := range cfg.SearchRelays {
				fmt.Printf("  %s\n", url)
			}

			fmt.Println("\n--- Known Relays ---")
			if len(cfg.KnownRelays) == 0 {
				fmt.Println("  (none)")
			}
			for _, url := range cfg.KnownRelays {
				fmt.Printf("  %s\n", url)
			}

			fmt.Println("\n--- Aliases ---")
			if cfg.Alias == nil || len(cfg.Alias) == 0 {
				fmt.Println("  (none)")
			}
			for k, v := range cfg.Alias {
				fmt.Printf("  %s -> %s\n", k, v)
			}

			fmt.Println("\n--- Proxy ---")
			fmt.Printf("  SOCKS: %s\n", maskString(cfg.Proxy.Socks, 0))
			fmt.Printf("  I2P SOCKS: %s\n", maskString(cfg.Proxy.I2PSocks, 0))

			fmt.Println()
		},
	}

	configSetCmd := &cobra.Command{
		Use:               "set <key> <value>",
		Short:             "Set a configuration value",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ConfigKeyCompletionFunc,
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			value := args[1]

			config.GetViper().Set(key, value)

			if err := config.GetViper().WriteConfig(); err != nil {
				handleError(newError("failed to write config", err))
			}

			reloadApp()
			fmt.Printf("Set %s = %s\n", key, value)
		},
	}

	configRelayCmd := &cobra.Command{
		Use:   "relay",
		Short: "Manage relays",
	}

	configRelayListCmd := &cobra.Command{
		Use:   "list",
		Short: "List relays",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			relays := getApp().ListRelays()
			if len(relays) == 0 {
				fmt.Println("No relays configured.")
				return
			}

			fmt.Println("=== Relays ===")
			for _, r := range relays {
				read := "false"
				if r.Read != nil && *r.Read {
					read = "true"
				}
				write := "false"
				if r.Write != nil && *r.Write {
					write = "true"
				}
				fmt.Printf("  %s  read=%s write=%s\n", r.URL, read, write)
			}
		},
	}

	configRelayAddCmd := &cobra.Command{
		Use:   "add <url>",
		Short: "Add a relay",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			url := args[0]

			if err := getApp().AddRelay(url, false, false); err != nil {
				handleError(err)
			}
			fmt.Printf("Relay added: %s\n", url)
		},
	}

	configRelayRemoveCmd := &cobra.Command{
		Use:   "remove <url>",
		Short: "Remove a relay",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			url := args[0]

			if err := getApp().RemoveRelay(url); err != nil {
				handleError(err)
			}
			fmt.Printf("Relay removed: %s\n", url)
		},
	}

	configRelaySyncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync relay list from network",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			app := getApp()

			pk, err := app.GetMyPubKey()
			if err != nil {
				handleError(newError("failed to get public key", err))
			}

			relays := app.WritableRelays()
			if len(relays) == 0 {
				relays = app.ReadableRelays()
			}
			if len(relays) == 0 {
				relays = app.Config().KnownRelays
			}

			filter := nostr.Filter{
				Kinds:   []nostr.Kind{nostr.KindRelayListMetadata},
				Authors: []nostr.PubKey{pk},
				Limit:   1,
			}

			result := app.Pool().QuerySingle(ctx, relays, filter, nostr.SubscriptionOptions{})
			if result == nil {
				handleError(newError("relay list not found on network", nil))
			}

			event := result.Event
			relayList := make([]config.Relay, 0)
			for _, tag := range event.Tags {
				if len(tag) >= 2 && tag[0] == "r" {
					url := tag[1]
					relay := config.Relay{
						URL:   url,
						Read:  config.BoolPtr(false),
						Write: config.BoolPtr(false),
					}
					for _, p := range tag[2:] {
						if p == "read" {
							relay.Read = config.BoolPtr(true)
						} else if p == "write" {
							relay.Write = config.BoolPtr(true)
						}
					}
					relayList = append(relayList, relay)
				}
			}

			app.SyncRelayList(relayList)
			fmt.Printf("Synced %d relays from network\n", len(relayList))
		},
	}

	configRelayCmd.AddCommand(configRelayListCmd)
	configRelayCmd.AddCommand(configRelayAddCmd)
	configRelayCmd.AddCommand(configRelayRemoveCmd)
	configRelayCmd.AddCommand(configRelaySyncCmd)

	configAliasCmd := &cobra.Command{
		Use:   "alias",
		Short: "Manage aliases",
	}

	configAliasListCmd := &cobra.Command{
		Use:   "list",
		Short: "List aliases",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			aliases := utils.ListAliases(getApp())
			if len(aliases) == 0 {
				fmt.Println("No aliases configured.")
				return
			}

			var keys []string
			for k := range aliases {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			fmt.Println("Configured aliases:")
			for _, alias := range keys {
				fmt.Printf("%-20s -> %s\n", alias, aliases[alias])
			}
		},
	}

	configAliasAddCmd := &cobra.Command{
		Use:   "add <alias> <value>",
		Short: "Add an alias",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			alias := args[0]
			value := args[1]
			getApp().AddAlias(alias, value)
			fmt.Printf("Alias added: %s -> %s\n", alias, value)
		},
	}

	configAliasRemoveCmd := &cobra.Command{
		Use:               "remove <alias>",
		Short:             "Remove an alias",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.AliasCompletionFunc,
		Run: func(cmd *cobra.Command, args []string) {
			alias := args[0]
			cfg := getApp().Config()
			if cfg.Alias == nil {
				handleError(newError("alias not found", nil))
			}
			delete(cfg.Alias, alias)
			config.GetViper().Set("alias", cfg.Alias)
			if err := config.GetViper().WriteConfig(); err != nil {
				handleError(newError("failed to write config", err))
			}
			reloadApp()
			fmt.Printf("Alias removed: %s\n", alias)
		},
	}

	configAliasCmd.AddCommand(configAliasListCmd)
	configAliasCmd.AddCommand(configAliasAddCmd)
	configAliasCmd.AddCommand(configAliasRemoveCmd)

	configSearchRelayCmd := &cobra.Command{
		Use:   "search-relay",
		Short: "Manage search relays",
	}

	configSearchRelayListCmd := &cobra.Command{
		Use:   "list",
		Short: "List search relays",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			relays := getApp().ListSearchRelays()
			if len(relays) == 0 {
				fmt.Println("No search relays configured.")
				return
			}
			fmt.Println("Search relays:")
			for _, url := range relays {
				fmt.Printf("  %s\n", url)
			}
		},
	}

	configSearchRelayAddCmd := &cobra.Command{
		Use:   "add <url>",
		Short: "Add search relay",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := getApp().AddSearchRelay(args[0]); err != nil {
				handleError(err)
			}
			fmt.Printf("Search relay added: %s\n", args[0])
		},
	}

	configSearchRelayRemoveCmd := &cobra.Command{
		Use:   "remove <url>",
		Short: "Remove search relay",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := getApp().RemoveSearchRelay(args[0]); err != nil {
				handleError(err)
			}
			fmt.Printf("Search relay removed: %s\n", args[0])
		},
	}

	configSearchRelayCmd.AddCommand(configSearchRelayListCmd)
	configSearchRelayCmd.AddCommand(configSearchRelayAddCmd)
	configSearchRelayCmd.AddCommand(configSearchRelayRemoveCmd)

	configDMRelayCmd := &cobra.Command{
		Use:   "dm-relay",
		Short: "Manage DM relays",
	}

	configDMRelayListCmd := &cobra.Command{
		Use:   "list",
		Short: "List DM relays",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			relays := getApp().ListDMRelays()
			if len(relays) == 0 {
				fmt.Println("No DM relays configured.")
				return
			}
			fmt.Println("DM relays:")
			for _, url := range relays {
				fmt.Printf("  %s\n", url)
			}
		},
	}

	configDMRelayAddCmd := &cobra.Command{
		Use:   "add <url>",
		Short: "Add DM relay",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := getApp().AddDMRelay(args[0]); err != nil {
				handleError(err)
			}
			fmt.Printf("DM relay added: %s\n", args[0])
		},
	}

	configDMRelayRemoveCmd := &cobra.Command{
		Use:   "remove <url>",
		Short: "Remove DM relay",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := getApp().RemoveDMRelay(args[0]); err != nil {
				handleError(err)
			}
			fmt.Printf("DM relay removed: %s\n", args[0])
		},
	}

	configDMRelayCmd.AddCommand(configDMRelayListCmd)
	configDMRelayCmd.AddCommand(configDMRelayAddCmd)
	configDMRelayCmd.AddCommand(configDMRelayRemoveCmd)

	configSyncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync all config from network",
		Long: `Sync profile, subscriptions, and relay list from the network.

This will overwrite your local configuration with data from relays.
- Profile (Kind 0)
- Subscriptions (Kind 3, 10004, 10015)
- Relay list (Kind 10002, 10050)`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			app := getApp()
			ctx := context.Background()

			if err := utils.SyncAll(ctx, app); err != nil {
				handleError(newError("failed to sync config", err))
			}

			fmt.Println("All config synced from network")
		},
	}

	configPublishCmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish all config to network",
		Long: `Publish profile, subscriptions, and relay list to the network.
- Profile (Kind 0)
- Subscriptions (Kind 3, 10004, 10015)
- Relay list (Kind 10002, 10050)`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			app := getApp()
			ctx := context.Background()

			if err := utils.PublishAll(ctx, app); err != nil {
				handleError(newError("failed to publish config", err))
			}

			fmt.Println("All config published to network")
		},
	}

	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPubkeyCmd)
	configCmd.AddCommand(profileCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configRelayCmd)
	configCmd.AddCommand(configAliasCmd)
	configCmd.AddCommand(configSearchRelayCmd)
	configCmd.AddCommand(configDMRelayCmd)
	configCmd.AddCommand(configSyncCmd)
	configCmd.AddCommand(configPublishCmd)

	RegisterCommandGroup("Config", "Configuration management", configCmd)
}

func maskString(s string, visible int) string {
	if s == "" {
		return "(not set)"
	}
	if visible == 0 || len(s) <= visible {
		return "***"
	}
	return s[:visible] + "***"
}
