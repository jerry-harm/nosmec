package main

import (
	"log"
	"testing"

	"github.com/spf13/viper"
)

func TestReadConfigFile(t *testing.T) {
	viper.SetConfigName("config.yml")    // 读取名为config的配置文件，没有设置特定的文件后缀名
	viper.SetConfigType("yaml")          // 当没有设置特定的文件后缀名时，必须要指定文件类型
	viper.AddConfigPath("./")            // 在当前文件夹下寻找
	viper.AddConfigPath("$HOME/.config") // 使用变量
	viper.AddConfigPath(".")             // 在工作目录下查找
	err := viper.ReadInConfig()          //读取配置
	if err != nil {
		log.Fatalln(err)
	}
}
