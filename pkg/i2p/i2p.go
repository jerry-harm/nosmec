package i2p

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
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

// GetHTTPClient 获取使用 I2P 网络的 HTTP 客户端
func (s *I2PServer) GetHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: s.DialContext,
		},
	}
}

// I2PTransport 智能 I2P 分流 Transport
type I2PTransport struct {
	i2pServer       *I2PServer
	directTransport http.RoundTripper
}

// NewI2PTransport 创建新的 I2P 分流 Transport
func NewI2PTransport(i2pServer *I2PServer) *I2PTransport {
	return &I2PTransport{
		i2pServer:       i2pServer,
		directTransport: http.DefaultTransport,
	}
}

// RoundTrip 实现 http.RoundTripper 接口
func (t *I2PTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 检查是否是 I2P 地址
	if IsI2PAddress(req.URL.Host) {
		// 使用 I2P 连接
		return t.i2pRoundTrip(req)
	} else {
		// 使用直连
		return t.directTransport.RoundTrip(req)
	}
}

// i2pRoundTrip 使用 I2P 网络处理请求
func (t *I2PTransport) i2pRoundTrip(req *http.Request) (*http.Response, error) {
	// 创建使用 I2P dialer 的 Transport
	i2pTransport := &http.Transport{
		DialContext: t.i2pServer.DialContext,
	}
	return i2pTransport.RoundTrip(req)
}
