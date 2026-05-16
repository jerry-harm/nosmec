package cmd

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/utils"
	"github.com/spf13/cobra"
)

func registerGossipCommands() {
	gossipCmd := &cobra.Command{
		Use:   "gossip",
		Short: "Batch fetch users' relay lists (NIP-65) and update KnownRelays",
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

	var relayCount int32

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	results := make(chan []string, len(subs))
	completed := 0

	for _, sub := range subs {
		go func(pubkeyHex string) {
			pk, err := hexToPubKey(pubkeyHex)
			if err != nil {
				results <- nil
				return
			}
			relays, err := utils.DiscoverUserRelays(ctx, app, pk)
			if err != nil || len(relays) == 0 {
				results <- nil
				return
			}
			atomic.AddInt32(&relayCount, int32(len(relays)))
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
				for _, rel := range r {
					relaySet[rel] = struct{}{}
				}
			}
			fmt.Printf("\rProcessing: %d/%d users, %d relays found", completed, len(subs), atomic.LoadInt32(&relayCount))
		case <-ticker.C:
		}
	}

	fmt.Printf("\nDiscovered %d unique relays from %d users\n", len(relaySet), len(subs))

	if len(relaySet) > 0 {
		relays := make([]string, 0, len(relaySet))
		for r := range relaySet {
			relays = append(relays, r)
		}
		app.TrackRelays(relays)
		fmt.Println("Relays tracked. They will be saved to config on next app close.")
	}
}

func hexToPubKey(hex string) (nostr.PubKey, error) {
	return nostr.PubKeyFromHex(hex)
}