package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"github.com/jerry-harm/nosmec/cmd/completion"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/utils"
	"github.com/spf13/cobra"
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
			fmt.Printf("Event ID: %s\n", nip19.EncodeNevent(event.ID, nil, event.PubKey))
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

			fmt.Printf("Data Directory: %s\n", cfg.DataDir)
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
			ctx := context.Background()
			if err := utils.PublishRelayList(ctx, getApp()); err != nil {
				handleError(newError("failed to publish relay list", err))
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
			ctx := context.Background()
			if err := utils.PublishRelayList(ctx, getApp()); err != nil {
				handleError(newError("failed to publish relay list", err))
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
			if err := utils.PublishRelayList(ctx, app); err != nil {
				handleError(newError("failed to publish relay list", err))
			}
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
			ctx := context.Background()
			if err := utils.PublishRelayList(ctx, getApp()); err != nil {
				handleError(newError("failed to publish relay list", err))
			}
			fmt.Printf("DM relay added: %s\n", args[0])
		},
	}

	configDMRelayRemoveCmd := &cobra.Command{
		Use:   "remove <url>",
		Short: "Remove a DM relay",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := getApp().RemoveDMRelay(args[0]); err != nil {
				handleError(err)
			}
			ctx := context.Background()
			if err := utils.PublishRelayList(ctx, getApp()); err != nil {
				handleError(newError("failed to publish relay list", err))
			}
			fmt.Printf("DM relay removed: %s\n", args[0])
		},
	}

	configDMRelayCmd.AddCommand(configDMRelayListCmd)
	configDMRelayCmd.AddCommand(configDMRelayAddCmd)
	configDMRelayCmd.AddCommand(configDMRelayRemoveCmd)

	configDMRelaySyncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync DM relays from network",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			if err := utils.SyncDMRelaysFromNetwork(ctx, getApp()); err != nil {
				handleError(newError("failed to sync DM relays", err))
			}
			fmt.Println("DM relays synced from network")
		},
	}
	configDMRelayCmd.AddCommand(configDMRelaySyncCmd)

	configSubscribeCmd := &cobra.Command{
		Use:   "subscribe",
		Short: "Manage subscriptions (followed users, communities, hashtags)",
	}

	configSubscribeListCmd := &cobra.Command{
		Use:   "list",
		Short: "List subscriptions",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			app := getApp()

			users := app.ListSubscriptions("user")
			communities := app.ListSubscriptions("community")
			hashtags := app.ListSubscriptions("hashtag")

			if len(users) == 0 && len(communities) == 0 && len(hashtags) == 0 {
				fmt.Println("No subscriptions.")
				return
			}

			fmt.Println("=== Subscriptions ===")

			if len(users) > 0 {
				fmt.Println("\n[Users]")
				for _, s := range users {
					petname := ""
					if s.Petname != "" {
						petname = " (" + s.Petname + ")"
					}
					relay := ""
					if s.Relay != "" {
						relay = " @ " + s.Relay
					}
					fmt.Printf("  - %s%s%s\n", s.ID, relay, petname)
				}
			}

			if len(communities) > 0 {
				fmt.Println("\n[Communities]")
				for _, s := range communities {
					relay := ""
					if s.Relay != "" {
						relay = " @ " + s.Relay
					}
					fmt.Printf("  - %s%s\n", s.ID, relay)
				}
			}

			if len(hashtags) > 0 {
				fmt.Println("\n[Hashtags]")
				for _, s := range hashtags {
					fmt.Printf("  - #%s\n", s.ID)
				}
			}
		},
	}

	configSubscribeAddUserCmd := &cobra.Command{
		Use:               "user <identifier> [relay] [petname]",
		Short:             "Subscribe to a user",
		Args:              cobra.RangeArgs(1, 3),
		ValidArgsFunction: completion.PubKeyCompletionFunc,
		Run: func(cmd *cobra.Command, args []string) {
			identifier := args[0]
			relay := ""
			petname := ""
			if len(args) > 1 {
				relay = args[1]
			}
			if len(args) > 2 {
				petname = args[2]
			}

			ctx := context.Background()
			if err := utils.FollowUser(ctx, getApp(), identifier, relay, petname); err != nil {
				handleError(newError("failed to subscribe", err))
			}
			if err := utils.PublishSubscriptions(ctx, getApp()); err != nil {
				handleError(newError("failed to publish subscriptions", err))
			}

			fmt.Printf("Subscribed to user: %s\n", identifier)
		},
	}

	configSubscribeAddCommunityCmd := &cobra.Command{
		Use:               "community <address> [relay]",
		Short:             "Subscribe to a community",
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: completion.CommunityCompletionFunc,
		Run: func(cmd *cobra.Command, args []string) {
			addr := args[0]
			relay := ""
			if len(args) > 1 {
				relay = args[1]
			}

			ctx := context.Background()
			if err := utils.FollowCommunity(ctx, getApp(), addr, relay); err != nil {
				handleError(newError("failed to subscribe", err))
			}
			if err := utils.PublishSubscriptions(ctx, getApp()); err != nil {
				handleError(newError("failed to publish subscriptions", err))
			}

			fmt.Printf("Subscribed to community: %s\n", addr)
		},
	}

	configSubscribeAddHashtagCmd := &cobra.Command{
		Use:               "hashtag <tag>",
		Short:             "Subscribe to a hashtag",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.HashtagCompletionFunc,
		Run: func(cmd *cobra.Command, args []string) {
			tag := args[0]
			ctx := context.Background()

			if err := utils.FollowHashtag(ctx, getApp(), tag); err != nil {
				handleError(newError("failed to subscribe", err))
			}
			if err := utils.PublishSubscriptions(ctx, getApp()); err != nil {
				handleError(newError("failed to publish subscriptions", err))
			}

			fmt.Printf("Subscribed to hashtag: %s\n", tag)
		},
	}

	configSubscribeRemoveCmd := &cobra.Command{
		Use:               "remove <id>",
		Short:             "Remove a subscription",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.CommunityCompletionFunc,
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			subType, _ := cmd.Flags().GetString("type")

			switch subType {
			case "user":
				ctx := context.Background()
				if err := utils.UnfollowUser(ctx, getApp(), id); err != nil {
					handleError(newError("failed to remove subscription", err))
				}
			case "community":
				ctx := context.Background()
				if err := utils.UnfollowCommunity(ctx, getApp(), id); err != nil {
					handleError(newError("failed to remove subscription", err))
				}
			case "hashtag":
				ctx := context.Background()
				if err := utils.UnfollowHashtag(ctx, getApp(), id); err != nil {
					handleError(newError("failed to remove subscription", err))
				}
			default:
				handleError(newError("invalid subscription type: "+subType, nil))
			}

			ctx := context.Background()
			if err := utils.PublishSubscriptions(ctx, getApp()); err != nil {
				handleError(newError("failed to publish subscriptions", err))
			}

			fmt.Printf("Removed subscription: %s (%s)\n", id, subType)
		},
	}
	configSubscribeRemoveCmd.Flags().String("type", "user", "Subscription type: user, community, hashtag")
	configSubscribeRemoveCmd.RegisterFlagCompletionFunc("type", completion.SubTypeCompletionFunc)

	configSubscribeSyncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync subscriptions from network",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			if err := utils.SyncSubscriptionsFromNetwork(ctx, getApp()); err != nil {
				handleError(newError("failed to sync subscriptions", err))
			}
			fmt.Println("Subscriptions synced from network")
		},
	}

	configSubscribePublishCmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish subscriptions to network",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			if err := utils.PublishSubscriptions(ctx, getApp()); err != nil {
				handleError(newError("failed to publish subscriptions", err))
			}
			fmt.Println("Subscriptions published to network")
		},
	}

	configSubscribeTimelineCmd := &cobra.Command{
		Use:   "timeline",
		Short: "Show followed timeline",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			limit := 10
			if l, err := cmd.Flags().GetInt("limit"); err == nil && l > 0 {
				limit = l
			}

			var hashtags []string
			if h, err := cmd.Flags().GetStringSlice("hashtag"); err == nil && len(h) > 0 {
				hashtags = h
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			app := getApp()

			// Collect followed pubkeys and community addresses
			var pubkeys []nostr.PubKey
			var communityAddrs []string
			for _, s := range app.ListSubscriptions("user") {
				if pk, err := utils.ResolveAliasToPubKey(app, s.ID); err == nil {
					pubkeys = append(pubkeys, pk)
				}
			}
			for _, s := range app.ListSubscriptions("community") {
				communityAddrs = append(communityAddrs, s.ID)
			}

			events, err := app.System().FetchFollowedTimelinePage(ctx, pubkeys, communityAddrs, limit, 0)
			if err != nil {
				handleError(newError("failed to fetch timeline", err))
				return
			}

			if len(events) == 0 {
				fmt.Println("No events found.")
				return
			}

			// Filter by hashtags if specified
			if len(hashtags) > 0 {
				filtered := make([]nostr.Event, 0)
			eventLoop:
				for _, ev := range events {
					for _, tag := range ev.Tags {
						if len(tag) >= 2 && tag[0] == "t" {
							for _, ht := range hashtags {
								if strings.EqualFold(tag[1], ht) {
									filtered = append(filtered, ev)
									continue eventLoop
								}
							}
						}
					}
				}
				events = filtered
			}

			if len(events) == 0 {
				fmt.Println("No events found.")
				return
			}

			var timelineEvents []timelineEvent
			for _, ev := range events {
				te := timelineEvent{Event: ev}
				if aTag := ev.Tags.Find("a"); len(aTag) > 1 {
					if strings.HasPrefix(aTag[1], "34550:") {
						te.IsCommunity = true
						te.CommunityID = aTag[1]
					}
				}
				timelineEvents = append(timelineEvents, te)
			}

			printTimeline(timelineEvents)
		},
	}
	configSubscribeTimelineCmd.Flags().IntP("limit", "n", 10, "Number of events")
	configSubscribeTimelineCmd.Flags().StringSliceP("hashtag", "t", nil, "Filter by hashtags")

	configSubscribeTimelineCmd.RegisterFlagCompletionFunc("limit", completion.LimitCompletionFunc)
	configSubscribeTimelineCmd.RegisterFlagCompletionFunc("hashtag", completion.HashtagCompletionFunc)

	configSubscribeCmd.AddCommand(configSubscribeListCmd)
	configSubscribeCmd.AddCommand(configSubscribeAddUserCmd)
	configSubscribeCmd.AddCommand(configSubscribeAddCommunityCmd)
	configSubscribeCmd.AddCommand(configSubscribeAddHashtagCmd)
	configSubscribeCmd.AddCommand(configSubscribeRemoveCmd)
	configSubscribeCmd.AddCommand(configSubscribeSyncCmd)
	configSubscribeCmd.AddCommand(configSubscribePublishCmd)
	configSubscribeCmd.AddCommand(configSubscribeTimelineCmd)

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
	configCmd.AddCommand(configSubscribeCmd)
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
