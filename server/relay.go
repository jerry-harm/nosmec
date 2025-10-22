package server

import (
	"context"
	"fmt"
	"iter"
	"log"
	"net/http"
	"path/filepath"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/eventstore/lmdb"
	"fiatjaf.com/nostr/khatru"
	"fiatjaf.com/nostr/nip19"
	"github.com/jerry-harm/nosmec/config"
)

func NewRelay() (*khatru.Relay, error) {
	// create the relay instance
	relay := khatru.NewRelay()

	prefix, decoded, err := nip19.Decode(config.Global.Client.PrivateKey)
	if err == nil || prefix == "nsec" {
		secretKey, ok := decoded.(nostr.SecretKey)
		pubKey := nostr.GetPublicKey(secretKey)
		if ok {
			relay.Info.PubKey = &pubKey
		} else {
			log.Println("error nip11 pub key")
		}
	}

	// set up some basic properties (will be returned on the NIP-11 endpoint)
	relay.Info.Name = config.Global.Server.NIP11.Name
	relay.Info.Description = config.Global.Server.NIP11.Description

	db := lmdb.LMDBBackend{Path: filepath.Join(config.Global.BasePath, "nosmec.db")}

	if err := db.Init(); err != nil {
		panic(err)
	}

	relay.StoreEvent = func(ctx context.Context, event nostr.Event) error {
		return db.SaveEvent(event)
	}

	relay.DeleteEvent = func(ctx context.Context, id nostr.ID) error {
		return db.DeleteEvent(id)
	}

	relay.QueryStored = func(ctx context.Context, filter nostr.Filter) iter.Seq[nostr.Event] {
		return func(yield func(nostr.Event) bool) {
			seq := db.QueryEvents(filter, 500) // 限制最大返回数量
			for event := range seq {
				if !yield(event) {
					break
				}
			}
		}
	}
	mux := relay.Router()
	// set up other http handlers
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/html")
		fmt.Fprintf(w, `<b>welcome</b> it is nosmec!`)
	})

	return relay, nil
}
