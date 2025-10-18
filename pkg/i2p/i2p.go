package i2p

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-i2p/onramp"
)

// I2PServer I2P 服务器和客户端
type I2PServer struct {
	server   *http.Server
	listener net.Listener
	garlic   *onramp.Garlic
}

// NewI2PServer 创建新的 I2P 服务器
func NewI2PServer(handler http.Handler, tunnelName string, basePath string, samAddr string) (*I2PServer, error) {
	// 在 garlic 初始化之前设置所有密钥存储路径
	if basePath != "" {
		// 设置 I2P 密钥存储路径
		i2pPath := filepath.Join(filepath.Dir(basePath), "i2pkeys")
		onramp.I2P_KEYSTORE_PATH = i2pPath

		// 设置 Onion 密钥存储路径（使用 i2pkeys 目录的父目录 + onionkeys）
		onionPath := filepath.Join(filepath.Dir(basePath), "onionkeys")
		onramp.ONION_KEYSTORE_PATH = onionPath

		// 设置 TLS 密钥存储路径（使用 i2pkeys 目录的父目录 + tlskeys）
		tlsPath := filepath.Join(filepath.Dir(basePath), "tlskeys")
		onramp.TLS_KEYSTORE_PATH = tlsPath

		// 调用 I2PKeystorePath 确保目录正确创建并开始使用
		if path, err := onramp.I2PKeystorePath(); err != nil {
			return nil, fmt.Errorf("failed to initialize I2P keystore path: %w", err)
		} else {
			log.Printf("I2P keystore path initialized: %s", path)
		}
	}

	// 创建 Garlic 隧道
	garlic, err := onramp.NewGarlic(tunnelName, samAddr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create I2P garlic tunnel: %w", err)
	}

	// 创建监听器
	listener, err := garlic.Listen()
	if err != nil {
		garlic.Close()
		return nil, fmt.Errorf("failed to create I2P listener: %w", err)
	}

	log.Printf("I2P listener created: %s", listener.Addr().String())

	// 创建 HTTP 服务器
	server := &http.Server{
		Handler: handler,
	}

	return &I2PServer{
		server:   server,
		listener: listener,
		garlic:   garlic,
	}, nil
}

// Dial 连接到 I2P 地址
func (s *I2PServer) Dial(network, addr string) (net.Conn, error) {
	return s.garlic.Dial(network, addr)
}

// DialContext 使用上下文连接到 I2P 地址
func (s *I2PServer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return s.garlic.DialContext(ctx, network, addr)
}

// IsI2PAddress 检查地址是否为 I2P 地址
func IsI2PAddress(addr string) bool {
	return strings.HasSuffix(addr, ".i2p") || strings.HasSuffix(addr, ".b32.i2p")
}

// Start 启动 I2P 服务器
func (s *I2PServer) Start() error {
	// 获取 I2P 目的地地址
	destination := s.GetDestination()
	log.Printf("I2P server started on: %s", s.listener.Addr().String())
	log.Printf("I2P destination: %s", destination)

	// 在 I2P 监听器上服务
	return s.server.Serve(s.listener)
}

// Stop 停止 I2P 服务器
func (s *I2PServer) Stop(ctx context.Context) error {
	// 停止 HTTP 服务器
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown I2P server: %w", err)
	}

	// 关闭监听器和 Garlic 隧道
	if s.listener != nil {
		s.listener.Close()
	}
	if s.garlic != nil {
		s.garlic.Close()
	}

	return nil
}

// GetDestination 获取 I2P 目的地地址
func (s *I2PServer) GetDestination() string {
	if s.garlic != nil {
		keys, err := s.garlic.Keys()
		if err != nil {
			log.Printf("Failed to get I2P keys: %v", err)
			return ""
		}
		return keys.Addr().Base32()
	}
	return ""
}

// GetListenerAddr 获取监听器地址
func (s *I2PServer) GetListenerAddr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return ""
}
