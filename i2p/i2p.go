package i2p

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	sam3 "github.com/go-i2p/go-sam-go"
	"github.com/jerry-harm/nosmec/config"
)

var Sam *sam3.SAM
var Session *sam3.StreamSession

func Init() {
	var err error
	Sam, err = sam3.NewSAM(fmt.Sprintf("%s:%d", config.Global.I2P.SamAddress, config.Global.I2P.SamPort))
	if err != nil {
		log.Fatal("SAM build fialed:", err)
	}

	//keys, err := Sam.EnsureKeyfile(filepath.Join(config.Global.BasePath, "sam.dat"))
	keys, err := Sam.NewKeys()
	if err != nil {
		log.Fatal("key generate fialed:", err)
	}

	Session, err = Sam.NewStreamSession("nosmec-tcp", keys, sam3.Options_Default)
	if err != nil {
		log.Fatal("session create fialed:", err)
	}
	// DatagramSession, err = Sam.NewDatagramSession("nosmec-udp", keys, sam3.Options_Default, 3819)
	if err != nil {
		log.Fatal("Datagramsession create fialed:", err)
	}
	http.DefaultClient = &http.Client{
		Transport: &http.Transport{
			DialContext: I2PDialContext,
		},
	}
}

func IsI2PAddress(addr string) bool {
	return strings.HasSuffix(addr, ".i2p") || strings.HasSuffix(addr, ".b32.i2p")
}

func I2PDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	url, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	hostname := url.Hostname()
	if strings.HasSuffix(hostname, ".i2p") {
		conn, err := Session.DialContext(ctx, addr)
		return conn, err
	}

	conn, err := net.Dial(network, addr)

	return conn, err
}
