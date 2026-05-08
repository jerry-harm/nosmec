package utils

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/jerry-harm/nosmec/config"
)

func buildProxyURL(addr string) (*url.URL, error) {
	if addr == "" {
		return nil, nil
	}
	return url.Parse("socks5://" + addr)
}

func ProxySelector(req *http.Request) (*url.URL, error) {
	cfg := config.GetProxyConfig()
	hostname := req.URL.Hostname()
	isI2P := strings.HasSuffix(hostname, ".i2p")

	switch {
	case cfg.Socks != "" && cfg.I2PSocks == "":
		return buildProxyURL(cfg.Socks)

	case cfg.I2PSocks != "" && cfg.Socks == "":
		if isI2P {
			return buildProxyURL(cfg.I2PSocks)
		}
		return nil, nil

	case cfg.Socks != "" && cfg.I2PSocks != "":
		if isI2P {
			return buildProxyURL(cfg.I2PSocks)
		}
		return buildProxyURL(cfg.Socks)

	default:
		return nil, nil
	}
}
