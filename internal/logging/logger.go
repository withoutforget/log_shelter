package logging

import (
	"log/slog"
	"log_shelter/internal/config"
	"os"

	"github.com/lmittmann/tint"
)

func InitLogger(cfg config.LoggerConfig) {

	var logger *slog.Logger

	if cfg.DeveloperMode {

		logger = slog.New(tint.NewHandler(os.Stderr, &tint.Options{
			Level:     slog.LevelDebug,
			AddSource: true,
		}))

	} else {
		logger = slog.New(
			slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelInfo,
			}))
	}

	slog.SetDefault(logger)
}
