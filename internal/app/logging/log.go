package logging

import (
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"
)

// Init initializes the logging configuration based on environment variables.
func Init() {
	var handler slog.Handler
	level := getLogLevel()
	opts := &slog.HandlerOptions{
		Level: level,
	}

	format := os.Getenv("LOG_FORMAT")
	if format == "test" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
	setRedisLogger()
}

// getLogLevel gets the slog.Level based on the LOG_LEVEL environment variable.
func getLogLevel() slog.Level {
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// setRedisLogger sets the Redis logger to use the SlogWrapper, which implements the redis.Logger interface using slog.
func setRedisLogger() {
	redis.SetLogger(&SlogWrapper{})
}
