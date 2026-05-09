package config

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/eventstore/lmdb"
	"fiatjaf.com/nostr/khatru"
	"fiatjaf.com/nostr/nip19"
	"github.com/jerry-harm/nosmec/logger"
	"github.com/spf13/viper"
)

var (
	globalPool      *nostr.Pool
	globalLMDB      StoreInterface
	globalConfig    Config
	configDir       string
	onceInit        sync.Once
	initialized     bool
	proxyConfig     ProxyConfig
	localRelayURL   string
	globalViper     *viper.Viper
)

type ProxyConfig struct {
	Socks  string
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

	globalViper.SetDefault("local_relay.enabled", true)
	globalViper.SetDefault("local_relay.port", "8989")
	globalViper.SetDefault("local_relay.data_dir", defaultDataDir)
	globalViper.SetDefault("known_relays", []string{})
	globalViper.SetDefault("private_key", "")
	globalViper.SetDefault("proxy.socks", "")
	globalViper.SetDefault("proxy.i2p_socks", "")
	globalViper.SetDefault("relay_list", []Relay{})
	globalViper.SetDefault("dm_relays", []string{})
	globalViper.SetDefault("search_relays", []string{})
	globalViper.SetDefault("private_relays", []string{})
	globalViper.SetDefault("cache_filters", []map[string]interface{}{
		{"kinds": []int{0, 3, 10002, 10050}},
	})
	globalViper.SetDefault("subscriptions", []Subscription{})

	err := globalViper.ReadInConfig()
	if err != nil {
		logger.Warn("could not read config file, using defaults", "error", err.Error())
	}

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		os.MkdirAll(configDir, 0755)
	}

	configFile := filepath.Join(configDir, "nosmec.yaml")

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if err := globalViper.WriteConfigAs(configFile); err != nil {
			logger.Warn("could not write config file", "error", err.Error())
		}
	}

	var config Config
	err = globalViper.Unmarshal(&config)
	if err != nil {
		logger.Fatal("unable to decode config into struct", "error", err.Error())
	}

	if config.CacheFilters == nil || len(config.CacheFilters) == 0 {
		config.CacheFilters = []nostr.Filter{
			{Kinds: []nostr.Kind{0, 3, 10002, 10050}},
		}
		globalViper.Set("cache_filters", []map[string]interface{}{
			{"kinds": []int{0, 3, 10002, 10050}},
		})
		if err := globalViper.WriteConfigAs(configFile); err != nil {
			logger.Warn("could not write config file", "error", err.Error())
		}
	}

	config.ConfigDir = configDir

	if config.LocalRelay.Port == "" {
		config.LocalRelay.Port = "8989"
	}
	if config.LocalRelay.DataDir == "" {
		config.LocalRelay.DataDir = defaultDataDir
	}
	os.MkdirAll(config.LocalRelay.DataDir, 0755)

	if config.PrivateKey == "" {
		sk := nostr.Generate()
		config.PrivateKey = nip19.EncodeNsec(sk)
		globalViper.Set("private_key", config.PrivateKey)
		if err := globalViper.WriteConfig(); err != nil {
			logger.Warn("could not save generated private key", "error", err.Error())
		} else {
			logger.Info("generated new private key and saved to config")
		}
	}

	if config.CacheFilters == nil {
		_, s, err := nip19.Decode(config.PrivateKey)
		if err == nil {
			if sk, ok := s.(nostr.SecretKey); ok {
				pubKey := sk.Public()
				var allKinds []nostr.Kind
				config.CacheFilters = []nostr.Filter{
					{Kinds: []nostr.Kind{0, 3, 10002, 10050}},
					{Kinds: allKinds, Authors: []nostr.PubKey{pubKey}},
				}
			}
		}
	}

	return &config
}

func NewPool() *nostr.Pool {
	return nostr.NewPool(nostr.PoolOptions{
		RelayOptions: nostr.RelayOptions{},
	})
}

func GlobalPool() *nostr.Pool {
	if globalPool != nil {
		return globalPool
	}
	globalPool = NewPool()
	return globalPool
}

func SetPool(p *nostr.Pool) {
	globalPool = p
}

func NewLMDB(dataDir string) (StoreInterface, error) {
	lmdbStore := &lmdb.LMDBBackend{Path: filepath.Join(dataDir, "nosmec.db")}
	if err := lmdbStore.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize LMDB: %w", err)
	}
	return lmdbStore, nil
}

func GlobalLMDB() StoreInterface {
	if globalLMDB != nil {
		return globalLMDB
	}

	dataDir := globalConfig.LocalRelay.DataDir
	lmdbStore, err := NewLMDB(dataDir)
	if err != nil {
		logger.Fatal("failed to initialize LMDB", "error", err.Error())
	}
	globalLMDB = lmdbStore

	if globalConfig.LocalRelay.Enabled {
		if err := StartLocalRelay(lmdbStore); err != nil {
			logger.Fatal("failed to start local relay", "error", err.Error())
		}
	}
	return globalLMDB
}

func SetStore(s StoreInterface) {
	globalLMDB = s
}

func LocalRelayEnabled() bool {
	return globalConfig.LocalRelay.Enabled
}

func StartLocalRelay(store StoreInterface) error {
	port := globalConfig.LocalRelay.Port
	if port == "" {
		port = "8989"
	}

	relay := khatru.NewRelay()
	relay.UseEventstore(store, 500)
	relay.Info.Name = "nosmec-local"
	relay.Info.Description = "Local relay for nosmec"

	go func() {
		addr := fmt.Sprintf(":%s", port)
		logger.Info("starting local relay", "addr", addr)
		if err := http.ListenAndServe(addr, relay); err != nil {
			logger.Error("local relay error", "error", err.Error())
		}
	}()

	localRelayURL = fmt.Sprintf("ws://localhost:%s", port)
	return nil
}

func GetLocalRelayURL() string {
	if !LocalRelayEnabled() {
		return ""
	}
	return localRelayURL
}

func ConfigDir() string {
	return configDir
}

func DataDir() string {
	return globalConfig.LocalRelay.DataDir
}
