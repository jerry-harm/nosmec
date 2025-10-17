package main

import (
	"log"

	"github.com/spf13/viper"
)

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
		Enabled bool   `mapstructure:"enabled"`
		Address string `mapstructure:"address"`
		Port    int    `mapstructure:"port"`
	} `mapstructure:"i2p"`
}

func LoadConfig() *Config {
	viper.SetConfigName("config")           // 读取名为config的配置文件
	viper.SetConfigType("yaml")             // 指定文件类型为yaml
	viper.AddConfigPath("./")               // 在当前文件夹下寻找
	viper.AddConfigPath("$XDG_CONFIG_HOME") // 使用变量
	viper.AddConfigPath(".")                // 在工作目录下查找

	// 设置默认值
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("storage.base_path", "$HOME/.local/share/nosmec")
	viper.SetDefault("storage.database.path", "nostr_relay.db")
	viper.SetDefault("i2p.enabled", false)
	viper.SetDefault("i2p.address", "127.0.0.1")
	viper.SetDefault("i2p.port", 7656)

	err := viper.ReadInConfig() // 读取配置
	if err != nil {
		log.Printf("Warning: Could not read config file, using defaults: %v", err)
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Unable to decode config into struct: %v", err)
	}

	return &config
}
