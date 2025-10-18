package nostr

import (
	"context"
	"log"

	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/pkg/config"
	"github.com/jerry-harm/nosmec/pkg/i2p"
)

// Client Nostr 客户端
type Client struct {
	config    *config.Config
	relays    map[string]*nostr.Relay
	i2pServer *i2p.I2PServer
}

// NewClient 创建新的 Nostr 客户端
func NewClient(cfg *config.Config, i2pServer *i2p.I2PServer) (*Client, error) {
	client := &Client{
		config:    cfg,
		relays:    make(map[string]*nostr.Relay),
		i2pServer: i2pServer,
	}

	return client, nil
}

// NewClientWithoutI2P 创建没有 I2P 支持的 Nostr 客户端
func NewClientWithoutI2P(cfg *config.Config) (*Client, error) {
	return NewClient(cfg, nil)
}

// Connect 连接到配置的 relay
func (c *Client) Connect() error {
	ctx := context.Background()
	for _, relayURL := range c.config.Client.DefaultRelays {
		relay, err := nostr.RelayConnect(ctx, relayURL, nostr.RelayOptions{})
		if err != nil {
			log.Printf("Failed to connect to relay %s: %v", relayURL, err)
			continue
		}
		c.relays[relayURL] = relay
		log.Printf("Connected to relay: %s", relayURL)
	}

	return nil
}

// Close 关闭所有连接
func (c *Client) Close() {
	for url, relay := range c.relays {
		relay.Close()
		log.Printf("Closed relay: %s", url)
	}
}

// GetRelays 获取连接的 relay 列表
func (c *Client) GetRelays() []string {
	relays := make([]string, 0, len(c.relays))
	for url := range c.relays {
		relays = append(relays, url)
	}
	return relays
}

// PublishEvent 发布事件
func (c *Client) PublishEvent(event nostr.Event) error {
	ctx := context.Background()
	// TODO: 从配置读取私钥并签名事件
	for _, relay := range c.relays {
		err := relay.Publish(ctx, event)
		if err != nil {
			log.Printf("Failed to publish event to relay: %v", err)
		}
	}
	return nil
}

// Subscribe 订阅事件
func (c *Client) Subscribe(filter nostr.Filter) (*nostr.Subscription, error) {
	ctx := context.Background()
	// 暂时只使用第一个 relay
	for _, relay := range c.relays {
		return relay.Subscribe(ctx, filter, nostr.SubscriptionOptions{})
	}
	return nil, nil
}
