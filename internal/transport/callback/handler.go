package callback

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/logging"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/orchestrator"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback/response"
)

// Handler handles callback requests.
// dispatcher: The dispatcher for handling management and flow actions.
func Handler(dispatcher *orchestrator.Dispatcher) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		requestID := c.Param("requestId")

		slog.DebugContext(ctx, "received callback request",
			slog.String("path", c.Request().URL.Path),
			slog.String(logging.FieldRequestID, requestID),
		)

		defer c.Request().Body.Close()
		data, bErr := io.ReadAll(c.Request().Body)
		if bErr != nil {
			slog.ErrorContext(ctx, "failed to read callback request body",
				slog.String(logging.FieldRequestID, requestID),
				slog.Any(logging.FieldError, bErr),
			)
			return echo.NewHTTPError(http.StatusBadRequest, response.NewError(response.ErrInvalidPayload))
		}
		body := json.RawMessage(data)
		slog.DebugContext(ctx, "callback request body",
			slog.String(logging.FieldRequestID, requestID),
			slog.String("body", string(body)),
		)

		return dispatcher.HandleCallback(ctx, c, requestID, body)
	}
}
