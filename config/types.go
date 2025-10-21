package config

// Config 应用配置
type Config struct {
	Server struct {
		Host  string `mapstructure:"host"`
		Port  int    `mapstructure:"port"`
		NIP11 struct {
			Name        string `mapstructure:"name"`
			Description string `mapstructure:"description"`
			PubKey      string `mapstructure:"pubkey"`
			Contact     string `mapstructure:"contact"`
			Software    string `mapstructure:"software"`
			Version     string `mapstructure:"version"`
		} `mapstructure:"nip11"`
	} `mapstructure:"server"`
	BasePath string `mapstructure:"base_path"`
	I2P      struct {
		Enabled    bool   `mapstructure:"enabled"`
		SamAddress string `mapstructure:"sam_address"`
		SamPort    int    `mapstructure:"sam_port"`
	} `mapstructure:"i2p"`
	Client struct {
		Relays     []Relay `mapstructure:"relays"`
		PrivateKey string  `mapstructure:"private_key"`
	} `mapstructure:"client"`
}
type Relay struct {
	Url   string `mapstructure:"url"`
	Read  bool   `mapstructure:"read"`
	Write bool   `mapstructure:"write"`
	Inbox bool   `mapstructure:"inbox"`
}
