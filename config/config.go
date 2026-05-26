package config

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/eventstore"
	eventstorebleve "fiatjaf.com/nostr/eventstore/bleve"
	eventstorelmdb "fiatjaf.com/nostr/eventstore/lmdb"
	"fiatjaf.com/nostr/nip19"
	"github.com/jerry-harm/nosmec/logger"
	"github.com/jerry-harm/nosmec/nostr_sdk"
	"github.com/jerry-harm/nosmec/nostr_sdk/hints"
	"github.com/jerry-harm/nosmec/nostr_sdk/hints/lmdbh"
	"github.com/jerry-harm/nosmec/nostr_sdk/kvstore"
	kvstorelmdb "github.com/jerry-harm/nosmec/nostr_sdk/kvstore/lmdb"
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

	globalViper.SetDefault("theme.primary", "#25A065")
	globalViper.SetDefault("theme.primary_dark", "#00875A")
	globalViper.SetDefault("theme.text_bright", "#00FF00")
	globalViper.SetDefault("theme.text_bright_alt", "#00875A")
	globalViper.SetDefault("theme.text", "#FFFFFF")
	globalViper.SetDefault("theme.text_dark", "#333333")
	globalViper.SetDefault("theme.text_muted", "#AAAAAA")
	globalViper.SetDefault("theme.text_muted_dark", "#6B6B6B")
	globalViper.SetDefault("theme.text_muted_alt", "#888888")
	globalViper.SetDefault("theme.selection", "#FFFF00")
	globalViper.SetDefault("theme.status_text", "#04B575")
	globalViper.SetDefault("theme.author_text", "#00AA00")
	globalViper.SetDefault("theme.author_text_alt", "#008800")
	globalViper.SetDefault("theme.error", "#FF4444")
	globalViper.SetDefault("theme.error_alt", "#FF6B6B")
	globalViper.SetDefault("theme.tag_color", "#00AAFF")
	globalViper.SetDefault("theme.community_addr", "#FFD700")
	globalViper.SetDefault("theme.overlay_bg", "#333333")
	globalViper.SetDefault("theme.title_text", "#FFFDF5")
	globalViper.SetDefault("theme.title_bg", "#25A065")
	globalViper.SetDefault("theme.border", "#25A065")
	globalViper.SetDefault("theme.border_dark", "#00875A")
	globalViper.SetDefault("theme.viewport_border", "#25A065")
	globalViper.SetDefault("theme.viewport_border_dark", "#00875A")
	globalViper.SetDefault("theme.input_placeholder", "#666666")
	globalViper.SetDefault("theme.spinner", "#00FF00")

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
			logger.Warn("could not save generated private key", "error", err.Error())
		} else {
			logger.Info("generated new private key and saved to config")
		}
	}

	return &config
}

func openHints(dataDir string) hints.HintsDB {
	hintsPath := filepath.Join(dataDir, "hints")
	bh, err := lmdbh.NewLMDBHints(hintsPath)
	if err != nil {
		logger.Error("failed to open hints db", "error", err.Error(), "path", hintsPath)
		return nil
	}
	return bh
}

func openKVStore(dataDir string) kvstore.KVStore {
	kvPath := filepath.Join(dataDir, "kvstore")
	store, err := kvstorelmdb.NewStore(kvPath)
	if err != nil {
		logger.Error("failed to open kvstore", "error", err.Error(), "path", kvPath)
		return nil
	}
	return store
}

func openStore(dataDir string) eventstore.Store {
	eventsPath := filepath.Join(dataDir, "events")
	lmdbStore := &eventstorelmdb.LMDBBackend{Path: eventsPath}
	if err := lmdbStore.Init(); err != nil {
		logger.Warn("failed to create LMDB event store, local cache disabled", "error", err.Error(), "path", eventsPath)
		return nil
	}

	searchIndexPath := filepath.Join(dataDir, "search_index")
	bleveStore := &eventstorebleve.BleveBackend{Path: searchIndexPath, RawEventStore: lmdbStore}
	if err := bleveStore.Init(); err != nil {
		logger.Warn("failed to create Bleve search index, search disabled", "error", err.Error(), "path", searchIndexPath)
		return lmdbStore
	}

	return bleveStore
}

func newPool(sys *nostr_sdk.System) *nostr.Pool {
	opts := nostr.PoolOptions{
		RelayOptions: nostr.RelayOptions{
			NoticeHandler: func(relay *nostr.Relay, notice string) {
				logger.Debug("NOTICE from %s: '%s'", relay.URL, notice)
			},
		},
	}
	if sys != nil {
		opts.EventMiddleware = sys.TrackEventHintsAndRelays
	}
	return nostr.NewPool(opts)
}

func DataDir() string {
	return globalConfig.DataDir
}

func getDataDir() string {
	if globalConfig.DataDir != "" {
		return globalConfig.DataDir
	}
	return InitConfig().DataDir
}

func resetGlobalRuntimeState() {
}

func GetEventRelay(eventID string) string {
	return ""
}
