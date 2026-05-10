package config

import (
	"fmt"
	"sync"
	"time"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"github.com/spf13/viper"
)

type AppContext struct {
	pool         *nostr.Pool
	store        StoreInterface
	cfg          Config
	mu           sync.RWMutex
	viper        *viper.Viper
	knownRelays  map[string]struct{}
}

func NewAppContext(pool *nostr.Pool, store StoreInterface, cfg Config, v *viper.Viper) *AppContext {
	return &AppContext{
		pool:        pool,
		store:       store,
		cfg:         cfg,
		viper:       v,
		knownRelays: make(map[string]struct{}),
	}
}

func (a *AppContext) Pool() *nostr.Pool {
	return a.pool
}

func (a *AppContext) Store() StoreInterface {
	return a.store
}

func (a *AppContext) Config() Config {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg
}

func (a *AppContext) GetProfile() ProfileConfig {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.Profile
}

func (a *AppContext) SetProfile(profile ProfileConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.Profile = profile
	a.viper.Set("profile", profile)
	return a.viper.WriteConfig()
}

func (a *AppContext) GetPrivateKey() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.PrivateKey
}

func (a *AppContext) SetPrivateKey(sk string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.PrivateKey = sk
	a.viper.Set("private_key", sk)
	return a.viper.WriteConfig()
}

func (a *AppContext) WritableRelays() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return GetWritableRelaysFromList(a.cfg.RelayList)
}

func (a *AppContext) ReadableRelays() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return GetReadableRelaysFromList(a.cfg.RelayList)
}

func (a *AppContext) AllWritableRelays() []string {
	relays := a.WritableRelays()
	if localURL := a.localRelayURL(); localURL != "" {
		relays = append([]string{localURL}, relays...)
	}
	return relays
}

func (a *AppContext) AllReadableRelays() []string {
	relays := a.ReadableRelays()
	if localURL := a.localRelayURL(); localURL != "" {
		relays = append([]string{localURL}, relays...)
	}
	return relays
}

func (a *AppContext) localRelayURL() string {
	if !a.LocalRelayEnabled() {
		return ""
	}
	port := a.cfg.LocalRelay.Port
	if port == "" {
		port = "8989"
	}
	return fmt.Sprintf("ws://localhost:%s", port)
}

func (a *AppContext) QueryTimeout() time.Duration {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.cfg.Query.Timeout <= 0 {
		return 5 * time.Second
	}
	return time.Duration(a.cfg.Query.Timeout) * time.Second
}

func (a *AppContext) LocalRelayEnabled() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.LocalRelay.Enabled
}

func (a *AppContext) GetMyPubKey() (nostr.PubKey, error) {
	sk, err := a.GetMySecretKey()
	if err != nil {
		return nostr.PubKey{}, err
	}
	return sk.Public(), nil
}

func (a *AppContext) GetMySecretKey() (nostr.SecretKey, error) {
	privKey := a.GetPrivateKey()

	if privKey == "" {
		return nostr.SecretKey{}, fmt.Errorf("no private key configured")
	}

	_, s, err := nip19.Decode(privKey)
	if err != nil {
		return nostr.SecretKey{}, err
	}

	sk, ok := s.(nostr.SecretKey)
	if !ok {
		return nostr.SecretKey{}, fmt.Errorf("invalid private key format")
	}
	return sk, nil
}

func (a *AppContext) GetRelay(url string) (Relay, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, r := range a.cfg.RelayList {
		if r.URL == url {
			return r, true
		}
	}
	return Relay{}, false
}

func (a *AppContext) ListRelays() []Relay {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.RelayList
}

func (a *AppContext) AddRelay(url string, read, write bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for i, r := range a.cfg.RelayList {
		if r.URL == url {
			a.cfg.RelayList[i].Read = &read
			a.cfg.RelayList[i].Write = &write
			a.viper.Set("relay_list", a.cfg.RelayList)
			return a.viper.WriteConfig()
		}
	}

	a.cfg.RelayList = append(a.cfg.RelayList, Relay{URL: url, Read: &read, Write: &write})
	a.viper.Set("relay_list", a.cfg.RelayList)
	return a.viper.WriteConfig()
}

func (a *AppContext) RemoveRelay(url string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	newList := make([]Relay, 0)
	for _, r := range a.cfg.RelayList {
		if r.URL != url {
			newList = append(newList, r)
		}
	}
	a.cfg.RelayList = newList
	a.viper.Set("relay_list", a.cfg.RelayList)
	return a.viper.WriteConfig()
}

func (a *AppContext) SetRelayRead(url string, read bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for i, r := range a.cfg.RelayList {
		if r.URL == url {
			a.cfg.RelayList[i].Read = &read
			a.viper.Set("relay_list", a.cfg.RelayList)
			return a.viper.WriteConfig()
		}
	}
	return fmt.Errorf("relay not found: %s", url)
}

func (a *AppContext) SetRelayWrite(url string, write bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for i, r := range a.cfg.RelayList {
		if r.URL == url {
			a.cfg.RelayList[i].Write = &write
			a.viper.Set("relay_list", a.cfg.RelayList)
			return a.viper.WriteConfig()
		}
	}
	return fmt.Errorf("relay not found: %s", url)
}

func (a *AppContext) SyncRelayList(relays []Relay) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.RelayList = relays
	a.viper.Set("relay_list", relays)
	a.viper.WriteConfig()
}

func (a *AppContext) ListDMRelays() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.DMRelays
}

func (a *AppContext) AddDMRelay(url string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, u := range a.cfg.DMRelays {
		if u == url {
			return nil
		}
	}
	a.cfg.DMRelays = append(a.cfg.DMRelays, url)
	a.viper.Set("dm_relays", a.cfg.DMRelays)
	return a.viper.WriteConfig()
}

func (a *AppContext) RemoveDMRelay(url string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	newList := make([]string, 0)
	for _, u := range a.cfg.DMRelays {
		if u != url {
			newList = append(newList, u)
		}
	}
	a.cfg.DMRelays = newList
	a.viper.Set("dm_relays", a.cfg.DMRelays)
	return a.viper.WriteConfig()
}

func (a *AppContext) SyncDMRelays(relays []string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.DMRelays = relays
	a.viper.Set("dm_relays", relays)
	a.viper.WriteConfig()
}

func (a *AppContext) ListSearchRelays() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.SearchRelays
}

func (a *AppContext) AddSearchRelay(url string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, u := range a.cfg.SearchRelays {
		if u == url {
			return nil
		}
	}
	a.cfg.SearchRelays = append(a.cfg.SearchRelays, url)
	a.viper.Set("search_relays", a.cfg.SearchRelays)
	return a.viper.WriteConfig()
}

func (a *AppContext) RemoveSearchRelay(url string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	newList := make([]string, 0)
	for _, u := range a.cfg.SearchRelays {
		if u != url {
			newList = append(newList, u)
		}
	}
	a.cfg.SearchRelays = newList
	a.viper.Set("search_relays", a.cfg.SearchRelays)
	return a.viper.WriteConfig()
}

func (a *AppContext) ListSubscriptions(subType string) []Subscription {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.cfg.Subscriptions == nil {
		return []Subscription{}
	}

	if subType == "" {
		return a.cfg.Subscriptions
	}

	result := make([]Subscription, 0)
	for _, s := range a.cfg.Subscriptions {
		if s.Type == subType {
			result = append(result, s)
		}
	}
	return result
}

func (a *AppContext) AddSubscription(sub Subscription) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cfg.Subscriptions == nil {
		a.cfg.Subscriptions = []Subscription{}
	}

	for _, s := range a.cfg.Subscriptions {
		if s.Type == sub.Type && s.ID == sub.ID {
			return fmt.Errorf("already subscribed: %s %s", sub.Type, sub.ID)
		}
	}

	a.cfg.Subscriptions = append(a.cfg.Subscriptions, sub)
	a.viper.Set("subscriptions", a.cfg.Subscriptions)
	return a.viper.WriteConfig()
}

func (a *AppContext) RemoveSubscription(subType, subID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cfg.Subscriptions == nil {
		return fmt.Errorf("no subscriptions found")
	}

	found := false
	newList := make([]Subscription, 0)
	for _, s := range a.cfg.Subscriptions {
		if s.Type == subType && s.ID == subID {
			found = true
			continue
		}
		newList = append(newList, s)
	}

	if !found {
		return fmt.Errorf("subscription not found: %s %s", subType, subID)
	}

	a.cfg.Subscriptions = newList
	a.viper.Set("subscriptions", a.cfg.Subscriptions)
	return a.viper.WriteConfig()
}

func (a *AppContext) ReplaceAllSubscriptions(subscriptions []Subscription) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.Subscriptions = subscriptions
	a.viper.Set("subscriptions", subscriptions)
	return a.viper.WriteConfig()
}

func (a *AppContext) AddAlias(k, v string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cfg.Alias == nil {
		a.cfg.Alias = make(map[string]string)
	}
	a.cfg.Alias[k] = v
	a.viper.Set("alias", a.cfg.Alias)
	a.viper.WriteConfig()
}

func (a *AppContext) TrackRelays(relays []string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, r := range relays {
		a.knownRelays[r] = struct{}{}
	}
}

func (a *AppContext) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.store != nil {
		if closer, ok := a.store.(interface{ Close() }); ok {
			closer.Close()
		}
	}

	relays := make([]string, 0, len(a.knownRelays))
	for r := range a.knownRelays {
		relays = append(relays, r)
	}

	if len(relays) > 0 {
		existing := make(map[string]struct{})
		for _, r := range a.cfg.KnownRelays {
			existing[r] = struct{}{}
		}
		for _, r := range relays {
			existing[r] = struct{}{}
		}
		merged := make([]string, 0, len(existing))
		for r := range existing {
			merged = append(merged, r)
		}
		a.cfg.KnownRelays = merged
		a.viper.Set("known_relays", merged)
		a.viper.WriteConfig()
	}

	return nil
}