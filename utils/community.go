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
	Event       *nostr.Event // raw Nostr event for event detail view
}

func FetchCommunityEvents(ctx context.Context, app *config.AppContext) ([]CommunityDefinition, error) {
	relays := app.AllReadableRelays()
	if len(relays) == 0 {
		relays = app.Config().KnownRelays
	}
	if len(relays) == 0 {
		return nil, fmt.Errorf("no relays available")
	}

	filter := nostr.Filter{
		Kinds: []nostr.Kind{nostr.KindCommunityDefinition},
	}

	ctxQuery, cancel := context.WithTimeout(ctx, app.QueryTimeout())
	defer cancel()

	results := app.Pool().FetchMany(ctxQuery, relays, filter, nostr.SubscriptionOptions{})

	seen := map[string]bool{}
	var communities []CommunityDefinition
	for relayEvent := range results {
		ev := relayEvent.Event
		addr := fmt.Sprintf("%d:%s:%s", ev.Kind, ev.PubKey.Hex(), ev.Tags.GetD())
		if seen[addr] {
			continue
		}
		seen[addr] = true

		def := CommunityDefinition{
			ID:     ev.Tags.GetD(),
			Relays: map[string]string{},
		}
		for _, tag := range ev.Tags {
			switch tag[0] {
			case "d":
				def.ID = tag[1]
			case "name":
				if len(tag) > 1 {
					def.Name = tag[1]
				}
			case "description":
				if len(tag) > 1 {
					def.Description = tag[1]
				}
			case "image":
				if len(tag) > 1 {
					def.ImageURL = tag[1]
				}
			case "p":
				if len(tag) > 1 {
					if pk, err := nostr.PubKeyFromHex(tag[1]); err == nil {
						def.Moderators = append(def.Moderators, pk)
					}
				}
			case "relay":
				if len(tag) > 1 {
					def.Relays[tag[1]] = tag[1]
				}
			}
		}
		def.Event = &ev
		communities = append(communities, def)
	}

	return communities, nil
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
	var err error
	pubKey, err = nostr.PubKeyFromHex(parts[1])
	if err != nil {
		return nostr.PubKey{}, "", fmt.Errorf("invalid community pubkey: %w", err)
	}
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
