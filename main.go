package main

import (
	"log"

	"github.com/go-i2p/onramp"
)

func main() {
	garlic := &onramp.Garlic{}
	defer garlic.Close()
	listener, err := garlic.Listen()
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
}
