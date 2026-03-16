package config

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/eventstore/lmdb"
	"github.com/spf13/viper"
)

var (
	globalPool   *nostr.Pool
	oncePool     sync.Once
	globalLMDB   *lmdb.LMDBBackend
	onceLMDB     sync.Once
	globalConfig Config = *LoadConfig()
)

func LoadConfig() *Config {
	cachedir, _ := os.UserCacheDir()

	dataDir := filepath.Join(cachedir, "nosmec")

	os.MkdirAll(dataDir, 0755)

	viper.SetConfigName("nosmec")           // 读取名为config的配置文件
	viper.SetConfigType("yaml")             // 指定文件类型为yaml
	viper.AddConfigPath("$XDG_CONFIG_HOME") // 使用变量
	viper.AddConfigPath("./")               // 在当前文件夹下寻找
	viper.AddConfigPath(".")                // 在工作目录下查找

	// 设置默认值
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("data_dir", dataDir)
	viper.SetDefault("i2p.enabled", true)
	viper.SetDefault("i2p.sam_address", "127.0.0.1")
	viper.SetDefault("i2p.sam_port", 7656)
	viper.SetDefault("known_relays", []string{
		"ws://nostr.jerryhome.i2p",
	})
	viper.SetDefault("private_key", "")
	viper.SetDefault("proxy.i2p_socks", "")
	viper.SetDefault("proxy.onion_socks", "")
	viper.SetDefault("proxy.socks", "")

	err := viper.ReadInConfig() // 读取配置
	if err != nil {
		log.Printf("Warning: Could not read config file, using defaults: %v", err)
		viper.SafeWriteConfig()
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Unable to decode config into struct: %v", err)
	}

	return &config
}

func GlobalPool() *nostr.Pool {
	oncePool.Do(func() {
		globalPool = nostr.NewPool(nostr.PoolOptions{
			RelayOptions: nostr.RelayOptions{},
		})
	})

	return globalPool
}

func GlobalLMDB() *lmdb.LMDBBackend {
	onceLMDB.Do(func() {
		globalLMDB = &lmdb.LMDBBackend{Path: filepath.Join(viper.GetString("data_dir"), "nosmec.db")}
		if err := globalLMDB.Init(); err != nil {
			log.Fatalf("Failed to initialize LMDB: %v", err)
		}
	})
	return globalLMDB
}

func GlobalConfig() Config {
	return globalConfig
}
