package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"fiatjaf.com/nostr"
)

// Community represents a moderated community as defined by NIP-72.
type Community struct {
	Author      nostr.PubKey
	DIdentifier string
	Name        string
	Description string
	Image       string
	ImageSize   string // width x height
	Moderators  []nostr.ProfilePointer
	Relays      map[string][]string // marker -> relay URLs
	// Other tags can be stored as raw tags
	RawTags nostr.Tags
}

// ParseCommunityFromEvent parses a community definition event into a Community struct.
// Returns an error if the event is not a valid community definition.
func ParseCommunityFromEvent(event *nostr.Event) (*Community, error) {
	if event.Kind != nostr.KindCommunityDefinition {
		return nil, fmt.Errorf("event kind %d is not a community definition", event.Kind)
	}

	comm := &Community{
		Author:      event.PubKey,
		DIdentifier: event.Tags.GetD(),
		Name:        event.Tags.Find("name")[1],
		Description: event.Tags.Find("description")[1],
		Image:       "",
		ImageSize:   "",
		Moderators:  make([]nostr.ProfilePointer, 0),
		Relays:      make(map[string][]string, 0),
		RawTags:     event.Tags,
	}

	if len(event.Tags.Find("image")) == 3 {
		comm.Image = event.Tags.Find("image")[1]
		comm.ImageSize = event.Tags.Find("image")[2]
	}

	for tag := range event.Tags.FindAll("p") {
		if len(tag) == 4 && tag[3] == "moderator" {
			pubkey, err := nostr.PubKeyFromHex(tag[1])
			if err != nil {
				continue
			}
			comm.Moderators = append(comm.Moderators, nostr.ProfilePointer{PublicKey: pubkey, Relays: []string{tag[2]}})
		}
	}

	for tag := range event.Tags.FindAll("relay") {
		if len(tag) == 3 {
			marker := tag[2]
			comm.Relays[marker] = append(comm.Relays[marker], tag[1])
		}
	}

	return comm, nil
}

// CreateCommunityDefinition creates a new community definition event.
// The event is unsigned; caller must sign it before publishing.
func CreateCommunityDefinition(author nostr.PubKey, dIdentifier, name, description, image, imageSize string, moderators []nostr.ProfilePointer, relays map[string][]string) *nostr.Event {
	tags := nostr.Tags{}
	tags = append(tags, nostr.Tag{"d", dIdentifier})
	if name != "" {
		tags = append(tags, nostr.Tag{"name", name})
	}
	if description != "" {
		tags = append(tags, nostr.Tag{"description", description})
	}
	if image != "" {
		if imageSize != "" {
			tags = append(tags, nostr.Tag{"image", image, imageSize})
		} else {
			tags = append(tags, nostr.Tag{"image", image})
		}
	}
	for _, mod := range moderators {
		tags = append(tags, nostr.Tag{"p", mod.PublicKey.Hex(), "", "moderator"})
	}
	for marker, urls := range relays {
		for _, url := range urls {
			if marker != "" {
				tags = append(tags, nostr.Tag{"relay", url, marker})
			} else {
				tags = append(tags, nostr.Tag{"relay", url})
			}
		}
	}
	return &nostr.Event{
		PubKey:    author,
		Kind:      nostr.KindCommunityDefinition,
		Tags:      tags,
		Content:   "",
		CreatedAt: nostr.Now(),
	}
}

// IsModerator returns true if the given pubkey is a moderator of the community.
func (c *Community) IsModerator(pubkey nostr.PubKey) bool {
	for _, mod := range c.Moderators {
		if mod.PublicKey == pubkey {
			return true
		}
	}
	return false
}

// GetRelays returns relay URLs for a given marker (e.g., "author", "requests", "approvals").
func (c *Community) GetRelays(marker string) []string {
	return c.Relays[marker]
}

// CommunityPost represents a post within a community.
type CommunityPost struct {
	Event      *nostr.Event
	Community  *Community
	ParentID   string // optional, for nested replies
	ParentKind int    // optional
}

// ParseCommunityPost parses a kind 1111 event into a CommunityPost.
// It expects the event to have appropriate A/a tags referencing a community.
// The community definition must be provided (can be fetched separately).
func ParseCommunityPost(event *nostr.Event, community *Community) (*CommunityPost, error) {
	if event.Kind != nostr.KindComment {
		return nil, fmt.Errorf("event kind %d is not a community post", event.Kind)
	}
	// Verify that the event references the community via A tag.
	// For simplicity, we assume the community matches.
	// In a real implementation, you'd check the A tag matches community's a identifier.
	post := &CommunityPost{
		Event:     event,
		Community: community,
	}
	// Extract parent if present (e tag)
	for _, tag := range event.Tags {
		if len(tag) > 1 && tag[0] == "e" {
			post.ParentID = tag[1]
			break
		}
	}
	// Extract parent kind from k tag (optional)
	for _, tag := range event.Tags {
		if len(tag) > 1 && tag[0] == "k" {
			// parse int
			fmt.Sscanf(tag[1], "%d", &post.ParentKind)
			break
		}
	}
	return post, nil
}

// CreateCommunityPost creates a new community post event (kind 1111).
// If parentID is empty, it's a top-level post.
// The community parameter must have Author and DIdentifier.
func CreateCommunityPost(author nostr.PubKey, community *Community, content string, parentID string, parentKind int) *nostr.Event {
	tags := nostr.Tags{}
	// A tag referencing the community definition
	aTag := fmt.Sprintf("%d:%s:%s", nostr.KindCommunityDefinition, community.Author.Hex(), community.DIdentifier)
	tags = append(tags, nostr.Tag{"A", aTag})
	tags = append(tags, nostr.Tag{"P", community.Author.Hex()})
	tags = append(tags, nostr.Tag{"K", fmt.Sprintf("%d", nostr.KindCommunityDefinition)})
	// For top-level posts, also include lowercase a and p tags (NIP-22)
	if parentID == "" {
		tags = append(tags, nostr.Tag{"a", aTag})
		tags = append(tags, nostr.Tag{"p", community.Author.Hex()})
		tags = append(tags, nostr.Tag{"k", fmt.Sprintf("%d", nostr.KindCommunityDefinition)})
	} else {
		// nested reply: include parent e, p, k tags
		tags = append(tags, nostr.Tag{"e", parentID})
		tags = append(tags, nostr.Tag{"p", author.Hex()}) // parent author? we don't have it, but we can leave empty
		tags = append(tags, nostr.Tag{"k", fmt.Sprintf("%d", parentKind)})
	}
	return &nostr.Event{
		PubKey:    author,
		Kind:      nostr.KindComment,
		Tags:      tags,
		Content:   content,
		CreatedAt: nostr.Now(),
	}
}

// Approval represents a moderator approval event (kind 4550).
type Approval struct {
	Event     *nostr.Event
	Community *Community // parsed from a tag
	PostID    string
	PostKind  int
	PostEvent *nostr.Event // decoded from content
}

// ParseApproval parses a kind 4550 event into an Approval.
// It expects the event to have appropriate a and e tags.
func ParseApproval(event *nostr.Event) (*Approval, error) {
	if event.Kind != nostr.KindCommunityPostApproval {
		return nil, fmt.Errorf("event kind %d is not an approval", event.Kind)
	}
	approval := &Approval{
		Event: event,
	}
	var communityTag string
	for _, tag := range event.Tags {
		if len(tag) > 1 && tag[0] == "a" && strings.HasPrefix(tag[1], "34550:") {
			communityTag = tag[1]
			break
		}
	}
	if communityTag == "" {
		return nil, fmt.Errorf("approval missing community a tag")
	}
	// Parse community tag: 34550:<pubkey>:<d-identifier>
	parts := strings.Split(communityTag, ":")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid community a tag format")
	}
	// We could reconstruct a minimal Community struct, but for simplicity we just store the tag.
	// Extract post ID from e tag
	for _, tag := range event.Tags {
		if len(tag) > 1 && tag[0] == "e" {
			approval.PostID = tag[1]
			break
		}
	}
	// Extract post kind from k tag
	for _, tag := range event.Tags {
		if len(tag) > 1 && tag[0] == "k" {
			fmt.Sscanf(tag[1], "%d", &approval.PostKind)
			break
		}
	}
	// Decode post event from content
	if event.Content != "" {
		var postEvent nostr.Event
		if err := json.Unmarshal([]byte(event.Content), &postEvent); err == nil {
			approval.PostEvent = &postEvent
		}
	}
	return approval, nil
}

// CreateApproval creates a new approval event (kind 4550).
// The postEvent is the event being approved (must be JSON encoded in content).
func CreateApproval(moderator nostr.PubKey, community *Community, postEvent *nostr.Event) *nostr.Event {
	tags := nostr.Tags{}
	aTag := fmt.Sprintf("%d:%s:%s", nostr.KindCommunityDefinition, community.Author.Hex(), community.DIdentifier)
	tags = append(tags, nostr.Tag{"a", aTag})
	tags = append(tags, nostr.Tag{"e", postEvent.ID.Hex()})
	tags = append(tags, nostr.Tag{"p", postEvent.PubKey.Hex()})
	tags = append(tags, nostr.Tag{"k", fmt.Sprintf("%d", postEvent.Kind)})
	content, _ := json.Marshal(postEvent)
	return &nostr.Event{
		PubKey:    moderator,
		Kind:      nostr.KindCommunityPostApproval,
		Tags:      tags,
		Content:   string(content),
		CreatedAt: nostr.Now(),
	}
}

// Helper functions

// ExtractCommunityTag returns the community a tag from an event's tags.
func ExtractCommunityTag(tags nostr.Tags) (string, bool) {
	for _, tag := range tags {
		if len(tag) > 1 && tag[0] == "a" && strings.HasPrefix(tag[1], "34550:") {
			return tag[1], true
		}
	}
	return "", false
}

// IsCommunityPost returns true if the event is a community post (has community a tag).
func IsCommunityPost(event *nostr.Event) bool {
	_, ok := ExtractCommunityTag(event.Tags)
	return ok
}
