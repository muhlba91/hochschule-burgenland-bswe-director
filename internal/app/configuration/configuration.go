package configuration

import (
	"log/slog"
	"os"

	"github.com/caarlos0/env/v11"
)

// Data defines the configuration for the director init command, which is read from environment variables.
type Data struct {
	// HealthzHost is the host for the healthz server to listen on
	HealthzHost string `env:"HEALTHZ_HOST" envDefault:"0.0.0.0"`
	// HealthzPort is the port for the healthz server to listen on
	HealthzPort int `env:"HEALTHZ_PORT" envDefault:"8080"`
	// ServerHost is the host for the echo server to listen on
	ServerHost string `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	// ServerPort is the port for the echo server to listen on
	ServerPort int `env:"SERVER_PORT" envDefault:"8888"`
	// RedisHost is the host for the redis server
	RedisHost string `env:"REDIS_HOST" envDefault:"localhost"`
	// RedisPort is the port for the redis server
	RedisPort int `env:"REDIS_PORT" envDefault:"6379"`
	// RedisPassword is the password for the redis server
	RedisPassword string `env:"REDIS_PASSWORD"`
	// BaseURL is the base URL for the director
	BaseURL string `env:"BASE_URL" envDefault:"http://localhost:8888"`
}

// Init reads the configuration from environment variables and returns a Config struct.
func Init() Data {
	cfg := Data{}
	if err := env.Parse(&cfg); err != nil {
		slog.Error("error reading configuration from environment", slog.Any("error", err))
		os.Exit(1)
	}
	return cfg
}
