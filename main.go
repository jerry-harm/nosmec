package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/jerry-harm/nosmec/pkg/config"
	"github.com/jerry-harm/nosmec/pkg/i2p"
	"github.com/jerry-harm/nosmec/server"
)

func main() {
	// 加载配置
	config := config.LoadConfig()

	// 处理存储路径
	basePath := expandPath(config.Storage.BasePath)
	databasePath := filepath.Join(basePath, config.Storage.Database.Path)

	log.Printf("Starting nosmec (client + server)...")
	log.Printf("Server: %s:%d", config.Server.Host, config.Server.Port)
	log.Printf("Client relays: %v", config.Client.DefaultRelays)
	log.Printf("Storage base path: %s", basePath)

	// 创建 relay 服务器
	relayServer, err := server.NewRelayServer(
		config.Server.Host,
		config.Server.Port,
		databasePath,
		config.Server.NIP11.Name,
		config.Server.NIP11.Description,
		config.Server.NIP11.PubKey,
		config.Server.NIP11.Contact,
		config.Server.NIP11.Software,
		config.Server.NIP11.Version,
	)
	if err != nil {
		log.Fatalf("Failed to create relay server: %v", err)
	}
	defer relayServer.Close()

	// 启动常规 HTTP 服务器
	go func() {
		if err := relayServer.Start(config.Server.Host, config.Server.Port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start relay server: %v", err)
		}
	}()

	// 如果启用 I2P，启动 I2P 服务器
	var i2pServer *i2p.I2PServer
	if config.I2P.Enabled {
		log.Printf("Starting I2P server...")

		// 构建 SAM 地址
		samAddr := fmt.Sprintf("%s:%d", config.I2P.Address, config.I2P.Port)
		log.Printf("Using SAM address: %s", samAddr)

		// 创建 I2P 服务器
		i2pServer, err = i2p.NewI2PServer(relayServer.GetHandler(), "nosmec-relay", basePath, samAddr)
		if err != nil {
			log.Fatalf("Failed to create I2P server: %v", err)
		}
		defer i2pServer.Stop(context.Background())

		// 设置智能 I2P 分流 Transport
		i2pTransport := i2p.NewI2PTransport(i2pServer)
		http.DefaultClient.Transport = i2pTransport
		log.Printf("I2P proxy transport enabled - .i2p addresses will use I2P network, others use direct connection")

		// 启动 I2P 服务器
		go func() {
			if err := i2pServer.Start(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start I2P server: %v", err)
			}
		}()
	}

	// TODO: 启动客户端UI
	log.Println("Client UI will be started here...")

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down nosmec...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 关闭 I2P 服务器
	if i2pServer != nil {
		if err := i2pServer.Stop(ctx); err != nil {
			log.Printf("Error stopping I2P server: %v", err)
		}
	}

	log.Println("nosmec stopped gracefully")
}

// expandPath 扩展路径中的环境变量
func expandPath(path string) string {
	return os.ExpandEnv(path)
}
