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
	Storage struct {
		BasePath string `mapstructure:"base_path"`
		Database struct {
			Path string `mapstructure:"path"`
		} `mapstructure:"database"`
	} `mapstructure:"storage"`
	I2P struct {
		Enabled    bool   `mapstructure:"enabled"`
		SamAddress string `mapstructure:"sam_address"`
		SamPort    int    `mapstructure:"sam_port"`
	} `mapstructure:"i2p"`
	Client struct {
		DefaultRelays []string `mapstructure:"default_relays"`
		Theme         string   `mapstructure:"theme"`
		PrivateKey    string   `mapstructure:"private_key"`
	} `mapstructure:"client"`
}
