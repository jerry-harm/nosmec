package cmd

import (
	"context"
	"fmt"
	"strings"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"github.com/jerry-harm/nosmec/cmd/completion"
	"github.com/jerry-harm/nosmec/sdkplus"
	"github.com/jerry-harm/nosmec/tui/community/discover"
	"github.com/jerry-harm/nosmec/tui/timeline"
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
			fmt.Printf("Event ID: %s\n", nip19.EncodeNevent(event.ID, nil, event.PubKey))
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
			fmt.Printf("Post ID: %s\n", nip19.EncodeNevent(event.ID, nil, event.PubKey))
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
			app := getApp()

			// Parse postID to hex event ID (supports nevent, note, or raw hex)
			eventIDStr := postID
			if _, decoded, err := nip19.Decode(postID); err == nil {
				switch v := decoded.(type) {
				case nostr.EventPointer:
					eventIDStr = v.ID.Hex()
				case nostr.ID:
					eventIDStr = v.Hex()
				}
			}

			// Fetch parent post
			wrapper := sdkplus.Wrap(app.System())
			parentEvent := wrapper.FetchNote(ctx, eventIDStr, app.QueryTimeoutms())
			if parentEvent == nil {
				handleError(newError("parent post not found", nil))
			}

			// Extract community addr from parent tags (a tag with "34550:" prefix)
			var communityAddr string
			for _, tag := range parentEvent.Tags {
				if tag[0] == "a" && strings.HasPrefix(tag[1], "34550:") {
					communityAddr = tag[1]
					break
				}
			}
			if communityAddr == "" {
				handleError(newError("parent post is not associated with a community", nil))
			}

			event, err := utils.PostToCommunity(ctx, app, communityAddr, content, parentEvent.ID.Hex())
			if err != nil {
				handleError(newError("failed to reply", err))
			}

			fmt.Printf("Replied!\n")
			fmt.Printf("Reply ID: %s\n", nip19.EncodeNevent(event.ID, nil, event.PubKey))
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

			// Get relay list for queries
			relays := app.AllReadableRelays()
			if len(relays) == 0 {
				relays = app.Config().KnownRelays
			}
			timeoutMs := app.QueryTimeoutms()

			// --- Following (Kind 10004) ---
			fmt.Println("[Following] (Kind 10004)")
			myPubKey, err := app.GetMyPubKey()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				followedFilter := nostr.Filter{
					Kinds:   []nostr.Kind{10004},
					Authors: []nostr.PubKey{myPubKey},
					Limit:   1,
				}
				followedEvent := sdkplus.Wrap(app.System()).FetchEventByFilter(ctx, followedFilter, timeoutMs)
				if followedEvent != nil {
					var followed []string
					for _, tag := range followedEvent.Tags {
						if tag[0] == "a" && strings.HasPrefix(tag[1], "34550:") {
							followed = append(followed, tag[1])
						}
					}
					if len(followed) == 0 {
						fmt.Println("  (none)")
					} else {
						for _, addr := range followed {
							fmt.Printf("  - %s\n", addr)
						}
					}
				} else {
					fmt.Println("  (none)")
				}
			}
			fmt.Println()

			// --- Created (Kind 34550) ---
			fmt.Println("[Created] (Kind 34550)")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				createdFilter := nostr.Filter{
					Kinds:   []nostr.Kind{nostr.KindCommunityDefinition},
					Authors: []nostr.PubKey{myPubKey},
				}
				ctxQuery, cancel := context.WithTimeout(ctx, app.QueryTimeout())
				defer cancel()
				var events []nostr.Event
				for ie := range app.Pool().FetchMany(ctxQuery, relays, createdFilter, nostr.SubscriptionOptions{}) {
					events = append(events, ie.Event)
				}
				if len(events) == 0 {
					fmt.Println("  (none)")
				} else {
					for _, event := range events {
						name := ""
						if nameTag := event.Tags.Find("name"); len(nameTag) > 1 {
							name = nameTag[1]
						} else if dTag := event.Tags.Find("d"); len(dTag) > 1 {
							name = dTag[1]
						}
						if name == "" {
							name = nip19.EncodeNevent(event.ID, nil, event.PubKey)[:32] + "..."
						}
						fmt.Printf("  - %s\n", name)
					}
				}
			}
			fmt.Println()

			// --- Posted (Kind 1111) ---
			fmt.Println("[Posted] (Kind 1111)")
			postedFilter := nostr.Filter{
				Kinds:   []nostr.Kind{nostr.KindComment},
				Authors: []nostr.PubKey{myPubKey},
			}
			ctxQuery2, cancel2 := context.WithTimeout(ctx, app.QueryTimeout())
			defer cancel2()
			var postedAddrs []string
			for ie := range app.Pool().FetchMany(ctxQuery2, relays, postedFilter, nostr.SubscriptionOptions{}) {
				for _, tag := range ie.Event.Tags {
					if tag[0] == "a" && strings.HasPrefix(tag[1], "34550:") {
						postedAddrs = append(postedAddrs, tag[1])
					}
				}
			}
			if len(postedAddrs) == 0 {
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
			filter := nostr.Filter{
				Kinds:   []nostr.Kind{nostr.KindCommunityDefinition},
				Authors: []nostr.PubKey{authorPubKey},
				Tags:    nostr.TagMap{"d": []string{communityID}},
			}
			event := sdkplus.Wrap(app.System()).FetchEventByFilter(ctx, filter, app.QueryTimeoutms())
			if event == nil {
				handleError(newError("community not found", nil))
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
			fmt.Printf("Author: %s\n", nip19.EncodeNpub(authorPubKey))
			fmt.Printf("Event ID: %s\n", nip19.EncodeNevent(event.ID, nil, event.PubKey))
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
		Short:             "Show community timeline with TUI",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.CommunityCompletionFunc,
		Run: func(cmd *cobra.Command, args []string) {
			communityAddr := args[0]

			limit := 10
			if l, err := cmd.Flags().GetInt("limit"); err == nil && l > 0 {
				limit = l
			}

			app := getApp()
			if err := timeline.RunTimeline(app, "community", nil, limit, communityAddr); err != nil {
				handleError(err)
			}
		},
	}
	communityTimelineCmd.Flags().IntP("limit", "n", 10, "Number of posts")
	communityTimelineCmd.RegisterFlagCompletionFunc("limit", completion.LimitCompletionFunc)

	communityDiscoverCmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover communities from relays (kind 34550)",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			app := getApp()
			if err := discover.RunCommunityDiscover(app); err != nil {
				handleError(err)
			}
		},
	}

	communityCmd.AddCommand(communityCreateCmd)
	communityCmd.AddCommand(communityPostCmd)
	communityCmd.AddCommand(communityReplyCmd)
	communityCmd.AddCommand(communityListCmd)
	communityCmd.AddCommand(communityInfoCmd)
	communityCmd.AddCommand(communityTimelineCmd)
	communityCmd.AddCommand(communityDiscoverCmd)

	RegisterCommandGroup("Community", "Community operations (NIP-72)", communityCmd)
}
