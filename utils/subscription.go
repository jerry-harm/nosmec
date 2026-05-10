package utils

import (
	"context"
	"fmt"
	"strings"
	"time"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/logger"
)

func FollowCommunity(ctx context.Context, app *config.AppContext, communityAddr string, relay string) error {
	sub := config.Subscription{
		Type:  "community",
		ID:    communityAddr,
		Relay: relay,
	}
	return app.AddSubscription(sub)
}

func FollowUser(ctx context.Context, app *config.AppContext, pubkeyStr string, relay string, petname string) error {
	_, err := ResolveAliasToPubKey(app, pubkeyStr)
	if err != nil {
		return fmt.Errorf("invalid pubkey: %w", err)
	}

	sub := config.Subscription{
		Type:    "user",
		ID:      pubkeyStr,
		Relay:   relay,
		Petname: petname,
	}
	return app.AddSubscription(sub)
}

func FollowHashtag(ctx context.Context, app *config.AppContext, hashtag string) error {
	hashtag = strings.TrimPrefix(hashtag, "#")
	sub := config.Subscription{
		Type: "hashtag",
		ID:   hashtag,
	}
	return app.AddSubscription(sub)
}

func UnfollowCommunity(ctx context.Context, app *config.AppContext, communityAddr string) error {
	return app.RemoveSubscription("community", communityAddr)
}

func UnfollowUser(ctx context.Context, app *config.AppContext, pubkeyStr string) error {
	return app.RemoveSubscription("user", pubkeyStr)
}

func UnfollowHashtag(ctx context.Context, app *config.AppContext, hashtag string) error {
	hashtag = strings.TrimPrefix(hashtag, "#")
	return app.RemoveSubscription("hashtag", hashtag)
}

func SyncSubscriptionsFromNetwork(ctx context.Context, app *config.AppContext) error {
	secretKey, err := app.GetMySecretKey()
	if err != nil {
		return fmt.Errorf("failed to get secret key: %w", err)
	}
	pubKey := secretKey.Public()

	communities, err := syncCommunitiesFromNetwork(ctx, app, pubKey)
	if err != nil {
		return fmt.Errorf("failed to sync communities: %w", err)
	}

	users, err := syncUsersFromNetwork(ctx, app, pubKey)
	if err != nil {
		return fmt.Errorf("failed to sync users: %w", err)
	}

	hashtags, err := syncHashtagsFromNetwork(ctx, app, pubKey)
	if err != nil {
		return fmt.Errorf("failed to sync hashtags: %w", err)
	}

	subscriptions := append(communities, users...)
	subscriptions = append(subscriptions, hashtags...)

	if err := app.ReplaceAllSubscriptions(subscriptions); err != nil {
		return fmt.Errorf("failed to save subscriptions: %w", err)
	}

	return nil
}

func syncCommunitiesFromNetwork(ctx context.Context, app *config.AppContext, pubKey nostr.PubKey) ([]config.Subscription, error) {
	filter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindCommunityList},
		Authors: []nostr.PubKey{pubKey},
		Limit:   1,
	}

	readableRelays := app.ReadableRelays()
	if len(readableRelays) == 0 {
		readableRelays = app.Config().KnownRelays
	}

	var event nostr.Event
	for _, relay := range readableRelays {
		ctxTimeout, cancel := context.WithTimeout(ctx, app.QueryTimeout())
		result := app.Pool().QuerySingle(ctxTimeout, []string{relay}, filter, nostr.SubscriptionOptions{})
		cancel()
		if result != nil && result.Event.ID != [32]byte{} {
			event = result.Event
			break
		}
	}

	if event.ID == [32]byte{} {
		return []config.Subscription{}, nil
	}

	subscriptions := make([]config.Subscription, 0)
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "a" && strings.HasPrefix(tag[1], "34550:") {
			relay := ""
			if len(tag) >= 3 {
				relay = tag[2]
			}
			subscriptions = append(subscriptions, config.Subscription{
				Type:  "community",
				ID:    tag[1],
				Relay: relay,
			})
		}
	}

	return subscriptions, nil
}

func syncUsersFromNetwork(ctx context.Context, app *config.AppContext, pubKey nostr.PubKey) ([]config.Subscription, error) {
	filter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindFollowList},
		Authors: []nostr.PubKey{pubKey},
		Limit:   1,
	}

	readableRelays := app.ReadableRelays()
	if len(readableRelays) == 0 {
		readableRelays = app.Config().KnownRelays
	}

	var event nostr.Event
	for _, relay := range readableRelays {
		ctxTimeout, cancel := context.WithTimeout(ctx, app.QueryTimeout())
		result := app.Pool().QuerySingle(ctxTimeout, []string{relay}, filter, nostr.SubscriptionOptions{})
		cancel()
		if result != nil && result.Event.ID != [32]byte{} {
			event = result.Event
			break
		}
	}

	if event.ID == [32]byte{} {
		return []config.Subscription{}, nil
	}

	subscriptions := make([]config.Subscription, 0)
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "p" {
			relay := ""
			petname := ""
			if len(tag) >= 3 {
				relay = tag[2]
			}
			if len(tag) >= 4 {
				petname = tag[3]
			}

			pubKeyHex := tag[1]
			var pk nostr.PubKey
			copy(pk[:], []byte(pubKeyHex))
			npub := nip19.EncodeNpub(pk)
			subscriptions = append(subscriptions, config.Subscription{
				Type:    "user",
				ID:      npub,
				Relay:   relay,
				Petname: petname,
			})
		}
	}

	return subscriptions, nil
}

func syncHashtagsFromNetwork(ctx context.Context, app *config.AppContext, pubKey nostr.PubKey) ([]config.Subscription, error) {
	filter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindInterestList},
		Authors: []nostr.PubKey{pubKey},
		Limit:   1,
	}

	readableRelays := app.ReadableRelays()
	if len(readableRelays) == 0 {
		readableRelays = app.Config().KnownRelays
	}

	var event nostr.Event
	for _, relay := range readableRelays {
		ctxTimeout, cancel := context.WithTimeout(ctx, app.QueryTimeout())
		result := app.Pool().QuerySingle(ctxTimeout, []string{relay}, filter, nostr.SubscriptionOptions{})
		cancel()
		if result != nil && result.Event.ID != [32]byte{} {
			event = result.Event
			break
		}
	}

	if event.ID == [32]byte{} {
		return []config.Subscription{}, nil
	}

	subscriptions := make([]config.Subscription, 0)
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "t" {
			subscriptions = append(subscriptions, config.Subscription{
				Type: "hashtag",
				ID:   tag[1],
			})
		}
	}

	return subscriptions, nil
}

func PublishSubscriptions(ctx context.Context, app *config.AppContext) error {
	secretKey, err := app.GetMySecretKey()
	if err != nil {
		return fmt.Errorf("failed to get secret key: %w", err)
	}

	if err := publishFollowList(ctx, app, secretKey); err != nil {
		return fmt.Errorf("failed to publish follow list: %w", err)
	}

	if err := publishCommunitiesList(ctx, app, secretKey); err != nil {
		return fmt.Errorf("failed to publish communities list: %w", err)
	}

	if err := publishInterestsList(ctx, app, secretKey); err != nil {
		return fmt.Errorf("failed to publish interests list: %w", err)
	}

	return nil
}

func publishFollowList(ctx context.Context, app *config.AppContext, secretKey nostr.SecretKey) error {
	subscriptions := app.ListSubscriptions("user")

	tags := nostr.Tags{}
	skipped := 0
	for _, sub := range subscriptions {
		pubKey, err := ResolveAliasToPubKey(app, sub.ID)
		if err != nil {
			logger.Warn("skipping user subscription", "id", sub.ID, "error", err.Error())
			skipped++
			continue
		}

		relay := sub.Relay
		petname := sub.Petname
		tags = append(tags, nostr.Tag{"p", pubKey.Hex(), relay, petname})
	}

	if skipped > 0 {
		logger.Warn("skipped user subscriptions during publish", "count", skipped)
	}

	event := &nostr.Event{
		Kind:      nostr.KindFollowList,
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

func publishCommunitiesList(ctx context.Context, app *config.AppContext, secretKey nostr.SecretKey) error {
	subscriptions := app.ListSubscriptions("community")

	tags := nostr.Tags{}
	for _, sub := range subscriptions {
		if sub.Relay != "" {
			tags = append(tags, nostr.Tag{"a", sub.ID, sub.Relay})
		} else {
			tags = append(tags, nostr.Tag{"a", sub.ID})
		}
	}

	event := &nostr.Event{
		Kind:      nostr.KindCommunityList,
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

func publishInterestsList(ctx context.Context, app *config.AppContext, secretKey nostr.SecretKey) error {
	subscriptions := app.ListSubscriptions("hashtag")

	tags := nostr.Tags{}
	for _, sub := range subscriptions {
		tags = append(tags, nostr.Tag{"t", sub.ID})
	}

	event := &nostr.Event{
		Kind:      nostr.KindInterestList,
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
