package log

import (
	"log/slog"
	"os"
)

var globalLogger *slog.Logger

func init() {
	globalLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
}

func WithModule(module string) *slog.Logger {
	return globalLogger.With("module", module)
}
