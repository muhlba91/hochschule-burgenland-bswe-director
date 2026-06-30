package logging

import (
	"context"

	"github.com/sirupsen/logrus"
)

// LogrusWrapper is a wrapper around logrus.Logger to implement the redis.Logger interface.
type LogrusWrapper struct{}

// Printf implements the redis.Logger interface using logrus.
// ctx: The context for the log entry, which can be used to extract trace IDs or other metadata if needed.
// format: The log message format string.
// v: The variadic arguments for the log message format string.
func (w *LogrusWrapper) Printf(_ context.Context, format string, v ...interface{}) {
	logrus.Infof(format, v...)
}
