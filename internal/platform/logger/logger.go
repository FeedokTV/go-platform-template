package logger

import (
	"go-platform-template/internal/platform/config"
	"io"
	"log/slog"
	"os"
)

func newWithWriter(writer io.Writer, cfg config.LoggerConfig) *slog.Logger {

	var level slog.Level
	switch cfg.Level {
	case config.LevelDebug:
		level = slog.LevelDebug
	case config.LevelInfo:
		level = slog.LevelInfo
	case config.LevelWarn:
		level = slog.LevelWarn
	case config.LevelError:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler

	switch cfg.Format {
	case config.FormatText:
		opts.AddSource = true
		handler = slog.NewTextHandler(writer, &opts)
	case config.FormatJSON:
		handler = slog.NewJSONHandler(writer, &opts)
	default:
		handler = slog.NewJSONHandler(writer, &opts)
	}

	return slog.New(handler)
}

func New(cfg config.LoggerConfig) *slog.Logger {
	return newWithWriter(os.Stderr, cfg)
}
