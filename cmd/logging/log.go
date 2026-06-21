package logging

import (
	"os"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

// Init initializes the logging configuration based on environment variables.
func Init() {
	setLogLevel()
	setLogFormat()
	setRedisLogger()
}

// setLogFormat sets the log format based on the LOG_FORMAT environment variable.
func setLogFormat() {
	format := os.Getenv("LOG_FORMAT")
	if format == "test" {
		log.SetFormatter(&log.TextFormatter{})
	} else {
		log.SetFormatter(&log.JSONFormatter{})
	}
}

// setLogLevel sets the log level based on the LOG_LEVEL environment variable.
func setLogLevel() {
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}

// setRedisLogger sets the Redis logger to use the LogrusWrapper, which implements the redis.Logger interface using logrus.
func setRedisLogger() {
	redis.SetLogger(&LogrusWrapper{})
}
