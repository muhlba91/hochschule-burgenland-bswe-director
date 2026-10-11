package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/configuration"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/logging"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/orchestrator"
)

// shutdownTimeout defines the duration for graceful shutdown of the echo server.
const shutdownTimeout = 5 * time.Second

// Server represents the echo web server.
type Server struct {
	Address    string
	Server     *echo.Echo
	Dispatcher *orchestrator.Dispatcher

	cancel context.CancelFunc
	done   chan struct{}
}

// NewServer creates a new echo server.
// configuration: The configuration data for the server.
// dispatcher: The websocket dispatcher instance.
func NewServer(
	configuration *configuration.Data,
	dispatcher *orchestrator.Dispatcher,
) *Server {
	s := &Server{
		Address:    fmt.Sprintf("%s:%d", configuration.ServerHost, configuration.ServerPort),
		Server:     echo.New(),
		Dispatcher: dispatcher,
	}

	configureServer(s, configuration)
	configureRoutes(s, configuration)

	return s
}

// Start starts the echo server.
func (s *Server) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.done = make(chan struct{})

	go func() {
		defer close(s.done)
		slog.Info("starting server",
			slog.String("address", s.Address),
		)
		sc := echo.StartConfig{
			Address:         s.Address,
			HideBanner:      true,
			HidePort:        true,
			GracefulTimeout: shutdownTimeout,
		}
		if err := sc.Start(ctx, s.Server); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start server",
				slog.String("address", s.Address),
				slog.Any(logging.FieldError, err),
			)
		}
	}()
}

// Stop stops the echo server gracefully.
func (s *Server) Stop() {
	if s.Server == nil {
		return
	}

	slog.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if s.cancel == nil {
		return
	}
	s.cancel()

	select {
	case <-s.done:
	case <-ctx.Done():
		slog.ErrorContext(ctx, "failed to shutdown server",
			slog.Any(logging.FieldError, ctx.Err()),
		)
	}
}
