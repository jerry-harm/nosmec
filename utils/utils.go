package utils

import (
	"context"
	"fmt"
	"time"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"github.com/jerry-harm/nosmec/config"
)

// GetMyPubKey 从配置的私钥获取公钥
func GetMyPubKey() (nostr.PubKey, error) {
	cfg := config.GlobalConfig()
	if cfg.PrivateKey == "" {
		return nostr.PubKey{}, fmt.Errorf("no private key configured")
	}

	// Decode private key (nsec) to get hex secret
	_, s, err := nip19.Decode(cfg.PrivateKey)
	if err != nil {
		return nostr.PubKey{}, err
	}

	priv := s.(nostr.SecretKey)

	// Get public key from secret key
	return priv.Public(), nil
}

// GetProfileWithPointer 查询profile
func GetProfile(ctx context.Context, pubKey nostr.PubKey) (*nostr.Event, error) {
	var event nostr.Event
	relays := config.GlobalConfig().KnownRelays

	// Query for kind 0 (profile metadata)
	filter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindProfileMetadata},
		Authors: []nostr.PubKey{pubKey},
		Limit:   1,
	}

	for e := range config.GlobalLMDB().QueryEvents(filter, 1) {
		event = e
	}
	if event.ID != [32]byte{} && time.Now().Unix()-int64(event.CreatedAt) <= 86400 {
		return &event, nil
	} else {
		// SubscriptionOptions does not have Timeout, so adjust to context timeout
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		relay_event := config.GlobalPool().QuerySingle(ctx, relays, filter, nostr.SubscriptionOptions{})
		if relay_event == nil {
			return nil, fmt.Errorf("profile not found for pubkey: %s", pubKey.Hex())
		}
		config.GlobalLMDB().SaveEvent(relay_event.Event)
		return &relay_event.Event, nil
	}
}
