package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jerry-harm/nosmec/client"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/i2p"
)

func main() {
	// 初始化配置和I2P
	os.MkdirAll(config.Global.BasePath, 0777)
	// go func() {
	// 	i2p.Init()
	// }()

	// 设置HTTP客户端使用I2PDial，设置30秒超时
	http.DefaultClient = &http.Client{
		Transport: &http.Transport{
			Proxy: i2p.I2PProxy,
		},
		Timeout: 30 * time.Second,
	}
	// 初始化客户端
	client.Init()
	defer client.Close()

	// // 启动服务器
	// relayServer, err := server.NewRelay()
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	serverStopped := make(chan bool, 1)

	// 创建信号通道
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// // 启动HTTP服务器
	// go func() {
	// 	log.Println("Starting Nostr relay server...")
	// 	if err := http.Serve(i2p.Listener, relayServer); err != nil {
	// 		log.Printf("Server error: %v\n", err)
	// 	}
	// 	serverStopped <- true
	// }()

	// 启动UI（默认启动）

	// 等待退出信号
	select {
	case <-sigChan:
		log.Println("Shutting down nosmec...")
	case <-serverStopped:
		log.Println("Server stopped unexpectedly")
	}

	// 优雅关闭
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer i2p.Sam.Close()
	// defer i2p.ListenerSession.Close()
	// defer i2p.Listener.Close()
	defer cancel()

	log.Println("nosmec stopped gracefully")
}
