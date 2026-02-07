package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"bigbrother/internal/bot"
	"bigbrother/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := bot.Run(ctx, cfg); err != nil {
		log.Fatalf("bot error: %v", err)
	}
}
