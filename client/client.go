package client

import (
	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
)

type Client struct {
	ReadRelays  []string
	WriteRelays []string
	InboxRelays []string
	Pool        *nostr.Pool
}

func (c *Client) Init() {

	c.Pool = nostr.NewPool(nostr.PoolOptions{})
	for _, relay := range config.Global.Client.Relays {
		relayUrl := nostr.NormalizeURL(relay.Url)
		if relay.Read {
			c.ReadRelays = append(c.ReadRelays, relayUrl)
		}
		if relay.Write {
			c.WriteRelays = append(c.WriteRelays, relayUrl)
		}
		if relay.Inbox {
			c.InboxRelays = append(c.InboxRelays, relayUrl)
		}
		go c.Pool.EnsureRelay(relayUrl)
	}

}
