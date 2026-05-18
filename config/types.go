package config

type Relay struct {
	URL   string `mapstructure:"url"`
	Read  *bool  `mapstructure:"read,omitempty"`
	Write *bool  `mapstructure:"write,omitempty"`
}

type Config struct {
	ConfigDir    string   `mapstructure:"config_dir"`
	DataDir      string   `mapstructure:"data_dir"`
	RelayList   []Relay  `mapstructure:"relay_list"`
	DMRelays    []string `mapstructure:"dm_relays"`
	SearchRelays []string `mapstructure:"search_relays"`
	PrivateKey  string   `mapstructure:"private_key"`

	KnownRelays []string `mapstructure:"known_relays"`

	Proxy struct {
		Socks    string `mapstructure:"socks"`
		I2PSocks string `mapstructure:"i2p_socks"`
	} `mapstructure:"proxy"`

	Alias map[string]string `mapstructure:"alias"`

	Subscriptions []Subscription `mapstructure:"subscriptions"`

	Profile ProfileConfig `mapstructure:"profile"`

	Query struct {
		Timeout int `mapstructure:"timeout"`
	} `mapstructure:"query"`
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
