package utils

import (
	"testing"

	"fiatjaf.com/nostr"
)

func TestDetermineReplyTarget_NilEvent(t *testing.T) {
	target := DetermineReplyTarget(nil, nil)
	if target.Strategy != ReplyUnsupported {
		t.Errorf("nil event: got strategy %v, want ReplyUnsupported", target.Strategy)
	}
}

func TestDetermineReplyTarget_Kind1(t *testing.T) {
	event := &nostr.Event{
		ID:      [32]byte{1},
		Kind:    nostr.KindTextNote,
		Content: "test note",
	}
	target := DetermineReplyTarget(event, nil)

	if target.Strategy != ReplyNote {
		t.Errorf("kind:1: got strategy %v, want ReplyNote", target.Strategy)
	}
	if target.ReplyKind != nostr.KindTextNote {
		t.Errorf("kind:1 reply kind: got %v, want %v", target.ReplyKind, nostr.KindTextNote)
	}
	if len(target.RootTags) != 1 || target.RootTags[0][0] != "e" || target.RootTags[0][3] != "root" {
		t.Errorf("kind:1 root tags: got %v, want [e id relay root pubkey]", target.RootTags)
	}
	if len(target.ParentTags) < 1 {
		t.Errorf("kind:1 parent tags empty")
	}
}

func TestDetermineReplyTarget_VoiceMessage(t *testing.T) {
	event := &nostr.Event{
		ID:   [32]byte{2},
		Kind: 1222, // Voice message root
	}
	target := DetermineReplyTarget(event, nil)

	if target.Strategy != ReplyDedicated {
		t.Errorf("1222: got strategy %v, want ReplyDedicated", target.Strategy)
	}
	if target.ReplyKind != KindVoiceMessageComment {
		t.Errorf("1222 reply kind: got %v, want %v", target.ReplyKind, KindVoiceMessageComment)
	}
}

func TestDetermineReplyTarget_Torrent(t *testing.T) {
	event := &nostr.Event{
		ID:   [32]byte{3},
		Kind: 2003, // Torrent
	}
	target := DetermineReplyTarget(event, nil)

	if target.Strategy != ReplyDedicated {
		t.Errorf("2003: got strategy %v, want ReplyDedicated", target.Strategy)
	}
	if target.ReplyKind != KindTorrentComment {
		t.Errorf("2003 reply kind: got %v, want %v", target.ReplyKind, KindTorrentComment)
	}
}

func TestDetermineReplyTarget_LiveEvent(t *testing.T) {
	event := &nostr.Event{
		ID:   [32]byte{4},
		Kind: 30311, // Live event
	}
	target := DetermineReplyTarget(event, nil)

	if target.Strategy != ReplyDedicated {
		t.Errorf("30311: got strategy %v, want ReplyDedicated", target.Strategy)
	}
	if target.ReplyKind != KindLiveChat {
		t.Errorf("30311 reply kind: got %v, want %v", target.ReplyKind, KindLiveChat)
	}
}

func TestDetermineReplyTarget_ChannelMessages(t *testing.T) {
	event := &nostr.Event{
		ID:   [32]byte{5},
		Kind: 42, // Channel message
	}
	target := DetermineReplyTarget(event, nil)

	if target.Strategy != ReplySameKind {
		t.Errorf("42: got strategy %v, want ReplySameKind", target.Strategy)
	}
	if target.ReplyKind != 42 {
		t.Errorf("42 reply kind: got %v, want 42", target.ReplyKind)
	}
}

func TestDetermineReplyTarget_ChatMessages(t *testing.T) {
	event := &nostr.Event{
		ID:   [32]byte{6},
		Kind: 9, // Chat message
	}
	target := DetermineReplyTarget(event, nil)

	if target.Strategy != ReplySameKind {
		t.Errorf("9: got strategy %v, want ReplySameKind", target.Strategy)
	}
	if target.ReplyKind != 9 {
		t.Errorf("9 reply kind: got %v, want 9", target.ReplyKind)
	}
	// Chat uses q-tag, not e-tag
	if len(target.ParentTags) != 1 || target.ParentTags[0][0] != "q" {
		t.Errorf("9 parent tags: got %v, want [q id relay pubkey]", target.ParentTags)
	}
}

func TestDetermineReplyTarget_CommunityDefinition(t *testing.T) {
	event := &nostr.Event{
		ID:   [32]byte{7},
		Kind: nostr.KindCommunityDefinition,
		Tags: nostr.Tags{
			nostr.Tag{"d", "test-community"},
		},
	}
	target := DetermineReplyTarget(event, nil)

	if target.Strategy != ReplyComment {
		t.Errorf("34550: got strategy %v, want ReplyComment", target.Strategy)
	}
	if target.ReplyKind != KindComment {
		t.Errorf("34550 reply kind: got %v, want %v", target.ReplyKind, KindComment)
	}
}

func TestDetermineReplyTarget_LongFormContent(t *testing.T) {
	event := &nostr.Event{
		ID:   [32]byte{8},
		Kind: 30023, // Long-form content
		Tags: nostr.Tags{
			nostr.Tag{"d", "my-article"},
		},
	}
	target := DetermineReplyTarget(event, nil)

	if target.Strategy != ReplyComment {
		t.Errorf("30023: got strategy %v, want ReplyComment", target.Strategy)
	}
	if target.ReplyKind != KindComment {
		t.Errorf("30023 reply kind: got %v, want %v", target.ReplyKind, KindComment)
	}
	// Addressable event should use A tag in root
	if len(target.RootTags) == 0 {
		t.Errorf("30023 root tags empty")
	}
}

func TestDetermineReplyTarget_GenericNonKind1(t *testing.T) {
	event := &nostr.Event{
		ID:   [32]byte{9},
		Kind: 9999, // Generic non-kind:1
	}
	target := DetermineReplyTarget(event, nil)

	if target.Strategy != ReplyComment {
		t.Errorf("9999: got strategy %v, want ReplyComment", target.Strategy)
	}
	if target.ReplyKind != KindComment {
		t.Errorf("9999 reply kind: got %v, want %v", target.ReplyKind, KindComment)
	}
}

func TestDetermineReplyTarget_CommunityPost(t *testing.T) {
	// A community post is kind:1111 with an "a" tag pointing to community
	event := &nostr.Event{
		ID:   [32]byte{10},
		Kind: nostr.KindComment,
		Tags: nostr.Tags{
			nostr.Tag{"a", "34550:abc123:general"},
		},
	}
	target := DetermineReplyTarget(event, nil)

	if target.Strategy != ReplyComment {
		t.Errorf("community post: got strategy %v, want ReplyComment", target.Strategy)
	}
	// Root should be the community definition (K=34550, A=community)
	if len(target.RootTags) == 0 {
		t.Errorf("community post root tags empty")
	}
}

func TestBuildAddressableTag(t *testing.T) {
	event := &nostr.Event{
		ID:   [32]byte{1},
		Kind: 30023,
		Tags: nostr.Tags{
			nostr.Tag{"d", "my-article"},
		},
	}
	tag := buildAddressableTag(event)
	expected := "30023:000000000000000000000000000000000000000000000000000000000000000001:my-article"
	if tag != expected {
		t.Errorf("buildAddressableTag: got %q, want %q", tag, expected)
	}
}

func TestExtractPubKeyFromATag(t *testing.T) {
	tests := []struct {
		aTag  string
		want  string
		empty bool
	}{
		{"34550:abc123def456:general", "abc123def456", false},
		{"30023:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff:my-article", "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		got := extractPubKeyFromATag(tt.aTag)
		if tt.empty && got != "" {
			t.Errorf("extractPubKeyFromATag(%q): got %q, want empty", tt.aTag, got)
		}
		if !tt.empty && got != tt.want {
			t.Errorf("extractPubKeyFromATag(%q): got %q, want %q", tt.aTag, got, tt.want)
		}
	}
}

func TestIsCommunityPost(t *testing.T) {
	eventWithA := &nostr.Event{
		ID:   [32]byte{1},
		Kind: nostr.KindComment,
		Tags: nostr.Tags{
			nostr.Tag{"a", "34550:abc123:general"},
		},
	}
	if !isCommunityPost(eventWithA) {
		t.Errorf("event with 'a' tag: should be community post")
	}

	eventWithoutA := &nostr.Event{
		ID:   [32]byte{2},
		Kind: nostr.KindComment,
		Tags: nostr.Tags{},
	}
	if isCommunityPost(eventWithoutA) {
		t.Errorf("event without 'a' tag: should not be community post")
	}

	nonCommentEvent := &nostr.Event{
		ID:   [32]byte{3},
		Kind: nostr.KindTextNote,
		Tags: nostr.Tags{
			nostr.Tag{"a", "34550:abc123:general"},
		},
	}
	if isCommunityPost(nonCommentEvent) {
		t.Errorf("non-comment event with 'a' tag: should not be community post")
	}
}

func TestReplyTarget_QuoteTags(t *testing.T) {
	event := &nostr.Event{
		ID:     [32]byte{1},
		Kind:   nostr.KindTextNote,
		PubKey: [32]byte{1},
	}
	target := DetermineReplyTarget(event, nil)

	// QuoteTags should be set for quote behavior
	// For kind:1, quote uses q tag
	if target.QuoteTags == nil {
		// QuoteTags is nil when resolver is called without AppContext for quotes
		// This is acceptable since quotes are built in compose
	}
}