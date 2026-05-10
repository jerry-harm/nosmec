package config

import (
	"encoding/json"

	"fiatjaf.com/nostr"
)

type CacheFilter struct {
	Kinds   []int  `json:"kinds,omitempty"`
	Authors []string `json:"authors,omitempty"`
}

func (f CacheFilter) ToNostr() nostr.Filter {
	filter := nostr.Filter{}
	if len(f.Kinds) > 0 {
		kinds := make([]nostr.Kind, len(f.Kinds))
		for i, k := range f.Kinds {
			kinds[i] = nostr.Kind(k)
		}
		filter.Kinds = kinds
	}
	if len(f.Authors) > 0 {
		authors := make([]nostr.PubKey, 0, len(f.Authors))
		for _, a := range f.Authors {
			if pk, err := nostr.PubKeyFromHex(a); err == nil {
				authors = append(authors, pk)
			}
		}
		filter.Authors = authors
	}
	return filter
}

func (f *CacheFilter) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if kinds, ok := raw["kinds"].([]interface{}); ok {
		f.Kinds = make([]int, 0, len(kinds))
		for _, k := range kinds {
			if n, ok := k.(float64); ok {
				f.Kinds = append(f.Kinds, int(n))
			}
		}
	}

	if authors, ok := raw["authors"].([]interface{}); ok {
		f.Authors = make([]string, 0, len(authors))
		for _, a := range authors {
			if s, ok := a.(string); ok {
				f.Authors = append(f.Authors, s)
			}
		}
	}

	return nil
}

func (f CacheFilter) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})
	if len(f.Kinds) > 0 {
		m["kinds"] = f.Kinds
	}
	if len(f.Authors) > 0 {
		m["authors"] = f.Authors
	}
	return json.Marshal(m)
}

type Relay struct {
	URL   string `mapstructure:"url"`
	Read  *bool  `mapstructure:"read,omitempty"`
	Write *bool  `mapstructure:"write,omitempty"`
}

type LocalRelayConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    string `mapstructure:"port"`
	DataDir string `mapstructure:"data_dir"`
}

type Config struct {
	ConfigDir    string   `mapstructure:"config_dir"`
	RelayList   []Relay  `mapstructure:"relay_list"`
	DMRelays    []string `mapstructure:"dm_relays"`
	SearchRelays []string `mapstructure:"search_relays"`
	PrivateKey  string   `mapstructure:"private_key"`

	KnownRelays []string `mapstructure:"known_relays"`

	LocalRelay LocalRelayConfig `mapstructure:"local_relay"`

	Proxy struct {
		Socks    string `mapstructure:"socks"`
		I2PSocks string `mapstructure:"i2p_socks"`
	} `mapstructure:"proxy"`

	Alias map[string]string `mapstructure:"alias"`

	Subscriptions []Subscription `mapstructure:"subscriptions"`

	Profile ProfileConfig `mapstructure:"profile"`

	CacheFilters []CacheFilter `mapstructure:"cache_filters"`
}

type ProfileConfig struct {
	Name        string `mapstructure:"name"`
	About       string `mapstructure:"about"`
	Picture     string `mapstructure:"picture"`
	DisplayName string `mapstructure:"display_name"`
	Website     string `mapstructure:"website"`
	Banner      string `mapstructure:"banner"`
	Bot         *bool  `mapstructure:"bot"`
	Birthday    string `mapstructure:"birthday"`
	NIP05       string `mapstructure:"nip05"`
	Lud06       string `mapstructure:"lud06"`
	Lud16       string `mapstructure:"lud16"`
}

type Subscription struct {
	Type    string `mapstructure:"type"`    // "community" | "user" | "hashtag"
	ID      string `mapstructure:"id"`      // community addr, pubkey, or hashtag
	Relay   string `mapstructure:"relay"`   // recommended relay URL (optional)
	Petname string `mapstructure:"petname"` // petname/alias (only for user)
}
