package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"fiatjaf.com/nostr/sdk"
	"github.com/jerry-harm/nosmec/utils"
	"github.com/spf13/cobra"
)

func registerSearchCommands() {
	searchCmd := &cobra.Command{
		Use:   "search",
		Short: "Search events (NIP-50)",
		Long: `Search for events using NIP-50 full-text search.

Examples:
  nosmec search "nostr apps"
  nosmec search "bitcoin" --kinds 1
  nosmec search "nostr" --limit 20

NIP-50 filter syntax:
  kinds:1,3       Filter by event kinds
  authors:npub1.. Filter by author npub or hex
  #t:hashtag     Filter by hashtag`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			query := args[0]

			kinds, _ := cmd.Flags().GetIntSlice("kinds")
			limit, _ := cmd.Flags().GetInt("limit")

			app := getApp()
			ctx, cancel := context.WithTimeout(context.Background(), app.QueryTimeout())
			defer cancel()

			results, err := utils.SearchEvents(ctx, app, query, limit)
			if err != nil {
				handleError(err)
				return
			}

			if len(results) == 0 {
				fmt.Println("No results found.")
				return
			}

			// Apply kinds filter if specified (NIP-50 relays may not support client-side filtering)
			if len(kinds) > 0 {
				filtered := make([]utils.SearchResult, 0)
				for _, r := range results {
					for _, fk := range kinds {
						if int(r.Event.Kind) == fk {
							filtered = append(filtered, r)
							break
						}
					}
				}
				results = filtered
			}

			// Sort by created_at descending (relay may have already sorted by relevance)
			sort.Slice(results, func(i, j int) bool {
				return results[i].Event.CreatedAt > results[j].Event.CreatedAt
			})

			fmt.Printf("Found %d result(s):\n\n", len(results))
			for i, r := range results {
				printSearchResult(r, i)
			}
		},
	}

	searchCmd.Flags().IntSlice("kinds", nil, "Filter by event kinds (e.g., --kinds 1,3)")
	searchCmd.Flags().IntP("limit", "n", 50, "Maximum number of results")

	RegisterCommandGroup("Search", "Search operations", searchCmd)
}

func printSearchResult(r utils.SearchResult, index int) {
	e := r.Event

	// Get author name
	authorName := nip19.EncodeNpub(e.PubKey)[:16] + "..."

	// Try to get profile name
	pm := getApp().System().FetchProfileMetadata(context.Background(), e.PubKey)
	if pm.Event != nil {
		if meta, err := sdk.ParseMetadata(*pm.Event); err == nil && meta.Name != "" {
			authorName = meta.Name
		}
	}

	fmt.Printf("[%d] %s @%s\n", index+1, formatSearchTime(e.CreatedAt), authorName)
	fmt.Printf("    ID: %s\n", nip19.EncodeNevent(e.ID, []string{r.Relay}, e.PubKey))
	fmt.Printf("    Relay: %s\n", r.Relay)
	fmt.Printf("    Kind: %d\n", e.Kind)

	// Print tags summary
	if len(e.Tags) > 0 {
		var tagSummary []string
		for _, tag := range e.Tags {
			if len(tag) >= 2 {
				tagSummary = append(tagSummary, fmt.Sprintf("%s:%s", tag[0], tag[1]))
			} else if len(tag) == 1 {
				tagSummary = append(tagSummary, tag[0])
			}
		}
		if len(tagSummary) > 0 {
			fmt.Printf("    Tags: %s\n", searchTruncate(strings.Join(tagSummary, ", "), 80))
		}
	}

	// Print content preview
	content := e.Content
	if len(content) > 200 {
		content = content[:200] + "..."
	}
	if content != "" {
		fmt.Printf("    %s\n", content)
	}

	fmt.Println()
}

func formatSearchTime(t nostr.Timestamp) string {
	return t.Time().Format("2006-01-02 15:04:05")
}

func searchTruncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}