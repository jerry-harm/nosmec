package i2p

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	sam3 "github.com/go-i2p/sam3"
	"github.com/jerry-harm/nosmec/config"
)

var Sam *sam3.SAM
var Session *sam3.StreamSession
var Listener *sam3.StreamListener

func Init() {
	var err error
	Sam, err = sam3.NewSAM(fmt.Sprintf("%s:%d", config.Global.I2P.SamAddress, config.Global.I2P.SamPort))
	if err != nil {
		log.Fatal("SAM build fialed:", err)
	}

	keys, err := Sam.EnsureKeyfile(filepath.Join(config.Global.BasePath, "sam.dat"))

	if err != nil {
		log.Fatal("key generate fialed:", err)
	}

	Session, err = Sam.NewStreamSession("nosmec", keys, sam3.Options_Default)
	if err != nil {
		log.Fatal("session create fialed:", err)
	}

	if err != nil {
		log.Fatal("Datagramsession create fialed:", err)
	}
	http.DefaultClient = &http.Client{
		Transport: &http.Transport{
			Dial: I2PDial,
		},
	}
	Listener, err = Session.Listen()
	if err != nil {
		log.Fatal("session listen error")
	}
}

func IsI2PAddress(addr string) bool {
	return strings.Contains(addr, ".i2p")
}

func I2PDial(network, addr string) (net.Conn, error) {
	if IsI2PAddress(addr) {
		conn, err := Session.Dial(network, addr)
		return conn, err
	}

	conn, err := net.Dial(network, addr)

	return conn, err
}

func I2PAndClearnetProxy(req *http.Request) (*url.URL, error) {
	hostname := req.URL.Hostname()
	if strings.HasSuffix(hostname, ".i2p") {
		socksurl, _ := url.Parse("socks5://127.0.0.1:4447")
		return socksurl, nil
	} else {
		return nil, nil
	}
}
