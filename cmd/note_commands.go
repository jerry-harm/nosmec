package cmd

import (
	"context"
	"fmt"

	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/cmd/completion"
	"github.com/jerry-harm/nosmec/tui/timeline"
	"github.com/jerry-harm/nosmec/utils"
	"github.com/spf13/cobra"
	"fiatjaf.com/nostr/nip19"
)

func registerNoteCommands() {
	noteCmd := &cobra.Command{
		Use:   "note",
		Short: "Note operations",
	}

	noteTimelineCmd := &cobra.Command{
		Use:   "timeline",
		Short: "Show timeline with TUI",
		Run: func(cmd *cobra.Command, args []string) {
			filter := "followed"
			if global, _ := cmd.Flags().GetBool("global"); global {
				filter = "global"
			} else if mine, _ := cmd.Flags().GetBool("mine"); mine {
				filter = "mine"
			}

			limit := 50
			if l, err := cmd.Flags().GetInt("limit"); err == nil && l > 0 {
				limit = l
			}

			var hashtags []string
			if h, err := cmd.Flags().GetStringSlice("hashtag"); err == nil && len(h) > 0 {
				hashtags = h
			}

			app := getApp()
			if err := timeline.RunTimeline(app, filter, hashtags, limit); err != nil {
				handleError(err)
			}
		},
	}

	noteTimelineCmd.Flags().Bool("follow", false, "Show followed timeline (default)")
	noteTimelineCmd.Flags().Bool("mine", false, "Show my timeline")
	noteTimelineCmd.Flags().Bool("global", false, "Show global timeline")
	noteTimelineCmd.Flags().IntP("limit", "n", 50, "Number of notes to show")
	noteTimelineCmd.Flags().StringSliceP("hashtag", "t", nil, "Filter by hashtags")

	noteTimelineCmd.RegisterFlagCompletionFunc("limit", completion.LimitCompletionFunc)
	noteTimelineCmd.RegisterFlagCompletionFunc("hashtag", completion.HashtagCompletionFunc)

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
			fmt.Printf("Note ID: %s\n", nip19.EncodeNevent(event.ID, nil, event.PubKey))
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
			fmt.Printf("Reply ID: %s\n", nip19.EncodeNevent(event.ID, nil, event.PubKey))
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
		name := nip19.EncodeNpub(e.PubKey)[:16] + "..."

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