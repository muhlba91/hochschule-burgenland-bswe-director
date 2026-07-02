package logging

import (
	"context"
	"fmt"
	"log/slog"
)

// SlogWrapper is a wrapper around slog to implement the redis.Logger interface.
type SlogWrapper struct{}

// Printf implements the redis.Logger interface using slog.
// ctx: The context for the log entry, which can be used to extract trace IDs or other metadata if needed.
// format: The log message format string.
// v: The variadic arguments for the log message format string.
func (w *SlogWrapper) Printf(ctx context.Context, format string, v ...interface{}) {
	slog.InfoContext(ctx, fmt.Sprintf(format, v...))
}
