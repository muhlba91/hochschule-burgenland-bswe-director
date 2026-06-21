package http

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/websocket"
)

// NewServer creates a new echo server.
// http: The echo server to configure.
func configureRoutes(
	http *Server,
) {
	http.Server.GET("/ws", websocket.Handler(http.Dispatcher))
}

// configureServer configures the echo server with middleware and settings.
// http: The echo server to configure.
func configureServer(http *Server) {
	http.Server.HideBanner = true
	http.Server.HidePort = true

	http.Server.Use(middleware.Recover())

	http.Server.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogMethod: true,
		LogError:  true,
		LogValuesFunc: func(_ echo.Context, v middleware.RequestLoggerValues) error {
			logrus.WithFields(logrus.Fields{
				"time":   v.StartTime.Format(time.RFC3339),
				"method": v.Method,
				"uri":    v.URI,
				"status": v.Status,
				"error":  v.Error,
			}).Info("request")

			return nil
		},
	}))
}
