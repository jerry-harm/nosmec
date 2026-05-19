package nostr_sdk

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/nostr_sdk/dataloader"
)

type EventResult dataloader.Result[*nostr.Event]

func (sys *System) initializeReplaceableDataloaders() {
	sys.replaceableLoaders = make(map[nostr.Kind]*dataloader.Loader[nostr.PubKey, nostr.Event])
	sys.RegisterReplaceableDataloader(0)
	sys.RegisterReplaceableDataloader(3)
	sys.RegisterReplaceableDataloader(10000)
	sys.RegisterReplaceableDataloader(10001)
	sys.RegisterReplaceableDataloader(10002)
	sys.RegisterReplaceableDataloader(10003)
	sys.RegisterReplaceableDataloader(10004)
	sys.RegisterReplaceableDataloader(10005)
	sys.RegisterReplaceableDataloader(10006)
	sys.RegisterReplaceableDataloader(10007)
	sys.RegisterReplaceableDataloader(10015)
	sys.RegisterReplaceableDataloader(10019)
	sys.RegisterReplaceableDataloader(10030)
}

func (sys *System) RegisterReplaceableDataloader(kind nostr.Kind) {
	if sys.replaceableLoaders == nil {
		sys.replaceableLoaders = make(map[nostr.Kind]*dataloader.Loader[nostr.PubKey, nostr.Event])
	}
	if _, exists := sys.replaceableLoaders[kind]; !exists {
		sys.replaceableLoaders[kind] = sys.createReplaceableDataloader(kind)
	}
}

func (sys *System) createReplaceableDataloader(kind nostr.Kind) *dataloader.Loader[nostr.PubKey, nostr.Event] {
	return dataloader.NewBatchedLoader(
		func(ctxs []context.Context, pubkeys []nostr.PubKey) map[nostr.PubKey]dataloader.Result[nostr.Event] {
			return sys.batchLoadReplaceableEvents(ctxs, kind, pubkeys)
		},
		dataloader.Options{
			Wait:         time.Millisecond * 110,
			MaxThreshold: 30,
		},
	)
}

func (sys *System) batchLoadReplaceableEvents(
	ctxs []context.Context,
	kind nostr.Kind,
	pubkeys []nostr.PubKey,
) map[nostr.PubKey]dataloader.Result[nostr.Event] {
	batchSize := len(pubkeys)
	results := make(map[nostr.PubKey]dataloader.Result[nostr.Event], batchSize)
	relayFilter := make([]nostr.DirectedFilter, 0, max(3, batchSize*2))
	relayFilterIndex := make(map[string]int, max(3, batchSize*2))

	wg := sync.WaitGroup{}
	wg.Add(len(pubkeys))
	cm := sync.Mutex{}

	aggregatedContext, aggregatedCancel := context.WithCancel(context.Background())
	waiting := atomic.Int32{}
	waiting.Add(int32(len(pubkeys)))

	for i, pubkey := range pubkeys {
		ctx, cancel := context.WithCancel(ctxs[i])
		defer cancel()

		// build batched queries for the external relays
		go func(i int, pubkey nostr.PubKey, ctx context.Context) {
			// gather relays we'll use for this pubkey
			relays := sys.determineRelaysToQuery(ctx, pubkey, kind)

			cm.Lock()
			for _, relay := range relays {
				// each relay will have a custom filter
				idx, ok := relayFilterIndex[relay]
				var dfilter nostr.DirectedFilter
				if ok {
					dfilter = relayFilter[idx]
				} else {
					dfilter = nostr.DirectedFilter{
						Relay: relay,
						Filter: nostr.Filter{
							Kinds:   []nostr.Kind{kind},
							Authors: make([]nostr.PubKey, 0, batchSize-i /* this and all pubkeys after this can be added */),
						},
					}
					idx = len(relayFilter)
					relayFilterIndex[relay] = idx
					relayFilter = append(relayFilter, dfilter)
				}
				dfilter.Authors = append(dfilter.Authors, pubkey)
				relayFilter[idx] = dfilter
			}
			cm.Unlock()
			wg.Done()

			<-ctx.Done()
			if waiting.Add(-1) == 0 {
				aggregatedCancel()
			}
		}(i, pubkey, ctx)
	}

	// query all relays with the prepared filters
	wg.Wait()
	multiSubs := sys.Pool.BatchedQueryMany(aggregatedContext, relayFilter, nostr.SubscriptionOptions{
		Label:          "repl~" + strconv.Itoa(int(kind)),
		MaxWaitForEOSE: time.Second * 3,
	})
	for {
		select {
		case ie, more := <-multiSubs:
			if !more {
				return results
			}

			// insert this event at the desired position
			if val, ok := results[ie.PubKey]; !ok || val.Data.ID == nostr.ZeroID || val.Data.CreatedAt < ie.CreatedAt {
				results[ie.PubKey] = dataloader.Result[nostr.Event]{Data: ie.Event}
			}
		case <-aggregatedContext.Done():
			return results
		}
	}
}

func (sys *System) determineRelaysToQuery(ctx context.Context, pubkey nostr.PubKey, kind nostr.Kind) []string {
	var relays []string

	// search in specific relays for user
	if kind == 10002 {
		// prevent infinite loops by jumping directly to this
		relays = sys.Hints.TopN(pubkey, 3)
	} else {
		ctx, cancel := context.WithTimeoutCause(ctx, time.Millisecond*2300,
			errors.New("fetching relays in subloader took too long"),
		)

		if kind == 0 || kind == 3 {
			// leave room for two hardcoded relays because people are stupid
			relays = sys.FetchOutboxRelays(ctx, pubkey, 1)
		} else {
			relays = sys.FetchOutboxRelays(ctx, pubkey, 3)
		}

		cancel()
	}

	// use a different set of extra relays depending on the kind
	needed := 3 - len(relays)
	for range needed {
		var next string
		switch kind {
		case 0:
			next = sys.MetadataRelays.Next()
		case 3:
			next = sys.FollowListRelays.Next()
		case 10002:
			next = sys.RelayListRelays.Next()
		default:
			next = sys.FallbackRelays.Next()
		}

		if !slices.Contains(relays, next) {
			relays = append(relays, next)
		}
	}

	return relays
}
