package utils

import (
	"context"
	"fmt"
	"time"

	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
)

func SyncRelaysFromNetwork(ctx context.Context, app *config.AppContext) error {
	pubKey, err := app.GetMyPubKey()
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	relays := app.WritableRelays()
	if len(relays) == 0 {
		relays = app.ReadableRelays()
	}
	if len(relays) == 0 {
		relays = app.Config().KnownRelays
	}
	if len(relays) == 0 {
		return fmt.Errorf("no relays available to query")
	}

	if err := syncRelayListFromNetwork(ctx, app, pubKey, relays); err != nil {
		return fmt.Errorf("failed to sync relay list: %w", err)
	}

	if err := syncDMRelaysFromNetwork(ctx, app, pubKey, relays); err != nil {
		return fmt.Errorf("failed to sync DM relays: %w", err)
	}

	return nil
}

func syncRelayListFromNetwork(ctx context.Context, app *config.AppContext, pubKey nostr.PubKey, relays []string) error {
	filter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindRelayListMetadata},
		Authors: []nostr.PubKey{pubKey},
		Limit:   1,
	}

	result := app.Pool().QuerySingle(ctx, relays, filter, nostr.SubscriptionOptions{})
	if result == nil {
		return nil
	}

	event := result.Event
	relayList := make([]config.Relay, 0)
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "r" {
			url := tag[1]
			relay := config.Relay{
				URL:   url,
				Read:  config.BoolPtr(false),
				Write: config.BoolPtr(false),
			}
			for _, p := range tag[2:] {
				if p == "read" {
					relay.Read = config.BoolPtr(true)
				} else if p == "write" {
					relay.Write = config.BoolPtr(true)
				}
			}
			relayList = append(relayList, relay)
		}
	}

	app.SyncRelayList(relayList)
	return nil
}

func syncDMRelaysFromNetwork(ctx context.Context, app *config.AppContext, pubKey nostr.PubKey, relays []string) error {
	filter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindDMRelayList},
		Authors: []nostr.PubKey{pubKey},
		Limit:   1,
	}

	result := app.Pool().QuerySingle(ctx, relays, filter, nostr.SubscriptionOptions{})
	if result == nil {
		return nil
	}

	event := result.Event
	var dmRelays []string
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "relay" {
			dmRelays = append(dmRelays, tag[1])
		}
	}

	app.SyncDMRelays(dmRelays)
	return nil
}

func PublishRelayList(ctx context.Context, app *config.AppContext) error {
	secretKey, err := app.GetMySecretKey()
	if err != nil {
		return fmt.Errorf("failed to get secret key: %w", err)
	}

	if err := publishRelayListMetadata(ctx, app, secretKey); err != nil {
		return fmt.Errorf("failed to publish relay list metadata: %w", err)
	}

	if err := publishDMRelayList(ctx, app, secretKey); err != nil {
		return fmt.Errorf("failed to publish DM relay list: %w", err)
	}

	return nil
}

func publishRelayListMetadata(ctx context.Context, app *config.AppContext, secretKey nostr.SecretKey) error {
	relayList := app.ListRelays()

	tags := nostr.Tags{}
	for _, relay := range relayList {
		tag := nostr.Tag{"r", relay.URL}
		if relay.Read != nil && *relay.Read {
			tag = append(tag, "read")
		}
		if relay.Write != nil && *relay.Write {
			tag = append(tag, "write")
		}
		tags = append(tags, tag)
	}

	event := &nostr.Event{
		Kind:      nostr.KindRelayListMetadata,
		CreatedAt: nostr.Timestamp(time.Now().Unix()),
		Tags:      tags,
		Content:   "",
		PubKey:    secretKey.Public(),
	}

	if err := event.Sign(secretKey); err != nil {
		return fmt.Errorf("failed to sign event: %w", err)
	}

	writableRelays := app.WritableRelays()
	if len(writableRelays) > 0 {
		resultChan := app.Pool().PublishMany(ctx, writableRelays, *event)
		for result := range resultChan {
			if result.Error != nil {
				return fmt.Errorf("failed to publish to %s: %w", result.RelayURL, result.Error)
			}
		}
	}

	return nil
}

func publishDMRelayList(ctx context.Context, app *config.AppContext, secretKey nostr.SecretKey) error {
	dmRelays := app.ListDMRelays()

	tags := nostr.Tags{}
	for _, relay := range dmRelays {
		tags = append(tags, nostr.Tag{"relay", relay})
	}

	event := &nostr.Event{
		Kind:      nostr.KindDMRelayList,
		CreatedAt: nostr.Timestamp(time.Now().Unix()),
		Tags:      tags,
		Content:   "",
		PubKey:    secretKey.Public(),
	}

	if err := event.Sign(secretKey); err != nil {
		return fmt.Errorf("failed to sign event: %w", err)
	}

	writableRelays := app.WritableRelays()
	if len(writableRelays) > 0 {
		resultChan := app.Pool().PublishMany(ctx, writableRelays, *event)
		for result := range resultChan {
			if result.Error != nil {
				return fmt.Errorf("failed to publish to %s: %w", result.RelayURL, result.Error)
			}
		}
	}

	return nil
}
