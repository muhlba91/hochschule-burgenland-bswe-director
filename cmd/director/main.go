package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/cmd/configuration"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/cmd/logging"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/dispatcher"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/health"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/http"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/cache"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/registry"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/requestor"
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

	logrus.Infof("version: %s (%s)", Version, Gitsha)

	cfg := configuration.Init()

	c := cache.NewCache(&cfg)
	req := requestor.NewRequestor(c)
	reg := registry.NewRegistry()
	d := dispatcher.NewDispatcher(c, reg)
	flows.Init(c, req, reg)

	healthServer := health.NewServer(&cfg, c)
	healthServer.Start()

	httpServer := http.NewServer(&cfg, d)
	httpServer.Start()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	<-sig
	logrus.Info("shutting down...")
	healthServer.Stop()
	httpServer.Stop()
	c.Stop()
}
