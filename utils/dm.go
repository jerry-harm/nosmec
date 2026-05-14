package utils

import (
	"context"
	"fmt"
	"sort"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/keyer"
	"fiatjaf.com/nostr/nip17"
	"fiatjaf.com/nostr/nip59"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/logger"
)

func SendDM(ctx context.Context, app *config.AppContext, recipientPubKey nostr.PubKey, content string) error {
	secretKey, err := app.GetMySecretKey()
	if err != nil {
		return err
	}

	kr := keyer.NewPlainKeySigner(secretKey)

	ourRelays := app.ListDMRelays()
	if len(ourRelays) == 0 {
		ourRelays = app.ReadableRelays()
	}

	var theirRelays []string
	theirRelays, err = FetchRecipientDMRelays(ctx, app, recipientPubKey)
	if err != nil {
		logger.Debug("failed to fetch recipient DM relays", "error", err.Error())
	}

	if len(theirRelays) == 0 {
		theirRelays, err = FetchRecipientReadRelays(ctx, app, recipientPubKey)
		if err != nil {
			logger.Debug("failed to fetch recipient read relays", "error", err.Error())
		}
	}

	if len(theirRelays) == 0 {
		logger.Debug("recipient has no published relay list, sending to our relays only")
		theirRelays = nil
	}

	return nip17.PublishMessage(
		ctx,
		content,
		nostr.Tags{{"p", recipientPubKey.Hex()}},
		app.Pool(),
		ourRelays,
		theirRelays,
		kr,
		recipientPubKey,
		nil,
	)
}

func ListenForDMs(ctx context.Context, app *config.AppContext, since nostr.Timestamp) chan nostr.Event {
	ourDMRelays := app.ListDMRelays()
	if len(ourDMRelays) == 0 {
		ourDMRelays = app.ReadableRelays()
	}

	if len(ourDMRelays) == 0 {
		ch := make(chan nostr.Event)
		close(ch)
		return ch
	}

	secretKey, err := app.GetMySecretKey()
	if err != nil {
		ch := make(chan nostr.Event)
		close(ch)
		return ch
	}

	kr := keyer.NewPlainKeySigner(secretKey)

	return nip17.ListenForMessages(ctx, app.Pool(), kr, ourDMRelays, since)
}

type Conversation struct {
	PubKey       string
	LatestDM     DMMessage
	LatestAt     nostr.Timestamp
}

type DMMessage struct {
	Content   string
	FromMe    bool
	Timestamp nostr.Timestamp
}

func ListDMConversations(ctx context.Context, app *config.AppContext, limit int) ([]Conversation, error) {
	secretKey, err := app.GetMySecretKey()
	if err != nil {
		return nil, err
	}

	kr := keyer.NewPlainKeySigner(secretKey)
	ourPubKey, err := kr.GetPublicKey(ctx)
	if err != nil {
		return nil, err
	}

	relays := app.ListDMRelays()
	if len(relays) == 0 {
		relays = app.ReadableRelays()
	}
	if len(relays) == 0 {
		relays = app.Config().KnownRelays
	}
	if len(relays) == 0 {
		relays = app.AllReadableRelays()
	}
	if len(relays) == 0 {
		return nil, fmt.Errorf("no relays available to query")
	}

	filter := nostr.Filter{
		Kinds: []nostr.Kind{nostr.KindGiftWrap},
		Tags:  nostr.TagMap{"p": []string{ourPubKey.Hex()}},
		Limit: limit * 3,
	}

	conversations := make(map[string]Conversation)
	seen := make(map[string]bool)

	for ie := range app.Pool().SubscribeMany(ctx, relays, filter, nostr.SubscriptionOptions{Label: "dmconversations"}) {
		fromMe := ie.Event.PubKey == ourPubKey

		var otherPubKey string
		if fromMe {
			for _, tag := range ie.Event.Tags {
				if len(tag) >= 2 && tag[0] == "p" && tag[1] != ourPubKey.Hex() {
					otherPubKey = tag[1]
					break
				}
			}
		} else {
			otherPubKey = ie.Event.PubKey.Hex()
		}

		if otherPubKey == "" || seen[otherPubKey] {
			continue
		}

		rumor, err := nip59.GiftUnwrap(
			ie.Event,
			func(otherpubkey nostr.PubKey, ciphertext string) (string, error) {
				return kr.Decrypt(ctx, ciphertext, otherpubkey)
			},
		)
		if err != nil {
			continue
		}

		seen[otherPubKey] = true
		preview := rumor.Content
		if len(preview) > 50 {
			preview = preview[:50] + "..."
		}

		conversations[otherPubKey] = Conversation{
			PubKey:   otherPubKey,
			LatestDM: DMMessage{Content: preview, FromMe: fromMe, Timestamp: rumor.CreatedAt},
			LatestAt: ie.Event.CreatedAt,
		}

		if len(conversations) >= limit {
			break
		}
	}

	result := make([]Conversation, 0, len(conversations))
	for _, conv := range conversations {
		result = append(result, conv)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].LatestAt > result[j].LatestAt
	})

	return result, nil
}

func QueryDMHistory(ctx context.Context, app *config.AppContext, recipientPubKey nostr.PubKey, limit int) ([]DMMessage, error) {
	secretKey, err := app.GetMySecretKey()
	if err != nil {
		return nil, err
	}

	kr := keyer.NewPlainKeySigner(secretKey)
	ourPubKey, err := kr.GetPublicKey(ctx)
	if err != nil {
		return nil, err
	}

	relays := app.ListDMRelays()
	if len(relays) == 0 {
		relays = app.ReadableRelays()
	}
	if len(relays) == 0 {
		relays = app.Config().KnownRelays
	}
	if len(relays) == 0 {
		relays = app.AllReadableRelays()
	}
	if len(relays) == 0 {
		return nil, fmt.Errorf("no relays available to query")
	}

	filter := nostr.Filter{
		Kinds: []nostr.Kind{nostr.KindGiftWrap},
		Tags:  nostr.TagMap{"p": []string{ourPubKey.Hex(), recipientPubKey.Hex()}},
		Limit: limit * 3,
	}

	var messages []DMMessage
	seen := make(map[string]bool)

	for ie := range app.Pool().SubscribeMany(ctx, relays, filter, nostr.SubscriptionOptions{Label: "dmhistory"}) {
		rumor, err := nip59.GiftUnwrap(
			ie.Event,
			func(otherpubkey nostr.PubKey, ciphertext string) (string, error) {
				return kr.Decrypt(ctx, ciphertext, otherpubkey)
			},
		)
		if err != nil {
			continue
		}

		fromMe := rumor.PubKey == ourPubKey
		msgID := rumor.ID.Hex()
		if seen[msgID] {
			continue
		}
		seen[msgID] = true

		messages = append(messages, DMMessage{
			Content:   rumor.Content,
			FromMe:   fromMe,
			Timestamp: rumor.CreatedAt,
		})
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp < messages[j].Timestamp
	})

	if len(messages) > limit {
		messages = messages[:limit]
	}

	return messages, nil
}
