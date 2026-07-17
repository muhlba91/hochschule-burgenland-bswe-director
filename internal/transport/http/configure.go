package http

import (
	"log/slog"
	nethttp "net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/configuration"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/transport/auth"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/transport/callback"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/transport/websocket"
)

// NewServer creates a new echo server.
// http: The echo server to configure.
func configureRoutes(
	http *Server,
	configuration *configuration.Data,
) {
	http.Server.POST("/auth/token", auth.Handler(configuration))
	http.Server.GET("/ws", websocket.Handler(http.Dispatcher), authMiddleware(configuration))
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

// authMiddleware is a middleware function that checks for a valid JWT in the request.
// cfg: The configuration data containing the auth secret and auth enabled flag.
func authMiddleware(cfg *configuration.Data) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !cfg.AuthEnabled {
				return next(c)
			}

			token := c.Request().Header.Get("Authorization")
			if token == "" {
				token = c.QueryParam("token")
			} else {
				token = strings.TrimPrefix(token, "Bearer ")
			}

			if token == "" {
				return echo.NewHTTPError(nethttp.StatusUnauthorized, auth.ErrMissingToken.Error())
			}

			claims, err := auth.ValidateJWT(token, cfg.AuthSecret)
			if err != nil {
				return echo.NewHTTPError(nethttp.StatusUnauthorized, auth.ErrInvalidToken.Error())
			}

			c.Set("clientID", claims.ClientID)

			return next(c)
		}
	}
}
