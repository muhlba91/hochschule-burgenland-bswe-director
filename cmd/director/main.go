package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/configuration"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/logging"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/flows"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/orchestrator"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store/redis"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/transport/health"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/transport/http"
)

// Version and Gitsha of the current build.
//
//nolint:gochecknoglobals // These variables are set at build time using ldflags.
var (
	Version = "local"
	Gitsha  = "?"
)

// main is the entry point of the application.
func main() {
	logging.Init()

	slog.Info("application starting",
		slog.String("version", Version),
		slog.String("gitsha", Gitsha),
	)

	cfg := configuration.Init()

	c := redis.NewCache(&cfg)
	req := orchestrator.NewRequestor(c, c, c, &cfg)
	reg := orchestrator.NewRegistry()
	d := orchestrator.NewDispatcher(c, c, c, reg, &cfg)
	flows.Init(c, c, c, c, req, c, reg)

	healthServer := health.NewServer(&cfg, c, c, c)
	healthServer.Start()

	httpServer := http.NewServer(&cfg, d)
	httpServer.Start()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	<-sig
	slog.Info("shutting down...")
	healthServer.Stop()
	httpServer.Stop()
	c.Stop()
}
