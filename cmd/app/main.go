package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"log_shelter/internal/config"
	"log_shelter/internal/logging"
	"log_shelter/internal/server"
)

func main() {
	configName := os.Getenv("CONFIG_PATH")

	cfg := config.GetConfig(configName)

	logging.InitLogger(cfg.Logger)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGABRT,
		syscall.SIGINT,
	)
	defer cancel()

	srv := server.NewServer(ctx, cfg)

	srv.Run()
}
