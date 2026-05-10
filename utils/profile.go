package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"github.com/jerry-harm/nosmec/config"
)

type ProfileMetadata struct {
	Name        string `json:"name,omitempty"`
	About       string `json:"about,omitempty"`
	Picture     string `json:"picture,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Website     string `json:"website,omitempty"`
	Banner      string `json:"banner,omitempty"`
	Bot         *bool  `json:"bot,omitempty"`
	Birthday    string `json:"birthday,omitempty"`
	NIP05       string `json:"nip05,omitempty"`
	Lud06       string `json:"lud06,omitempty"`
	Lud16       string `json:"lud16,omitempty"`
}

type RelayInfo struct {
	URL   string `json:"url"`
	Read  bool   `json:"read"`
	Write bool   `json:"write"`
}

type FollowInfo struct {
	PubKey  string `json:"pubkey"`
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
	NPub       string           `json:"npub"`
	PubKey     string           `json:"pubkey"`
	Metadata   *ProfileMetadata `json:"metadata,omitempty"`
	Relays     []RelayInfo      `json:"relays,omitempty"`
	DMRelays   []string         `json:"dm_relays,omitempty"`
	Follows    []FollowInfo     `json:"follows,omitempty"`
	Communities []CommunityInfo `json:"communities,omitempty"`
	Hashtags   []HashtagInfo    `json:"hashtags,omitempty"`
}

func profileConfigToMetadata(pc config.ProfileConfig) ProfileMetadata {
	return ProfileMetadata{
		Name:        pc.Name,
		About:       pc.About,
		Picture:     pc.Picture,
		DisplayName: pc.DisplayName,
		Website:     pc.Website,
		Banner:      pc.Banner,
		Bot:         pc.Bot,
		Birthday:    pc.Birthday,
		NIP05:       pc.NIP05,
		Lud06:       pc.Lud06,
		Lud16:       pc.Lud16,
	}
}

func metadataToProfileConfig(pm ProfileMetadata) config.ProfileConfig {
	return config.ProfileConfig{
		Name:        pm.Name,
		About:       pm.About,
		Picture:     pm.Picture,
		DisplayName: pm.DisplayName,
		Website:     pm.Website,
		Banner:      pm.Banner,
		Bot:         pm.Bot,
		Birthday:    pm.Birthday,
		NIP05:       pm.NIP05,
		Lud06:       pm.Lud06,
		Lud16:       pm.Lud16,
	}
}

func SetProfile(ctx context.Context, app *config.AppContext, publishOnly bool, name, about, picture, displayName, website, banner, bot, birthday, nip05, lud06, lud16 string) (*nostr.Event, error) {
	secretKey, err := app.GetMySecretKey()
	if err != nil {
		return nil, err
	}

	cfg := app.Config()
	metadata := profileConfigToMetadata(cfg.Profile)

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
		metadata.Bot = &b
		profileChanged = true
	}
	if isSet(birthday) {
		metadata.Birthday = birthday
		profileChanged = true
	}
	if isSet(nip05) {
		metadata.NIP05 = nip05
		profileChanged = true
	}
	if isSet(lud06) {
		metadata.Lud06 = lud06
		profileChanged = true
	}
	if isSet(lud16) {
		metadata.Lud16 = lud16
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
		newCfg.Bot = metadata.Bot
		newCfg.Birthday = metadata.Birthday
		newCfg.NIP05 = metadata.NIP05
		newCfg.Lud06 = metadata.Lud06
		newCfg.Lud16 = metadata.Lud16

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

	event := GetProfile(ctx, pubKey, &GetOptions{App: app})
	if event == nil {
		return nil
	}

	var metadata ProfileMetadata
	if err := json.Unmarshal([]byte(event.Content), &metadata); err != nil {
		return fmt.Errorf("failed to parse profile: %w", err)
	}

	newCfg := metadataToProfileConfig(metadata)

	if err := app.SetProfile(newCfg); err != nil {
		return fmt.Errorf("failed to save profile config: %w", err)
	}

	return nil
}

func isSet(s string) bool {
	return s != "" && s != "<nil>"
}

func FetchRecipientDMRelays(ctx context.Context, app *config.AppContext, recipientPubKey nostr.PubKey) ([]string, error) {
	knownRelays := app.Config().KnownRelays
	if len(knownRelays) == 0 {
		return nil, fmt.Errorf("no known relays to query")
	}

	filter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindDMRelayList},
		Authors: []nostr.PubKey{recipientPubKey},
		Limit:   1,
	}

	ctx, cancel := context.WithTimeout(ctx, app.QueryTimeout())
	defer cancel()

	result := app.Pool().QuerySingle(ctx, knownRelays, filter, nostr.SubscriptionOptions{})
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

func FetchRecipientReadRelays(ctx context.Context, app *config.AppContext, recipientPubKey nostr.PubKey) ([]string, error) {
	knownRelays := app.Config().KnownRelays
	if len(knownRelays) == 0 {
		return nil, fmt.Errorf("no known relays to query")
	}

	filter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindRelayListMetadata},
		Authors: []nostr.PubKey{recipientPubKey},
		Limit:   1,
	}

	ctx, cancel := context.WithTimeout(ctx, app.QueryTimeout())
	defer cancel()

	result := app.Pool().QuerySingle(ctx, knownRelays, filter, nostr.SubscriptionOptions{})
	if result == nil || result.Event.ID == [32]byte{} {
		return nil, nil
	}

	var readRelays []string
	for _, tag := range result.Event.Tags {
		if len(tag) >= 2 && tag[0] == "r" {
			url := tag[1]
			for _, p := range tag[2:] {
				if p == "read" {
					readRelays = append(readRelays, url)
					break
				}
			}
		}
	}

	return readRelays, nil
}

func GetFullProfile(ctx context.Context, pubKey nostr.PubKey, opts *GetOptions) (*FullProfile, error) {
	if opts == nil || opts.App == nil {
		return nil, fmt.Errorf("nil options")
	}

	knownRelays := opts.App.Config().KnownRelays
	if len(knownRelays) == 0 {
		return nil, fmt.Errorf("no known relays")
	}

	fp := &FullProfile{
		NPub:   nip19.EncodeNpub(pubKey),
		PubKey: pubKey.Hex(),
	}

	metadataEvent := GetProfile(ctx, pubKey, opts)
	if metadataEvent != nil {
		var m ProfileMetadata
		if err := json.Unmarshal([]byte(metadataEvent.Content), &m); err == nil {
			fp.Metadata = &m
		}
	}

	filter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindRelayListMetadata},
		Authors: []nostr.PubKey{pubKey},
		Limit:   1,
	}
	result := opts.App.Pool().QuerySingle(ctx, knownRelays, filter, nostr.SubscriptionOptions{})
	if result != nil && result.Event.ID != [32]byte{} {
		for _, tag := range result.Event.Tags {
			if len(tag) >= 2 && tag[0] == "r" {
				r := RelayInfo{URL: tag[1]}
				for _, p := range tag[2:] {
					if p == "read" {
						r.Read = true
					} else if p == "write" {
						r.Write = true
					}
				}
				fp.Relays = append(fp.Relays, r)
			}
		}
	}

	dmRelays, _ := FetchRecipientDMRelays(ctx, opts.App, pubKey)
	if dmRelays != nil {
		fp.DMRelays = dmRelays
	}

	followFilter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindFollowList},
		Authors: []nostr.PubKey{pubKey},
		Limit:   1,
	}
	followResult := opts.App.Pool().QuerySingle(ctx, knownRelays, followFilter, nostr.SubscriptionOptions{})
	if followResult != nil && followResult.Event.ID != [32]byte{} {
		for _, tag := range followResult.Event.Tags {
			if len(tag) >= 2 && tag[0] == "p" {
				pkHex := tag[1]
				var pk nostr.PubKey
				copy(pk[:], []byte(pkHex))
				fi := FollowInfo{
					PubKey: pkHex,
					NPub:   nip19.EncodeNpub(pk),
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
	communityResult := opts.App.Pool().QuerySingle(ctx, knownRelays, communityFilter, nostr.SubscriptionOptions{})
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
	hashtagResult := opts.App.Pool().QuerySingle(ctx, knownRelays, hashtagFilter, nostr.SubscriptionOptions{})
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