package config

import (
	"log/slog"
	"os"
)

func InitLogger() (*slog.Logger, error) {
	logOpts := &slog.HandlerOptions{
		AddSource: !Production,
		Level:     slog.Level(LogLevel),
	}

	var logger *slog.Logger
	if Production {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, logOpts))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, logOpts))
	}

	return logger, nil
}
