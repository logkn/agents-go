package utils

import (
	"log/slog"
	"os"
	"strings"
)

// SetupLogger creates a structured logger with the specified log level from environment variable.
// Falls back to INFO level if LOG_LEVEL is not set or invalid.
func SetupLogger() *slog.Logger {
	logLevel := slog.LevelInfo // default to INFO
	
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		switch strings.ToUpper(level) {
		case "DEBUG":
			logLevel = slog.LevelDebug
		case "WARN":
			logLevel = slog.LevelWarn
		case "ERROR":
			logLevel = slog.LevelError
		case "INFO":
			logLevel = slog.LevelInfo
		default:
			// Invalid LOG_LEVEL value, fall back to INFO and log a warning
			logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}))
			logger.Warn("invalid LOG_LEVEL value, falling back to INFO", "provided_level", level)
		}
	}
	
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))
}