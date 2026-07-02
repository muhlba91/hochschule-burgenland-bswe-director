package health

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/configuration"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store"
)

type healthCheckResult struct {
	name   string
	status string
	up     bool
}

// shutdownTimeout defines the duration for graceful shutdown of the health check server.
const shutdownTimeout = 5 * time.Second

// readHeaderTimeout defines the duration for reading the request headers.
const readHeaderTimeout = 3 * time.Second

// Server represents the health check server.
type Server struct {
	address        string
	server         *http.Server
	sessionStore   store.SessionStore
	requestStore   store.RequestStore
	broadcastStore store.BroadcastStore
}

// NewServer creates a new health check server.
// configuration: The configuration data for the server.
// cache: The cache instance.
func NewServer(
	configuration *configuration.Data,
	sessionStore store.SessionStore,
	requestStore store.RequestStore,
	broadcastStore store.BroadcastStore,
) *Server {
	return &Server{
		address:        fmt.Sprintf("%s:%d", configuration.HealthzHost, configuration.HealthzPort),
		sessionStore:   sessionStore,
		requestStore:   requestStore,
		broadcastStore: broadcastStore,
	}
}

// Start starts the health check server.
func (s *Server) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/livez", s.livezHandler)
	mux.HandleFunc("/healthz", s.healthzHandler)
	mux.HandleFunc("/startupz", s.startupzHandler)

	s.server = &http.Server{
		Addr:              s.address,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	go func() {
		slog.InfoContext(context.Background(), "starting healthz server",
			slog.String("address", s.address),
		)
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.ErrorContext(context.Background(), "failed to start healthz server",
				slog.Any("error", err),
			)
		}
	}()
}

// Stop stops the health check server gracefully.
func (s *Server) Stop() {
	if s.server == nil {
		return
	}

	slog.InfoContext(context.Background(), "shutting down healthz server")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to shutdown healthz server",
			slog.Any("error", err),
		)
	}
}
