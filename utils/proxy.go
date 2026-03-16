package utils

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/jerry-harm/nosmec/config"
)

// buildProxyURL 将ip:port地址转换为代理URL
func buildProxyURL(addr string) (*url.URL, error) {
	if addr == "" {
		return nil, nil
	}
	// 验证格式为ip:port，使用url库进行匹配
	return url.Parse("socks5://" + addr)
}

// ProxySelector 根据请求的hostname选择相应的代理
func ProxySelector(req *http.Request) (*url.URL, error) {
	cfg := config.GlobalConfig()
	hostname := req.URL.Hostname()

	// 检查域名后缀
	if strings.HasSuffix(hostname, ".i2p") && cfg.Proxy.I2PSocks != "" {
		return buildProxyURL(cfg.Proxy.I2PSocks)
	}
	if strings.HasSuffix(hostname, ".onion") && cfg.Proxy.OnionSocks != "" {
		return buildProxyURL(cfg.Proxy.OnionSocks)
	}

	// 如果设置了通用socks代理
	if cfg.Proxy.Socks != "" {
		return buildProxyURL(cfg.Proxy.Socks)
	}

	// 没有代理
	return nil, nil
}
