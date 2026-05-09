package cmd

import (
	"context"
	"fmt"
	"strings"

	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/cmd/completion"
	"github.com/jerry-harm/nosmec/utils"
	"github.com/spf13/cobra"
)

func registerCommunityCommands() {
	communityCmd := &cobra.Command{
		Use:   "community",
		Short: "Community operations (NIP-72)",
	}

	communityCreateCmd := &cobra.Command{
		Use:   "create <name> <id> [description]",
		Short: "Create a community",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			communityID := args[1]
			description := ""
			if len(args) > 2 {
				description = args[2]
			}

			imageURL, _ := cmd.Flags().GetString("image")

			def := utils.CommunityDefinition{
				Name:        name,
				Description: description,
				ImageURL:    imageURL,
				Moderators:  []nostr.PubKey{},
				Relays:      make(map[string]string),
				ID:          communityID,
			}

			app := getApp()
			pubKey, err := app.GetMyPubKey()
			if err != nil {
				handleError(newError("failed to get public key", err))
			}
			def.Moderators = append(def.Moderators, pubKey)

			ctx := context.Background()
			event, err := utils.CreateCommunity(ctx, app, def)
			if err != nil {
				handleError(newError("failed to create community", err))
			}

			fmt.Printf("Community created!\n")
			if dTag := event.Tags.Find("d"); len(dTag) > 1 {
				fmt.Printf("ID: %s\n", dTag[1])
			}
			fmt.Printf("Name: %s\n", name)
			if description != "" {
				fmt.Printf("Description: %s\n", description)
			}
			fmt.Printf("Event ID: %s\n", event.ID.Hex())
		},
	}
	communityCreateCmd.Flags().String("image", "", "Community image URL")

	communityPostCmd := &cobra.Command{
		Use:               "post <community-addr> <content>",
		Short:             "Post to a community",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.CommunityCompletionFunc,
		Run: func(cmd *cobra.Command, args []string) {
			communityAddr := args[0]
			content := args[1]

			app := getApp()
			ctx := context.Background()

			event, err := utils.PostToCommunity(ctx, app, communityAddr, content, "")
			if err != nil {
				handleError(newError("failed to post", err))
			}

			fmt.Printf("Posted to community!\n")
			fmt.Printf("Post ID: %s\n", event.ID.Hex())
		},
	}

	communityReplyCmd := &cobra.Command{
		Use:   "reply <post-id> <content>",
		Short: "Reply to a post",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			postID := args[0]
			content := args[1]

			ctx := context.Background()
			event, err := utils.ReplyToCommunity(ctx, getApp(), postID, content)
			if err != nil {
				handleError(newError("failed to reply", err))
			}

			fmt.Printf("Replied!\n")
			fmt.Printf("Reply ID: %s\n", event.ID.Hex())
		},
	}

	communityListCmd := &cobra.Command{
		Use:   "list",
		Short: "List communities",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			app := getApp()

			fmt.Println("=== My Communities ===")
			fmt.Println()

			fmt.Println("[Following] (Kind 10004)")
			followed, err := utils.GetFollowedCommunities(ctx, app)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else if len(followed) == 0 {
				fmt.Println("  (none)")
			} else {
				for _, addr := range followed {
					fmt.Printf("  - %s\n", addr)
				}
			}
			fmt.Println()

			fmt.Println("[Created] (Kind 34550)")
			myPubKey, err := app.GetMyPubKey()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				createdEvents, err := utils.GetMyCreatedCommunities(ctx, app, myPubKey)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				} else if len(createdEvents) == 0 {
					fmt.Println("  (none)")
				} else {
					for _, event := range createdEvents {
						name := ""
						if nameTag := event.Tags.Find("name"); len(nameTag) > 1 {
							name = nameTag[1]
						} else if dTag := event.Tags.Find("d"); len(dTag) > 1 {
							name = dTag[1]
						}
						if name == "" {
							name = event.ID.Hex()[:16] + "..."
						}
						fmt.Printf("  - %s\n", name)
					}
				}
			}
			fmt.Println()

			fmt.Println("[Posted] (Kind 1111)")
			postedAddrs, err := utils.GetPostedCommunities(ctx, app, myPubKey)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else if len(postedAddrs) == 0 {
				fmt.Println("  (none)")
			} else {
				seen := make(map[string]bool)
				for _, addr := range postedAddrs {
					if !seen[addr] {
						seen[addr] = true
						fmt.Printf("  - %s\n", addr)
					}
				}
			}
		},
	}

	communityInfoCmd := &cobra.Command{
		Use:               "info <community-addr>",
		Short:             "Show community info",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.CommunityCompletionFunc,
		Run: func(cmd *cobra.Command, args []string) {
			communityAddr := args[0]

			app := getApp()
			authorPubKey, communityID, err := utils.ParseCommunityAddr(communityAddr)
			if err != nil {
				handleError(newError("failed to parse community address", err))
			}

			ctx := context.Background()
			event, err := utils.GetCommunity(ctx, app, authorPubKey, communityID)
			if err != nil {
				handleError(newError("failed to get community", err))
			}

			fmt.Printf("Community Information:\n")
			if nameTag := event.Tags.Find("name"); len(nameTag) > 1 {
				fmt.Printf("Name: %s\n", nameTag[1])
			}
			if descTag := event.Tags.Find("description"); len(descTag) > 1 {
				fmt.Printf("Description: %s\n", descTag[1])
			}
			if imageTag := event.Tags.Find("image"); len(imageTag) > 1 {
				fmt.Printf("Image: %s\n", imageTag[1])
			}
			fmt.Printf("ID: %s\n", communityID)
			fmt.Printf("Author: %s\n", authorPubKey.Hex())
			fmt.Printf("Event ID: %s\n", event.ID.Hex())
			fmt.Printf("Created: %v\n", event.CreatedAt.Time())

			fmt.Printf("\nModerators:\n")
			for tag := range event.Tags.FindAll("p") {
				if len(tag) >= 4 && tag[3] == "moderator" {
					fmt.Printf("  - %s\n", tag[1])
				}
			}
		},
	}

	communityTimelineCmd := &cobra.Command{
		Use:               "timeline <community-addr>",
		Short:             "Show community timeline",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.CommunityCompletionFunc,
		Run: func(cmd *cobra.Command, args []string) {
			communityAddr := args[0]

			app := getApp()
			authorPubKey, communityID, err := utils.ParseCommunityAddr(communityAddr)
			if err != nil {
				handleError(newError("failed to parse community address", err))
			}

			limit := 10
			if l, err := cmd.Flags().GetInt("limit"); err == nil && l > 0 {
				limit = l
			}

			ctx := context.Background()
			events, err := utils.GetCommunityPosts(ctx, app, authorPubKey, communityID, limit)
			if err != nil {
				handleError(newError("failed to get posts", err))
			}

			if len(events) == 0 {
				fmt.Println("No posts found.")
				return
			}

			fmt.Printf("Recent posts in %s:\n", communityAddr)
			fmt.Println(strings.Repeat("=", 50))
			for i, event := range events {
				fmt.Printf("\n[%d] %s\n", i+1, event.ID.Hex()[:16]+"...")
				fmt.Printf("Author: %s\n", event.PubKey.Hex()[:16]+"...")
				fmt.Printf("Time: %v\n", event.CreatedAt.Time())
				fmt.Printf("Content: %s\n", event.Content)
				if i < len(events)-1 {
					fmt.Println(strings.Repeat("-", 30))
				}
			}
		},
	}
	communityTimelineCmd.Flags().IntP("limit", "n", 10, "Number of posts")
	communityTimelineCmd.RegisterFlagCompletionFunc("limit", completion.LimitCompletionFunc)

	communityCmd.AddCommand(communityCreateCmd)
	communityCmd.AddCommand(communityPostCmd)
	communityCmd.AddCommand(communityReplyCmd)
	communityCmd.AddCommand(communityListCmd)
	communityCmd.AddCommand(communityInfoCmd)
	communityCmd.AddCommand(communityTimelineCmd)

	RegisterCommandGroup("Community", "Community operations (NIP-72)", communityCmd)
}
