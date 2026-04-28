package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/arkhe-systems/senddock/internal/config"
	"github.com/arkhe-systems/senddock/pkg/app"
)

func main() {
	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer application.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := application.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
