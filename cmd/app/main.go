package main

import (
	"context"
	"log_shelter/internal/config"
	"log_shelter/internal/logging"
	"log_shelter/internal/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configName := os.Getenv("CONFIG_PATH")

	cfg := config.GetConfig(configName)

	logging.InitLogger(cfg.Logger)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGABRT, syscall.SIGINT)
	defer cancel()

	srv := server.NewServer(ctx, cfg)

	srv.Run()
}
