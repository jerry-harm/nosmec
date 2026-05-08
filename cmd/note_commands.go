package cmd

import (
	"context"
	"fmt"
	"time"

	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/cmd/completion"
	"github.com/jerry-harm/nosmec/utils"
	"github.com/spf13/cobra"
)

func registerNoteCommands() {
	noteCmd := &cobra.Command{
		Use:   "note",
		Short: "Note operations",
	}

	noteTimelineCmd := &cobra.Command{
		Use:   "timeline",
		Short: "Show timeline",
		Run: func(cmd *cobra.Command, args []string) {
			filter := "followed"
			if global, _ := cmd.Flags().GetBool("global"); global {
				filter = "global"
			} else if f, err := cmd.Flags().GetString("filter"); err == nil && f != "" {
				filter = f
			}

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
			opts := &utils.GetOptions{App: app}

			var events []utils.TimelineEvent
			var err error

			switch filter {
			case "global":
				nostrEvents, e := utils.GetGlobalTimeline(ctx, limit, opts)
				if e != nil {
					handleError(e)
				}
				for _, e := range nostrEvents {
					events = append(events, utils.TimelineEvent{Event: e})
				}
			case "mine":
				nostrEvents, e := utils.GetMyTimeline(ctx, limit, opts)
				if e != nil {
					handleError(e)
				}
				for _, e := range nostrEvents {
					events = append(events, utils.TimelineEvent{Event: e})
				}
			default:
				events, err = utils.GetFollowedTimeline(ctx, limit, hashtags, opts)
				if err != nil {
					handleError(err)
				}
			}

			if len(events) == 0 {
				fmt.Println("No notes found.")
				return
			}

			printTimeline(events)
		},
	}

	noteTimelineCmd.Flags().Bool("global", false, "Show global timeline")
	noteTimelineCmd.Flags().StringP("filter", "f", "followed", "Timeline filter: followed, global, mine")
	noteTimelineCmd.Flags().IntP("limit", "n", 10, "Number of notes to show")
	noteTimelineCmd.Flags().StringSliceP("hashtag", "t", nil, "Filter by hashtags")

	noteTimelineCmd.RegisterFlagCompletionFunc("filter", completion.FilterCompletionFunc)
	noteTimelineCmd.RegisterFlagCompletionFunc("limit", completion.LimitCompletionFunc)
	noteTimelineCmd.RegisterFlagCompletionFunc("hashtag", completion.HashtagCompletionFunc)
	noteTimelineCmd.RegisterFlagCompletionFunc("global", completion.GlobalCompletionFunc)

	notePostCmd := &cobra.Command{
		Use:   "post <content>",
		Short: "Post a note",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			content := args[0]

			ctx := context.Background()
			app := getApp()

			event, err := utils.PostNote(ctx, app, content)
			if err != nil {
				handleError(err)
			}

			fmt.Printf("Posted successfully!\n")
			fmt.Printf("Note ID: %s\n", event.ID.Hex())
		},
	}

	noteReplyCmd := &cobra.Command{
		Use:   "reply <note-id> <content>",
		Short: "Reply to a note",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			noteID := args[0]
			content := args[1]

			ctx := context.Background()
			app := getApp()

			event, err := utils.ReplyToNote(ctx, app, noteID, content)
			if err != nil {
				handleError(err)
			}

			fmt.Printf("Replied successfully!\n")
			fmt.Printf("Reply ID: %s\n", event.ID.Hex())
		},
	}

	noteCmd.AddCommand(noteTimelineCmd)
	noteCmd.AddCommand(notePostCmd)
	noteCmd.AddCommand(noteReplyCmd)

	RegisterCommandGroup("Notes", "Note operations", noteCmd)
}

func printTimeline(events []utils.TimelineEvent) {
	for i, te := range events {
		e := te.Event
		name := e.PubKey.Hex()[:8] + "..."

		if profile := utils.GetProfileName(context.Background(), e.PubKey, &utils.GetOptions{App: getApp()}); profile != "" {
			name = profile
		}

		fmt.Printf("[%s] @%s\n", formatTime(e.CreatedAt), name)
		fmt.Printf("  %s\n", truncate(e.Content, 100))

		if te.IsCommunity {
			fmt.Printf("  (community: %s)\n", te.CommunityID)
		}

		if i < len(events)-1 {
			fmt.Println()
		}
	}
}

func formatTime(t nostr.Timestamp) string {
	return t.Time().Format("2006-01-02 15:04")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
