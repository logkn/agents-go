package utils

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// SetupLogger creates a structured logger with the specified log level from environment variable.
// Falls back to INFO level if LOG_LEVEL is not set or invalid.
func SetupLogger() *slog.Logger {
	logLevel := envLevel()

	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))
}

func NilLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelError + 1, // Above error level to suppress all logs
	}))
}

func envLevel() slog.Level {
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		switch strings.ToUpper(level) {
		case "DEBUG":
			return slog.LevelDebug
		case "WARN":
			return slog.LevelWarn
		case "ERROR":
			return slog.LevelError
		case "INFO":
			return slog.LevelInfo
		}
	}
	// Default to INFO level
	return slog.LevelInfo
}

func FileLogger(filename string) *slog.Logger {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	logLevel := envLevel()

	return slog.New(slog.NewTextHandler(file, &slog.HandlerOptions{
		Level: logLevel,
	}))
}
