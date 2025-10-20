package nostr

import (
	"context"

	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
)

type Client struct {
	pool *nostr.Pool
	ctx  context.Context
}

func (c *Client) Init() {
	c.ctx = context.Background()
	c.pool = nostr.NewPool(nostr.PoolOptions{})
	for _, url := range config.Global.Client.DefaultRelays {
		relayUrl := nostr.NormalizeURL(url)
		c.pool.Relays.LoadAndStore(relayUrl, nostr.NewRelay(c.ctx, relayUrl, nostr.RelayOptions{}))
	}
}
