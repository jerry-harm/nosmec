package config

// Config 应用配置
type Config struct {
	Server struct {
		Host  string `mapstructure:"host"`
		Port  int    `mapstructure:"port"`
		NIP11 struct {
			Name        string `mapstructure:"name"`
			Description string `mapstructure:"description"`
			Contact     string `mapstructure:"contact"`
		} `mapstructure:"nip11"`
	} `mapstructure:"server"`
	BasePath string `mapstructure:"base_path"`
	I2P      struct {
		Enabled    bool   `mapstructure:"enabled"`
		SamAddress string `mapstructure:"sam_address"`
		SamPort    int    `mapstructure:"sam_port"`
	} `mapstructure:"i2p"`
	Client struct {
		Relays     []string `mapstructure:"relays"`
		PrivateKey string   `mapstructure:"private_key"`
	} `mapstructure:"client"`
}
