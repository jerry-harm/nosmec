package utils

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"github.com/jerry-harm/nosmec/config"
)

// SearchResult holds a search result with its source relay
type SearchResult struct {
	Event  nostr.Event
	Relay  string
}

// ParseSearchFilter parses a NIP-50 search string and extracts filter parameters.
// Supports: kinds:1,3 authors:npub1... #t:hashtag
// Remaining text becomes the search query.
func ParseSearchFilter(input string) (nostr.Filter, string) {
	filter := nostr.Filter{}

	// Match kinds:1,3 or kinds: 1,3
	kindsRegex := regexp.MustCompile(`(?i)\bkinds:([\d, ]+)\b`)
	if match := kindsRegex.FindStringSubmatch(input); len(match) > 1 {
		kindsStr := strings.ReplaceAll(match[1], " ", "")
		parts := strings.Split(kindsStr, ",")
		for _, p := range parts {
			if k, err := strconv.Atoi(p); err == nil {
				filter.Kinds = append(filter.Kinds, nostr.Kind(k))
			}
		}
		// Remove matched portion from input
		input = kindsRegex.ReplaceAllString(input, "")
	}

	// Match authors:npub1... or authors: hexpubkey
	authorsRegex := regexp.MustCompile(`(?i)\bauthors:([^\s#]+)\b`)
	if match := authorsRegex.FindStringSubmatch(input); len(match) > 1 {
		authorStr := match[1]
		_, decoded, err := nip19.Decode(authorStr)
		if err == nil {
			if pk, ok := decoded.(nostr.PubKey); ok {
				filter.Authors = append(filter.Authors, pk)
			}
		}
		input = authorsRegex.ReplaceAllString(input, "")
	}

	// Match #t:hashtag or #t hashtag
	tagRegex := regexp.MustCompile(`(?i)#t:(\S+)`)
	if match := tagRegex.FindStringSubmatch(input); len(match) > 1 {
		tag := strings.ToLower(match[1])
		if filter.Tags == nil {
			filter.Tags = make(nostr.TagMap)
		}
		filter.Tags["t"] = append(filter.Tags["t"], tag)
		input = tagRegex.ReplaceAllString(input, "")
	}

	// Remaining text is the search query
	searchText := strings.TrimSpace(input)
	if searchText != "" {
		filter.Search = searchText
	}

	return filter, filter.Search
}

// SearchEvents searches for events across multiple relays using NIP-50
func SearchEvents(ctx context.Context, app *config.AppContext, query string, limit int) ([]SearchResult, error) {
	// Parse the search string
	filter, searchText := ParseSearchFilter(query)

	// Set limit
	if limit > 0 {
		filter.Limit = limit
	} else {
		filter.Limit = 50
	}

	// Build relay list: search_relays + local relay
	relays := buildSearchRelayList(app)

	if len(relays) == 0 {
		return nil, fmt.Errorf("no search relays configured")
	}

	if searchText == "" && filter.Search == "" && len(filter.Kinds) == 0 && len(filter.Authors) == 0 {
		return nil, fmt.Errorf("no search terms provided")
	}

	var results []SearchResult
	seen := make(map[nostr.ID]bool)

	// Query all relays and collect results
	ch := app.Pool().FetchMany(ctx, relays, filter, nostr.SubscriptionOptions{})

	for re := range ch {
		if seen[re.Event.ID] {
			continue
		}
		seen[re.Event.ID] = true
		results = append(results, SearchResult{
			Event: re.Event,
			Relay: re.Relay.URL,
		})
	}

	return results, nil
}

// buildSearchRelayList builds the list of relays to search: search_relays + local relay
func buildSearchRelayList(app *config.AppContext) []string {
	relaySet := make(map[string]struct{})

	// Add configured search relays
	for _, url := range app.ListSearchRelays() {
		relaySet[url] = struct{}{}
	}

	relays := make([]string, 0, len(relaySet))
	for r := range relaySet {
		relays = append(relays, r)
	}

	return relays
}

// GetSearchRelays returns the configured search relays
func GetSearchRelays(app *config.AppContext) []string {
	return app.ListSearchRelays()
}