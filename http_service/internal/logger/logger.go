/*
 * Shared project logger.
 * 1. Build a single structured logger for the whole service.
 * 2. Normalize output format based on runtime environment.
 * 3. Keep logging setup isolated from business packages.
 */
package logger

import (
	"log/slog"
	"os"

	"ajoliving_web/http_service/internal/config"
)

// 1. New creates a structured logger instance.
func New(cfg *config.Config) *slog.Logger {
	options := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if cfg.AppEnv == "development" {
		return slog.New(slog.NewTextHandler(os.Stdout, options))
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, options))
}
