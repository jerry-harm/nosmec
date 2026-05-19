package nostr_sdk

import (
	"testing"

	"fiatjaf.com/nostr"
)

func TestExtractCommunityScope_PrefersUppercaseA(t *testing.T) {
	event := &nostr.Event{
		Kind: nostr.KindComment,
		Tags: nostr.Tags{
			{"A", "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:cats"},
			{"a", "1111:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb:parent"},
		},
	}

	scope := ExtractCommunityScope(event)
	if scope != "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:cats" {
		t.Fatalf("ExtractCommunityScope() = %q", scope)
	}
}

func TestExtractCommunityScope_FallsBackToLegacyLowercaseA(t *testing.T) {
	event := &nostr.Event{
		Kind: nostr.KindTextNote,
		Tags: nostr.Tags{{"a", "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:cats"}},
	}

	scope := ExtractCommunityScope(event)
	if scope != "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:cats" {
		t.Fatalf("ExtractCommunityScope() = %q", scope)
	}
}

func TestExtractCommunityScope_IgnoresNonCommunityLowercaseA(t *testing.T) {
	event := &nostr.Event{
		Kind: nostr.KindComment,
		Tags: nostr.Tags{{"a", "1111:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb:parent"}},
	}

	scope := ExtractCommunityScope(event)
	if scope != "" {
		t.Fatalf("ExtractCommunityScope() = %q, want empty", scope)
	}
}

func TestMatchesCommunityScope(t *testing.T) {
	scope := "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:cats"
	event := &nostr.Event{
		Kind: nostr.KindComment,
		Tags: nostr.Tags{{"A", scope}},
	}

	if !MatchesCommunityScope(event, scope) {
		t.Fatal("MatchesCommunityScope() = false, want true")
	}
	if MatchesCommunityScope(event, "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:dogs") {
		t.Fatal("MatchesCommunityScope() = true, want false")
	}
	if !MatchesCommunityScope(event, "") {
		t.Fatal("MatchesCommunityScope() with empty scope = false, want true")
	}
}

func TestFetchSpecificEventInScope_NilWhenScopeMismatches(t *testing.T) {
	sys := NewSystem()
	rootID, _ := nostr.IDFromHex("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	event := nostr.Event{
		ID:   rootID,
		Kind: nostr.KindComment,
		Tags: nostr.Tags{{"A", "34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:cats"}},
	}
	sys.Publisher.Publish(t.Context(), event)

	got, _, err := sys.FetchSpecificEventInScope(
		t.Context(),
		nostr.EventPointer{ID: rootID},
		"34550:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:dogs",
		FetchSpecificEventParameters{},
	)
	if err != nil {
		t.Fatalf("FetchSpecificEventInScope() error = %v", err)
	}
	if got != nil {
		t.Fatal("FetchSpecificEventInScope() returned event, want nil")
	}
}
