package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/arkhe-systems/senddock/pkg/app"
	"github.com/arkhe-systems/senddock/pkg/config"
)

type slogBridge struct{}

func (slogBridge) Write(p []byte) (int, error) {
	slog.Info(strings.TrimRight(string(p), "\n"))
	return len(p), nil
}

func initLogging() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
	log.SetFlags(0)
	log.SetOutput(slogBridge{})
}

func main() {
	initLogging()

	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		slog.Error("startup failed", "error", err)
		os.Exit(1)
	}
	defer application.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := application.Run(ctx); err != nil {
		slog.Error("server exited with error", "error", err)
		application.Close()
		os.Exit(1)
	}
}
