package config

// Config 应用配置
type Config struct {
	LocalServer struct {
		Enabled bool   `mapstructure:"enabled"`
		Host    string `mapstructure:"host"`
		Port    int    `mapstructure:"port"`
		NIP11   struct {
			Name        string `mapstructure:"name"`
			Description string `mapstructure:"description"`
			Contact     string `mapstructure:"contact"`
		} `mapstructure:"nip11"`
		I2P struct {
			Enabled    bool   `mapstructure:"enabled"`
			SamAddress string `mapstructure:"sam_address"`
			SamPort    int    `mapstructure:"sam_port"`
		} `mapstructure:"i2p"`
	} `mapstructure:"local_server"`

	DataDir    string           `mapstructure:"data_dir"`
	RelayList  map[string]Relay `mapstructure:"relay_list"`
	PrivateKey string           `mapstructure:"private_key"`

	KnownRelays []string `mapstructure:"known_relays"`

	Proxy struct {
		I2PSocks   string `mapstructure:"i2p_socks"`
		OnionSocks string `mapstructure:"onion_socks"`
		Socks      string `mapstructure:"socks"`
	} `mapstructure:"proxy"`
}

type Relay struct {
	Read   *bool `mapstructure:"read,omitempty"`
	Write  *bool `mapstructure:"write,omitempty"`
	DM     *bool `mapstructure:"dm,omitempty"`
	Search *bool `mapstructure:"search,omitempty"`
}
