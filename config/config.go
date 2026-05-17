package config

import (
	"context"
	"fmt"
	"iter"
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
	"fiatjaf.com/nostr/eventstore/bleve"
	"fiatjaf.com/nostr/eventstore/boltdb"
	"fiatjaf.com/nostr/khatru"
	"fiatjaf.com/nostr/nip19"
	sdk_hints "fiatjaf.com/nostr/sdk/hints"
	"fiatjaf.com/nostr/sdk/hints/bbolth"
	"github.com/jerry-harm/nosmec/logger"
	"github.com/spf13/viper"
)

var (
	globalPool      *nostr.Pool
	globalHints     sdk_hints.HintsDB
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

func GlobalHints() sdk_hints.HintsDB {
	if globalHints != nil {
		return globalHints
	}
	hintsPath := filepath.Join(globalConfig.LocalRelay.DataDir, "hints.db")
	bh, err := bbolth.NewBoltHints(hintsPath)
	if err != nil {
		logger.Error("failed to open hints db", "error", err.Error(), "path", hintsPath)
		return nil
	}
	globalHints = bh
	return globalHints
}

func NewPool(h sdk_hints.HintsDB) *nostr.Pool {
	opts := nostr.PoolOptions{
		RelayOptions: nostr.RelayOptions{
			NoticeHandler: func(relay *nostr.Relay, notice string) {
				logger.Debug("NOTICE from %s: '%s'", relay.URL, notice)
			},
		},
	}
	if h != nil {
		opts.EventMiddleware = func(ie nostr.RelayEvent) {
			ev := ie.Event

			// HintsDB: learn relay→pubkey associations
			if ev.PubKey != [32]byte{} {
				h.Save(ev.PubKey, ie.Relay.URL, sdk_hints.MostRecentEventFetched, nostr.Now())
			}
			for tag := range ev.Tags.FindAll("p") {
				if len(tag) >= 3 && tag[1] != "" && tag[2] != "" {
					if pk, err := nostr.PubKeyFromHex(tag[1]); err == nil {
						h.Save(pk, tag[2], sdk_hints.LastInHint, nostr.Now())
					}
				}
			}
			if ev.Kind == nostr.KindRelayListMetadata {
				for tag := range ev.Tags.FindAll("r") {
					if len(tag) >= 2 {
						h.Save(ev.PubKey, tag[1], sdk_hints.LastInRelayList, nostr.Now())
					}
				}
			}

			// Local relay cache: persist events for offline/fallback
			cacheEvent(ie)
		}
	}
	return nostr.NewPool(opts)
}

// cacheEvent publishes an event to the local relay for offline/fallback caching.
// Called from the pool EventMiddleware so every incoming event gets cached.
func cacheEvent(ie nostr.RelayEvent) {
	if localRelayURL == "" {
		return
	}
	ev := ie.Event
	if ev.ID == [32]byte{} {
		return
	}
	// Only cache events matching configured CacheFilters
	if !shouldCacheEvent(ev.Kind) {
		return
	}
	go func() {
		if p := GlobalPool(); p != nil {
			p.PublishMany(context.Background(), []string{localRelayURL}, ev)
		}
	}()
}

func shouldCacheEvent(kind nostr.Kind) bool {
	if globalConfig.CacheFilters == nil {
		return false
	}
	for _, f := range globalConfig.CacheFilters {
		for _, k := range f.Kinds {
			if nostr.Kind(k) == kind {
				return true
			}
		}
	}
	return false
}

func GlobalPool() *nostr.Pool {
	if globalPool != nil {
		return globalPool
	}
	globalPool = NewPool(GlobalHints())
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

	// Add NIP-50 support via Bleve search index
	searchIndexPath := filepath.Join(globalConfig.LocalRelay.DataDir, "search_index")
	if err := os.MkdirAll(searchIndexPath, 0755); err != nil {
		logger.Warn("failed to create search index directory", "path", searchIndexPath, "error", err.Error())
	}

	bleveStore := &bleve.BleveBackend{
		Path:         searchIndexPath,
		RawEventStore: store,
	}
	if err := bleveStore.Init(); err != nil {
		logger.Warn("failed to initialize bleve search index, search will not be available", "error", err.Error())
	} else {
		// Configure query to use Bleve for search queries
		relay.QueryStored = func(ctx context.Context, filter nostr.Filter) iter.Seq[nostr.Event] {
			if filter.Search != "" {
				// Use Bleve for search queries
				return bleveStore.QueryEvents(filter, 1000)
			}
			// Use BoltDB for regular queries
			return store.QueryEvents(filter, 1000)
		}

		// Store events in both stores
		relay.StoreEvent = func(ctx context.Context, event nostr.Event) error {
			if err := store.SaveEvent(event); err != nil {
				return err
			}
			return bleveStore.SaveEvent(event)
		}

		// Delete events from both stores
		relay.DeleteEvent = func(ctx context.Context, id nostr.ID) error {
			if err := store.DeleteEvent(id); err != nil {
				return err
			}
			return bleveStore.DeleteEvent(id)
		}

		// Add NIP-50 to supported NIPs
		relay.Info.SupportedNIPs = append(relay.Info.SupportedNIPs, 50)
		logger.Info("NIP-50 search enabled on local relay")
	}

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
