package utils

import (
	"context"
	"fmt"
	"strings"
	"time"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip10"
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

	ext := app.System()
	parentEvent := ext.FetchNote(ctx, parentID, app.QueryTimeoutms())
	if parentEvent == nil {
		return nil, fmt.Errorf("parent note not found: %s", parentID)
	}

	rootID, isRoot := findRootEvent(parentEvent)
	var rootPubKey string
	if !isRoot && rootID != parentEvent.ID {
		if rootEvent := ext.FetchNote(ctx, rootID.Hex(), app.QueryTimeoutms()); rootEvent != nil {
			rootPubKey = rootEvent.PubKey.Hex()
		}
	}

	tags := BuildReplyTagsWithRoot(parentEvent, rootPubKey)
	for _, p := range parentEvent.Tags {
		if len(p) >= 2 && p[0] == "p" {
			tags = append(tags, p)
		}
	}
	tags = append(tags, nostr.Tag{"p", parentEvent.PubKey.Hex()})

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

func BuildReplyTags(parentEvent *nostr.Event) nostr.Tags {
	return BuildReplyTagsWithRoot(parentEvent, "")
}

func BuildReplyTagsWithRoot(parentEvent *nostr.Event, rootPubKey string) nostr.Tags {
	rootID, isRoot := findRootEvent(parentEvent)
	rootRelay := config.GetEventRelay(rootID.Hex())
	parentRelay := config.GetEventRelay(parentEvent.ID.Hex())

	if isRoot {
		return nostr.Tags{{"e", parentEvent.ID.Hex(), parentRelay, "root", parentEvent.PubKey.Hex()}}
	}
	return nostr.Tags{
		{"e", rootID.Hex(), rootRelay, "root", rootPubKey},
		{"e", parentEvent.ID.Hex(), parentRelay, "reply", parentEvent.PubKey.Hex()},
	}
}

func QuoteNote(ctx context.Context, app *config.AppContext, quotedID, quotedAuthorPubkey, content string) (*nostr.Event, error) {
	secretKey, err := app.GetMySecretKey()
	if err != nil {
		return nil, err
	}

	relay := config.GetEventRelay(quotedID)
	tags := nostr.Tags{
		{"q", quotedID, relay, quotedAuthorPubkey},
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

func DeleteNote(ctx context.Context, app *config.AppContext, eventID, authorPubkey string) (*nostr.Event, error) {
	secretKey, err := app.GetMySecretKey()
	if err != nil {
		return nil, err
	}

	event := &nostr.Event{
		Kind:      nostr.KindDeletion,
		CreatedAt: nostr.Timestamp(time.Now().Unix()),
		Tags:      nostr.Tags{{"e", eventID}, {"p", authorPubkey}},
		Content:   "",
		PubKey:    secretKey.Public(),
	}

	if err := event.Sign(secretKey); err != nil {
		return nil, err
	}

	writableRelays := app.AllWritableRelays()
	if len(writableRelays) == 0 {
		return event, nil
	}

	var failed []string
	resultChan := app.Pool().PublishMany(ctx, writableRelays, *event)
	for result := range resultChan {
		if result.Error != nil {
			failed = append(failed, result.RelayURL)
		}
	}
	if len(failed) > 0 {
		return event, fmt.Errorf("partial failure on %d/%d relays: %s", len(failed), len(writableRelays), strings.Join(failed, ", "))
	}

	return event, nil
}

func findRootEvent(event *nostr.Event) (rootID nostr.ID, isRoot bool) {
	ptr := nip10.GetThreadRoot(event.Tags)
	if ptr == nil {
		return event.ID, true
	}
	if ep, ok := ptr.(nostr.EventPointer); ok {
		return ep.ID, false
	}
	return event.ID, true
}
