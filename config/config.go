package config

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/eventstore/boltdb"
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
	lockFile       *os.File
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

	if config.CacheFilters == nil || len(config.CacheFilters) == 0 {
		defaultFilters := []CacheFilter{
			{Kinds: []int{0, 3, 10002, 10050}},
		}

		_, s, err := nip19.Decode(config.PrivateKey)
		if err == nil {
			if sk, ok := s.(nostr.SecretKey); ok {
				pubKey := sk.Public()
				defaultFilters = append(defaultFilters, CacheFilter{
					Authors: []string{pubKey.Hex()},
				})
			}
		}

		config.CacheFilters = defaultFilters
		globalViper.Set("cache_filters", config.CacheFilters)
		if err := globalViper.WriteConfigAs(configFile); err != nil {
			logger.Warn("could not write config file", "error", err.Error())
		}
	}

	return &config
}

func NewPool() *nostr.Pool {
	return nostr.NewPool(nostr.PoolOptions{
		RelayOptions: nostr.RelayOptions{
			NoticeHandler: func(relay *nostr.Relay, notice string) {
				logger.Debug("NOTICE from %s: '%s'", relay.URL, notice)
			},
		},
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
	dbPath := filepath.Join(dataDir, "nosmec.db")
	lockPath := filepath.Join(dataDir, "nosmec.lock")

	// Try to clean up stale lock file from crashed process
	if data, err := os.ReadFile(lockPath); err == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
			if !isProcessRunning(pid) {
				logger.Warn("removing stale lock file from dead process", "pid", pid)
				os.Remove(lockPath)
			}
		}
	}

	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		if os.IsExist(err) {
			return nil, fmt.Errorf("secondary")
		}
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if err := os.WriteFile(lockPath, []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
		lockFile.Close()
		os.Remove(lockPath)
		return nil, fmt.Errorf("failed to write pid to lock file: %w", err)
	}

	boltStore := &boltdb.BoltBackend{Path: dbPath}
	if err := boltStore.Init(); err != nil {
		lockFile.Close()
		os.Remove(lockPath)
		return nil, fmt.Errorf("failed to initialize BoltDB: %w", err)
	}

	return boltStore, nil
}

func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func GlobalLMDB() StoreInterface {
	if globalLMDB != nil {
		return globalLMDB
	}

	dataDir := globalConfig.LocalRelay.DataDir
	lmdbStore, err := NewLMDB(dataDir)
	if err != nil {
		if err.Error() == "secondary" {
			logger.Info("another instance is running, connecting to existing local relay")
			if globalConfig.LocalRelay.Enabled {
				port := globalConfig.LocalRelay.Port
				if port == "" {
					port = "8989"
				}
				localRelayURL = fmt.Sprintf("ws://localhost:%s", port)
			}
			return nil
		}
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

	if IsPortInUse(port) {
		logger.Info("local relay already running on port", "port", port)
		localRelayURL = fmt.Sprintf("ws://localhost:%s", port)
		return nil
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

func IsPortInUse(port string) bool {
	addr := fmt.Sprintf("localhost:%s", port)
	conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
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
