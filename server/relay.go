package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"iter"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/eventstore/lmdb"
	"fiatjaf.com/nostr/khatru"
	"fiatjaf.com/nostr/nip19"
)

// RelayServer nostr relay 服务器
type RelayServer struct {
	relay   *khatru.Relay
	db      *lmdb.LMDBBackend
	handler http.Handler
}

// NewRelayServer 创建新的 relay 服务器
func NewRelayServer(host string, port int, dbPath string, nip11Name, nip11Description, nip11PubKey, nip11Contact, nip11Software, nip11Version string) (*RelayServer, error) {
	// 创建 LMDB 存储
	db := &lmdb.LMDBBackend{Path: dbPath}
	if err := db.Init(); err != nil {
		return nil, fmt.Errorf("failed to init LMDB: %w", err)
	}

	// 创建 relay
	relay := khatru.NewRelay()

	// 设置 relay 基本信息 (NIP-11)
	relay.Info.Name = nip11Name
	relay.Info.Description = nip11Description
	relay.Info.Contact = nip11Contact
	relay.Info.Software = nip11Software
	relay.Info.Version = nip11Version

	// 解析 npub 格式的 pubkey
	if nip11PubKey != "" {
		prefix, decoded, err := nip19.Decode(nip11PubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decode npub: %w", err)
		}
		if prefix != "npub" {
			return nil, fmt.Errorf("expected npub prefix, got %s", prefix)
		}
		pubKey, ok := decoded.(nostr.PubKey)
		if !ok {
			return nil, fmt.Errorf("decoded value is not a pubkey")
		}
		relay.Info.PubKey = &pubKey
	}

	// 使用 LMDB 存储
	relay.StoreEvent = func(ctx context.Context, event nostr.Event) error {
		return db.SaveEvent(event)
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

	relay.DeleteEvent = func(ctx context.Context, id nostr.ID) error {
		return db.DeleteEvent(id)
	}

	// 创建处理器
	mux := relay.Router()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/html")
		fmt.Fprintf(w, `<h1>Welcome to nosmec nostr relay</h1>
		<p>This is a simple nostr relay using LMDB storage.</p>
		<p>Connect via WebSocket to use the relay.</p>`)
	})

	return &RelayServer{
		relay:   relay,
		db:      db,
		handler: mux,
	}, nil
}

// Start 启动 relay 服务器
func (s *RelayServer) Start(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)

	log.Printf("Starting nostr relay server on %s", addr)
	return http.ListenAndServe(addr, s.GetHandler())
}

// GetHandler 获取 HTTP 处理器
func (s *RelayServer) GetHandler() http.Handler {
	return s.handler
}

// Close 关闭服务器和存储
func (s *RelayServer) Close() error {
	if s.db != nil {
		s.db.Close()
	}
	return nil
}
