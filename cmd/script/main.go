package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"fanapi/internal/config"
	"fanapi/internal/mq"
	"fanapi/internal/script"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if err := mq.Init(&cfg.NATS); err != nil {
		log.Fatalf("nats: %v", err)
	}
	log.Println("nats connected")
	if err := mq.EnsureStream(); err != nil {
		log.Fatalf("nats ensure stream: %v", err)
	}

	if err := script.StartWorkers(cfg.Worker); err != nil {
		log.Fatalf("start workers: %v", err)
	}

	// Block until signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("script worker shutting down")
}
