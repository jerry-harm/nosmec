package server

import (
	"net"
	"net/http"
)

// I2PServer I2P 服务器和客户端
type I2PServer struct {
	server   *http.Server
	listener net.Listener
}
