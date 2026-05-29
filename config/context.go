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
	pool *nostr.Pool
	cfg  Config
	mu   sync.RWMutex
}

func NewAppContext(pool *nostr.Pool, cfg Config, v *viper.Viper) *AppContext {
	if pool == nil {
		pool = nostr.NewPool(nostr.PoolOptions{})
	}

	return &AppContext{
		pool: pool,
		cfg:  cfg,
	}
}

func (a *AppContext) Pool() *nostr.Pool {
	return a.pool
}

func (a *AppContext) Config() Config {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg
}

func (a *AppContext) GetPrivateKey() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.PrivateKey
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

func (a *AppContext) ListRelays() []Relay {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.RelayList
}

func (a *AppContext) ListDMRelays() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.DMRelays
}

func (a *AppContext) ListSearchRelays() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.SearchRelays
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

func (a *AppContext) QueryTimeout() time.Duration {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.cfg.Query.Timeout <= 0 {
		return 5 * time.Second
	}
	return time.Duration(a.cfg.Query.Timeout) * time.Second
}

func (a *AppContext) QueryTimeoutms() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.cfg.Query.Timeout <= 0 {
		return 5000
	}
	return a.cfg.Query.Timeout * 1000
}

func (a *AppContext) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.pool != nil {
		a.pool.Close("app context closed")
	}
	a.pool = nil
	return nil
}
