package logger

import (
	"log/slog"
	"os"
)

func CreateDefaultLogger() {
	opts := prettyHandlerOptions{
		SlogOpts: slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}
	handler := newPrettyHandler(os.Stdout, opts)
	slog.SetDefault(slog.New(handler))
}
