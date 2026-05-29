package gui

import (
	"reflect"
	"testing"

	"github.com/jerry-harm/nosmec/config"
)

func TestNormalizeLocale(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "exact locale", input: "zh", want: "zh"},
		{name: "locale with region", input: "zh-CN", want: "zh"},
		{name: "locale with underscore", input: "en_US", want: "en"},
		{name: "unsupported falls back", input: "fr-FR", want: "en"},
		{name: "empty falls back", input: "", want: "en"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeLocale(tt.input); got != tt.want {
				t.Fatalf("normalizeLocale(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCommunityNamesFromSubscriptions(t *testing.T) {
	fallback := []string{"nostr-lang-zh", "music", "tech"}

	subs := []config.Subscription{
		{Type: "user", ID: "ignored-user"},
		{Type: "community", ID: "tech", Petname: "Tech Corner"},
		{Type: "community", ID: "music"},
		{Type: "community", ID: "tech", Petname: "Duplicate should be ignored"},
	}

	want := []string{"Tech Corner", "music"}
	if got := communityNamesFromSubscriptions(subs, fallback); !reflect.DeepEqual(got, want) {
		t.Fatalf("communityNamesFromSubscriptions() = %v, want %v", got, want)
	}

	if got := communityNamesFromSubscriptions(nil, fallback); !reflect.DeepEqual(got, fallback) {
		t.Fatalf("communityNamesFromSubscriptions(nil) = %v, want fallback %v", got, fallback)
	}
}

func TestPostsForScope(t *testing.T) {
	posts := []Post{
		{ID: "1", Community: "tech"},
		{ID: "2", Community: "music"},
		{ID: "3", Community: "tech"},
	}

	if got := postsForScope(posts, scopeMyFeed, ""); len(got) != 3 {
		t.Fatalf("postsForScope(my feed) returned %d posts, want 3", len(got))
	}

	got := postsForScope(posts, scopeCommunity, "tech")
	want := []Post{{ID: "1", Community: "tech"}, {ID: "3", Community: "tech"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("postsForScope(community) = %v, want %v", got, want)
	}

	if got := postsForScope(posts, scopeCommunity, "missing"); len(got) != 0 {
		t.Fatalf("postsForScope(missing) returned %d posts, want 0", len(got))
	}
}
