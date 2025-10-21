package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

// LoadConfig 加载应用配置
func LoadConfig() *Config {
	homedir, _ := os.UserHomeDir()
	viper.SetConfigName("nosmec")           // 读取名为config的配置文件
	viper.SetConfigType("yaml")             // 指定文件类型为yaml
	viper.AddConfigPath("$XDG_CONFIG_HOME") // 使用变量
	viper.AddConfigPath("./")               // 在当前文件夹下寻找
	viper.AddConfigPath(".")                // 在工作目录下查找

	// 设置默认值
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("base_path", homedir+"/.local/share/nosmec")
	viper.SetDefault("i2p.enabled", true)
	viper.SetDefault("i2p.sam_address", "127.0.0.1")
	viper.SetDefault("i2p.sam_port", 7656)
	viper.SetDefault("client.relays", []Relay{{
		Url:   "ws://nostr.jerryhome.i2p",
		Read:  true,
		Write: true,
		Inbox: true,
	},
		{
			Url:   "wss://bostr.shop",
			Read:  true,
			Write: true,
			Inbox: true,
		}})

	err := viper.ReadInConfig() // 读取配置
	if err != nil {
		log.Printf("Warning: Could not read config file, using defaults: %v", err)
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Unable to decode config into struct: %v", err)
	}

	viper.SafeWriteConfig()

	return &config
}

var Global *Config = LoadConfig()
