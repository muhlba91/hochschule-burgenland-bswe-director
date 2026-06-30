package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/cmd/configuration"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/dispatcher"
)

// shutdownTimeout defines the duration for graceful shutdown of the echo server.
const shutdownTimeout = 5 * time.Second

// Server represents the echo web server.
type Server struct {
	Address    string
	Server     *echo.Echo
	Dispatcher *dispatcher.Dispatcher
}

// NewServer creates a new echo server.
// configuration: The configuration data for the server.
// dispatcher: The websocket dispatcher instance.
func NewServer(
	configuration *configuration.Data,
	dispatcher *dispatcher.Dispatcher,
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
		logrus.Infof("starting server on %s", s.Address)
		if err := s.Server.Start(s.Address); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Errorf("failed to start server: %v", err)
		}
	}()
}

// Stop stops the echo server gracefully.
func (s *Server) Stop() {
	if s.Server == nil {
		return
	}

	logrus.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := s.Server.Shutdown(ctx); err != nil {
		logrus.Errorf("failed to shutdown server: %v", err)
	}
}
