package utils

import (
	"context"
	"fmt"
	"time"

	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
)

func PostNote(ctx context.Context, app *config.AppContext, content string) (*nostr.Event, error) {
	secretKey, err := app.GetMySecretKey()
	if err != nil {
		return nil, err
	}

	event := &nostr.Event{
		Kind:      nostr.KindTextNote,
		CreatedAt: nostr.Timestamp(time.Now().Unix()),
		Tags:      nostr.Tags{},
		Content:   content,
		PubKey:    secretKey.Public(),
	}

	if err := event.Sign(secretKey); err != nil {
		return nil, err
	}

	writableRelays := app.AllWritableRelays()
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

func ReplyToNote(ctx context.Context, app *config.AppContext, parentID, content string) (*nostr.Event, error) {
	secretKey, err := app.GetMySecretKey()
	if err != nil {
		return nil, err
	}

	opts := &GetOptions{App: app}
	parentEvent := GetNote(ctx, parentID, opts)
	if parentEvent == nil {
		return nil, fmt.Errorf("parent note not found: %s", parentID)
	}

	tags := nostr.Tags{
		{"e", parentID, "", "reply"},
		{"p", parentEvent.PubKey.Hex()},
	}

	event := &nostr.Event{
		Kind:      nostr.KindTextNote,
		CreatedAt: nostr.Timestamp(time.Now().Unix()),
		Tags:      tags,
		Content:   content,
		PubKey:    secretKey.Public(),
	}

	if err := event.Sign(secretKey); err != nil {
		return nil, err
	}

	writableRelays := app.AllWritableRelays()
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

func QuoteNote(ctx context.Context, app *config.AppContext, quotedID, content string) (*nostr.Event, error) {
	secretKey, err := app.GetMySecretKey()
	if err != nil {
		return nil, err
	}

	tags := nostr.Tags{
		{"q", quotedID},
	}

	event := &nostr.Event{
		Kind:      nostr.KindTextNote,
		CreatedAt: nostr.Timestamp(time.Now().Unix()),
		Tags:      tags,
		Content:   content,
		PubKey:    secretKey.Public(),
	}

	if err := event.Sign(secretKey); err != nil {
		return nil, err
	}

	writableRelays := app.AllWritableRelays()
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

// DeleteNote sends a deletion request (Kind 5) for a given event ID.
func DeleteNote(ctx context.Context, app *config.AppContext, eventID string) (*nostr.Event, error) {
	secretKey, err := app.GetMySecretKey()
	if err != nil {
		return nil, err
	}

	event := &nostr.Event{
		Kind:      nostr.KindDeletion,
		CreatedAt: nostr.Timestamp(time.Now().Unix()),
		Tags:      nostr.Tags{{"e", eventID}},
		Content:   "",
		PubKey:    secretKey.Public(),
	}

	if err := event.Sign(secretKey); err != nil {
		return nil, err
	}

	writableRelays := app.AllWritableRelays()
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
