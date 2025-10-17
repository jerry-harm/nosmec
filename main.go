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

	"github.com/jerry-harm/nosmec/server"
)

func main() {
	// 加载配置
	config := LoadConfig()

	// 处理存储路径
	basePath := expandPath(config.Storage.BasePath)
	databasePath := filepath.Join(basePath, config.Storage.Database.Path)

	// 使用默认的目录名称
	i2pPath := filepath.Join(basePath, "i2pkeys")
	torPath := filepath.Join(basePath, "onionkeys")
	tlsPath := filepath.Join(basePath, "tlskeys")

	// 创建存储目录
	if err := createStorageDirs(basePath, i2pPath, torPath, tlsPath); err != nil {
		log.Fatalf("Failed to create storage directories: %v", err)
	}

	log.Printf("Starting nostr relay server...")
	log.Printf("Host: %s, Port: %d", config.Server.Host, config.Server.Port)
	log.Printf("Storage base path: %s", basePath)
	log.Printf("Database path: %s", databasePath)

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
	var i2pServer *server.I2PServer
	if config.I2P.Enabled {
		log.Printf("Starting I2P server...")

		// 构建 SAM 地址
		samAddr := fmt.Sprintf("%s:%d", config.I2P.Address, config.I2P.Port)
		log.Printf("Using SAM address: %s", samAddr)

		// 创建 I2P 服务器
		i2pServer, err = server.NewI2PServer(relayServer.GetHandler(), "nosmec-relay", i2pPath, samAddr)
		if err != nil {
			log.Fatalf("Failed to create I2P server: %v", err)
		}
		defer i2pServer.Stop(context.Background())

		// 启动 I2P 服务器
		go func() {
			if err := i2pServer.Start(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start I2P server: %v", err)
			}
		}()
	}

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down servers...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 关闭 I2P 服务器
	if i2pServer != nil {
		if err := i2pServer.Stop(ctx); err != nil {
			log.Printf("Error stopping I2P server: %v", err)
		}
	}

	log.Println("Servers stopped gracefully")
}

// expandPath 扩展路径中的环境变量
func expandPath(path string) string {
	return os.ExpandEnv(path)
}

// createStorageDirs 创建存储目录
func createStorageDirs(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		log.Printf("Created directory: %s", dir)
	}
	return nil
}
