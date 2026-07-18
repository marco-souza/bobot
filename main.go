package main

import (
	"log/slog"
	"os"

	"github.com/marco-souza/bobot/internal/discord"
)

func main() {
	// One structured logger for the whole process; default level is INFO,
	// reading from LOG_LEVEL env keeps ops tuning without a code change.
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, handlerOpts())))

	if err := discord.Run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

// handlerOpts returns the level from $LOG_LEVEL (debug/info/warn/error), default Info.
func handlerOpts() *slog.HandlerOptions {
	opts := &slog.HandlerOptions{AddSource: true}
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		opts.Level = slog.LevelDebug
	case "warn":
		opts.Level = slog.LevelWarn
	case "error":
		opts.Level = slog.LevelError
	default:
		opts.Level = slog.LevelInfo
	}
	return opts
}