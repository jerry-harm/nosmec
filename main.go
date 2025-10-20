package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jerry-harm/nosmec/config"
)

func main() {

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down nosmec...")

	// 优雅关闭
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("nosmec stopped gracefully")
}
