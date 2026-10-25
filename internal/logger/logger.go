package logger

import (
	"log/slog"
	"os"
)

func New(levelStr string) *slog.Logger {
	var level slog.Level

	if err := level.UnmarshalText([]byte(levelStr)); err != nil {
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}
