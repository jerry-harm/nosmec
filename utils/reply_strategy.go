package utils

import (
	"fmt"

	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
)

// ReplyStrategy describes how a reply should be formed for a given event.
type ReplyStrategy int

const (
	// ReplyUnsupported indicates no standard reply path exists for this event kind.
	ReplyUnsupported ReplyStrategy = iota
	// ReplyNote indicates kind:1 with NIP-10 e-tag threading.
	ReplyNote
	// ReplyComment indicates kind:1111 with NIP-22 root/parent tags.
	ReplyComment
	// ReplySameKind indicates same kind as target with kind-specific threading.
	ReplySameKind
	// ReplyDedicated indicates a dedicated reply kind (1244, 2004, 1311, etc.).
	ReplyDedicated
)

// ReplyTarget holds the strategy and tags for composing a reply to an event.
type ReplyTarget struct {
	Strategy   ReplyStrategy
	ReplyKind  nostr.Kind
	RootTags   nostr.Tags
	ParentTags nostr.Tags
	QuoteTags  nostr.Tags
}

// Kind constants for specialized reply events.
const (
	KindVoiceMessageComment nostr.Kind = 1244
	KindTorrentComment      nostr.Kind = 2004
	KindLiveChat            nostr.Kind = 1311
	KindComment             nostr.Kind = 1111
)

// DetermineReplyTarget returns the appropriate reply strategy and tags for a given event.
func DetermineReplyTarget(event *nostr.Event, app *config.AppContext) ReplyTarget {
	if event == nil {
		return ReplyTarget{Strategy: ReplyUnsupported}
	}

	k := event.Kind

	// 1. Dedicated reply kinds from README table
	switch k {
	case 1222: // Voice message root -> dedicated voice message comment
		return ReplyTarget{
			Strategy:  ReplyDedicated,
			ReplyKind: KindVoiceMessageComment,
			RootTags:  buildNIP22RootTags(app, event, "E"),
			ParentTags: nostr.Tags{
				nostr.Tag{"e", event.ID.Hex(), app.GetEventRelay(event.ID.Hex())},
				nostr.Tag{"k", fmt.Sprintf("%d", k)},
				nostr.Tag{"p", event.PubKey.Hex()},
			},
		}
	case 2003: // Torrent -> dedicated torrent comment
		return ReplyTarget{
			Strategy:  ReplyDedicated,
			ReplyKind: KindTorrentComment,
			RootTags:  buildNIP10RootTags(app, event),
			ParentTags: nostr.Tags{
				nostr.Tag{"e", event.ID.Hex(), app.GetEventRelay(event.ID.Hex()), "reply", event.PubKey.Hex()},
				nostr.Tag{"p", event.PubKey.Hex()},
			},
		}
	case 30311: // Live event -> dedicated live chat message
		return ReplyTarget{
			Strategy:  ReplyDedicated,
			ReplyKind: KindLiveChat,
			RootTags:  buildNIP22RootTags(app, event, "A"),
			ParentTags: nostr.Tags{
				nostr.Tag{"e", event.ID.Hex(), app.GetEventRelay(event.ID.Hex())},
				nostr.Tag{"k", fmt.Sprintf("%d", k)},
				nostr.Tag{"p", event.PubKey.Hex()},
			},
		}
	case 42: // Channel messages stay same kind with NIP-10 threading
		return ReplyTarget{
			Strategy:  ReplySameKind,
			ReplyKind: 42,
			RootTags:  buildNIP10RootTags(app, event),
			ParentTags: nostr.Tags{
				nostr.Tag{"e", event.ID.Hex(), app.GetEventRelay(event.ID.Hex()), "reply", event.PubKey.Hex()},
				nostr.Tag{"p", event.PubKey.Hex()},
			},
		}
	case 9: // Chat messages reply with same kind using q-tag citation
		return ReplyTarget{
			Strategy:  ReplySameKind,
			ReplyKind: 9,
			ParentTags: nostr.Tags{
				nostr.Tag{"q", event.ID.Hex(), app.GetEventRelay(event.ID.Hex()), event.PubKey.Hex()},
			},
		}
	}

	// 2. NIP-10 for kind:1
	if k == nostr.KindTextNote {
		return ReplyTarget{
			Strategy:  ReplyNote,
			ReplyKind: nostr.KindTextNote,
			RootTags:  buildNIP10RootTags(app, event),
			ParentTags: nostr.Tags{
				nostr.Tag{"e", event.ID.Hex(), app.GetEventRelay(event.ID.Hex()), "reply", event.PubKey.Hex()},
				nostr.Tag{"p", event.PubKey.Hex()},
			},
		}
	}

	// 3. NIP-22 / NIP-72 for generic non-kind:1 replyable scopes
	// This covers: kind:30023, addressable events (30000-39999), community posts,
	// and other non-kind:1 content with a standard reply path.
	switch {
	case isCommunityPost(event):
		// NIP-72 community-rooted reply structure
		return buildNIP72ReplyTarget(event, app)
	case k == nostr.KindCommunityDefinition:
		// Replying to community definition itself uses NIP-72
		return buildNIP72ReplyTarget(event, app)
	case k >= 30000 && k < 40000:
		// Addressable events — use NIP-22 with A tag
		return ReplyTarget{
			Strategy:  ReplyComment,
			ReplyKind: KindComment,
			RootTags:  buildNIP22RootTags(app, event, "A"),
			ParentTags: nostr.Tags{
				nostr.Tag{"a", buildAddressableTag(event)},
				nostr.Tag{"k", fmt.Sprintf("%d", k)},
				nostr.Tag{"p", event.PubKey.Hex()},
			},
		}
	default:
		// Generic non-kind:1 event — use NIP-22 with E tag
		return ReplyTarget{
			Strategy:  ReplyComment,
			ReplyKind: KindComment,
			RootTags:  buildNIP22RootTags(app, event, "E"),
			ParentTags: nostr.Tags{
				nostr.Tag{"e", event.ID.Hex(), app.GetEventRelay(event.ID.Hex())},
				nostr.Tag{"k", fmt.Sprintf("%d", k)},
				nostr.Tag{"p", event.PubKey.Hex()},
			},
		}
	}
}

// buildNIP10RootTags builds NIP-10 root e-tags for kind:1 replies.
func buildNIP10RootTags(app *config.AppContext, event *nostr.Event) nostr.Tags {
	relay := app.GetEventRelay(event.ID.Hex())
	return nostr.Tags{
		nostr.Tag{"e", event.ID.Hex(), relay, "root", event.PubKey.Hex()},
	}
}

// buildNIP22RootTags builds NIP-22 uppercase root tags.
// scopeType is "E", "A", or "I".
func buildNIP22RootTags(app *config.AppContext, event *nostr.Event, scopeType string) nostr.Tags {
	relay := app.GetEventRelay(event.ID.Hex())
	switch scopeType {
	case "E":
		return nostr.Tags{
			nostr.Tag{"K", fmt.Sprintf("%d", event.Kind)},
			nostr.Tag{"E", event.ID.Hex(), relay},
			nostr.Tag{"P", event.PubKey.Hex()},
		}
	case "A":
		return nostr.Tags{
			nostr.Tag{"K", fmt.Sprintf("%d", event.Kind)},
			nostr.Tag{"A", buildAddressableTag(event), relay},
			nostr.Tag{"P", event.PubKey.Hex()},
		}
	default:
		return nostr.Tags{
			nostr.Tag{"K", fmt.Sprintf("%d", event.Kind)},
			nostr.Tag{"E", event.ID.Hex(), relay},
			nostr.Tag{"P", event.PubKey.Hex()},
		}
	}
}

// buildAddressableTag builds an addressable event "a" tag.
// Format: "kind:pubkey:d"
func buildAddressableTag(event *nostr.Event) string {
	d := event.Tags.GetD()
	return fmt.Sprintf("%d:%s:%s", event.Kind, event.PubKey.Hex(), d)
}

// isCommunityPost returns true if the event is a community post (kind:1111 within a community).
func isCommunityPost(event *nostr.Event) bool {
	if event.Kind != nostr.KindComment {
		return false
	}
	// Community posts have an "a" tag pointing to the community definition.
	return event.Tags.Has("a")
}

// buildNIP72ReplyTarget builds a NIP-72 community-rooted reply target.
func buildNIP72ReplyTarget(event *nostr.Event, app *config.AppContext) ReplyTarget {
	// Find the community definition tag ("a" tag).
	var communityA string
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "a" {
			communityA = tag[1]
			break
		}
	}

	relay := app.GetEventRelay(event.ID.Hex())

	if communityA == "" {
		// No community tag found; fall back to generic NIP-22.
		return ReplyTarget{
			Strategy:  ReplyComment,
			ReplyKind: KindComment,
			RootTags: nostr.Tags{
				nostr.Tag{"K", fmt.Sprintf("%d", event.Kind)},
				nostr.Tag{"E", event.ID.Hex(), relay},
				nostr.Tag{"P", event.PubKey.Hex()},
			},
			ParentTags: nostr.Tags{
				nostr.Tag{"e", event.ID.Hex(), relay},
				nostr.Tag{"k", fmt.Sprintf("%d", event.Kind)},
				nostr.Tag{"p", event.PubKey.Hex()},
			},
		}
	}

	// Community-rooted: root is the community definition, parent is this post.
	return ReplyTarget{
		Strategy:  ReplyComment,
		ReplyKind: KindComment,
		RootTags: nostr.Tags{
			nostr.Tag{"K", "34550"},
			nostr.Tag{"A", communityA, relay},
			nostr.Tag{"P", extractPubKeyFromATag(communityA)},
		},
		ParentTags: nostr.Tags{
			nostr.Tag{"e", event.ID.Hex(), relay},
			nostr.Tag{"k", fmt.Sprintf("%d", event.Kind)},
			nostr.Tag{"p", event.PubKey.Hex()},
		},
	}
}

// extractPubKeyFromATag extracts the pubkey from an addressable "a" tag.
// Format: "kind:pubkey:d"
func extractPubKeyFromATag(aTag string) string {
	var kind int
	var pubkey string
	_, err := fmt.Sscanf(aTag, "%d:%s:", &kind, &pubkey)
	if err != nil {
		return ""
	}
	return pubkey
}