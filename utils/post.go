package utils

import (
	"context"
	"fmt"
	"strings"
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

	rootID, isRoot, _ := FindRootEvent(parentEvent)
	var rootPubKey string
	if !isRoot && rootID != parentEvent.ID {
		if rootEvent := GetNote(ctx, rootID.Hex(), opts); rootEvent != nil {
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

// BuildReplyTags is a convenience wrapper for BuildReplyTagsWithRoot.
// rootPubKey defaults to "" — use BuildReplyTagsWithRoot when root pubkey is available.
// Used by TUI compose display (which only needs e tag structure, not full NIP-10).
func BuildReplyTags(parentEvent *nostr.Event) nostr.Tags {
	return BuildReplyTagsWithRoot(parentEvent, "")
}

// BuildReplyTagsWithRoot creates NIP-10 marked e tags for a reply to a parent event.
// rootPubKey is the hex pubkey of the root event author (required for nested replies).
// Tags follow the full format: ["e", <id>, <relay>, <marker>, <pubkey>]
func BuildReplyTagsWithRoot(parentEvent *nostr.Event, rootPubKey string) nostr.Tags {
	rootID, isRoot, _ := FindRootEvent(parentEvent)
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
// Per NIP-09, only the author of an event can delete it.
// Publishes to writable relays and local relay to keep cache consistent.
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

	// Publish to local relay first so cache reflects deletion immediately
	if localURL := config.GetLocalRelayURL(); localURL != "" {
		go app.Pool().PublishMany(context.Background(), []string{localURL}, *event)
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