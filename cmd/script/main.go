package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"fanapi/internal/cache"
	"fanapi/internal/config"
	"fanapi/internal/db"
	"fanapi/internal/mq"
	"fanapi/internal/script"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if err := db.Init(&cfg.DB); err != nil {
		log.Fatalf("db: %v", err)
	}
	log.Println("db connected")

	if err := cache.Init(&cfg.Redis); err != nil {
		log.Fatalf("redis: %v", err)
	}
	log.Println("redis connected")

	if err := mq.Init(&cfg.NATS); err != nil {
		log.Fatalf("nats: %v", err)
	}
	log.Println("nats connected")

	if err := script.StartWorkers(); err != nil {
		log.Fatalf("start workers: %v", err)
	}

	// Block until signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("script worker shutting down")
}
