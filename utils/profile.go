package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"fiatjaf.com/nostr/sdk"
	"github.com/jerry-harm/nosmec/config"
)

type RelayInfo struct {
	URL   string `json:"url"`
	Read  bool   `json:"read"`
	Write bool   `json:"write"`
}

type FollowInfo struct {
	NPub    string `json:"npub"`
	Relay   string `json:"relay,omitempty"`
	Petname string `json:"petname,omitempty"`
}

type CommunityInfo struct {
	Addr  string `json:"addr"`
	Relay string `json:"relay,omitempty"`
}

type HashtagInfo struct {
	Tag string `json:"tag"`
}

type FullProfile struct {
	NPub       string            `json:"npub"`
	PubKey     string            `json:"pubkey"`
	Metadata   *sdk.ProfileMetadata `json:"metadata,omitempty"`
	Relays     []RelayInfo       `json:"relays,omitempty"`
	DMRelays   []string          `json:"dm_relays,omitempty"`
	Follows    []FollowInfo      `json:"follows,omitempty"`
	Communities []CommunityInfo  `json:"communities,omitempty"`
	Hashtags   []HashtagInfo     `json:"hashtags,omitempty"`
}

func profileConfigToMetadata(pc config.ProfileConfig) sdk.ProfileMetadata {
	return sdk.ProfileMetadata{
		Name:        pc.Name,
		About:       pc.About,
		Picture:     pc.Picture,
		DisplayName: pc.DisplayName,
		Website:     pc.Website,
		Banner:      pc.Banner,
		NIP05:       pc.NIP05,
		LUD16:       pc.Lud16,
	}
}

func metadataToProfileConfig(pm sdk.ProfileMetadata, bot *bool, birthday string) config.ProfileConfig {
	return config.ProfileConfig{
		Name:        pm.Name,
		About:       pm.About,
		Picture:     pm.Picture,
		DisplayName: pm.DisplayName,
		Website:     pm.Website,
		Banner:      pm.Banner,
		Bot:         bot,
		Birthday:    birthday,
		NIP05:       pm.NIP05,
		Lud16:       pm.LUD16,
	}
}

func SetProfile(ctx context.Context, app *config.AppContext, publishOnly bool, name, about, picture, displayName, website, banner, bot, birthday, nip05, lud06, lud16 string) (*nostr.Event, error) {
	secretKey, err := app.GetMySecretKey()
	if err != nil {
		return nil, err
	}

	cfg := app.Config()
	metadata := profileConfigToMetadata(cfg.Profile)
	var profileBot *bool
	var profileBirthday string
	profileNIP05 := metadata.NIP05

	profileChanged := false

	if isSet(name) {
		metadata.Name = name
		profileChanged = true
	}
	if isSet(about) {
		metadata.About = about
		profileChanged = true
	}
	if isSet(picture) {
		metadata.Picture = picture
		profileChanged = true
	}
	if isSet(displayName) {
		metadata.DisplayName = displayName
		profileChanged = true
	}
	if isSet(website) {
		metadata.Website = website
		profileChanged = true
	}
	if isSet(banner) {
		metadata.Banner = banner
		profileChanged = true
	}
	if isSet(bot) {
		b := bot == "true" || bot == "1"
		profileBot = &b
		profileChanged = true
	}
	if isSet(birthday) {
		profileBirthday = birthday
		profileChanged = true
	}
	if isSet(nip05) {
		metadata.NIP05 = nip05
		profileChanged = true
		profileNIP05 = nip05
	}
	if isSet(lud16) {
		metadata.LUD16 = lud16
		profileChanged = true
	}

	if !publishOnly && profileChanged {
		newCfg := cfg.Profile
		newCfg.Name = metadata.Name
		newCfg.About = metadata.About
		newCfg.Picture = metadata.Picture
		newCfg.DisplayName = metadata.DisplayName
		newCfg.Website = metadata.Website
		newCfg.Banner = metadata.Banner
		newCfg.Bot = profileBot
		newCfg.Birthday = profileBirthday
		newCfg.NIP05 = profileNIP05
		newCfg.Lud16 = metadata.LUD16

		if err := app.SetProfile(newCfg); err != nil {
			return nil, fmt.Errorf("failed to save profile config: %w", err)
		}
	}

	content, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	event := &nostr.Event{
		Kind:      nostr.KindProfileMetadata,
		CreatedAt: nostr.Timestamp(time.Now().Unix()),
		Tags:      nostr.Tags{},
		Content:   string(content),
		PubKey:    secretKey.Public(),
	}

	if err := event.Sign(secretKey); err != nil {
		return nil, err
	}

	writableRelays := app.WritableRelays()
	if len(writableRelays) > 0 {
		resultChan := app.Pool().PublishMany(ctx, writableRelays, *event)
		for result := range resultChan {
			if result.Error != nil {
				return nil, fmt.Errorf("failed to publish profile: %w", result.Error)
			}
		}
	}

	return event, nil
}

func SyncProfile(ctx context.Context, app *config.AppContext) error {
	pubKey, err := app.GetMyPubKey()
	if err != nil {
		return err
	}

	pm := app.System().FetchProfileMetadata(ctx, pubKey)
	if pm.Event == nil {
		return nil
	}

	metadata, err := sdk.ParseMetadata(*pm.Event)
	if err != nil {
		return fmt.Errorf("failed to parse profile: %w", err)
	}

	cfg := app.Config()
	newCfg := metadataToProfileConfig(metadata, cfg.Profile.Bot, cfg.Profile.Birthday)

	if err := app.SetProfile(newCfg); err != nil {
		return fmt.Errorf("failed to save profile config: %w", err)
	}

	return nil
}

func isSet(s string) bool {
	return s != "" && s != "<nil>"
}

func FetchRecipientDMRelays(ctx context.Context, app *config.AppContext, recipientPubKey nostr.PubKey, relays []string) ([]string, error) {
	if len(relays) == 0 {
		return nil, fmt.Errorf("no known relays to query")
	}

	filter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindDMRelayList},
		Authors: []nostr.PubKey{recipientPubKey},
		Limit:   1,
	}

	ctx, cancel := context.WithTimeout(ctx, app.QueryTimeout())
	defer cancel()

	result := app.Pool().QuerySingle(ctx, relays, filter, nostr.SubscriptionOptions{})
	if result == nil || result.Event.ID == [32]byte{} {
		return nil, nil
	}

	var dmRelays []string
	for _, tag := range result.Event.Tags {
		if len(tag) >= 2 && tag[0] == "relay" {
			dmRelays = append(dmRelays, tag[1])
		}
	}

	return dmRelays, nil
}

func FetchRecipientReadRelays(ctx context.Context, app *config.AppContext, recipientPubKey nostr.PubKey, relays []string) ([]string, error) {
	if len(relays) == 0 {
		return nil, fmt.Errorf("no known relays to query")
	}

	filter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindRelayListMetadata},
		Authors: []nostr.PubKey{recipientPubKey},
		Limit:   1,
	}

	ctx, cancel := context.WithTimeout(ctx, app.QueryTimeout())
	defer cancel()

	result := app.Pool().QuerySingle(ctx, relays, filter, nostr.SubscriptionOptions{})
	if result == nil || result.Event.ID == [32]byte{} {
		return nil, nil
	}

	var readRelays []string
	seen := make(map[string]bool)
	for _, tag := range result.Event.Tags {
		if len(tag) >= 2 && tag[0] == "r" {
			url := tag[1]
			if seen[url] {
				continue
			}
			for _, p := range tag[2:] {
				if p == "read" {
					readRelays = append(readRelays, url)
					seen[url] = true
					break
				}
			}
		}
	}

	return readRelays, nil
}

func GetFullProfile(ctx context.Context, pubKey nostr.PubKey, app *config.AppContext) (*FullProfile, error) {
	if app == nil {
		return nil, fmt.Errorf("nil app")
	}

	allRelays := app.AllReadableRelays()
	knownRelays := app.Config().KnownRelays
	seen := make(map[string]bool)
	relays := make([]string, 0, len(allRelays)+len(knownRelays))
	for _, r := range allRelays {
		if !seen[r] {
			relays = append(relays, r)
			seen[r] = true
		}
	}
	for _, r := range knownRelays {
		if !seen[r] {
			relays = append(relays, r)
			seen[r] = true
		}
	}
	if len(relays) == 0 {
		return nil, fmt.Errorf("no known relays")
	}

	fp := &FullProfile{
		NPub:   nip19.EncodeNpub(pubKey),
		PubKey: pubKey.Hex(),
	}

	pm := app.System().FetchProfileMetadata(ctx, pubKey)
	if pm.Event != nil {
		m, err := sdk.ParseMetadata(*pm.Event)
		if err == nil {
			fp.Metadata = &m
		}
	}

	filter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindRelayListMetadata},
		Authors: []nostr.PubKey{pubKey},
		Limit:   1,
	}
	result := app.Pool().QuerySingle(ctx, relays, filter, nostr.SubscriptionOptions{})
	if result != nil && result.Event.ID != [32]byte{} {
		relayMap := make(map[string]struct{ read, write bool })
		for _, tag := range result.Event.Tags {
			if len(tag) >= 2 && tag[0] == "r" {
				url := tag[1]
				r := relayMap[url]
				if len(tag) == 2 {
					r.read = true
					r.write = true
				} else {
					for _, p := range tag[2:] {
						if p == "read" {
							r.read = true
						} else if p == "write" {
							r.write = true
						}
					}
				}
				relayMap[url] = r
			}
		}
		for url, r := range relayMap {
			fp.Relays = append(fp.Relays, RelayInfo{URL: url, Read: r.read, Write: r.write})
		}
	}

	dmRelays, _ := FetchRecipientDMRelays(ctx, app, pubKey, relays)
	if dmRelays != nil {
		fp.DMRelays = dmRelays
	}

	followFilter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindFollowList},
		Authors: []nostr.PubKey{pubKey},
		Limit:   1,
	}
	followResult := app.Pool().QuerySingle(ctx, relays, followFilter, nostr.SubscriptionOptions{})
	if followResult != nil && followResult.Event.ID != [32]byte{} {
		for _, tag := range followResult.Event.Tags {
			if len(tag) >= 2 && tag[0] == "p" {
				pkHex := tag[1]
				pk, err := nostr.PubKeyFromHex(pkHex)
				if err != nil {
					continue
				}
				fi := FollowInfo{
					NPub: nip19.EncodeNpub(pk),
				}
				if len(tag) >= 3 {
					fi.Relay = tag[2]
				}
				if len(tag) >= 4 {
					fi.Petname = tag[3]
				}
				fp.Follows = append(fp.Follows, fi)
			}
		}
	}

	communityFilter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindCommunityList},
		Authors: []nostr.PubKey{pubKey},
		Limit:   1,
	}
	communityResult := app.Pool().QuerySingle(ctx, relays, communityFilter, nostr.SubscriptionOptions{})
	if communityResult != nil && communityResult.Event.ID != [32]byte{} {
		for _, tag := range communityResult.Event.Tags {
			if len(tag) >= 2 && tag[0] == "a" && len(tag[1]) > 0 {
				ci := CommunityInfo{Addr: tag[1]}
				if len(tag) >= 3 {
					ci.Relay = tag[2]
				}
				fp.Communities = append(fp.Communities, ci)
			}
		}
	}

	hashtagFilter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindInterestList},
		Authors: []nostr.PubKey{pubKey},
		Limit:   1,
	}
	hashtagResult := app.Pool().QuerySingle(ctx, relays, hashtagFilter, nostr.SubscriptionOptions{})
	if hashtagResult != nil && hashtagResult.Event.ID != [32]byte{} {
		for _, tag := range hashtagResult.Event.Tags {
			if len(tag) >= 2 && tag[0] == "t" {
				fp.Hashtags = append(fp.Hashtags, HashtagInfo{Tag: tag[1]})
			}
		}
	}

	return fp, nil
}

func SerializeProfile(fp *FullProfile) ([]byte, error) {
	return json.MarshalIndent(fp, "", "  ")
}