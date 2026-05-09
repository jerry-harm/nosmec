package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
)

type CommunityDefinition struct {
	Name        string
	Description string
	ImageURL    string
	Moderators  []nostr.PubKey
	Relays      map[string]string
	ID          string
}

func CreateCommunity(ctx context.Context, app *config.AppContext, def CommunityDefinition) (*nostr.Event, error) {
	secretKey, err := app.GetMySecretKey()
	if err != nil {
		return nil, err
	}

	if def.ID == "" {
		return nil, fmt.Errorf("community ID is required")
	}

	tags := nostr.Tags{
		{"d", def.ID},
		{"name", def.Name},
		{"description", def.Description},
	}

	if def.ImageURL != "" {
		tags = append(tags, nostr.Tag{"image", def.ImageURL, "256x256"})
	}

	for _, moderator := range def.Moderators {
		tags = append(tags, nostr.Tag{"p", moderator.Hex(), "", "moderator"})
	}

	for purpose, relayURL := range def.Relays {
		if purpose != "" {
			tags = append(tags, nostr.Tag{"relay", relayURL, purpose})
		} else {
			tags = append(tags, nostr.Tag{"relay", relayURL})
		}
	}

	event := &nostr.Event{
		Kind:      nostr.KindCommunityDefinition,
		CreatedAt: nostr.Timestamp(time.Now().Unix()),
		Tags:      tags,
		Content:   "",
		PubKey:    secretKey.Public(),
	}

	if err := event.Sign(secretKey); err != nil {
		return nil, fmt.Errorf("failed to sign community event: %v", err)
	}

	writableRelays := app.WritableRelays()
	if len(writableRelays) > 0 {
		resultChan := app.Pool().PublishMany(ctx, writableRelays, *event)
		for result := range resultChan {
			if result.Error != nil {
				return nil, fmt.Errorf("failed to publish to %s: %w", result.RelayURL, result.Error)
			}
		}
	}

	privateRelays := app.PrivateRelays()
	if len(privateRelays) > 0 {
		go func() {
			app.Pool().PublishMany(context.Background(), privateRelays, *event)
		}()
	}

	return event, nil
}

func ParseCommunityAddr(addr string) (nostr.PubKey, string, error) {
	parts := strings.Split(addr, ":")
	if len(parts) != 3 {
		return nostr.PubKey{}, "", fmt.Errorf("invalid community address format: %s", addr)
	}
	if parts[0] != "34550" {
		return nostr.PubKey{}, "", fmt.Errorf("not a community address: %s", addr)
	}

	var pubKey nostr.PubKey
	copy(pubKey[:], []byte(parts[1]))
	return pubKey, parts[2], nil
}

func PostToCommunity(ctx context.Context, app *config.AppContext, communityAddr string, content string, parentID string) (*nostr.Event, error) {
	secretKey, err := app.GetMySecretKey()
	if err != nil {
		return nil, err
	}

	communityAuthor, _, err := ParseCommunityAddr(communityAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse community address: %w", err)
	}

	tags := nostr.Tags{
		{"A", communityAddr},
		{"a", communityAddr},
		{"P", communityAuthor.Hex()},
		{"p", communityAuthor.Hex()},
		{"K", "34550"},
		{"k", "34550"},
	}

	if parentID != "" {
		tags = append(tags, nostr.Tag{"e", parentID})
	}

	event := &nostr.Event{
		Kind:      nostr.KindComment,
		CreatedAt: nostr.Timestamp(time.Now().Unix()),
		Tags:      tags,
		Content:   content,
		PubKey:    secretKey.Public(),
	}

	if err := event.Sign(secretKey); err != nil {
		return nil, fmt.Errorf("failed to sign community post: %v", err)
	}

	writableRelays := app.WritableRelays()
	if len(writableRelays) > 0 {
		resultChan := app.Pool().PublishMany(ctx, writableRelays, *event)
		for result := range resultChan {
			if result.Error != nil {
				return nil, fmt.Errorf("failed to publish to %s: %w", result.RelayURL, result.Error)
			}
		}
	}

	privateRelays := app.PrivateRelays()
	if len(privateRelays) > 0 {
		go func() {
			app.Pool().PublishMany(context.Background(), privateRelays, *event)
		}()
	}

	return event, nil
}

func ApproveCommunityPost(ctx context.Context, app *config.AppContext, communityAuthor nostr.PubKey, communityID string, postEvent *nostr.Event) (*nostr.Event, error) {
	secretKey, err := app.GetMySecretKey()
	if err != nil {
		return nil, err
	}

	communityAddr := fmt.Sprintf("%s:%s:%s", nostr.KindCommunityDefinition.String(), communityAuthor.Hex(), communityID)

	postJSON, err := json.Marshal(postEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal post event: %v", err)
	}

	tags := nostr.Tags{
		{"a", communityAddr},
		{"e", postEvent.ID.Hex()},
		{"p", postEvent.PubKey.Hex()},
		{"k", fmt.Sprintf("%d", postEvent.Kind)},
	}

	event := &nostr.Event{
		Kind:      nostr.KindCommunityPostApproval,
		CreatedAt: nostr.Timestamp(time.Now().Unix()),
		Tags:      tags,
		Content:   string(postJSON),
		PubKey:    secretKey.Public(),
	}

	if err := event.Sign(secretKey); err != nil {
		return nil, fmt.Errorf("failed to sign approval event: %v", err)
	}

	return event, nil
}

func GetCommunity(ctx context.Context, app *config.AppContext, communityAuthor nostr.PubKey, communityID string) (*nostr.Event, error) {
	filter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindCommunityDefinition},
		Authors: []nostr.PubKey{communityAuthor},
		Tags:    nostr.TagMap{"d": []string{communityID}},
		Limit:   1,
	}
	opts := &GetOptions{App: app}
	event := GetEvent(ctx, filter, opts)
	if event == nil {
		return nil, fmt.Errorf("community not found: %s by %s", communityID, communityAuthor.Hex())
	}
	return event, nil
}

func GetCommunityPosts(ctx context.Context, app *config.AppContext, communityAuthor nostr.PubKey, communityID string, limit int) chan *nostr.Event {
	communityAddr := fmt.Sprintf("%d:%s:%s", nostr.KindCommunityDefinition, communityAuthor.Hex(), communityID)

	relays := app.Config().KnownRelays
	privateRelays := app.PrivateRelays()
	relays = append(relays, privateRelays...)

	filter := nostr.Filter{
		Kinds: []nostr.Kind{nostr.KindComment},
		Tags:  nostr.TagMap{"a": []string{communityAddr}},
		Limit: limit,
	}

	out := make(chan *nostr.Event)
	go func() {
		ch := app.Pool().FetchMany(ctx, relays, filter, nostr.SubscriptionOptions{})
		for relayEvent := range ch {
			CacheEvent(&relayEvent.Event, app)
			out <- &relayEvent.Event
		}
		close(out)
	}()
	return out
}

func GetFollowedCommunities(ctx context.Context, app *config.AppContext) ([]string, error) {
	pubKey, err := app.GetMyPubKey()
	if err != nil {
		return nil, err
	}

	filter := nostr.Filter{
		Kinds:   []nostr.Kind{10004},
		Authors: []nostr.PubKey{pubKey},
		Limit:   1,
	}

	opts := &GetOptions{App: app}
	event := GetEvent(ctx, filter, opts)
	if event == nil {
		return []string{}, nil
	}

	var communities []string
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "a" && strings.HasPrefix(tag[1], "34550:") {
			communities = append(communities, tag[1])
		}
	}

	return communities, nil
}

func GetPost(ctx context.Context, app *config.AppContext, postID string) (*nostr.Event, error) {
	var id nostr.ID
	if len(postID) == 64 {
		copy(id[:], postID)
	} else {
		return nil, fmt.Errorf("invalid post ID length")
	}

	filter := nostr.Filter{
		IDs: []nostr.ID{id},
		Limit: 1,
	}

	opts := &GetOptions{App: app}
	event := GetEvent(ctx, filter, opts)
	return event, nil
}

func GetParentPostInfo(ctx context.Context, app *config.AppContext, parentPostID string) (communityAddr string, authorPubKey nostr.PubKey, err error) {
	parentEvent, err := GetPost(ctx, app, parentPostID)
	if err != nil || parentEvent == nil {
		return "", nostr.PubKey{}, fmt.Errorf("parent post not found")
	}

	for _, tag := range parentEvent.Tags {
		if len(tag) >= 2 && tag[0] == "a" && strings.HasPrefix(tag[1], "34550:") {
			communityAddr = tag[1]
		}
		if len(tag) >= 2 && tag[0] == "p" {
			copy(authorPubKey[:], tag[1])
		}
	}

	if communityAddr == "" {
		return "", nostr.PubKey{}, fmt.Errorf("community address not found in parent post")
	}

	return communityAddr, authorPubKey, nil
}

func ReplyToCommunity(ctx context.Context, app *config.AppContext, parentPostID string, content string) (*nostr.Event, error) {
	communityAddr, _, err := GetParentPostInfo(ctx, app, parentPostID)
	if err != nil {
		return nil, err
	}

	return PostToCommunity(ctx, app, communityAddr, content, parentPostID)
}

func GetMyCreatedCommunities(ctx context.Context, app *config.AppContext, pubKey nostr.PubKey) chan *nostr.Event {
	relays := app.Config().KnownRelays
	privateRelays := app.PrivateRelays()
	relays = append(relays, privateRelays...)

	filter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindCommunityDefinition},
		Authors: []nostr.PubKey{pubKey},
		Limit:   100,
	}

	out := make(chan *nostr.Event)
	go func() {
		ch := app.Pool().FetchMany(ctx, relays, filter, nostr.SubscriptionOptions{})
		for relayEvent := range ch {
			CacheEvent(&relayEvent.Event, app)
			out <- &relayEvent.Event
		}
		close(out)
	}()
	return out
}

func GetPostedCommunities(ctx context.Context, app *config.AppContext, pubKey nostr.PubKey) chan string {
	relays := app.Config().KnownRelays
	privateRelays := app.PrivateRelays()
	relays = append(relays, privateRelays...)

	filter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindComment},
		Authors: []nostr.PubKey{pubKey},
		Limit:   500,
	}

	out := make(chan string)
	go func() {
		seen := make(map[nostr.ID]bool)
		ch := app.Pool().FetchMany(ctx, relays, filter, nostr.SubscriptionOptions{})
		for relayEvent := range ch {
			if seen[relayEvent.Event.ID] {
				continue
			}
			seen[relayEvent.Event.ID] = true
			CacheEvent(&relayEvent.Event, app)
			for _, tag := range relayEvent.Event.Tags {
				if len(tag) >= 2 && tag[0] == "a" && strings.HasPrefix(tag[1], "34550:") {
					out <- tag[1]
				}
			}
		}
		close(out)
	}()
	return out
}
