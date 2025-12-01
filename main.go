package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jerry-harm/nosmec/client/ui"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/i2p"
	"github.com/jerry-harm/nosmec/server"
)

func main() {
	// 初始化配置和I2P
	os.MkdirAll(config.Global.BasePath, 0777)
	i2p.Init()
	defer i2p.Sam.Close()
	defer i2p.ListenerSession.Close()
	defer i2p.DialSession.Close()
	defer i2p.Listener.Close()

	// 启动服务器
	relayServer, err := server.NewRelay()
	if err != nil {
		log.Fatalln(err)
	}

	serverStopped := make(chan bool, 1)

	// 创建信号通道
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 启动HTTP服务器
	go func() {
		log.Println("Starting Nostr relay server...")
		if err := http.Serve(i2p.Listener, relayServer); err != nil {
			log.Printf("Server error: %v\n", err)
		}
		serverStopped <- true
	}()

	// 启动UI（默认启动）
	go func() {
		log.Println("Starting UI...")
		ui.StartMenu()
		// UI退出时发送信号
		sigChan <- os.Interrupt
	}()

	// 等待退出信号
	select {
	case <-sigChan:
		log.Println("Shutting down nosmec...")
	case <-serverStopped:
		log.Println("Server stopped unexpectedly")
	}

	// 优雅关闭
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("nosmec stopped gracefully")
}
