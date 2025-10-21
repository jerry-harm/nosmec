package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jerry-harm/nosmec/client"
	_ "github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/i2p"
)

func main() {
	i2p.Init()
	defer i2p.Sam.Close()
	defer i2p.Session.Close()
	log.Println("i2p loaded")

	newclient := &client.Client{}
	newclient.Init()
	defer newclient.Pool.Close("down")
	log.Println("client loaded")

	go newclient.SubscribeAllShortNote()

	log.Println("waiting")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down nosmec...")

	// 优雅关闭
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("nosmec stopped gracefully")
}
