package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/jerry-harm/nosmec/cmd/completion"
	"github.com/jerry-harm/nosmec/utils"
	"github.com/spf13/cobra"
)

func registerSubscribeCommands() {
	subCmd := &cobra.Command{
		Use:   "subscribe",
		Short: "Subscription operations",
	}

	subListCmd := &cobra.Command{
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

	subAddUserCmd := &cobra.Command{
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

			fmt.Printf("Subscribed to user: %s\n", identifier)
		},
	}

	subAddCommunityCmd := &cobra.Command{
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

			fmt.Printf("Subscribed to community: %s\n", addr)
		},
	}

	subAddHashtagCmd := &cobra.Command{
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

			fmt.Printf("Subscribed to hashtag: %s\n", tag)
		},
	}

	subRemoveCmd := &cobra.Command{
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

			fmt.Printf("Removed subscription: %s (%s)\n", id, subType)
		},
	}
	subRemoveCmd.Flags().String("type", "user", "Subscription type: user, community, hashtag")
	subRemoveCmd.RegisterFlagCompletionFunc("type", completion.SubTypeCompletionFunc)

	subSyncCmd := &cobra.Command{
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

	subPublishCmd := &cobra.Command{
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

	subTimelineCmd := &cobra.Command{
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
			events, err := utils.GetFollowedTimeline(ctx, limit, hashtags, &utils.GetOptions{App: app})
			if err != nil {
				handleError(newError("failed to get timeline", err))
			}

			if len(events) == 0 {
				fmt.Println("No events found.")
				return
			}

			printTimeline(events)
		},
	}
	subTimelineCmd.Flags().IntP("limit", "n", 10, "Number of events")
	subTimelineCmd.Flags().StringSliceP("hashtag", "t", nil, "Filter by hashtags")

	subTimelineCmd.RegisterFlagCompletionFunc("limit", completion.LimitCompletionFunc)
	subTimelineCmd.RegisterFlagCompletionFunc("hashtag", completion.HashtagCompletionFunc)

	subCmd.AddCommand(subListCmd)
	subCmd.AddCommand(subAddUserCmd)
	subCmd.AddCommand(subAddCommunityCmd)
	subCmd.AddCommand(subAddHashtagCmd)
	subCmd.AddCommand(subRemoveCmd)
	subCmd.AddCommand(subSyncCmd)
	subCmd.AddCommand(subPublishCmd)
	subCmd.AddCommand(subTimelineCmd)

	RegisterCommandGroup("Subscribe", "Subscription operations", subCmd)
}
