package utils

import (
	"testing"

	"fiatjaf.com/nostr"
)

func TestParseSearchFilter_BasicKeywords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantText string
		wantLen  int
	}{
		{"plain text", "hello world", "hello world", 1},
		{"text with spaces", "search term here", "search term here", 1},
		{"empty string", "", "", 0},
		{"only whitespace", "   ", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, searchText := ParseSearchFilter(tt.input)
			if searchText != tt.wantText {
				t.Errorf("searchText = %q, want %q", searchText, tt.wantText)
			}
			if filter.Search != tt.wantText {
				t.Errorf("filter.Search = %q, want %q", filter.Search, tt.wantText)
			}
		})
	}
}

func TestParseSearchFilter_Kinds(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLen  int
		wantKind []int
	}{
		{"single kind", "kinds:1", 1, []int{1}},
		{"multiple kinds", "kinds:1,3", 2, []int{1, 3}},
		{"kinds with space", "kinds: 1, 3", 2, []int{1, 3}},
		{"kinds at start", "kinds:1 hello", 1, []int{1}},
		{"kinds at end", "hello kinds:1", 1, []int{1}},
		{"kinds middle", "hello kinds:1 world", 1, []int{1}},
		{"no kinds", "hello world", 0, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, _ := ParseSearchFilter(tt.input)
			if len(filter.Kinds) != tt.wantLen {
				t.Errorf("len(filter.Kinds) = %d, want %d", len(filter.Kinds), tt.wantLen)
			}
			for i, wantK := range tt.wantKind {
				if i < len(filter.Kinds) && int(filter.Kinds[i]) != wantK {
					t.Errorf("filter.Kinds[%d] = %d, want %d", i, filter.Kinds[i], wantK)
				}
			}
		})
	}
}

func TestParseSearchFilter_NoAuthor(t *testing.T) {
	filter, _ := ParseSearchFilter("hello world")
	if len(filter.Authors) != 0 {
		t.Errorf("len(filter.Authors) = %d, want 0", len(filter.Authors))
	}
}

func TestParseSearchFilter_CaseInsensitive(t *testing.T) {
	tests := []string{
		"kinds:1",
		"KINDS:1",
		"Kinds:1",
		"kInDs:1",
	}

	for _, input := range tests {
		filter, _ := ParseSearchFilter(input)
		if len(filter.Kinds) != 1 || int(filter.Kinds[0]) != 1 {
			t.Errorf("input %q: filter.Kinds = %v, want [1]", input, filter.Kinds)
		}
	}
}

func TestParseSearchFilter_Tags(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
		wantTag string
	}{
		{"hashtag with hash", "#t:nostr", 1, "nostr"},
		{"hashtag with tag prefix", "tags:#t:nostr", 1, "nostr"},
		{"no tag", "hello world", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, _ := ParseSearchFilter(tt.input)
			if tt.wantLen > 0 {
				if len(filter.Tags) == 0 {
					t.Errorf("filter.Tags is empty, want non-empty")
					return
				}
				if tags, ok := filter.Tags["t"]; !ok || len(tags) == 0 {
					t.Errorf("filter.Tags[\"t\"] not found or empty")
				} else if tags[0] != tt.wantTag {
					t.Errorf("tag = %q, want %q", tags[0], tt.wantTag)
				}
			}
		})
	}
}

func TestParseSearchFilter_Combined(t *testing.T) {
	input := "kinds:1,3 #t:nostr hello world"

	filter, searchText := ParseSearchFilter(input)

	if searchText != "hello world" {
		t.Errorf("searchText = %q, want %q", searchText, "hello world")
	}
	if len(filter.Kinds) != 2 {
		t.Errorf("len(filter.Kinds) = %d, want 2", len(filter.Kinds))
	}
	if int(filter.Kinds[0]) != 1 || int(filter.Kinds[1]) != 3 {
		t.Errorf("filter.Kinds = %v, want [1, 3]", filter.Kinds)
	}
	if filter.Tags == nil || len(filter.Tags["t"]) == 0 {
		t.Errorf("filter.Tags['t'] missing")
	} else if filter.Tags["t"][0] != "nostr" {
		t.Errorf("filter.Tags['t'][0] = %q, want %q", filter.Tags["t"][0], "nostr")
	}
}

func TestParseSearchFilter_InvalidKinds(t *testing.T) {
	filter, _ := ParseSearchFilter("kinds:abc")

	if len(filter.Kinds) != 0 {
		t.Errorf("len(filter.Kinds) = %d, want 0 for invalid kind", len(filter.Kinds))
	}
}

func TestParseSearchFilter_ValidKind(t *testing.T) {
	filter, _ := ParseSearchFilter("kinds:1,3")

	if len(filter.Kinds) != 2 {
		t.Errorf("len(filter.Kinds) = %d, want 2", len(filter.Kinds))
	}
	if int(filter.Kinds[0]) != 1 || int(filter.Kinds[1]) != 3 {
		t.Errorf("filter.Kinds = %v, want [1, 3]", filter.Kinds)
	}
}

func TestParseSearchFilter_EmptyKinds(t *testing.T) {
	filter, searchText := ParseSearchFilter("kinds: ")

	if searchText != "kinds:" && searchText != "" {
		t.Errorf("searchText = %q, expected cleared or empty", searchText)
	}
	if len(filter.Kinds) != 0 {
		t.Errorf("len(filter.Kinds) = %d, want 0", len(filter.Kinds))
	}
}

func TestSearchResult_Struct(t *testing.T) {
	result := SearchResult{
		Event: nostr.Event{Kind: 1, Content: "test"},
		Relay: "wss://relay.example.com",
	}

	if result.Event.Content != "test" {
		t.Errorf("result.Event.Content = %q, want %q", result.Event.Content, "test")
	}
	if result.Relay != "wss://relay.example.com" {
		t.Errorf("result.Relay = %q, want %q", result.Relay, "wss://relay.example.com")
	}
}