package callback

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

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

		logrus.WithFields(logrus.Fields{
			"path":                 c.Request().URL.Path,
			logging.FieldRequestID: requestID,
		}).Debug("received callback request")

		defer c.Request().Body.Close()
		data, bErr := io.ReadAll(c.Request().Body)
		if bErr != nil {
			logrus.WithFields(logrus.Fields{
				logging.FieldRequestID: requestID,
				logging.FieldError:     bErr,
			}).Error("failed to read callback request body")
			return echo.NewHTTPError(http.StatusBadRequest, response.NewError(response.ErrInvalidPayload))
		}
		body := json.RawMessage(data)
		logrus.WithFields(logrus.Fields{
			logging.FieldRequestID: requestID,
			"body":                 string(body),
		}).Debug("callback request body")

		return dispatcher.HandleCallback(ctx, c, requestID, body)
	}
}
