package health

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/cmd/configuration"
)

// shutdownTimeout defines the duration for graceful shutdown of the health check server.
const shutdownTimeout = 5 * time.Second

// readHeaderTimeout defines the duration for reading the request headers.
const readHeaderTimeout = 3 * time.Second

// Server represents the health check server.
type Server struct {
	address string
	server  *http.Server
}

// NewServer creates a new health check server.
// configuration: The configuration data for the server.
func NewServer(
	configuration *configuration.Data,
) *Server {
	return &Server{
		address: fmt.Sprintf("%s:%d", configuration.HealthzHost, configuration.HealthzPort),
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
		logrus.Infof("starting healthz server on %s", s.address)
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Errorf("failed to start healthz server: %v", err)
		}
	}()
}

// Stop stops the health check server gracefully.
func (s *Server) Stop() {
	if s.server == nil {
		return
	}

	logrus.Info("shutting down healthz server")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		logrus.Errorf("failed to shutdown healthz server: %v", err)
	}
}
