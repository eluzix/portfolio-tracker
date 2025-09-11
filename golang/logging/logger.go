package logging

import (
	"log/slog"
	"os"
	"sync"
)

var defaultLogger *slog.Logger
var once sync.Once

func Get() *slog.Logger {
	once.Do(func() {
		handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: false,
			Level:     slog.LevelInfo,
		})

		defaultLogger = slog.New(handler)
	})

	return defaultLogger
}
