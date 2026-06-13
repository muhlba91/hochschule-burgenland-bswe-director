package configuration

import (
	"github.com/caarlos0/env/v11"
	log "github.com/sirupsen/logrus"
)

// Data defines the configuration for the director init command, which is read from environment variables.
type Data struct {
	// HealthzHost is the host for the healthz server to listen on
	HealthzHost string `env:"HEALTHZ_HOST" envDefault:"0.0.0.0"`
	// HealthzPort is the port for the healthz server to listen on
	HealthzPort int `env:"HEALTHZ_PORT" envDefault:"8080"`
}

// Init reads the configuration from environment variables and returns a Config struct.
func Init() Data {
	cfg := Data{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("error reading configuration from environment: %v", err)
	}
	return cfg
}
