package i2p

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	sam3 "github.com/go-i2p/sam3"
	"github.com/jerry-harm/nosmec/config"
)

var Sam *sam3.SAM
var ListenerSession *sam3.StreamSession // 用于listener的session，使用固定key
var Listener *sam3.StreamListener

// generateSessionName 生成带有随机字符串的session名称
func generateSessionName(base string) string {
	// 使用时间戳和随机数确保唯一性
	rand.Seed(time.Now().UnixNano())
	randomSuffix := strconv.FormatInt(time.Now().UnixNano(), 36) + strconv.Itoa(rand.Intn(1000))
	return base + "-" + randomSuffix
}

func Init() {
	log.Println("Initializing I2P... (this may take a moment)")
	var err error
	Sam, err = sam3.NewSAM(fmt.Sprintf("%s:%d", config.Global.I2P.SamAddress, config.Global.I2P.SamPort))
	if err != nil {
		log.Fatal("SAM build fialed:", err)
	}

	// 为listener生成/读取固定key（从文件）
	listenerKeys, err := Sam.EnsureKeyfile(filepath.Join(config.Global.BasePath, "sam.dat"))
	if err != nil {
		log.Fatal("listener key generate fialed:", err)
	}

	if err != nil {
		log.Fatal("dial key generate fialed:", err)
	}

	// 生成唯一的session名称
	listenerSessionName := generateSessionName("nosmec-listener")

	// 创建listener session（使用固定key）
	ListenerSession, err = Sam.NewStreamSession(listenerSessionName, listenerKeys, sam3.Options_Default)
	if err != nil {
		log.Fatal("listener session create fialed:", err)
	}

	// 使用listener session创建监听器
	Listener, err = ListenerSession.Listen()
	if err != nil {
		log.Fatal("session listen error")
	}
}

func IsI2PAddress(addr string) bool {
	return strings.Contains(addr, ".i2p")
}

// func I2PDial(network, addr string) (net.Conn, error) {
// 	if IsI2PAddress(addr) {
// 		conn, err := DialSession.Dial(network, addr)
// 		return conn, err
// 	}

// 	conn, err := net.Dial(network, addr)

// 	return conn, err
// }

func I2PProxy(req *http.Request) (*url.URL, error) {
	hostname := req.URL.Hostname()
	if strings.HasSuffix(hostname, ".i2p") {
		socksurl, _ := url.Parse("socks5://127.0.0.1:4447")
		return socksurl, nil
	} else {
		return nil, nil
	}
}
