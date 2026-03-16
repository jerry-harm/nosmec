package utils

import (
	"context"
	"fmt"
	"iter"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/eventstore/lmdb"
	"fiatjaf.com/nostr/khatru"
	"fiatjaf.com/nostr/nip19"
	"github.com/go-i2p/sam3"
	"github.com/jerry-harm/nosmec/config"
)

var sam *sam3.SAM
var i2pListenerSession *sam3.StreamSession
var i2pListener *sam3.StreamListener

func generateSessionName(base string) string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	randomSuffix := strconv.FormatInt(time.Now().UnixNano(), 36) + strconv.Itoa(rand.Intn(1000))
	return base + "-" + randomSuffix
}

func sam_init() {
	log.Println("Initializing I2P... (this may take a moment)")
	var err error
	sam, err = sam3.NewSAM(fmt.Sprintf("%s:%d", config.GlobalConfig().LocalServer.I2P.SamAddress, config.GlobalConfig().LocalServer.I2P.SamPort))
	if err != nil {
		log.Fatal("SAM build fialed:", err)
	}

	// 为listener生成/读取固定key（从文件）
	listenerKeys, err := sam.EnsureKeyfile(filepath.Join(config.GlobalConfig().DataDir, "sam.dat"))
	if err != nil {
		log.Fatal("listener key generate fialed:", err)
	}

	// 生成唯一的session名称
	listenerSessionName := generateSessionName("nosmec-listener")

	// 创建listener session（使用固定key）
	i2pListenerSession, err = sam.NewStreamSession(listenerSessionName, listenerKeys, sam3.Options_Default)
	if err != nil {
		log.Fatal("listener session create fialed:", err)
	}

	// 使用listener session创建监听器
	i2pListener, err = i2pListenerSession.Listen()
	if err != nil {
		log.Fatal("session listen error")
	}
}

// func I2PDial(network, addr string) (net.Conn, error) {
// 	if IsI2PAddress(addr) {
// 		conn, err := dialSession.Dial(network, addr)
// 		return conn, err
// 	}

// 	conn, err := net.Dial(network, addr)

// 	return conn, err
// }

func newRelay() (*khatru.Relay, error) {
	// create the relay instance
	relay := khatru.NewRelay()

	prefix, decoded, err := nip19.Decode(config.GlobalConfig().PrivateKey)
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
	relay.Info.Name = config.GlobalConfig().LocalServer.NIP11.Name
	relay.Info.Description = config.GlobalConfig().LocalServer.NIP11.Description

	db := lmdb.LMDBBackend{Path: filepath.Join(config.GlobalConfig().DataDir, "nosmec.db")}

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

// for standalone i2p server
func Run() {

	go func() {
		sam_init()
	}()

	relayServer, err := newRelay()
	if err != nil {
		log.Fatalln(err)
	}

	serverStopped := make(chan bool, 1)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Starting Nostr relay server...")
		if err := http.Serve(i2pListener, relayServer); err != nil {
			log.Printf("Server error: %v\n", err)
		}
		serverStopped <- true
	}()

	select {
	case <-sigChan:
		log.Println("Shutting down server...")
	case <-serverStopped:
		log.Println("Server stopped unexpectedly")
	}

	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer sam.Close()
	defer i2pListenerSession.Close()
	defer i2pListener.Close()
	defer cancel()
}
