package config

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"github.com/spf13/viper"
)

var (
	globalConfig Config
	configDir    string
	onceInit     sync.Once
	initialized  bool
	proxyConfig  ProxyConfig
	globalViper  *viper.Viper
)

type ProxyConfig struct {
	Socks    string
	I2PSocks string
}

func SetProxyConfig(pc ProxyConfig) {
	proxyConfig = pc
}

func GetProxyConfig() ProxyConfig {
	return proxyConfig
}

func GetViper() *viper.Viper {
	return globalViper
}

func SetViper(v *viper.Viper) {
	globalViper = v
}

func InitConfig() Config {
	onceInit.Do(func() {
		globalConfig = *loadConfig()
		initialized = true
	})
	return globalConfig
}

func IsInitialized() bool {
	return initialized
}

func loadConfig() *Config {
	if globalViper == nil {
		globalViper = viper.New()
	}

	cachedir, _ := os.UserCacheDir()
	defaultDataDir := filepath.Join(cachedir, "nosmec")

	configDir = filepath.Join(os.Getenv("HOME"), ".config", "nosmec")
	os.MkdirAll(configDir, 0755)

	globalViper.SetConfigName("nosmec")
	globalViper.SetConfigType("yaml")
	globalViper.AddConfigPath(configDir)
	globalViper.AddConfigPath("$HOME/.config")
	globalViper.AddConfigPath("./")
	globalViper.AddConfigPath(".")

	globalViper.SetEnvPrefix("NOSMEC")
	globalViper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	globalViper.AutomaticEnv()

	globalViper.SetDefault("data_dir", defaultDataDir)
	globalViper.SetDefault("private_key", "")
	globalViper.SetDefault("proxy.socks", "")
	globalViper.SetDefault("proxy.i2p_socks", "")
	globalViper.SetDefault("relay_list", []Relay{})
	globalViper.SetDefault("dm_relays", []string{})
	globalViper.SetDefault("search_relays", []string{})
	globalViper.SetDefault("private_relays", []string{})
	globalViper.SetDefault("subscriptions", []Subscription{})

	err := globalViper.ReadInConfig()
	if err != nil {
		// using defaults
	}

	configFile := filepath.Join(configDir, "nosmec.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if err := globalViper.WriteConfigAs(configFile); err != nil {
			// cannot write config
		}
	}

	var config Config
	err = globalViper.Unmarshal(&config)
	if err != nil {
		// unable to decode config
	}

	config.ConfigDir = configDir

	if config.DataDir == "" {
		config.DataDir = defaultDataDir
	}
	os.MkdirAll(config.DataDir, 0755)

	if config.PrivateKey == "" {
		sk := nostr.Generate()
		config.PrivateKey = nip19.EncodeNsec(sk)
		globalViper.Set("private_key", config.PrivateKey)
		if err := globalViper.WriteConfig(); err != nil {
			// cannot save key
		}
	}

	return &config
}

func DataDir() string {
	return globalConfig.DataDir
}

func GetEventRelay(eventID string) string {
	return ""
}