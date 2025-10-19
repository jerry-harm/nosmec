package nostr

import (
	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
)

// Client Nostr 客户端
type Client struct {
	config *config.Config
	relays map[string]*nostr.Relay
}

// NewClient 创建新的 Nostr 客户端
func NewClient(cfg *config.Config) (*Client, error) {
	client := &Client{
		config: cfg,
		relays: make(map[string]*nostr.Relay),
	}

	return client, nil
}
