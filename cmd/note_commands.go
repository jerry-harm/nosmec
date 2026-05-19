package cmd

import (
	"context"
	"fmt"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"github.com/jerry-harm/nosmec/cmd/completion"
	"github.com/jerry-harm/nosmec/tui/compose"
	"github.com/jerry-harm/nosmec/tui/timeline"
	"github.com/jerry-harm/nosmec/utils"
	"github.com/spf13/cobra"
)

type timelineEvent struct {
	Event       nostr.Event
	IsCommunity bool
	CommunityID string
}

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
			if err := timeline.RunTimeline(app, filter, hashtags, limit, ""); err != nil {
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
		Use:   "reply <note-id>",
		Short: "Reply to a note via TUI compose",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			noteID := args[0]

			_, decoded, err := nip19.Decode(noteID)
			if err != nil {
				handleError(newError("invalid note ID", err))
				return
			}

			eventIDStr := ""
			switch v := decoded.(type) {
			case nostr.EventPointer:
				eventIDStr = v.ID.Hex()
			default:
				handleError(newError("expected nevent or note ID", nil))
				return
			}

			ctx := context.Background()
			app := getApp()

			parentEvent := app.System().FetchNote(ctx, eventIDStr, 5000)
			if parentEvent == nil {
				handleError(newError("note not found", nil))
				return
			}

			if err := compose.RunReplyCompose(app, parentEvent); err != nil {
				handleError(err)
			}
		},
	}

	noteComposeCmd := &cobra.Command{
		Use:   "compose",
		Short: "Open compose TUI to write a note",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			app := getApp()
			if err := compose.RunNoteCompose(app); err != nil {
				handleError(err)
			}
		},
	}

	noteCmd.AddCommand(noteTimelineCmd)
	noteCmd.AddCommand(notePostCmd)
	noteCmd.AddCommand(noteReplyCmd)
	noteCmd.AddCommand(noteComposeCmd)

	RegisterCommandGroup("Notes", "Note operations", noteCmd)
}

func printTimeline(events []timelineEvent) {
	for i, te := range events {
		e := te.Event
		name := nip19.EncodeNpub(e.PubKey)[:16] + "..."

		pm := getApp().System().FetchProfileMetadata(context.Background(), e.PubKey)
		if pm.Name != "" {
			name = pm.Name
		} else if pm.DisplayName != "" {
			name = pm.DisplayName
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
