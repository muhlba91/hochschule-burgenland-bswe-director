package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

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

	configureServer(s)
	configureRoutes(s)

	return s
}

// Start starts the echo server.
func (s *Server) Start() {
	go func() {
		slog.Info("starting server",
			slog.String("address", s.Address),
		)
		if err := s.Server.Start(s.Address); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

	if err := s.Server.Shutdown(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to shutdown server",
			slog.Any(logging.FieldError, err),
		)
	}
}
