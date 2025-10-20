package i2p

import (
	"fmt"
	"log"
	"path/filepath"

	sam3 "github.com/go-i2p/go-sam-go"
	"github.com/go-i2p/i2pkeys"
	"github.com/jerry-harm/nosmec/config"
)

var Sam *sam3.SAM

func Init() {
	var err error
	Sam, err = sam3.NewSAM(fmt.Sprintf("%s:%d", config.Global.I2P.SamAddress, config.Global.I2P.SamPort))
	if err != nil {
		log.Fatal("SAM build fialed:", err)
	}

	keys, err := i2pkeys.LoadKeys(filepath.Join(config.Global.Storage.BasePath, "sam.dat"))

	if err != nil {
		log.Fatal("key generate fialed:", err)
	}
	Sam.DestinationKeys = &keys
}
