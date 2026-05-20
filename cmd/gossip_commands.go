package cmd

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"github.com/jerry-harm/nosmec/logger"
	"github.com/jerry-harm/nosmec/utils"
	"github.com/spf13/cobra"
)

func registerGossipCommands() {
	gossipCmd := &cobra.Command{
		Use:   "gossip",
		Short: "Batch fetch users' relay lists (NIP-65) and ensure them in the pool",
		Args:  cobra.NoArgs,
		Run:   runGossip,
	}

	RegisterCommandGroup("Gossip", "Relay discovery", gossipCmd)
}

func runGossip(cmd *cobra.Command, args []string) {
	app := getApp()

	subs := app.ListSubscriptions("user")
	if len(subs) == 0 {
		fmt.Println("No user subscriptions found. Following users first.")
		return
	}

	logger.Debug("runGossip: starting", "users", len(subs))

	var relayCount int32

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	results := make(chan []string, len(subs))
	completed := 0

	for _, sub := range subs {
		go func(subID string) {
			prefix, value, err := nip19.Decode(subID)
			if err != nil || prefix != "npub" {
				results <- nil
				return
			}
			pk, ok := value.(nostr.PubKey)
			if !ok {
				results <- nil
				return
			}
			relays, err := utils.DiscoverUserRelays(ctx, app, pk)
			if err != nil || len(relays) == 0 {
				results <- nil
				return
			}
			results <- relays
		}(sub.ID)
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	relaySet := make(map[string]struct{})

	for completed < len(subs) {
		select {
		case r := <-results:
			completed++
			if r != nil {
				logger.Debug("runGossip: got relays from user", "completed", completed, "relayCount", len(r))
				for _, rel := range r {
					relaySet[rel] = struct{}{}
				}
				atomic.AddInt32(&relayCount, int32(len(r)))
			}
			fmt.Printf("\rProcessing: %d/%d users, %d relays found", completed, len(subs), atomic.LoadInt32(&relayCount))
		case <-ticker.C:
		}
	}

	fmt.Printf("\nDiscovered %d unique relays from %d users\n", len(relaySet), len(subs))

	if len(relaySet) > 0 {
		fmt.Printf("Ensured %d relays in pool for this session.\n", len(relaySet))
	}
}
