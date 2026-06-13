package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/cmd/configuration"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/cmd/logging"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/health"
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

	healthServer := health.NewServer(&cfg)
	healthServer.Start()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	<-sig
	logrus.Info("shutting down...")
	healthServer.Stop()
}
