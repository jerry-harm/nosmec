package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/nip72"
	"github.com/jerry-harm/nosmec/nostr_sdk"
)

type CommunityDefinition struct {
	Name        string
	Description string
	ImageURL    string
	Moderators  []nostr.PubKey
	Relays      []nip72.CommunityRelay
	ID          string
	Event       *nostr.Event // raw Nostr event for event detail view
}

func FetchCommunityEvents(ctx context.Context, app *config.AppContext) ([]CommunityDefinition, error) {
	filter := nostr.Filter{Kinds: []nostr.Kind{nostr.KindCommunityDefinition}}
	events, err := app.System().FetchEventsByFilter(ctx, filter, nostr_sdk.FetchEventsOptions{
		SaveToLocalStore: true,
	})
	if err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	communities := make([]CommunityDefinition, 0, len(events))
	for _, ev := range events {
		addr := fmt.Sprintf("%d:%s:%s", ev.Kind, ev.PubKey.Hex(), ev.Tags.GetD())
		if seen[addr] {
			continue
		}
		seen[addr] = true

		def := CommunityDefinition{
			ID:          nip72.GetDefinitionIdentifier(&ev),
			Name:        nip72.GetDefinitionName(&ev),
			Description: nip72.GetDefinitionDescription(&ev),
			ImageURL:    nip72.GetDefinitionImage(&ev),
			Moderators:  nip72.GetDefinitionModerators(&ev),
			Relays:      nip72.GetDefinitionRelays(&ev),
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

	for _, relay := range def.Relays {
		if relay.Purpose != "" {
			tags = append(tags, nostr.Tag{"relay", relay.URL, relay.Purpose})
		} else {
			tags = append(tags, nostr.Tag{"relay", relay.URL})
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
