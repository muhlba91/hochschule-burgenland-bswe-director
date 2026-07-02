package http

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/configuration"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/transport/callback"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/transport/websocket"
)

// NewServer creates a new echo server.
// http: The echo server to configure.
func configureRoutes(
	http *Server,
) {
	http.Server.GET("/ws", websocket.Handler(http.Dispatcher))
	http.Server.POST("/callback/:requestId", callback.Handler(http.Dispatcher))
}

// configureServer configures the echo server with middleware and settings.
// http: The echo server to configure.
// configuration: The configuration data for the server.
func configureServer(http *Server, configuration *configuration.Data) {
	http.Server.HideBanner = true
	http.Server.HidePort = true

	http.Server.Use(middleware.Recover())
	http.Server.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStoreWithConfig(
		middleware.RateLimiterMemoryStoreConfig{
			Rate:      rate.Limit(configuration.RateLimit),
			Burst:     configuration.BurstLimit,
			ExpiresIn: configuration.RateLimitWindow,
		},
	)))

	http.Server.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogMethod: true,
		LogError:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			slog.DebugContext(c.Request().Context(), "request",
				slog.String("time", v.StartTime.Format(time.RFC3339)),
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.Int("status", v.Status),
				slog.Any("error", v.Error),
			)

			return nil
		},
	}))
}
